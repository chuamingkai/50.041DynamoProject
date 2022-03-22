// protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative internalcomm.proto

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.17.3
// source: pkg/internalcomm/internalcomm.proto

package grpc

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type PutRepRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key  []byte `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Data []byte `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *PutRepRequest) Reset() {
	*x = PutRepRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_internalcomm_internalcomm_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PutRepRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PutRepRequest) ProtoMessage() {}

func (x *PutRepRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_internalcomm_internalcomm_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PutRepRequest.ProtoReflect.Descriptor instead.
func (*PutRepRequest) Descriptor() ([]byte, []int) {
	return file_pkg_internalcomm_internalcomm_proto_rawDescGZIP(), []int{0}
}

func (x *PutRepRequest) GetKey() []byte {
	if x != nil {
		return x.Key
	}
	return nil
}

func (x *PutRepRequest) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

type PutRepResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	IsDone bool `protobuf:"varint,1,opt,name=isDone,proto3" json:"isDone,omitempty"`
}

func (x *PutRepResponse) Reset() {
	*x = PutRepResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_internalcomm_internalcomm_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PutRepResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PutRepResponse) ProtoMessage() {}

func (x *PutRepResponse) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_internalcomm_internalcomm_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PutRepResponse.ProtoReflect.Descriptor instead.
func (*PutRepResponse) Descriptor() ([]byte, []int) {
	return file_pkg_internalcomm_internalcomm_proto_rawDescGZIP(), []int{1}
}

func (x *PutRepResponse) GetIsDone() bool {
	if x != nil {
		return x.IsDone
	}
	return false
}

type GetRepRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key []byte `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
}

func (x *GetRepRequest) Reset() {
	*x = GetRepRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_internalcomm_internalcomm_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetRepRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetRepRequest) ProtoMessage() {}

func (x *GetRepRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_internalcomm_internalcomm_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetRepRequest.ProtoReflect.Descriptor instead.
func (*GetRepRequest) Descriptor() ([]byte, []int) {
	return file_pkg_internalcomm_internalcomm_proto_rawDescGZIP(), []int{2}
}

func (x *GetRepRequest) GetKey() []byte {
	if x != nil {
		return x.Key
	}
	return nil
}

type GetRepResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data []byte `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *GetRepResponse) Reset() {
	*x = GetRepResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_internalcomm_internalcomm_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetRepResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetRepResponse) ProtoMessage() {}

func (x *GetRepResponse) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_internalcomm_internalcomm_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetRepResponse.ProtoReflect.Descriptor instead.
func (*GetRepResponse) Descriptor() ([]byte, []int) {
	return file_pkg_internalcomm_internalcomm_proto_rawDescGZIP(), []int{3}
}

func (x *GetRepResponse) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

var File_pkg_internalcomm_internalcomm_proto protoreflect.FileDescriptor

var file_pkg_internalcomm_internalcomm_proto_rawDesc = []byte{
	0x0a, 0x23, 0x70, 0x6b, 0x67, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x63, 0x6f,
	0x6d, 0x6d, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x63, 0x6f, 0x6d, 0x6d, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x35, 0x0a, 0x0d, 0x50, 0x75, 0x74, 0x52, 0x65, 0x70, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x22, 0x28, 0x0a, 0x0e,
	0x50, 0x75, 0x74, 0x52, 0x65, 0x70, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x16,
	0x0a, 0x06, 0x69, 0x73, 0x44, 0x6f, 0x6e, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x06,
	0x69, 0x73, 0x44, 0x6f, 0x6e, 0x65, 0x22, 0x21, 0x0a, 0x0d, 0x47, 0x65, 0x74, 0x52, 0x65, 0x70,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x22, 0x24, 0x0a, 0x0e, 0x47, 0x65, 0x74,
	0x52, 0x65, 0x70, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x64,
	0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x32,
	0x6b, 0x0a, 0x0b, 0x52, 0x65, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x2d,
	0x0a, 0x0a, 0x50, 0x75, 0x74, 0x52, 0x65, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x12, 0x0e, 0x2e, 0x50,
	0x75, 0x74, 0x52, 0x65, 0x70, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0f, 0x2e, 0x50,
	0x75, 0x74, 0x52, 0x65, 0x70, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2d, 0x0a,
	0x0a, 0x47, 0x65, 0x74, 0x52, 0x65, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x12, 0x0e, 0x2e, 0x47, 0x65,
	0x74, 0x52, 0x65, 0x70, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0f, 0x2e, 0x47, 0x65,
	0x74, 0x52, 0x65, 0x70, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x35, 0x5a, 0x33,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x63, 0x68, 0x75, 0x61, 0x6d,
	0x69, 0x6e, 0x67, 0x6b, 0x61, 0x69, 0x2f, 0x35, 0x30, 0x2e, 0x30, 0x34, 0x31, 0x44, 0x79, 0x6e,
	0x61, 0x6d, 0x6f, 0x50, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x67,
	0x72, 0x70, 0x63, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_pkg_internalcomm_internalcomm_proto_rawDescOnce sync.Once
	file_pkg_internalcomm_internalcomm_proto_rawDescData = file_pkg_internalcomm_internalcomm_proto_rawDesc
)

func file_pkg_internalcomm_internalcomm_proto_rawDescGZIP() []byte {
	file_pkg_internalcomm_internalcomm_proto_rawDescOnce.Do(func() {
		file_pkg_internalcomm_internalcomm_proto_rawDescData = protoimpl.X.CompressGZIP(file_pkg_internalcomm_internalcomm_proto_rawDescData)
	})
	return file_pkg_internalcomm_internalcomm_proto_rawDescData
}

var file_pkg_internalcomm_internalcomm_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_pkg_internalcomm_internalcomm_proto_goTypes = []interface{}{
	(*PutRepRequest)(nil),  // 0: PutRepRequest
	(*PutRepResponse)(nil), // 1: PutRepResponse
	(*GetRepRequest)(nil),  // 2: GetRepRequest
	(*GetRepResponse)(nil), // 3: GetRepResponse
}
var file_pkg_internalcomm_internalcomm_proto_depIdxs = []int32{
	0, // 0: Replication.PutReplica:input_type -> PutRepRequest
	2, // 1: Replication.GetReplica:input_type -> GetRepRequest
	1, // 2: Replication.PutReplica:output_type -> PutRepResponse
	3, // 3: Replication.GetReplica:output_type -> GetRepResponse
	2, // [2:4] is the sub-list for method output_type
	0, // [0:2] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_pkg_internalcomm_internalcomm_proto_init() }
func file_pkg_internalcomm_internalcomm_proto_init() {
	if File_pkg_internalcomm_internalcomm_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pkg_internalcomm_internalcomm_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PutRepRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_internalcomm_internalcomm_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PutRepResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_internalcomm_internalcomm_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetRepRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_internalcomm_internalcomm_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetRepResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_pkg_internalcomm_internalcomm_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_pkg_internalcomm_internalcomm_proto_goTypes,
		DependencyIndexes: file_pkg_internalcomm_internalcomm_proto_depIdxs,
		MessageInfos:      file_pkg_internalcomm_internalcomm_proto_msgTypes,
	}.Build()
	File_pkg_internalcomm_internalcomm_proto = out.File
	file_pkg_internalcomm_internalcomm_proto_rawDesc = nil
	file_pkg_internalcomm_internalcomm_proto_goTypes = nil
	file_pkg_internalcomm_internalcomm_proto_depIdxs = nil
}