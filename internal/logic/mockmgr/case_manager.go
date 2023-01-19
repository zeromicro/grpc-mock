package mockmgr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Knetic/govaluate"
	"github.com/golang/protobuf/jsonpb"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
	"go.uber.org/atomic"

	"github.com/zeromicro/grpc-mock/internal/config"
	"github.com/zeromicro/grpc-mock/internal/logic/parser"
	proxy2 "github.com/zeromicro/grpc-mock/internal/logic/proxy"
	"github.com/zeromicro/grpc-mock/internal/types"
)

// 包含所有的 upstreams
func formatAllUpstreamsKey() string {
	return fmt.Sprintf("fht-grpc-mock-all_upstreams")
}

// 包含所有的 app_name
func formatAllTestedAppNameKey() string {
	return fmt.Sprintf("fht-grpc-mock-all_tested_app")
}

// appName -> methods
func formatApp2MethodsKey(appName string) string {
	return fmt.Sprintf("fht-grpc-mock-app2methods_%s", appName)
}

// appName_method -> cases
func formatMethod2CasesKey(appName, fullMethod string) string {
	return fmt.Sprintf("fht-grpc-mock-app_method2cases_%s_%s", appName, fullMethod)
}

// cases k-v
func formatCaseBodyKey(appName, fullMethod, caseName string) string {
	return fmt.Sprintf("fht-grpc-mock-case_%s_%s_%s", appName, fullMethod, caseName)
}

// k-v: hash(appName-service_method-caseName) -> caseBody
// map[service_method][appName] -> map[case-name]caseBody
// caseName 可通过 meta-data, req-fields 匹配规则

type CaseMgr struct {
	conf     *config.Config
	redisCli *redis.Redis

	caseAndMethodsAtomic atomic.Value
}

type caseAndMethods struct {
	ss               map[string]map[string]*parser.ServiceDesc
	methods          map[string]*parser.MethodDesc
	endPoint2Methods map[string][]string
	cases            map[string]map[string]map[string]*CaseRuleBody

	upstreams []*zrpc.RpcClientConf
}

func NewCaseMgr(conf *config.Config, redisCli *redis.Redis) (*CaseMgr, error) {
	mgr := &CaseMgr{
		conf:     conf,
		redisCli: redisCli,
	}

	caseAndMethods, err := reloadCases(redisCli)
	if err == nil {
		mgr.caseAndMethodsAtomic.Store(caseAndMethods)
	}

	GO(func() {
		for {
			caseAndMethods, err := reloadCases(redisCli)
			if err == nil {
				mgr.caseAndMethodsAtomic.Store(caseAndMethods)
				time.Sleep(3 * time.Second)
			} else {
				time.Sleep(1 * time.Second)
			}
		}
	}, "reload_case_mgr")

	return mgr, err
}

var endPoint2Client = make(map[string]*zrpc.Client)

func getUpstreamClient(endPoint string) *zrpc.Client {
	if cli, ok := endPoint2Client[endPoint]; ok {
		return cli
	}

	cli := proxy2.NewClient(&zrpc.RpcClientConf{Endpoints: []string{endPoint}})
	endPoint2Client[endPoint] = cli

	return cli
}

func reloadCases(redisCli *redis.Redis) (res *caseAndMethods, err error) {
	res = &caseAndMethods{
		ss:               make(map[string]map[string]*parser.ServiceDesc),
		methods:          make(map[string]*parser.MethodDesc),
		endPoint2Methods: make(map[string][]string),
	}

	upstreams, err := getAllUpstreams(redisCli)
	if err != nil {
		return nil, err
	}

	var method2Cli = make(map[string]*zrpc.Client)

	for _, upstream := range upstreams {
		serviceDescs, err := parser.Parser(*upstream) // todo bob.rao 已经 parse 过的不需要重复 parse
		if err != nil {
			logx.Errorw("parser upstream error", logx.Field("upstream", *upstream), logx.Field("error", err))
			continue
		}

		endPoint := upstream.Endpoints[0]
		cli := getUpstreamClient(endPoint)
		for _, serviceDesc := range serviceDescs {
			serviceDesc := serviceDesc
			if serviceDesc.FullName == "grpc.reflection.v1alpha.ServerReflection" {
				continue
			}

			if _, ok := res.ss[endPoint]; !ok {
				res.ss[endPoint] = make(map[string]*parser.ServiceDesc)
			}
			res.ss[endPoint][serviceDesc.FullName] = &serviceDesc
			for _, m := range serviceDesc.Methods {
				m := m
				res.methods[m.FullName] = &m
				res.endPoint2Methods[endPoint] = append(res.endPoint2Methods[endPoint], m.FullName)
				method2Cli[m.FullName] = cli
			}
		}
	}

	proxy2.SetMethod2CliAtomic(method2Cli)

	res.cases = make(map[string]map[string]map[string]*CaseRuleBody)

	appNames, err := getAllTestedAppNames(redisCli)
	if err != nil {
		logx.Errorf("getAllTestedAppNames err: %s", err.Error())
		return nil, err
	}

	for _, appName := range appNames {
		fullMethods, err := getAllMethodsOfApp(redisCli, appName)
		if err != nil {
			logx.Errorf("getAllMethodsOfApp err: %s", err.Error())
			return nil, err
		}

		for _, fullMethod := range fullMethods {
			caseNames, err := getAllCaseNameOfAppMethods(redisCli, appName, fullMethod)
			if err != nil {
				logx.Errorf("getAllCaseNameOfAppMethods err: %s", err.Error())
				return nil, err
			}

			for _, caseName := range caseNames {
				caseBody, err := getCaseBody(redisCli, appName, fullMethod, caseName)
				if err != nil {
					logx.Errorf("matchByMetadata err: %s", err.Error())
					return nil, err
				}

				if caseBody.Rule != "" {
					rule := removeJPath(caseBody.Rule)
					expression, err := govaluate.NewEvaluableExpression(rule)
					if err != nil {
						logx.Errorf("govaluate.NewEvaluableExpression err: %s", err.Error())
						return nil, err
					}

					caseBody.Expression = expression
				}

				if _, ok := res.cases[fullMethod]; !ok {
					res.cases[fullMethod] = make(map[string]map[string]*CaseRuleBody)
				}
				if _, ok := res.cases[fullMethod][appName]; !ok {
					res.cases[fullMethod][appName] = make(map[string]*CaseRuleBody)
				}

				res.cases[fullMethod][appName][caseName] = caseBody
			}
		}
	}

	return res, err
}

func (m *CaseMgr) GetUpstreams() ([]*zrpc.RpcClientConf, error) {
	return getAllUpstreams(m.redisCli)
}

func (m *CaseMgr) DelUpstreams(upstreams []string) (err error) {
	_, err = m.redisCli.Srem(formatAllUpstreamsKey(), upstreams)
	return
}

func (m *CaseMgr) SetUpstreams(upstreams []string) (err error) {
	_, err = m.redisCli.Sadd(formatAllUpstreamsKey(), upstreams)
	return
}

func (m *CaseMgr) GetApps() ([]string, error) {
	return getAllTestedAppNames(m.redisCli)
}

func (m *CaseMgr) DelApp(appNames []string) (err error) {
	for _, appName := range appNames {
		methods, err := m.redisCli.Smembers(formatApp2MethodsKey(appName))
		if err != nil {
			logx.Errorf("DelApp redisCli.Smembers err", err)
			continue
		}

		for _, method := range methods {
			caseNames, err := m.redisCli.Smembers(formatMethod2CasesKey(appName, method))
			if err != nil {
				logx.Errorf("DelApp redisCli.Smembers err", err)
				continue
			}

			var keys []string
			for _, caseName := range caseNames {
				keys = append(keys, formatCaseBodyKey(appName, method, caseName))
			}

			_, err = m.redisCli.Del(keys...)
			if err != nil {
				logx.Errorf("DelApp redisCli.Del err", err)
				continue
			}

			_, err = m.redisCli.Del(formatMethod2CasesKey(appName, method))
			if err != nil {
				logx.Errorf("DelApp redisCli.Srem err", err)
				continue
			}
		}

		_, err = m.redisCli.Del(formatApp2MethodsKey(appName))
		if err != nil {
			logx.Errorf("DelApp redisCli.Del err", err)
			continue
		}
	}

	return
}

func (m *CaseMgr) GetCases(_ context.Context) map[string]map[string]map[string]*CaseRuleBody {
	caseAndMethods, ok := m.caseAndMethodsAtomic.Load().(*caseAndMethods)
	if !ok {
		logx.Errorf("load caseAndMethodsAtomic fail")
		return nil
	}

	return caseAndMethods.cases
}

func (m *CaseMgr) GetServiceDesc(_ context.Context, endPoint string) []*parser.ServiceDesc {
	var res []*parser.ServiceDesc
	caseAndMethods, ok := m.caseAndMethodsAtomic.Load().(*caseAndMethods)
	if !ok {
		logx.Errorf("load caseAndMethodsAtomic fail")
		return nil
	}

	for _, s := range caseAndMethods.ss[endPoint] {
		res = append(res, s)
	}

	return res
}

func (m *CaseMgr) GetMethodDesc(_ context.Context, name string) *parser.MethodDesc {
	caseAndMethods, ok := m.caseAndMethodsAtomic.Load().(*caseAndMethods)
	if !ok {
		logx.Errorf("load caseAndMethodsAtomic fail")
		return nil
	}

	return caseAndMethods.methods[name]
}

func (m *CaseMgr) GetMethodsByEndPoint(_ context.Context, endPoint string) []string {
	caseAndMethods, ok := m.caseAndMethodsAtomic.Load().(*caseAndMethods)
	if !ok {
		logx.Errorf("load caseAndMethodsAtomic fail")
		return nil
	}

	return caseAndMethods.endPoint2Methods[endPoint]
}

func (m *CaseMgr) CasesDetail(_ context.Context, appName, methodName, caseName string) (cs types.Case, err error) {
	bs, err := m.redisCli.Get(formatCaseBodyKey(appName, methodName, caseName))
	if err != nil {
		logx.Errorf("ListCases redisCli.get CaseBody err: %s", err.Error())
		return
	}

	err = json.Unmarshal([]byte(bs), &cs)
	return
}

func (m *CaseMgr) ListCases(_ context.Context, appName string) (cases []types.Case, err error) {
	methods, err := m.redisCli.Smembers(formatApp2MethodsKey(appName))
	if err != nil {
		logx.Errorf("ListCases redisCli.Smembers App2Methods err: %s", err.Error())
		return
	}

	for _, methodName := range methods {
		caseNames, err := m.redisCli.Smembers(formatMethod2CasesKey(appName, methodName))
		if err != nil {
			logx.Errorf("ListCases redisCli.Smembers Method2Cases err: %s", err.Error())
			continue
		}

		for _, caseName := range caseNames {
			bs, err := m.redisCli.Get(formatCaseBodyKey(appName, methodName, caseName)) // todo bob.rao MGET
			if err != nil {
				logx.Errorf("ListCases get cases err: %s", err.Error(), logx.Field("app_name", appName), logx.Field("full_method", methodName), logx.Field("case_name", caseName))
				continue
			}

			var cs types.Case
			err = json.Unmarshal([]byte(bs), &cs)
			if err != nil {
				logx.Errorf("ListCases json unmarshal cases err: %s", err.Error(), logx.Field("app_name", appName), logx.Field("full_method", methodName), logx.Field("case_name", caseName))
				continue
			}

			cases = append(cases, cs)
		}
	}

	return
}

func (m *CaseMgr) DelCases(_ context.Context, cases []types.Case) error {
	var err error
	for _, cs := range cases {
		_, err = m.redisCli.Del(formatCaseBodyKey(cs.TestedAppName, cs.MethodName, cs.Name))
		if err != nil {
			logx.Errorf("DelCases redisCli.Del CaseBody err: %s", err.Error())
			continue
		}

		_, err = m.redisCli.Srem(formatMethod2CasesKey(cs.TestedAppName, cs.MethodName), cs.Name)
		if err != nil {
			logx.Errorf("DelCases redisCli.Srem Method2Cases err: %s", err.Error())
			continue
		}
	}

	return err
}

func (m *CaseMgr) SetCases(_ context.Context, cases []types.Case) error {
	for _, cs := range cases {
		if cs.Body == "" || cs.Name == "" || cs.MethodName == "" || cs.TestedAppName == "" {
			return fmt.Errorf("body, name, method_name, tested_app_name should not be empty")
		}

		desc := m.GetMethodDesc(context.Background(), cs.MethodName)
		if desc == nil {
			return fmt.Errorf("tested_app: %s, case: %s, method: %s, not supported, please add upstream first", cs.TestedAppName, cs.Name, cs.MethodName)
		}

		msg := dynamic.NewMessageFactoryWithDefaults().NewMessage(desc.Out.RawDesc)
		err := jsonpb.Unmarshal(bytes.NewBufferString(cs.Body), msg)
		if err != nil {
			return fmt.Errorf("tested_app: %s, case: %s, method: %s, body illegal, body: %s, err: %s", cs.TestedAppName, cs.Name, cs.MethodName, cs.Body, err.Error())
		}
	}

	for _, cs := range cases {
		_, err := m.redisCli.Sadd(formatAllTestedAppNameKey(), cs.TestedAppName)
		if err != nil {
			logx.Errorf("SetCases redisCli.Sadd all app err: %s", err.Error())
			continue
		}

		_, err = m.redisCli.Sadd(formatApp2MethodsKey(cs.TestedAppName), cs.MethodName)
		if err != nil {
			logx.Errorf("SetCases redisCli.Sadd App2Methods err: %s", err.Error())
			continue
		}

		_, err = m.redisCli.Sadd(formatMethod2CasesKey(cs.TestedAppName, cs.MethodName), cs.Name)
		if err != nil {
			logx.Errorf("SetCases redisCli.Sadd Method2Cases err: %s", err.Error())
			continue
		}

		bs, err := json.Marshal(cs)
		if err != nil {
			logx.Errorf("SetCases json marshal err: %s", err.Error())
			continue
		}

		err = m.redisCli.Set(formatCaseBodyKey(cs.TestedAppName, cs.MethodName, cs.Name), string(bs))
		if err != nil {
			logx.Errorf("SetCases set case err: %s", err.Error())
			continue
		}
	}

	return nil
}

// 将 . 用魔法字符串代替
func removeJPath(rule string) string {
	rule = strings.Replace(rule, dotSymbol, magicNumDot, -1)

	return rule
}

func getAllUpstreams(redisCli *redis.Redis) ([]*zrpc.RpcClientConf, error) {
	endPoints, err := redisCli.Smembers(formatAllUpstreamsKey())
	if err != nil {
		logx.Errorf("getAllUpstreams redisCli.Smembers err: %s", err.Error())
		return nil, err
	}

	res := make([]*zrpc.RpcClientConf, 0, len(endPoints))
	for _, endPoint := range endPoints {
		res = append(res, &zrpc.RpcClientConf{Endpoints: []string{endPoint}})
	}

	return res, nil
}

func getAllTestedAppNames(redisCli *redis.Redis) ([]string, error) {
	appNames, err := redisCli.Smembers(formatAllTestedAppNameKey())
	if err != nil {
		logx.Errorf("getAllTestedAppNames redisCli.Smembers err: %s", err.Error())
		return nil, err
	}

	return appNames, nil
}

// 对于每个 app，有一个 hset: app2methods_APPNAME -> methods
func getAllMethodsOfApp(redisCli *redis.Redis, appName string) ([]string, error) {
	methods, err := redisCli.Smembers(formatApp2MethodsKey(appName))
	if err != nil {
		logx.Errorf("getAllMethodsOfApp redisCli.Smembers err: %s", err.Error(), logx.Field("app_name", appName))
		return nil, err
	}

	return methods, nil
}

// 对于每个 app mock 的每个 method，有一个 hset: method2cases__METHOD -> cases
func getAllCaseNameOfAppMethods(redisCli *redis.Redis, appName, fullMethod string) ([]string, error) {
	cases, err := redisCli.Smembers(formatMethod2CasesKey(appName, fullMethod))
	if err != nil {
		logx.Errorf("getAllCaseNameOfAppMethods redisCli.Smembers err: %s", err.Error(), logx.Field("method_name", fullMethod))
		return nil, err
	}

	return cases, nil
}

func getCaseBody(redisCli *redis.Redis, appName, fullMethod, caseName string) (*CaseRuleBody, error) {
	val, err := redisCli.Get(formatCaseBodyKey(appName, fullMethod, caseName))
	if err != nil {
		logx.Errorf("matchByMetadata redisCli.Get err: %s", err.Error(), logx.Field("app_name", appName), logx.Field("full_method", fullMethod), logx.Field("case_name", caseName))
		return nil, err
	}

	var res CaseRuleBody
	err = json.Unmarshal([]byte(val), &res)

	return &res, err
}
