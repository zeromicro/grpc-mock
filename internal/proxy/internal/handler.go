// Based on https://github.com/trusch/grpc-proxy
// Copyright Michal Witkowski. Licensed under Apache2 license: https://github.com/trusch/grpc-proxy/blob/master/LICENSE.txt

package internal

import (
	"io"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/zeromicro/grpc-mock/internal/match"
	"github.com/zeromicro/grpc-mock/internal/proxy/internal/codec"
)

var clientStreamDescForProxying = &grpc.StreamDesc{
	ServerStreams: true,
	ClientStreams: true,
}

// TransparentHandler returns a handler that attempts to proxy all requests that are not registered in the server.
// The indented use here is as a transparent proxy, where the server doesn't know about the services implemented by the
// backends. It should be used as a `grpc.UnknownServiceHandler`.
//
// This can *only* be used if the `server` also uses grpcproxy.CodecForServer() ServerOption.
func TransparentHandler(director func(ctx context.Context, fullMethodName string) (grpc.ClientConnInterface, error),
	match func(ctx context.Context, req match.Request) (*match.Response, error)) grpc.StreamHandler {
	streamer := &handler{
		director: director,
		match:    match,
	}
	return streamer.handler
}

type ReqMeta struct {
	FullMethodName string
	TestedAppName  string
	MD             metadata.MD
}

type handler struct {
	director func(ctx context.Context, fullMethodName string) (grpc.ClientConnInterface, error)
	match    func(ctx context.Context, req match.Request) (*match.Response, error)
}

// handler is where the real magic of proxying happens.
// It is invoked like any gRPC server stream and uses the gRPC server framing to get and receive bytes from the wire,
// forwarding it to a ClientStream established against the relevant ClientConn.
func (h *handler) handler(srv interface{}, serverStream grpc.ServerStream) error {
	ctx := serverStream.Context()

	logger := logx.WithContext(ctx)

	// little bit of gRPC internals never hurt anyone
	fullMethodName, ok := grpc.MethodFromServerStream(serverStream)
	if !ok {
		return status.Errorf(codes.Internal, "lowLevelServerStream not exists in context")
	}
	logger.Infow("handler full method name", logx.Field("method_name", fullMethodName))

	md, ok := metadata.FromIncomingContext(serverStream.Context())
	if !ok {
		logger.Error("handler get metadata from incoming context fail")
		return status.Errorf(codes.Internal, "metadata not exists in context")
	}
	logger.Infow("handler get md", logx.Field("md", md))

	// We require that the director's returned context inherits from the serverStream.Context().
	backendConn, err := h.director(ctx, fullMethodName)
	if err != nil {
		return err
	}

	// forward user client metadata to upstream server
	ctx = metadata.NewOutgoingContext(ctx, md.Copy())

	var (
		reqBytes []byte
		mt       match.MatchedType
	)

	f := &codec.Frame{}
	for i := 0; ; i++ {
		if err := serverStream.RecvMsg(f); err != nil {
			break
		}
	}
	reqBytes = f.GetBytes()

	reqMeta := &ReqMeta{
		FullMethodName: fullMethodName,
		MD:             md,
	}

	mt = h.doMock(serverStream, reqBytes, reqMeta, logger)
	if mt != match.MatchedTypeNone {
		logger.Infow("matched succeed.", logx.Field("match_type", mt))
		return nil
	}

	logger.Infof("grpc-mock act as a proxy")

	clientCtx, clientCancel := context.WithCancel(ctx)
	defer clientCancel()

	clientStream, err := grpc.NewClientStream(clientCtx, clientStreamDescForProxying, backendConn.(*grpc.ClientConn), fullMethodName)
	if err != nil {
		return err
	}
	// Explicitly *do not close* s2cErrChan and c2sErrChan, otherwise the select below will not terminate.
	// Channels do not have to be closed, it is just a control flow mechanism, see
	// https://groups.google.com/forum/#!msg/golang-nuts/pZwdYRGxCIk/qpbHxRRPJdUJ
	s2cErrChan := h.forwardClientToServer(serverStream, clientStream, reqBytes)
	c2sErrChan := h.forwardServerToClient(clientStream, serverStream)
	// We don't know which side is going to stop sending first, so we need a select between the two.
	for i := 0; i < 2; i++ {
		select {
		case s2cErr := <-s2cErrChan:
			if s2cErr == io.EOF {
				// this is the happy case where the sender has encountered io.EOF, and won't be sending anymore./
				// the clientStream>serverStream may continue pumping though.
				_ = clientStream.CloseSend() //nolint
			} else {
				// however, we may have gotten a receive error (stream disconnected, a read error etc) in which case we need
				// to cancel the clientStream to the backend, let all of its goroutines be freed up by the CancelFunc and
				// exit with an error to the stack
				clientCancel()
				return status.Errorf(codes.Internal, "failed proxying s2c: %v", s2cErr)
			}
		case c2sErr := <-c2sErrChan:
			// This happens when the clientStream has nothing else to offer (io.EOF), returned a gRPC error. In those two
			// cases we may have received Trailers as part of the call. In case of other errors (stream closed) the trailers
			// will be nil.
			serverStream.SetTrailer(clientStream.Trailer())
			// c2sErr will contain RPC error from client code. If not io.EOF return the RPC error as server stream error.
			if c2sErr != io.EOF {
				return c2sErr
			}
			return nil
		}
	}
	return status.Errorf(codes.Internal, "gRPC proxying should never reach this stage.")
}

func getMetadata(key string, md metadata.MD) string {
	vs := md.Get(key)
	if len(vs) == 0 {
		return ""
	}
	return vs[0]
}

func (h *handler) forwardServerToClient(src grpc.ClientStream, dst grpc.ServerStream) chan error {
	ret := make(chan error, 1)
	go func() {
		f := &codec.Frame{}
		for i := 0; ; i++ {
			if err := src.RecvMsg(f); err != nil {
				ret <- err // this can be io.EOF which is happy case
				break
			}
			if i == 0 {
				// This is a bit of a hack, but client to server headers are only readable after first client msg is
				// received but must be written to server stream before the first msg is flushed.
				// This is the only place to do it nicely.
				md, err := src.Header()
				if err != nil {
					ret <- err
					break
				}
				if err := dst.SendHeader(md); err != nil {
					ret <- err
					break
				}
			}
			if err := dst.SendMsg(f); err != nil {
				ret <- err
				break
			}
		}
	}()
	return ret
}

func (h *handler) forwardClientToServer(src grpc.ServerStream, dst grpc.ClientStream, reqBytes []byte) chan error {
	ret := make(chan error, 1)
	go func() {
		f := codec.NewFrame(reqBytes)
		for i := 0; ; i++ {
			if err := dst.SendMsg(f); err != nil {
				ret <- err
				break
			}

			if err := src.RecvMsg(f); err != nil {
				ret <- err // this can be io.EOF which is happy case
				break
			}

		}
	}()

	return ret
}

func (h *handler) doMock(src grpc.ServerStream, reqBytes []byte, meta *ReqMeta, logger logx.Logger) match.MatchedType {
	var (
		matched  bool
		response interface{}
	)

	defer func() {
		if matched {
			src.SetTrailer(metadata.MD{})

			header := map[string][]string{"mock": {"matched"}}
			src.SetHeader(header)

			err := src.SendMsg(response)
			if err != nil {
				logger.Errorw("send msg err", logx.Field("error", err))
			}

			logger.Infow("matched succeed.", logx.Field("response", response), logx.Field("header", header))
		}
	}()

	resp, err := h.match(context.Background(), match.Request{
		FullMethodName: meta.FullMethodName,
		MD:             meta.MD,
		RawReq:         reqBytes,
	})
	if err != nil {
		logger.Errorw("match err", logx.Field("error", err))
		return resp.MatchType
	}

	if resp.MatchType != match.MatchedTypeNone {
		matched = true
		response = resp.MockResp
		return resp.MatchType
	}

	return match.MatchedTypeNone
}
