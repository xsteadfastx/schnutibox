// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package v1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// IdentifierServiceClient is the client API for IdentifierService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type IdentifierServiceClient interface {
	Identify(ctx context.Context, in *IdentifyRequest, opts ...grpc.CallOption) (*IdentifyResponse, error)
}

type identifierServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewIdentifierServiceClient(cc grpc.ClientConnInterface) IdentifierServiceClient {
	return &identifierServiceClient{cc}
}

func (c *identifierServiceClient) Identify(ctx context.Context, in *IdentifyRequest, opts ...grpc.CallOption) (*IdentifyResponse, error) {
	out := new(IdentifyResponse)
	err := c.cc.Invoke(ctx, "/schnutibox.v1.IdentifierService/Identify", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// IdentifierServiceServer is the server API for IdentifierService service.
// All implementations should embed UnimplementedIdentifierServiceServer
// for forward compatibility
type IdentifierServiceServer interface {
	Identify(context.Context, *IdentifyRequest) (*IdentifyResponse, error)
}

// UnimplementedIdentifierServiceServer should be embedded to have forward compatible implementations.
type UnimplementedIdentifierServiceServer struct {
}

func (UnimplementedIdentifierServiceServer) Identify(context.Context, *IdentifyRequest) (*IdentifyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Identify not implemented")
}

// UnsafeIdentifierServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to IdentifierServiceServer will
// result in compilation errors.
type UnsafeIdentifierServiceServer interface {
	mustEmbedUnimplementedIdentifierServiceServer()
}

func RegisterIdentifierServiceServer(s grpc.ServiceRegistrar, srv IdentifierServiceServer) {
	s.RegisterService(&IdentifierService_ServiceDesc, srv)
}

func _IdentifierService_Identify_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IdentifyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IdentifierServiceServer).Identify(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/schnutibox.v1.IdentifierService/Identify",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IdentifierServiceServer).Identify(ctx, req.(*IdentifyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// IdentifierService_ServiceDesc is the grpc.ServiceDesc for IdentifierService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var IdentifierService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "schnutibox.v1.IdentifierService",
	HandlerType: (*IdentifierServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Identify",
			Handler:    _IdentifierService_Identify_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "schnutibox.proto",
}
