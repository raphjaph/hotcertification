// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.15.8
// source: signing.proto

package signing

import (
	_ "github.com/relab/gorums"
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

type TBS struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	CSRHash     string `protobuf:"bytes,1,opt,name=CSRHash,proto3" json:"CSRHash,omitempty"`
	Certificate []byte `protobuf:"bytes,2,opt,name=Certificate,proto3" json:"Certificate,omitempty"`
}

func (x *TBS) Reset() {
	*x = TBS{}
	if protoimpl.UnsafeEnabled {
		mi := &file_signing_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TBS) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TBS) ProtoMessage() {}

func (x *TBS) ProtoReflect() protoreflect.Message {
	mi := &file_signing_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TBS.ProtoReflect.Descriptor instead.
func (*TBS) Descriptor() ([]byte, []int) {
	return file_signing_proto_rawDescGZIP(), []int{0}
}

func (x *TBS) GetCSRHash() string {
	if x != nil {
		return x.CSRHash
	}
	return ""
}

func (x *TBS) GetCertificate() []byte {
	if x != nil {
		return x.Certificate
	}
	return nil
}

type SigShare struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Xi []byte `protobuf:"bytes,1,opt,name=Xi,proto3" json:"Xi,omitempty"`
	C  []byte `protobuf:"bytes,2,opt,name=C,proto3" json:"C,omitempty"`
	Z  []byte `protobuf:"bytes,3,opt,name=Z,proto3" json:"Z,omitempty"`
	Id uint32 `protobuf:"varint,4,opt,name=Id,proto3" json:"Id,omitempty"`
}

func (x *SigShare) Reset() {
	*x = SigShare{}
	if protoimpl.UnsafeEnabled {
		mi := &file_signing_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SigShare) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SigShare) ProtoMessage() {}

func (x *SigShare) ProtoReflect() protoreflect.Message {
	mi := &file_signing_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SigShare.ProtoReflect.Descriptor instead.
func (*SigShare) Descriptor() ([]byte, []int) {
	return file_signing_proto_rawDescGZIP(), []int{1}
}

func (x *SigShare) GetXi() []byte {
	if x != nil {
		return x.Xi
	}
	return nil
}

func (x *SigShare) GetC() []byte {
	if x != nil {
		return x.C
	}
	return nil
}

func (x *SigShare) GetZ() []byte {
	if x != nil {
		return x.Z
	}
	return nil
}

func (x *SigShare) GetId() uint32 {
	if x != nil {
		return x.Id
	}
	return 0
}

type ThresholdOf struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SigShares []*SigShare `protobuf:"bytes,1,rep,name=SigShares,proto3" json:"SigShares,omitempty"`
}

func (x *ThresholdOf) Reset() {
	*x = ThresholdOf{}
	if protoimpl.UnsafeEnabled {
		mi := &file_signing_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ThresholdOf) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ThresholdOf) ProtoMessage() {}

func (x *ThresholdOf) ProtoReflect() protoreflect.Message {
	mi := &file_signing_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ThresholdOf.ProtoReflect.Descriptor instead.
func (*ThresholdOf) Descriptor() ([]byte, []int) {
	return file_signing_proto_rawDescGZIP(), []int{2}
}

func (x *ThresholdOf) GetSigShares() []*SigShare {
	if x != nil {
		return x.SigShares
	}
	return nil
}

var File_signing_proto protoreflect.FileDescriptor

var file_signing_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x73, 0x69, 0x67, 0x6e, 0x69, 0x6e, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x07, 0x73, 0x69, 0x67, 0x6e, 0x69, 0x6e, 0x67, 0x1a, 0x0c, 0x67, 0x6f, 0x72, 0x75, 0x6d, 0x73,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x41, 0x0a, 0x03, 0x54, 0x42, 0x53, 0x12, 0x18, 0x0a,
	0x07, 0x43, 0x53, 0x52, 0x48, 0x61, 0x73, 0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07,
	0x43, 0x53, 0x52, 0x48, 0x61, 0x73, 0x68, 0x12, 0x20, 0x0a, 0x0b, 0x43, 0x65, 0x72, 0x74, 0x69,
	0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0b, 0x43, 0x65,
	0x72, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x65, 0x22, 0x46, 0x0a, 0x08, 0x53, 0x69, 0x67,
	0x53, 0x68, 0x61, 0x72, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x58, 0x69, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x02, 0x58, 0x69, 0x12, 0x0c, 0x0a, 0x01, 0x43, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c,
	0x52, 0x01, 0x43, 0x12, 0x0c, 0x0a, 0x01, 0x5a, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x01,
	0x5a, 0x12, 0x0e, 0x0a, 0x02, 0x49, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x02, 0x49,
	0x64, 0x22, 0x3e, 0x0a, 0x0b, 0x54, 0x68, 0x72, 0x65, 0x73, 0x68, 0x6f, 0x6c, 0x64, 0x4f, 0x66,
	0x12, 0x2f, 0x0a, 0x09, 0x53, 0x69, 0x67, 0x53, 0x68, 0x61, 0x72, 0x65, 0x73, 0x18, 0x01, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x73, 0x69, 0x67, 0x6e, 0x69, 0x6e, 0x67, 0x2e, 0x53, 0x69,
	0x67, 0x53, 0x68, 0x61, 0x72, 0x65, 0x52, 0x09, 0x53, 0x69, 0x67, 0x53, 0x68, 0x61, 0x72, 0x65,
	0x73, 0x32, 0x50, 0x0a, 0x07, 0x53, 0x69, 0x67, 0x6e, 0x69, 0x6e, 0x67, 0x12, 0x45, 0x0a, 0x0d,
	0x47, 0x65, 0x74, 0x50, 0x61, 0x72, 0x74, 0x69, 0x61, 0x6c, 0x53, 0x69, 0x67, 0x12, 0x0c, 0x2e,
	0x73, 0x69, 0x67, 0x6e, 0x69, 0x6e, 0x67, 0x2e, 0x54, 0x42, 0x53, 0x1a, 0x11, 0x2e, 0x73, 0x69,
	0x67, 0x6e, 0x69, 0x6e, 0x67, 0x2e, 0x53, 0x69, 0x67, 0x53, 0x68, 0x61, 0x72, 0x65, 0x22, 0x13,
	0xa0, 0xb5, 0x18, 0x01, 0xf2, 0xb6, 0x18, 0x0b, 0x54, 0x68, 0x72, 0x65, 0x73, 0x68, 0x6f, 0x6c,
	0x64, 0x4f, 0x66, 0x42, 0x1d, 0x5a, 0x1b, 0x2e, 0x2f, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65,
	0x2f, 0x74, 0x68, 0x72, 0x65, 0x73, 0x68, 0x6f, 0x6c, 0x64, 0x2f, 0x73, 0x69, 0x67, 0x6e, 0x69,
	0x6e, 0x67, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_signing_proto_rawDescOnce sync.Once
	file_signing_proto_rawDescData = file_signing_proto_rawDesc
)

func file_signing_proto_rawDescGZIP() []byte {
	file_signing_proto_rawDescOnce.Do(func() {
		file_signing_proto_rawDescData = protoimpl.X.CompressGZIP(file_signing_proto_rawDescData)
	})
	return file_signing_proto_rawDescData
}

var file_signing_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_signing_proto_goTypes = []interface{}{
	(*TBS)(nil),         // 0: signing.TBS
	(*SigShare)(nil),    // 1: signing.SigShare
	(*ThresholdOf)(nil), // 2: signing.ThresholdOf
}
var file_signing_proto_depIdxs = []int32{
	1, // 0: signing.ThresholdOf.SigShares:type_name -> signing.SigShare
	0, // 1: signing.Signing.GetPartialSig:input_type -> signing.TBS
	1, // 2: signing.Signing.GetPartialSig:output_type -> signing.SigShare
	2, // [2:3] is the sub-list for method output_type
	1, // [1:2] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_signing_proto_init() }
func file_signing_proto_init() {
	if File_signing_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_signing_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TBS); i {
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
		file_signing_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SigShare); i {
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
		file_signing_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ThresholdOf); i {
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
			RawDescriptor: file_signing_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_signing_proto_goTypes,
		DependencyIndexes: file_signing_proto_depIdxs,
		MessageInfos:      file_signing_proto_msgTypes,
	}.Build()
	File_signing_proto = out.File
	file_signing_proto_rawDesc = nil
	file_signing_proto_goTypes = nil
	file_signing_proto_depIdxs = nil
}
