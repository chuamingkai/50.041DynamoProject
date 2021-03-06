// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package grpc

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

// ReplicationClient is the client API for Replication service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ReplicationClient interface {
	PutReplica(ctx context.Context, in *PutRepRequest, opts ...grpc.CallOption) (*PutRepResponse, error)
	GetReplica(ctx context.Context, in *GetRepRequest, opts ...grpc.CallOption) (*GetRepResponse, error)
	PutMultiple(ctx context.Context, opts ...grpc.CallOption) (Replication_PutMultipleClient, error)
	Heartbeat(ctx context.Context, in *HeartbeatRequest, opts ...grpc.CallOption) (*HeartbeatResponse, error)
	HintedHandoff(ctx context.Context, in *HintHandoffRequest, opts ...grpc.CallOption) (*HintHandoffResponse, error)
}

type replicationClient struct {
	cc grpc.ClientConnInterface
}

func NewReplicationClient(cc grpc.ClientConnInterface) ReplicationClient {
	return &replicationClient{cc}
}

func (c *replicationClient) PutReplica(ctx context.Context, in *PutRepRequest, opts ...grpc.CallOption) (*PutRepResponse, error) {
	out := new(PutRepResponse)
	err := c.cc.Invoke(ctx, "/Replication/PutReplica", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *replicationClient) GetReplica(ctx context.Context, in *GetRepRequest, opts ...grpc.CallOption) (*GetRepResponse, error) {
	out := new(GetRepResponse)
	err := c.cc.Invoke(ctx, "/Replication/GetReplica", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *replicationClient) PutMultiple(ctx context.Context, opts ...grpc.CallOption) (Replication_PutMultipleClient, error) {
	stream, err := c.cc.NewStream(ctx, &Replication_ServiceDesc.Streams[0], "/Replication/PutMultiple", opts...)
	if err != nil {
		return nil, err
	}
	x := &replicationPutMultipleClient{stream}
	return x, nil
}

type Replication_PutMultipleClient interface {
	Send(*MultiPutRequest) error
	CloseAndRecv() (*PutRepResponse, error)
	grpc.ClientStream
}

type replicationPutMultipleClient struct {
	grpc.ClientStream
}

func (x *replicationPutMultipleClient) Send(m *MultiPutRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *replicationPutMultipleClient) CloseAndRecv() (*PutRepResponse, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(PutRepResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *replicationClient) Heartbeat(ctx context.Context, in *HeartbeatRequest, opts ...grpc.CallOption) (*HeartbeatResponse, error) {
	out := new(HeartbeatResponse)
	err := c.cc.Invoke(ctx, "/Replication/Heartbeat", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *replicationClient) HintedHandoff(ctx context.Context, in *HintHandoffRequest, opts ...grpc.CallOption) (*HintHandoffResponse, error) {
	out := new(HintHandoffResponse)
	err := c.cc.Invoke(ctx, "/Replication/HintedHandoff", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ReplicationServer is the server API for Replication service.
// All implementations must embed UnimplementedReplicationServer
// for forward compatibility
type ReplicationServer interface {
	PutReplica(context.Context, *PutRepRequest) (*PutRepResponse, error)
	GetReplica(context.Context, *GetRepRequest) (*GetRepResponse, error)
	PutMultiple(Replication_PutMultipleServer) error
	Heartbeat(context.Context, *HeartbeatRequest) (*HeartbeatResponse, error)
	HintedHandoff(context.Context, *HintHandoffRequest) (*HintHandoffResponse, error)
	mustEmbedUnimplementedReplicationServer()
}

// UnimplementedReplicationServer must be embedded to have forward compatible implementations.
type UnimplementedReplicationServer struct {
}

func (UnimplementedReplicationServer) PutReplica(context.Context, *PutRepRequest) (*PutRepResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PutReplica not implemented")
}
func (UnimplementedReplicationServer) GetReplica(context.Context, *GetRepRequest) (*GetRepResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetReplica not implemented")
}
func (UnimplementedReplicationServer) PutMultiple(Replication_PutMultipleServer) error {
	return status.Errorf(codes.Unimplemented, "method PutMultiple not implemented")
}
func (UnimplementedReplicationServer) Heartbeat(context.Context, *HeartbeatRequest) (*HeartbeatResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Heartbeat not implemented")
}
func (UnimplementedReplicationServer) HintedHandoff(context.Context, *HintHandoffRequest) (*HintHandoffResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method HintedHandoff not implemented")
}
func (UnimplementedReplicationServer) mustEmbedUnimplementedReplicationServer() {}

// UnsafeReplicationServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ReplicationServer will
// result in compilation errors.
type UnsafeReplicationServer interface {
	mustEmbedUnimplementedReplicationServer()
}

func RegisterReplicationServer(s grpc.ServiceRegistrar, srv ReplicationServer) {
	s.RegisterService(&Replication_ServiceDesc, srv)
}

func _Replication_PutReplica_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PutRepRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReplicationServer).PutReplica(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Replication/PutReplica",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReplicationServer).PutReplica(ctx, req.(*PutRepRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Replication_GetReplica_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRepRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReplicationServer).GetReplica(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Replication/GetReplica",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReplicationServer).GetReplica(ctx, req.(*GetRepRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Replication_PutMultiple_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ReplicationServer).PutMultiple(&replicationPutMultipleServer{stream})
}

type Replication_PutMultipleServer interface {
	SendAndClose(*PutRepResponse) error
	Recv() (*MultiPutRequest, error)
	grpc.ServerStream
}

type replicationPutMultipleServer struct {
	grpc.ServerStream
}

func (x *replicationPutMultipleServer) SendAndClose(m *PutRepResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *replicationPutMultipleServer) Recv() (*MultiPutRequest, error) {
	m := new(MultiPutRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Replication_Heartbeat_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HeartbeatRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReplicationServer).Heartbeat(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Replication/Heartbeat",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReplicationServer).Heartbeat(ctx, req.(*HeartbeatRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Replication_HintedHandoff_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HintHandoffRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReplicationServer).HintedHandoff(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Replication/HintedHandoff",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReplicationServer).HintedHandoff(ctx, req.(*HintHandoffRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Replication_ServiceDesc is the grpc.ServiceDesc for Replication service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Replication_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "Replication",
	HandlerType: (*ReplicationServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "PutReplica",
			Handler:    _Replication_PutReplica_Handler,
		},
		{
			MethodName: "GetReplica",
			Handler:    _Replication_GetReplica_Handler,
		},
		{
			MethodName: "Heartbeat",
			Handler:    _Replication_Heartbeat_Handler,
		},
		{
			MethodName: "HintedHandoff",
			Handler:    _Replication_HintedHandoff_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "PutMultiple",
			Handler:       _Replication_PutMultiple_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "pkg/internalcomm/internalcomm.proto",
}
