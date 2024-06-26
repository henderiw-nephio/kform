// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v4.24.3
// source: kfplugin.proto

package kfplugin1

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

// ProviderClient is the client API for Provider service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ProviderClient interface {
	Capabilities(ctx context.Context, in *Capabilities_Request, opts ...grpc.CallOption) (*Capabilities_Response, error)
	Configure(ctx context.Context, in *Configure_Request, opts ...grpc.CallOption) (*Configure_Response, error)
	ReadDataSource(ctx context.Context, in *ReadDataSource_Request, opts ...grpc.CallOption) (*ReadDataSource_Response, error)
	ListDataSource(ctx context.Context, in *ListDataSource_Request, opts ...grpc.CallOption) (*ListDataSource_Response, error)
	ReadResource(ctx context.Context, in *ReadResource_Request, opts ...grpc.CallOption) (*ReadResource_Response, error)
	CreateResource(ctx context.Context, in *CreateResource_Request, opts ...grpc.CallOption) (*CreateResource_Response, error)
	UpdateResource(ctx context.Context, in *UpdateResource_Request, opts ...grpc.CallOption) (*UpdateResource_Response, error)
	DeleteResource(ctx context.Context, in *DeleteResource_Request, opts ...grpc.CallOption) (*DeleteResource_Response, error)
	StopProvider(ctx context.Context, in *StopProvider_Request, opts ...grpc.CallOption) (*StopProvider_Response, error)
}

type providerClient struct {
	cc grpc.ClientConnInterface
}

func NewProviderClient(cc grpc.ClientConnInterface) ProviderClient {
	return &providerClient{cc}
}

func (c *providerClient) Capabilities(ctx context.Context, in *Capabilities_Request, opts ...grpc.CallOption) (*Capabilities_Response, error) {
	out := new(Capabilities_Response)
	err := c.cc.Invoke(ctx, "/kfplugin1.Provider/Capabilities", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *providerClient) Configure(ctx context.Context, in *Configure_Request, opts ...grpc.CallOption) (*Configure_Response, error) {
	out := new(Configure_Response)
	err := c.cc.Invoke(ctx, "/kfplugin1.Provider/Configure", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *providerClient) ReadDataSource(ctx context.Context, in *ReadDataSource_Request, opts ...grpc.CallOption) (*ReadDataSource_Response, error) {
	out := new(ReadDataSource_Response)
	err := c.cc.Invoke(ctx, "/kfplugin1.Provider/ReadDataSource", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *providerClient) ListDataSource(ctx context.Context, in *ListDataSource_Request, opts ...grpc.CallOption) (*ListDataSource_Response, error) {
	out := new(ListDataSource_Response)
	err := c.cc.Invoke(ctx, "/kfplugin1.Provider/ListDataSource", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *providerClient) ReadResource(ctx context.Context, in *ReadResource_Request, opts ...grpc.CallOption) (*ReadResource_Response, error) {
	out := new(ReadResource_Response)
	err := c.cc.Invoke(ctx, "/kfplugin1.Provider/ReadResource", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *providerClient) CreateResource(ctx context.Context, in *CreateResource_Request, opts ...grpc.CallOption) (*CreateResource_Response, error) {
	out := new(CreateResource_Response)
	err := c.cc.Invoke(ctx, "/kfplugin1.Provider/CreateResource", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *providerClient) UpdateResource(ctx context.Context, in *UpdateResource_Request, opts ...grpc.CallOption) (*UpdateResource_Response, error) {
	out := new(UpdateResource_Response)
	err := c.cc.Invoke(ctx, "/kfplugin1.Provider/UpdateResource", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *providerClient) DeleteResource(ctx context.Context, in *DeleteResource_Request, opts ...grpc.CallOption) (*DeleteResource_Response, error) {
	out := new(DeleteResource_Response)
	err := c.cc.Invoke(ctx, "/kfplugin1.Provider/DeleteResource", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *providerClient) StopProvider(ctx context.Context, in *StopProvider_Request, opts ...grpc.CallOption) (*StopProvider_Response, error) {
	out := new(StopProvider_Response)
	err := c.cc.Invoke(ctx, "/kfplugin1.Provider/StopProvider", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ProviderServer is the server API for Provider service.
// All implementations must embed UnimplementedProviderServer
// for forward compatibility
type ProviderServer interface {
	Capabilities(context.Context, *Capabilities_Request) (*Capabilities_Response, error)
	Configure(context.Context, *Configure_Request) (*Configure_Response, error)
	ReadDataSource(context.Context, *ReadDataSource_Request) (*ReadDataSource_Response, error)
	ListDataSource(context.Context, *ListDataSource_Request) (*ListDataSource_Response, error)
	ReadResource(context.Context, *ReadResource_Request) (*ReadResource_Response, error)
	CreateResource(context.Context, *CreateResource_Request) (*CreateResource_Response, error)
	UpdateResource(context.Context, *UpdateResource_Request) (*UpdateResource_Response, error)
	DeleteResource(context.Context, *DeleteResource_Request) (*DeleteResource_Response, error)
	StopProvider(context.Context, *StopProvider_Request) (*StopProvider_Response, error)
	mustEmbedUnimplementedProviderServer()
}

// UnimplementedProviderServer must be embedded to have forward compatible implementations.
type UnimplementedProviderServer struct {
}

func (UnimplementedProviderServer) Capabilities(context.Context, *Capabilities_Request) (*Capabilities_Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Capabilities not implemented")
}
func (UnimplementedProviderServer) Configure(context.Context, *Configure_Request) (*Configure_Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Configure not implemented")
}
func (UnimplementedProviderServer) ReadDataSource(context.Context, *ReadDataSource_Request) (*ReadDataSource_Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReadDataSource not implemented")
}
func (UnimplementedProviderServer) ListDataSource(context.Context, *ListDataSource_Request) (*ListDataSource_Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListDataSource not implemented")
}
func (UnimplementedProviderServer) ReadResource(context.Context, *ReadResource_Request) (*ReadResource_Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReadResource not implemented")
}
func (UnimplementedProviderServer) CreateResource(context.Context, *CreateResource_Request) (*CreateResource_Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateResource not implemented")
}
func (UnimplementedProviderServer) UpdateResource(context.Context, *UpdateResource_Request) (*UpdateResource_Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateResource not implemented")
}
func (UnimplementedProviderServer) DeleteResource(context.Context, *DeleteResource_Request) (*DeleteResource_Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteResource not implemented")
}
func (UnimplementedProviderServer) StopProvider(context.Context, *StopProvider_Request) (*StopProvider_Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StopProvider not implemented")
}
func (UnimplementedProviderServer) mustEmbedUnimplementedProviderServer() {}

// UnsafeProviderServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ProviderServer will
// result in compilation errors.
type UnsafeProviderServer interface {
	mustEmbedUnimplementedProviderServer()
}

func RegisterProviderServer(s grpc.ServiceRegistrar, srv ProviderServer) {
	s.RegisterService(&Provider_ServiceDesc, srv)
}

func _Provider_Capabilities_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Capabilities_Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProviderServer).Capabilities(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kfplugin1.Provider/Capabilities",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProviderServer).Capabilities(ctx, req.(*Capabilities_Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _Provider_Configure_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Configure_Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProviderServer).Configure(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kfplugin1.Provider/Configure",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProviderServer).Configure(ctx, req.(*Configure_Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _Provider_ReadDataSource_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReadDataSource_Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProviderServer).ReadDataSource(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kfplugin1.Provider/ReadDataSource",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProviderServer).ReadDataSource(ctx, req.(*ReadDataSource_Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _Provider_ListDataSource_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListDataSource_Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProviderServer).ListDataSource(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kfplugin1.Provider/ListDataSource",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProviderServer).ListDataSource(ctx, req.(*ListDataSource_Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _Provider_ReadResource_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReadResource_Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProviderServer).ReadResource(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kfplugin1.Provider/ReadResource",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProviderServer).ReadResource(ctx, req.(*ReadResource_Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _Provider_CreateResource_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateResource_Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProviderServer).CreateResource(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kfplugin1.Provider/CreateResource",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProviderServer).CreateResource(ctx, req.(*CreateResource_Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _Provider_UpdateResource_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateResource_Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProviderServer).UpdateResource(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kfplugin1.Provider/UpdateResource",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProviderServer).UpdateResource(ctx, req.(*UpdateResource_Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _Provider_DeleteResource_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteResource_Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProviderServer).DeleteResource(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kfplugin1.Provider/DeleteResource",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProviderServer).DeleteResource(ctx, req.(*DeleteResource_Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _Provider_StopProvider_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StopProvider_Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProviderServer).StopProvider(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kfplugin1.Provider/StopProvider",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProviderServer).StopProvider(ctx, req.(*StopProvider_Request))
	}
	return interceptor(ctx, in, info, handler)
}

// Provider_ServiceDesc is the grpc.ServiceDesc for Provider service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Provider_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "kfplugin1.Provider",
	HandlerType: (*ProviderServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Capabilities",
			Handler:    _Provider_Capabilities_Handler,
		},
		{
			MethodName: "Configure",
			Handler:    _Provider_Configure_Handler,
		},
		{
			MethodName: "ReadDataSource",
			Handler:    _Provider_ReadDataSource_Handler,
		},
		{
			MethodName: "ListDataSource",
			Handler:    _Provider_ListDataSource_Handler,
		},
		{
			MethodName: "ReadResource",
			Handler:    _Provider_ReadResource_Handler,
		},
		{
			MethodName: "CreateResource",
			Handler:    _Provider_CreateResource_Handler,
		},
		{
			MethodName: "UpdateResource",
			Handler:    _Provider_UpdateResource_Handler,
		},
		{
			MethodName: "DeleteResource",
			Handler:    _Provider_DeleteResource_Handler,
		},
		{
			MethodName: "StopProvider",
			Handler:    _Provider_StopProvider_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "kfplugin.proto",
}
