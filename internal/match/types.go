package match

import (
	"google.golang.org/grpc/metadata"
)

type Request struct {
	FullMethodName string
	MD             metadata.MD
	RawReq         []byte
}

type MatchedType int

const (
	MatchedTypeNone     MatchedType = 0
	MatchedTypeMetaData MatchedType = 1
	MatchedTypeBody     MatchedType = 2
)

type Response struct {
	MatchType MatchedType
	MockResp  interface{}
}
