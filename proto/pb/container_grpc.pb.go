// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v3.12.4
// source: proto/container.proto

package pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	ContainerAdmService_GetAllContainers_FullMethodName           = "/container_adm_service.ContainerAdmService/GetAllContainers"
	ContainerAdmService_GetContainerInformation_FullMethodName    = "/container_adm_service.ContainerAdmService/GetContainerInformation"
	ContainerAdmService_GetContainerUptimeDuration_FullMethodName = "/container_adm_service.ContainerAdmService/GetContainerUptimeDuration"
)

// ContainerAdmServiceClient is the client API for ContainerAdmService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ContainerAdmServiceClient interface {
	GetAllContainers(ctx context.Context, in *EmptyRequest, opts ...grpc.CallOption) (*ContainerResponse, error)
	GetContainerInformation(ctx context.Context, in *GetContainerInfomationRequest, opts ...grpc.CallOption) (*GetContainerInfomationResponse, error)
	GetContainerUptimeDuration(ctx context.Context, in *GetContainerInfomationRequest, opts ...grpc.CallOption) (*GetContainerUptimeDurationResponse, error)
}

type containerAdmServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewContainerAdmServiceClient(cc grpc.ClientConnInterface) ContainerAdmServiceClient {
	return &containerAdmServiceClient{cc}
}

func (c *containerAdmServiceClient) GetAllContainers(ctx context.Context, in *EmptyRequest, opts ...grpc.CallOption) (*ContainerResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ContainerResponse)
	err := c.cc.Invoke(ctx, ContainerAdmService_GetAllContainers_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *containerAdmServiceClient) GetContainerInformation(ctx context.Context, in *GetContainerInfomationRequest, opts ...grpc.CallOption) (*GetContainerInfomationResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetContainerInfomationResponse)
	err := c.cc.Invoke(ctx, ContainerAdmService_GetContainerInformation_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *containerAdmServiceClient) GetContainerUptimeDuration(ctx context.Context, in *GetContainerInfomationRequest, opts ...grpc.CallOption) (*GetContainerUptimeDurationResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetContainerUptimeDurationResponse)
	err := c.cc.Invoke(ctx, ContainerAdmService_GetContainerUptimeDuration_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ContainerAdmServiceServer is the server API for ContainerAdmService service.
// All implementations must embed UnimplementedContainerAdmServiceServer
// for forward compatibility.
type ContainerAdmServiceServer interface {
	GetAllContainers(context.Context, *EmptyRequest) (*ContainerResponse, error)
	GetContainerInformation(context.Context, *GetContainerInfomationRequest) (*GetContainerInfomationResponse, error)
	GetContainerUptimeDuration(context.Context, *GetContainerInfomationRequest) (*GetContainerUptimeDurationResponse, error)
	mustEmbedUnimplementedContainerAdmServiceServer()
}

// UnimplementedContainerAdmServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedContainerAdmServiceServer struct{}

func (UnimplementedContainerAdmServiceServer) GetAllContainers(context.Context, *EmptyRequest) (*ContainerResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAllContainers not implemented")
}
func (UnimplementedContainerAdmServiceServer) GetContainerInformation(context.Context, *GetContainerInfomationRequest) (*GetContainerInfomationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetContainerInformation not implemented")
}
func (UnimplementedContainerAdmServiceServer) GetContainerUptimeDuration(context.Context, *GetContainerInfomationRequest) (*GetContainerUptimeDurationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetContainerUptimeDuration not implemented")
}
func (UnimplementedContainerAdmServiceServer) mustEmbedUnimplementedContainerAdmServiceServer() {}
func (UnimplementedContainerAdmServiceServer) testEmbeddedByValue()                             {}

// UnsafeContainerAdmServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ContainerAdmServiceServer will
// result in compilation errors.
type UnsafeContainerAdmServiceServer interface {
	mustEmbedUnimplementedContainerAdmServiceServer()
}

func RegisterContainerAdmServiceServer(s grpc.ServiceRegistrar, srv ContainerAdmServiceServer) {
	// If the following call pancis, it indicates UnimplementedContainerAdmServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&ContainerAdmService_ServiceDesc, srv)
}

func _ContainerAdmService_GetAllContainers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EmptyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ContainerAdmServiceServer).GetAllContainers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ContainerAdmService_GetAllContainers_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ContainerAdmServiceServer).GetAllContainers(ctx, req.(*EmptyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ContainerAdmService_GetContainerInformation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetContainerInfomationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ContainerAdmServiceServer).GetContainerInformation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ContainerAdmService_GetContainerInformation_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ContainerAdmServiceServer).GetContainerInformation(ctx, req.(*GetContainerInfomationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ContainerAdmService_GetContainerUptimeDuration_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetContainerInfomationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ContainerAdmServiceServer).GetContainerUptimeDuration(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ContainerAdmService_GetContainerUptimeDuration_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ContainerAdmServiceServer).GetContainerUptimeDuration(ctx, req.(*GetContainerInfomationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ContainerAdmService_ServiceDesc is the grpc.ServiceDesc for ContainerAdmService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ContainerAdmService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "container_adm_service.ContainerAdmService",
	HandlerType: (*ContainerAdmServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetAllContainers",
			Handler:    _ContainerAdmService_GetAllContainers_Handler,
		},
		{
			MethodName: "GetContainerInformation",
			Handler:    _ContainerAdmService_GetContainerInformation_Handler,
		},
		{
			MethodName: "GetContainerUptimeDuration",
			Handler:    _ContainerAdmService_GetContainerUptimeDuration_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/container.proto",
}
