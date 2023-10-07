package match

import (
	"bytes"
	"context"

	"github.com/antonmedv/expr"
	"github.com/golang/protobuf/jsonpb"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/tidwall/gjson"
	"github.com/zeromicro/go-zero/core/logc"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/metadata"

	"github.com/zeromicro/grpc-mock/internal/controlapi/types"
	"github.com/zeromicro/grpc-mock/internal/svc"
)

type Matcher struct {
	svcCtx *svc.ServiceContext
}

func NewMatcher(svcCtx *svc.ServiceContext) *Matcher {
	return &Matcher{
		svcCtx: svcCtx,
	}
}

func (m *Matcher) Match(ctx context.Context, req Request) (*Response, error) {
	// 1. match with metadata
	resp, err := m.matchWithMetadata(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp.MatchType != MatchedTypeNone {
		return resp, nil
	}
	// 2. match with request body
	return m.matchWithRequestBody(ctx, req)
}

func (m *Matcher) matchWithMetadata(ctx context.Context, req Request) (*Response, error) {
	// 0. check config
	if m.svcCtx.Config.MatchConf.MockEnableKey == "" ||
		m.svcCtx.Config.MatchConf.MockEnableValue == "" ||
		m.svcCtx.Config.MatchConf.MockCaseKey == "" {
		return &Response{
			MatchType: MatchedTypeNone,
		}, nil
	}

	// 1. check enable key
	enable := getMetadata(m.svcCtx.Config.MatchConf.MockEnableKey, req.MD)
	if enable != m.svcCtx.Config.MatchConf.MockEnableValue {
		return &Response{
			MatchType: MatchedTypeNone,
		}, nil
	}

	// 2. check case name
	caseName := getMetadata(m.svcCtx.Config.MatchConf.MockCaseKey, req.MD)
	if caseName == "" {
		return &Response{
			MatchType: MatchedTypeNone,
		}, nil
	}

	// 3. get mock case
	_case, err := m.svcCtx.CaseManager.CaseGet(ctx, req.FullMethodName, caseName)
	if err != nil {
		return nil, err
	}
	if _case.Body == "" {
		return &Response{
			MatchType: MatchedTypeNone,
		}, nil
	}

	// 4. generate mock response
	desc, err := m.svcCtx.DialManager.MethodDetail(ctx, req.FullMethodName)
	if err != nil {
		return nil, err
	}

	msg := dynamic.NewMessageFactoryWithDefaults().NewMessage(desc.Out.RawDesc)

	err = jsonpb.Unmarshal(bytes.NewBufferString(_case.Body), msg)
	if err != nil {
		logc.Errorf(ctx, "GetMockResponseByMeta jsonpb.Unmarshal body err: %s", err.Error())
		return nil, err
	}

	return &Response{
		MatchType: MatchedTypeMetaData,
		MockResp:  msg,
	}, nil
}

func (m *Matcher) matchWithRequestBody(ctx context.Context, req Request) (*Response, error) {
	cases, err := m.svcCtx.CaseManager.CaseList(ctx, req.FullMethodName)
	if err != nil {
		return nil, err
	}

	var ruleCases []types.Case
	for _, _case := range cases {
		if _case.Rule != "" {
			ruleCases = append(ruleCases, _case)
		}
	}
	if len(ruleCases) == 0 {
		return &Response{
			MatchType: MatchedTypeNone,
		}, nil
	}

	desc, err := m.svcCtx.DialManager.MethodDetail(ctx, req.FullMethodName)
	if err != nil {
		return nil, err
	}

	in := dynamic.NewMessage(desc.In.RawDesc)

	err = encoding.GetCodec("proto").Unmarshal(req.RawReq, in)
	if err != nil {
		return nil, err
	}

	js, err := in.MarshalJSONPB(&jsonpb.Marshaler{OrigName: true, EnumsAsInts: true, EmitDefaults: true})
	if err != nil {
		return nil, err
	}

	get := func(path string) interface{} {
		return gjson.Get(string(js), path).Value()
	}

	env := map[string]interface{}{
		"json": get,
	}

	for _, _case := range ruleCases {
		output, err := expr.Eval(_case.Rule, env)
		if err != nil {
			continue
		}
		if v, ok := output.(bool); ok && v {
			msg := dynamic.NewMessageFactoryWithDefaults().NewMessage(desc.Out.RawDesc)

			err = jsonpb.Unmarshal(bytes.NewBufferString(_case.Body), msg)
			if err != nil {
				logc.Errorf(ctx, "GetMockResponseByMeta jsonpb.Unmarshal body err: %s", err.Error())
				continue
			}
			return &Response{
				MatchType: MatchedTypeBody,
				MockResp:  msg,
			}, nil
		}
	}

	return &Response{
		MatchType: MatchedTypeNone,
	}, nil
}

func getMetadata(key string, md metadata.MD) string {
	vs := md.Get(key)
	if len(vs) == 0 {
		return ""
	}
	return vs[0]
}
