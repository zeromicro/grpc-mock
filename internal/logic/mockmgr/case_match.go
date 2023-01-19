package mockmgr

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Knetic/govaluate"
	"github.com/golang/protobuf/jsonpb"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/spyzhov/ajson"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/metadata"

	"github.com/zeromicro/grpc-mock/internal/logic/proxy"
)

const (
	dollarSymbol = "$"
	dotSymbol    = "."
	magicNumDot  = "a0a0"
)

type CaseRuleBody struct {
	Rule       string
	Expression *govaluate.EvaluableExpression
	Body       string
}

type expressionBody struct {
	expression *govaluate.EvaluableExpression
	body       string
}

func (m *CaseMgr) getExpressions(methodName, appName string) (res []*expressionBody) {
	res = make([]*expressionBody, 0)
	if appName == "" {
		for _, appCaseRuleBody := range m.GetCases(context.Background())[methodName] {
			for _, caseRuleBody := range appCaseRuleBody {
				if caseRuleBody.Expression == nil {
					continue
				}

				res = append(res, &expressionBody{
					expression: caseRuleBody.Expression,
					body:       caseRuleBody.Body,
				})
			}
		}
	} else {
		for _, caseRuleBody := range m.GetCases(context.Background())[methodName][appName] {
			if caseRuleBody.Expression == nil {
				continue
			}

			res = append(res, &expressionBody{
				expression: caseRuleBody.Expression,
				body:       caseRuleBody.Body,
			})
		}
	}

	return res
}

// GetMockResponseByMeta 根据规则匹配预定的响应
// 匹配原则：若 end-point 未进行配置，则默认打到后端（为将来通过 sidecar 代理作准备）
//
// 当请求里指定了 mock = false 时、未匹配到规则时，grpc-mock 充当 proxy，请求真实的 server
//
// 1. 若 meta-data 里的 mock 明确指定为 false，则 grpc-mock 充当 proxy，不进行 mock
// 2. 若 meta-data 里指定了 custom-resp，则直接返回 custom-resp
// 3. 若 meta-data 里指定了 case-name，则直接匹配 case-name
// 4. 用请求体里的字段进行规则匹配
// 5. 透传请求到下游
func (m *CaseMgr) GetMockResponseByMeta(serverStream grpc.ServerStream, reqMeta *proxy.ReqMeta, logger logx.Logger) (bool, interface{}) {
	md := reqMeta.MD
	fullMethodName := reqMeta.FullMethodName

	desc := m.GetMethodDesc(context.Background(), fullMethodName)
	if desc == nil {
		logger.Errorf("GetMockResponseByMeta illegal method name: %s", fullMethodName)
		return false, nil
	}

	testedAppName := reqMeta.TestedAppName
	msg := dynamic.NewMessageFactoryWithDefaults().NewMessage(desc.Out.RawDesc)

	logger.Infof("GetMockResponseByMeta req metadata: %v, tested_app_name: %s,  full_method_name: %s", md, testedAppName, fullMethodName)

	var err error
	ok, body := m.matchByMetadata(testedAppName, fullMethodName, md, logger)
	if ok {
		err = jsonpb.Unmarshal(bytes.NewBufferString(body), msg)
		if err != nil {
			logger.Errorf("GetMockResponseByMeta jsonpb.Unmarshal body err: %s", err.Error())
			return false, nil
		}

		return true, msg
	}

	logger.Infof("GetMockResponseByMeta do not matched by metadata")

	return false, nil
}

func (m *CaseMgr) match(appName, methodName string, js []byte, logger logx.Logger) (bool, string) {
	logger.Infof("start match with field in request, method_name: %s, js: %s", methodName, js)

	var err error
	expressions := m.getExpressions(methodName, appName)
	for _, exp := range expressions {
		parameters := make(map[string]interface{}, len(exp.expression.Vars()))

		for _, k := range exp.expression.Vars() {
			parameters[k], err = getValue(k, js)
			if err != nil {
				logger.Infof("can not get value for key: %s, err: %s", k, err.Error())
				continue
			}
		}

		var result interface{}
		result, err = exp.expression.Evaluate(parameters)
		if err != nil {
			logger.Infof("Evaluate err: %s", err.Error())
			continue
		}

		if result.(bool) == true {
			logger.Infof("matched. rule: %s, parameters: %v, request: %s", exp.expression.String(), parameters, js)
			return true, exp.body
		}

		logger.Infof("do not match, rule: %s, parameters: %v, request: %s", exp.expression.String(), parameters, js)
	}

	return false, ""
}

func getValue(name string, js []byte) (val interface{}, err error) {
	name = restoreJPath(name)

	nodes, err := ajson.JSONPath(js, name)
	if len(nodes) <= 0 || len(nodes) > 1 {
		return "", fmt.Errorf("wrong nodes num: %d, now only support 1 node, name: %s, js: %s", len(nodes), name, string(js))
	}

	node := nodes[0]

	// int64, uint64 ---> String
	// float, double ----> number.
	// https://github.com/grpc-ecosystem/grpc-gateway/issues/438
	if node.Type() == ajson.String {
		fmt.Println(node.String())
		sVal, err := node.Value()
		oVal, err := strconv.ParseInt(sVal.(string), 10, 64)
		if err == nil {
			return oVal, err
		}

		oVal2, err := strconv.ParseUint(sVal.(string), 10, 64)
		if err == nil {
			return oVal2, err
		}
	}

	val, err = node.Value()
	return
}

func restoreJPath(name string) string {
	name = dollarSymbol + dotSymbol + name
	name = strings.Replace(name, magicNumDot, dotSymbol, -1)

	return name
}

func (m *CaseMgr) matchByMetadata(appName, fullMethodName string, md metadata.MD, logger logx.Logger) (bool, string) {
	if len(m.conf.MockCustomCaseKey) != 0 {
		body := getMetadata(m.conf.MockCustomCaseKey, md)
		if len(body) != 0 {
			logger.Infof("matched custom-case, body: %s", body)
			return true, body
		}
	}

	caseK := getMetadata(m.conf.MockCaseKey, md)
	if len(caseK) == 0 {
		logger.Infof("do not specify case name")
		return false, ""
	}

	cases := m.GetCases(context.Background())
	if len(cases) == 0 {
		logger.Infof("no cases loaded")
		return false, ""
	}

	if caseRuleBody, ok := cases[fullMethodName][appName][caseK]; ok {
		logger.Infof("matched by metadata. case_name: %s", caseK)
		return true, caseRuleBody.Body
	}

	logger.Infof("case_name: %s not found in the mock cases", caseK)
	return false, ""
}

func getMetadata(key string, md metadata.MD) string {
	vs := md.Get(key)
	if len(vs) == 0 {
		return ""
	}
	return vs[0]
}

func (m *CaseMgr) GetMockResponseByBody(serverStream grpc.ServerStream, req []byte, reqMeta *proxy.ReqMeta, logger logx.Logger) (matched bool, resp interface{}) {
	testedAppName := reqMeta.TestedAppName
	fullMethodName := reqMeta.FullMethodName

	desc := m.GetMethodDesc(context.Background(), fullMethodName)
	if desc == nil {
		logger.Errorf("GetMockResponseByBody illegal method name: %s", fullMethodName)
		return false, nil
	}

	in := dynamic.NewMessage(desc.In.RawDesc)

	err := encoding.GetCodec("proto").Unmarshal(req, in)
	if err != nil {
		logger.Errorf("GetMockResponseByBody unmarshal err", err.Error())
		return false, nil
	}

	js, err := in.MarshalJSONPB(&jsonpb.Marshaler{OrigName: true, EnumsAsInts: true, EmitDefaults: true})
	if err != nil {
		logger.Error("GetMockResponseByBody error in.MarshalJSONPB: %v", err)
		return false, nil
	}

	jsStr := string(js)
	jsStr = strings.Replace(jsStr, "\\", "", -1)
	logger.Infof("GetMockResponseByBody received body after replace: %s", jsStr)

	msg := dynamic.NewMessageFactoryWithDefaults().NewMessage(desc.Out.RawDesc)
	ok, body := m.match(testedAppName, fullMethodName, []byte(jsStr), logger)
	if ok {
		err = jsonpb.Unmarshal(bytes.NewBufferString(body), msg)
		if err != nil {
			logger.Errorf("GetMockResponseByBody jsonpb.Unmarshal body err: %s", err.Error())
			return false, nil
		}

		return true, msg
	}

	logger.Infof("GetMockResponseByBody do not matched by body")
	return false, nil
}
