// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.12.4
// source: avoid/avoid.proto

package avoid

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

// ManagementClient is the client API for Management service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ManagementClient interface {
	ListConnections(ctx context.Context, in *ListRequest, opts ...grpc.CallOption) (*ListReply, error)
	GetStats(ctx context.Context, in *StatsRequest, opts ...grpc.CallOption) (*StatsReply, error)
	Migrate(ctx context.Context, in *MigrateRequest, opts ...grpc.CallOption) (*MigrateReply, error)
	Shutdown(ctx context.Context, in *ShutdownRequest, opts ...grpc.CallOption) (*ShutdownReply, error)
}

type managementClient struct {
	cc grpc.ClientConnInterface
}

func NewManagementClient(cc grpc.ClientConnInterface) ManagementClient {
	return &managementClient{cc}
}

func (c *managementClient) ListConnections(ctx context.Context, in *ListRequest, opts ...grpc.CallOption) (*ListReply, error) {
	out := new(ListReply)
	err := c.cc.Invoke(ctx, "/avoid.Management/ListConnections", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *managementClient) GetStats(ctx context.Context, in *StatsRequest, opts ...grpc.CallOption) (*StatsReply, error) {
	out := new(StatsReply)
	err := c.cc.Invoke(ctx, "/avoid.Management/GetStats", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *managementClient) Migrate(ctx context.Context, in *MigrateRequest, opts ...grpc.CallOption) (*MigrateReply, error) {
	out := new(MigrateReply)
	err := c.cc.Invoke(ctx, "/avoid.Management/Migrate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *managementClient) Shutdown(ctx context.Context, in *ShutdownRequest, opts ...grpc.CallOption) (*ShutdownReply, error) {
	out := new(ShutdownReply)
	err := c.cc.Invoke(ctx, "/avoid.Management/Shutdown", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ManagementServer is the server API for Management service.
// All implementations must embed UnimplementedManagementServer
// for forward compatibility
type ManagementServer interface {
	ListConnections(context.Context, *ListRequest) (*ListReply, error)
	GetStats(context.Context, *StatsRequest) (*StatsReply, error)
	Migrate(context.Context, *MigrateRequest) (*MigrateReply, error)
	Shutdown(context.Context, *ShutdownRequest) (*ShutdownReply, error)
	mustEmbedUnimplementedManagementServer()
}

// UnimplementedManagementServer must be embedded to have forward compatible implementations.
type UnimplementedManagementServer struct {
}

func (UnimplementedManagementServer) ListConnections(context.Context, *ListRequest) (*ListReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListConnections not implemented")
}
func (UnimplementedManagementServer) GetStats(context.Context, *StatsRequest) (*StatsReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetStats not implemented")
}
func (UnimplementedManagementServer) Migrate(context.Context, *MigrateRequest) (*MigrateReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Migrate not implemented")
}
func (UnimplementedManagementServer) Shutdown(context.Context, *ShutdownRequest) (*ShutdownReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Shutdown not implemented")
}
func (UnimplementedManagementServer) mustEmbedUnimplementedManagementServer() {}

// UnsafeManagementServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ManagementServer will
// result in compilation errors.
type UnsafeManagementServer interface {
	mustEmbedUnimplementedManagementServer()
}

func RegisterManagementServer(s grpc.ServiceRegistrar, srv ManagementServer) {
	s.RegisterService(&Management_ServiceDesc, srv)
}

func _Management_ListConnections_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ManagementServer).ListConnections(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/avoid.Management/ListConnections",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ManagementServer).ListConnections(ctx, req.(*ListRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Management_GetStats_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StatsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ManagementServer).GetStats(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/avoid.Management/GetStats",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ManagementServer).GetStats(ctx, req.(*StatsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Management_Migrate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MigrateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ManagementServer).Migrate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/avoid.Management/Migrate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ManagementServer).Migrate(ctx, req.(*MigrateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Management_Shutdown_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ShutdownRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ManagementServer).Shutdown(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/avoid.Management/Shutdown",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ManagementServer).Shutdown(ctx, req.(*ShutdownRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Management_ServiceDesc is the grpc.ServiceDesc for Management service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Management_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "avoid.Management",
	HandlerType: (*ManagementServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ListConnections",
			Handler:    _Management_ListConnections_Handler,
		},
		{
			MethodName: "GetStats",
			Handler:    _Management_GetStats_Handler,
		},
		{
			MethodName: "Migrate",
			Handler:    _Management_Migrate_Handler,
		},
		{
			MethodName: "Shutdown",
			Handler:    _Management_Shutdown_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "avoid/avoid.proto",
}

// TunnelClient is the client API for Tunnel service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TunnelClient interface {
	Register(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*RegisterReply, error)
	TokenReplace(ctx context.Context, in *ConnectionRequest, opts ...grpc.CallOption) (*RegisterReply, error)
	HealthCheck(ctx context.Context, in *HealthRequest, opts ...grpc.CallOption) (*HealthReply, error)
	Watch(ctx context.Context, in *ConnectionRequest, opts ...grpc.CallOption) (Tunnel_WatchClient, error)
}

type tunnelClient struct {
	cc grpc.ClientConnInterface
}

func NewTunnelClient(cc grpc.ClientConnInterface) TunnelClient {
	return &tunnelClient{cc}
}

func (c *tunnelClient) Register(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*RegisterReply, error) {
	out := new(RegisterReply)
	err := c.cc.Invoke(ctx, "/avoid.Tunnel/Register", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tunnelClient) TokenReplace(ctx context.Context, in *ConnectionRequest, opts ...grpc.CallOption) (*RegisterReply, error) {
	out := new(RegisterReply)
	err := c.cc.Invoke(ctx, "/avoid.Tunnel/TokenReplace", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tunnelClient) HealthCheck(ctx context.Context, in *HealthRequest, opts ...grpc.CallOption) (*HealthReply, error) {
	out := new(HealthReply)
	err := c.cc.Invoke(ctx, "/avoid.Tunnel/HealthCheck", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tunnelClient) Watch(ctx context.Context, in *ConnectionRequest, opts ...grpc.CallOption) (Tunnel_WatchClient, error) {
	stream, err := c.cc.NewStream(ctx, &Tunnel_ServiceDesc.Streams[0], "/avoid.Tunnel/Watch", opts...)
	if err != nil {
		return nil, err
	}
	x := &tunnelWatchClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Tunnel_WatchClient interface {
	Recv() (*ConnectionReply, error)
	grpc.ClientStream
}

type tunnelWatchClient struct {
	grpc.ClientStream
}

func (x *tunnelWatchClient) Recv() (*ConnectionReply, error) {
	m := new(ConnectionReply)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// TunnelServer is the server API for Tunnel service.
// All implementations must embed UnimplementedTunnelServer
// for forward compatibility
type TunnelServer interface {
	Register(context.Context, *RegisterRequest) (*RegisterReply, error)
	TokenReplace(context.Context, *ConnectionRequest) (*RegisterReply, error)
	HealthCheck(context.Context, *HealthRequest) (*HealthReply, error)
	Watch(*ConnectionRequest, Tunnel_WatchServer) error
	mustEmbedUnimplementedTunnelServer()
}

// UnimplementedTunnelServer must be embedded to have forward compatible implementations.
type UnimplementedTunnelServer struct {
}

func (UnimplementedTunnelServer) Register(context.Context, *RegisterRequest) (*RegisterReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Register not implemented")
}
func (UnimplementedTunnelServer) TokenReplace(context.Context, *ConnectionRequest) (*RegisterReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TokenReplace not implemented")
}
func (UnimplementedTunnelServer) HealthCheck(context.Context, *HealthRequest) (*HealthReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method HealthCheck not implemented")
}
func (UnimplementedTunnelServer) Watch(*ConnectionRequest, Tunnel_WatchServer) error {
	return status.Errorf(codes.Unimplemented, "method Watch not implemented")
}
func (UnimplementedTunnelServer) mustEmbedUnimplementedTunnelServer() {}

// UnsafeTunnelServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TunnelServer will
// result in compilation errors.
type UnsafeTunnelServer interface {
	mustEmbedUnimplementedTunnelServer()
}

func RegisterTunnelServer(s grpc.ServiceRegistrar, srv TunnelServer) {
	s.RegisterService(&Tunnel_ServiceDesc, srv)
}

func _Tunnel_Register_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TunnelServer).Register(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/avoid.Tunnel/Register",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TunnelServer).Register(ctx, req.(*RegisterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Tunnel_TokenReplace_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ConnectionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TunnelServer).TokenReplace(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/avoid.Tunnel/TokenReplace",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TunnelServer).TokenReplace(ctx, req.(*ConnectionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Tunnel_HealthCheck_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HealthRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TunnelServer).HealthCheck(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/avoid.Tunnel/HealthCheck",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TunnelServer).HealthCheck(ctx, req.(*HealthRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Tunnel_Watch_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ConnectionRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(TunnelServer).Watch(m, &tunnelWatchServer{stream})
}

type Tunnel_WatchServer interface {
	Send(*ConnectionReply) error
	grpc.ServerStream
}

type tunnelWatchServer struct {
	grpc.ServerStream
}

func (x *tunnelWatchServer) Send(m *ConnectionReply) error {
	return x.ServerStream.SendMsg(m)
}

// Tunnel_ServiceDesc is the grpc.ServiceDesc for Tunnel service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Tunnel_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "avoid.Tunnel",
	HandlerType: (*TunnelServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Register",
			Handler:    _Tunnel_Register_Handler,
		},
		{
			MethodName: "TokenReplace",
			Handler:    _Tunnel_TokenReplace_Handler,
		},
		{
			MethodName: "HealthCheck",
			Handler:    _Tunnel_HealthCheck_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Watch",
			Handler:       _Tunnel_Watch_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "avoid/avoid.proto",
}
