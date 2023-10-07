package parser

import (
	"context"
	"fmt"

	"github.com/fullstorydev/grpcurl"
	"github.com/golang/protobuf/jsonpb"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoprint"
	"github.com/jhump/protoreflect/grpcreflect"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

type (
	ServiceDesc struct {
		Name     string
		FullName string
		Methods  []MethodDesc
	}
	MethodDesc struct {
		Name      string
		FullName  string
		ProtoDesc string
		In        FieldDesc
		Out       FieldDesc
	}
	FieldDesc struct {
		Name      string
		FullName  string
		JsonDesc  string
		ProtoDesc string
		RawDesc   *desc.MessageDescriptor
	}
)

func Parser(cc grpc.ClientConnInterface) ([]ServiceDesc, error) {
	source, err := createDescriptorSource(cc)
	if err != nil {
		return nil, err
	}

	svcs, err := source.ListServices()
	if err != nil {
		return nil, err
	}
	ss := make([]ServiceDesc, 0, len(svcs))
	for _, svc := range svcs {
		d, err := source.FindSymbol(svc)
		if err != nil {
			return nil, err
		}
		s := ServiceDesc{
			Name:     d.GetName(),
			FullName: d.GetFullyQualifiedName(),
		}
		switch val := d.(type) {
		case *desc.ServiceDescriptor:

			svcMethods := val.GetMethods()
			s.Methods = make([]MethodDesc, 0, len(svcMethods))
			pt := &protoprint.Printer{}

			for _, method := range svcMethods {
				mProto, _ := pt.PrintProtoToString(method)

				inJson, _ := NewMessage(method.GetInputType(), false).MarshalJSONPB(
					&jsonpb.Marshaler{EnumsAsInts: true, EmitDefaults: true})
				outJson, _ := NewMessage(method.GetOutputType(), false).MarshalJSONPB(
					&jsonpb.Marshaler{EnumsAsInts: true, EmitDefaults: true})
				inProto, _ := pt.PrintProtoToString(method.GetInputType())
				outProto, _ := pt.PrintProtoToString(method.GetOutputType())

				m := MethodDesc{
					Name:      method.GetName(),
					FullName:  fmt.Sprintf("/%s/%s", svc, method.GetName()),
					ProtoDesc: mProto,
					In: FieldDesc{
						Name:      method.GetInputType().GetName(),
						FullName:  method.GetInputType().GetFullyQualifiedName(),
						JsonDesc:  string(inJson),
						ProtoDesc: inProto,
						RawDesc:   method.GetInputType(),
					},
					Out: FieldDesc{
						Name:      method.GetOutputType().GetName(),
						FullName:  method.GetOutputType().GetFullyQualifiedName(),
						JsonDesc:  string(outJson),
						ProtoDesc: outProto,
						RawDesc:   method.GetOutputType(),
					},
				}

				s.Methods = append(s.Methods, m)
			}
		}
		ss = append(ss, s)
	}
	return ss, nil
}

func createDescriptorSource(cc grpc.ClientConnInterface) (grpcurl.DescriptorSource, error) {
	var source grpcurl.DescriptorSource

	refCli := grpc_reflection_v1alpha.NewServerReflectionClient(cc)
	client := grpcreflect.NewClient(context.Background(), refCli)
	source = grpcurl.DescriptorSourceFromServer(context.Background(), client)

	return source, nil
}
