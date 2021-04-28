// Code generated by protoc-gen-go. DO NOT EDIT.
// source: hashmail.proto

// We can't rename this to auctioneerrpc, otherwise it would be a breaking
// change since the package name is also contained in the HTTP URIs and old
// clients would call the wrong endpoints. Luckily with the go_package option we
// can have different golang and RPC package names.

package auctioneerrpc

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type PoolAccountAuth struct {
	// The account key being used to authenticate.
	AcctKey []byte `protobuf:"bytes,1,opt,name=acct_key,json=acctKey,proto3" json:"acct_key,omitempty"`
	// A valid signature over the stream ID being used.
	StreamSig            []byte   `protobuf:"bytes,2,opt,name=stream_sig,json=streamSig,proto3" json:"stream_sig,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PoolAccountAuth) Reset()         { *m = PoolAccountAuth{} }
func (m *PoolAccountAuth) String() string { return proto.CompactTextString(m) }
func (*PoolAccountAuth) ProtoMessage()    {}
func (*PoolAccountAuth) Descriptor() ([]byte, []int) {
	return fileDescriptor_165b784e4d2471a2, []int{0}
}

func (m *PoolAccountAuth) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PoolAccountAuth.Unmarshal(m, b)
}
func (m *PoolAccountAuth) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PoolAccountAuth.Marshal(b, m, deterministic)
}
func (m *PoolAccountAuth) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PoolAccountAuth.Merge(m, src)
}
func (m *PoolAccountAuth) XXX_Size() int {
	return xxx_messageInfo_PoolAccountAuth.Size(m)
}
func (m *PoolAccountAuth) XXX_DiscardUnknown() {
	xxx_messageInfo_PoolAccountAuth.DiscardUnknown(m)
}

var xxx_messageInfo_PoolAccountAuth proto.InternalMessageInfo

func (m *PoolAccountAuth) GetAcctKey() []byte {
	if m != nil {
		return m.AcctKey
	}
	return nil
}

func (m *PoolAccountAuth) GetStreamSig() []byte {
	if m != nil {
		return m.StreamSig
	}
	return nil
}

type SidecarAuth struct {
	//
	//A valid sidecar ticket that has been signed (offered) by a Pool account in
	//the active state.
	Ticket               string   `protobuf:"bytes,1,opt,name=ticket,proto3" json:"ticket,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SidecarAuth) Reset()         { *m = SidecarAuth{} }
func (m *SidecarAuth) String() string { return proto.CompactTextString(m) }
func (*SidecarAuth) ProtoMessage()    {}
func (*SidecarAuth) Descriptor() ([]byte, []int) {
	return fileDescriptor_165b784e4d2471a2, []int{1}
}

func (m *SidecarAuth) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SidecarAuth.Unmarshal(m, b)
}
func (m *SidecarAuth) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SidecarAuth.Marshal(b, m, deterministic)
}
func (m *SidecarAuth) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SidecarAuth.Merge(m, src)
}
func (m *SidecarAuth) XXX_Size() int {
	return xxx_messageInfo_SidecarAuth.Size(m)
}
func (m *SidecarAuth) XXX_DiscardUnknown() {
	xxx_messageInfo_SidecarAuth.DiscardUnknown(m)
}

var xxx_messageInfo_SidecarAuth proto.InternalMessageInfo

func (m *SidecarAuth) GetTicket() string {
	if m != nil {
		return m.Ticket
	}
	return ""
}

type CipherBoxInit struct {
	// A description of the stream one is attempting to initialize.
	Desc *CipherBoxDesc `protobuf:"bytes,1,opt,name=desc,proto3" json:"desc,omitempty"`
	// Types that are valid to be assigned to Auth:
	//	*CipherBoxInit_AcctAuth
	//	*CipherBoxInit_SidecarAuth
	Auth                 isCipherBoxInit_Auth `protobuf_oneof:"auth"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *CipherBoxInit) Reset()         { *m = CipherBoxInit{} }
func (m *CipherBoxInit) String() string { return proto.CompactTextString(m) }
func (*CipherBoxInit) ProtoMessage()    {}
func (*CipherBoxInit) Descriptor() ([]byte, []int) {
	return fileDescriptor_165b784e4d2471a2, []int{2}
}

func (m *CipherBoxInit) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CipherBoxInit.Unmarshal(m, b)
}
func (m *CipherBoxInit) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CipherBoxInit.Marshal(b, m, deterministic)
}
func (m *CipherBoxInit) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CipherBoxInit.Merge(m, src)
}
func (m *CipherBoxInit) XXX_Size() int {
	return xxx_messageInfo_CipherBoxInit.Size(m)
}
func (m *CipherBoxInit) XXX_DiscardUnknown() {
	xxx_messageInfo_CipherBoxInit.DiscardUnknown(m)
}

var xxx_messageInfo_CipherBoxInit proto.InternalMessageInfo

func (m *CipherBoxInit) GetDesc() *CipherBoxDesc {
	if m != nil {
		return m.Desc
	}
	return nil
}

type isCipherBoxInit_Auth interface {
	isCipherBoxInit_Auth()
}

type CipherBoxInit_AcctAuth struct {
	AcctAuth *PoolAccountAuth `protobuf:"bytes,2,opt,name=acct_auth,json=acctAuth,proto3,oneof"`
}

type CipherBoxInit_SidecarAuth struct {
	SidecarAuth *SidecarAuth `protobuf:"bytes,3,opt,name=sidecar_auth,json=sidecarAuth,proto3,oneof"`
}

func (*CipherBoxInit_AcctAuth) isCipherBoxInit_Auth() {}

func (*CipherBoxInit_SidecarAuth) isCipherBoxInit_Auth() {}

func (m *CipherBoxInit) GetAuth() isCipherBoxInit_Auth {
	if m != nil {
		return m.Auth
	}
	return nil
}

func (m *CipherBoxInit) GetAcctAuth() *PoolAccountAuth {
	if x, ok := m.GetAuth().(*CipherBoxInit_AcctAuth); ok {
		return x.AcctAuth
	}
	return nil
}

func (m *CipherBoxInit) GetSidecarAuth() *SidecarAuth {
	if x, ok := m.GetAuth().(*CipherBoxInit_SidecarAuth); ok {
		return x.SidecarAuth
	}
	return nil
}

// XXX_OneofWrappers is for the internal use of the proto package.
func (*CipherBoxInit) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*CipherBoxInit_AcctAuth)(nil),
		(*CipherBoxInit_SidecarAuth)(nil),
	}
}

type CipherChallenge struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CipherChallenge) Reset()         { *m = CipherChallenge{} }
func (m *CipherChallenge) String() string { return proto.CompactTextString(m) }
func (*CipherChallenge) ProtoMessage()    {}
func (*CipherChallenge) Descriptor() ([]byte, []int) {
	return fileDescriptor_165b784e4d2471a2, []int{3}
}

func (m *CipherChallenge) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CipherChallenge.Unmarshal(m, b)
}
func (m *CipherChallenge) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CipherChallenge.Marshal(b, m, deterministic)
}
func (m *CipherChallenge) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CipherChallenge.Merge(m, src)
}
func (m *CipherChallenge) XXX_Size() int {
	return xxx_messageInfo_CipherChallenge.Size(m)
}
func (m *CipherChallenge) XXX_DiscardUnknown() {
	xxx_messageInfo_CipherChallenge.DiscardUnknown(m)
}

var xxx_messageInfo_CipherChallenge proto.InternalMessageInfo

type CipherError struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CipherError) Reset()         { *m = CipherError{} }
func (m *CipherError) String() string { return proto.CompactTextString(m) }
func (*CipherError) ProtoMessage()    {}
func (*CipherError) Descriptor() ([]byte, []int) {
	return fileDescriptor_165b784e4d2471a2, []int{4}
}

func (m *CipherError) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CipherError.Unmarshal(m, b)
}
func (m *CipherError) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CipherError.Marshal(b, m, deterministic)
}
func (m *CipherError) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CipherError.Merge(m, src)
}
func (m *CipherError) XXX_Size() int {
	return xxx_messageInfo_CipherError.Size(m)
}
func (m *CipherError) XXX_DiscardUnknown() {
	xxx_messageInfo_CipherError.DiscardUnknown(m)
}

var xxx_messageInfo_CipherError proto.InternalMessageInfo

type CipherSuccess struct {
	Desc                 *CipherBoxDesc `protobuf:"bytes,1,opt,name=desc,proto3" json:"desc,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *CipherSuccess) Reset()         { *m = CipherSuccess{} }
func (m *CipherSuccess) String() string { return proto.CompactTextString(m) }
func (*CipherSuccess) ProtoMessage()    {}
func (*CipherSuccess) Descriptor() ([]byte, []int) {
	return fileDescriptor_165b784e4d2471a2, []int{5}
}

func (m *CipherSuccess) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CipherSuccess.Unmarshal(m, b)
}
func (m *CipherSuccess) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CipherSuccess.Marshal(b, m, deterministic)
}
func (m *CipherSuccess) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CipherSuccess.Merge(m, src)
}
func (m *CipherSuccess) XXX_Size() int {
	return xxx_messageInfo_CipherSuccess.Size(m)
}
func (m *CipherSuccess) XXX_DiscardUnknown() {
	xxx_messageInfo_CipherSuccess.DiscardUnknown(m)
}

var xxx_messageInfo_CipherSuccess proto.InternalMessageInfo

func (m *CipherSuccess) GetDesc() *CipherBoxDesc {
	if m != nil {
		return m.Desc
	}
	return nil
}

type CipherInitResp struct {
	// Types that are valid to be assigned to Resp:
	//	*CipherInitResp_Success
	//	*CipherInitResp_Challenge
	//	*CipherInitResp_Error
	Resp                 isCipherInitResp_Resp `protobuf_oneof:"resp"`
	XXX_NoUnkeyedLiteral struct{}              `json:"-"`
	XXX_unrecognized     []byte                `json:"-"`
	XXX_sizecache        int32                 `json:"-"`
}

func (m *CipherInitResp) Reset()         { *m = CipherInitResp{} }
func (m *CipherInitResp) String() string { return proto.CompactTextString(m) }
func (*CipherInitResp) ProtoMessage()    {}
func (*CipherInitResp) Descriptor() ([]byte, []int) {
	return fileDescriptor_165b784e4d2471a2, []int{6}
}

func (m *CipherInitResp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CipherInitResp.Unmarshal(m, b)
}
func (m *CipherInitResp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CipherInitResp.Marshal(b, m, deterministic)
}
func (m *CipherInitResp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CipherInitResp.Merge(m, src)
}
func (m *CipherInitResp) XXX_Size() int {
	return xxx_messageInfo_CipherInitResp.Size(m)
}
func (m *CipherInitResp) XXX_DiscardUnknown() {
	xxx_messageInfo_CipherInitResp.DiscardUnknown(m)
}

var xxx_messageInfo_CipherInitResp proto.InternalMessageInfo

type isCipherInitResp_Resp interface {
	isCipherInitResp_Resp()
}

type CipherInitResp_Success struct {
	Success *CipherSuccess `protobuf:"bytes,1,opt,name=success,proto3,oneof"`
}

type CipherInitResp_Challenge struct {
	Challenge *CipherChallenge `protobuf:"bytes,2,opt,name=challenge,proto3,oneof"`
}

type CipherInitResp_Error struct {
	Error *CipherError `protobuf:"bytes,3,opt,name=error,proto3,oneof"`
}

func (*CipherInitResp_Success) isCipherInitResp_Resp() {}

func (*CipherInitResp_Challenge) isCipherInitResp_Resp() {}

func (*CipherInitResp_Error) isCipherInitResp_Resp() {}

func (m *CipherInitResp) GetResp() isCipherInitResp_Resp {
	if m != nil {
		return m.Resp
	}
	return nil
}

func (m *CipherInitResp) GetSuccess() *CipherSuccess {
	if x, ok := m.GetResp().(*CipherInitResp_Success); ok {
		return x.Success
	}
	return nil
}

func (m *CipherInitResp) GetChallenge() *CipherChallenge {
	if x, ok := m.GetResp().(*CipherInitResp_Challenge); ok {
		return x.Challenge
	}
	return nil
}

func (m *CipherInitResp) GetError() *CipherError {
	if x, ok := m.GetResp().(*CipherInitResp_Error); ok {
		return x.Error
	}
	return nil
}

// XXX_OneofWrappers is for the internal use of the proto package.
func (*CipherInitResp) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*CipherInitResp_Success)(nil),
		(*CipherInitResp_Challenge)(nil),
		(*CipherInitResp_Error)(nil),
	}
}

type CipherBoxDesc struct {
	StreamId             []byte   `protobuf:"bytes,1,opt,name=stream_id,json=streamId,proto3" json:"stream_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CipherBoxDesc) Reset()         { *m = CipherBoxDesc{} }
func (m *CipherBoxDesc) String() string { return proto.CompactTextString(m) }
func (*CipherBoxDesc) ProtoMessage()    {}
func (*CipherBoxDesc) Descriptor() ([]byte, []int) {
	return fileDescriptor_165b784e4d2471a2, []int{7}
}

func (m *CipherBoxDesc) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CipherBoxDesc.Unmarshal(m, b)
}
func (m *CipherBoxDesc) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CipherBoxDesc.Marshal(b, m, deterministic)
}
func (m *CipherBoxDesc) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CipherBoxDesc.Merge(m, src)
}
func (m *CipherBoxDesc) XXX_Size() int {
	return xxx_messageInfo_CipherBoxDesc.Size(m)
}
func (m *CipherBoxDesc) XXX_DiscardUnknown() {
	xxx_messageInfo_CipherBoxDesc.DiscardUnknown(m)
}

var xxx_messageInfo_CipherBoxDesc proto.InternalMessageInfo

func (m *CipherBoxDesc) GetStreamId() []byte {
	if m != nil {
		return m.StreamId
	}
	return nil
}

type CipherBox struct {
	Desc                 *CipherBoxDesc `protobuf:"bytes,1,opt,name=desc,proto3" json:"desc,omitempty"`
	Msg                  []byte         `protobuf:"bytes,2,opt,name=msg,proto3" json:"msg,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *CipherBox) Reset()         { *m = CipherBox{} }
func (m *CipherBox) String() string { return proto.CompactTextString(m) }
func (*CipherBox) ProtoMessage()    {}
func (*CipherBox) Descriptor() ([]byte, []int) {
	return fileDescriptor_165b784e4d2471a2, []int{8}
}

func (m *CipherBox) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CipherBox.Unmarshal(m, b)
}
func (m *CipherBox) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CipherBox.Marshal(b, m, deterministic)
}
func (m *CipherBox) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CipherBox.Merge(m, src)
}
func (m *CipherBox) XXX_Size() int {
	return xxx_messageInfo_CipherBox.Size(m)
}
func (m *CipherBox) XXX_DiscardUnknown() {
	xxx_messageInfo_CipherBox.DiscardUnknown(m)
}

var xxx_messageInfo_CipherBox proto.InternalMessageInfo

func (m *CipherBox) GetDesc() *CipherBoxDesc {
	if m != nil {
		return m.Desc
	}
	return nil
}

func (m *CipherBox) GetMsg() []byte {
	if m != nil {
		return m.Msg
	}
	return nil
}

func init() {
	proto.RegisterType((*PoolAccountAuth)(nil), "poolrpc.PoolAccountAuth")
	proto.RegisterType((*SidecarAuth)(nil), "poolrpc.SidecarAuth")
	proto.RegisterType((*CipherBoxInit)(nil), "poolrpc.CipherBoxInit")
	proto.RegisterType((*CipherChallenge)(nil), "poolrpc.CipherChallenge")
	proto.RegisterType((*CipherError)(nil), "poolrpc.CipherError")
	proto.RegisterType((*CipherSuccess)(nil), "poolrpc.CipherSuccess")
	proto.RegisterType((*CipherInitResp)(nil), "poolrpc.CipherInitResp")
	proto.RegisterType((*CipherBoxDesc)(nil), "poolrpc.CipherBoxDesc")
	proto.RegisterType((*CipherBox)(nil), "poolrpc.CipherBox")
}

func init() { proto.RegisterFile("hashmail.proto", fileDescriptor_165b784e4d2471a2) }

var fileDescriptor_165b784e4d2471a2 = []byte{
	// 480 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x53, 0x5f, 0x6f, 0xd3, 0x3e,
	0x14, 0x6d, 0x7e, 0xdb, 0xaf, 0x7f, 0x6e, 0xba, 0x0d, 0x2c, 0x34, 0x4a, 0x11, 0x12, 0x8a, 0x84,
	0x34, 0xc1, 0x68, 0x51, 0x79, 0xe0, 0xdf, 0x03, 0x5a, 0x07, 0x52, 0xaa, 0x09, 0x84, 0x9c, 0x37,
	0x5e, 0x2a, 0xd7, 0xb9, 0x4a, 0xac, 0xa5, 0x71, 0x64, 0x3b, 0xc0, 0xbe, 0x17, 0x12, 0x1f, 0x80,
	0x2f, 0x86, 0xec, 0xa4, 0x69, 0xd5, 0xae, 0x0f, 0x7b, 0x8b, 0x6f, 0xce, 0x39, 0x39, 0xe7, 0xf8,
	0x06, 0x8e, 0x53, 0xa6, 0xd3, 0x25, 0x13, 0xd9, 0xa8, 0x50, 0xd2, 0x48, 0xd2, 0x29, 0xa4, 0xcc,
	0x54, 0xc1, 0x83, 0x2b, 0x38, 0xf9, 0x26, 0x65, 0x76, 0xc1, 0xb9, 0x2c, 0x73, 0x73, 0x51, 0x9a,
	0x94, 0x3c, 0x82, 0x2e, 0xe3, 0xdc, 0xcc, 0xaf, 0xf1, 0x66, 0xe0, 0x3d, 0xf5, 0xce, 0xfa, 0xb4,
	0x63, 0xcf, 0x57, 0x78, 0x43, 0x9e, 0x00, 0x68, 0xa3, 0x90, 0x2d, 0xe7, 0x5a, 0x24, 0x83, 0xff,
	0xdc, 0xcb, 0x5e, 0x35, 0x89, 0x44, 0x12, 0x3c, 0x03, 0x3f, 0x12, 0x31, 0x72, 0xa6, 0x9c, 0xd0,
	0x29, 0xb4, 0x8d, 0xe0, 0xd7, 0x68, 0x9c, 0x4c, 0x8f, 0xd6, 0xa7, 0xe0, 0x8f, 0x07, 0x47, 0x97,
	0xa2, 0x48, 0x51, 0x4d, 0xe5, 0xaf, 0x59, 0x2e, 0x0c, 0x79, 0x0e, 0x87, 0x31, 0x6a, 0xee, 0x70,
	0xfe, 0xe4, 0x74, 0x54, 0xbb, 0x1b, 0x35, 0xa8, 0x4f, 0xa8, 0x39, 0x75, 0x18, 0xf2, 0x06, 0x7a,
	0xce, 0x1e, 0x2b, 0x4d, 0xea, 0x2c, 0xf8, 0x93, 0x41, 0x43, 0xd8, 0xca, 0x12, 0xb6, 0xa8, 0xcb,
	0xe2, 0xec, 0xbc, 0x83, 0xbe, 0xae, 0xdc, 0x55, 0xdc, 0x03, 0xc7, 0x7d, 0xd0, 0x70, 0x37, 0xac,
	0x87, 0x2d, 0xea, 0xeb, 0xf5, 0x71, 0xda, 0x86, 0x43, 0x4b, 0x09, 0xee, 0xc3, 0x49, 0x65, 0xe9,
	0x32, 0x65, 0x59, 0x86, 0x79, 0x82, 0xc1, 0x11, 0xf8, 0xd5, 0xe8, 0xb3, 0x52, 0x52, 0x05, 0x1f,
	0x56, 0xd1, 0xa2, 0x92, 0x73, 0xd4, 0xfa, 0x2e, 0xd1, 0x82, 0xdf, 0x1e, 0x1c, 0x57, 0x73, 0xdb,
	0x0a, 0x45, 0x5d, 0x90, 0x09, 0x74, 0x74, 0xa5, 0xb4, 0x47, 0xa1, 0xfe, 0x4e, 0xd8, 0xa2, 0x2b,
	0x20, 0x79, 0x0b, 0x3d, 0xbe, 0xf2, 0xb7, 0xd3, 0xd0, 0x96, 0xff, 0xb0, 0x45, 0xd7, 0x60, 0x72,
	0x0e, 0xff, 0xa3, 0x8d, 0xb1, 0xd3, 0xcd, 0x46, 0xc4, 0xb0, 0x45, 0x2b, 0x90, 0x6d, 0x45, 0xa1,
	0x2e, 0x82, 0xf3, 0x8d, 0xeb, 0xb4, 0x69, 0xc8, 0x63, 0xa8, 0x97, 0x62, 0x2e, 0xe2, 0x7a, 0x85,
	0xba, 0xd5, 0x60, 0x16, 0x07, 0x33, 0xe8, 0x35, 0xe8, 0x3b, 0x5d, 0xfc, 0x3d, 0x38, 0x58, 0xea,
	0xd5, 0xd6, 0xd9, 0xc7, 0xc9, 0x5f, 0x0f, 0xba, 0x21, 0xd3, 0xe9, 0x17, 0x26, 0x32, 0xf2, 0x11,
	0xfa, 0x5f, 0xf1, 0xe7, 0x5a, 0xfa, 0x16, 0x31, 0xdb, 0xea, 0xf0, 0xe1, 0xd6, 0xbc, 0xa9, 0xfa,
	0x3d, 0x40, 0x84, 0x79, 0x1c, 0x39, 0xa3, 0x84, 0xec, 0xd2, 0x87, 0x7b, 0xfc, 0x9d, 0x79, 0x96,
	0x4b, 0x91, 0xff, 0xa8, 0xb9, 0x7b, 0x70, 0xc3, 0x5b, 0x34, 0x5f, 0x79, 0xd3, 0x97, 0xdf, 0x5f,
	0x24, 0xc2, 0xa4, 0xe5, 0x62, 0xc4, 0xe5, 0x72, 0x9c, 0x89, 0x24, 0x35, 0xb9, 0xc8, 0x93, 0x8c,
	0x2d, 0xf4, 0xd8, 0xe2, 0xc7, 0xac, 0xe4, 0x46, 0xc8, 0x1c, 0x51, 0xa9, 0x82, 0x2f, 0xda, 0xee,
	0x0f, 0x7e, 0xfd, 0x2f, 0x00, 0x00, 0xff, 0xff, 0x31, 0x19, 0x39, 0x7c, 0xd3, 0x03, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// HashMailClient is the client API for HashMail service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type HashMailClient interface {
	//
	//NewCipherBox creates a new cipher box pipe/stream given a valid
	//authentication mechanism. If the authentication mechanism has been revoked,
	//or needs to be changed, then a CipherChallenge message is returned.
	//Otherwise the method will either be accepted or rejected.
	NewCipherBox(ctx context.Context, in *CipherBoxInit, opts ...grpc.CallOption) (*CipherInitResp, error)
	//
	//SendStream opens up the write side of the passed CipherBox pipe. Writes
	//will be non-blocking up to the buffer size of the pipe. Beyond that writes
	//will block until completed.
	SendStream(ctx context.Context, opts ...grpc.CallOption) (HashMail_SendStreamClient, error)
	//
	//RecvStream opens up the read side of the passed CipherBox pipe. This method
	//will block until a full message has been read as this is a message based
	//pipe/stream abstraction.
	RecvStream(ctx context.Context, in *CipherBoxDesc, opts ...grpc.CallOption) (HashMail_RecvStreamClient, error)
}

type hashMailClient struct {
	cc *grpc.ClientConn
}

func NewHashMailClient(cc *grpc.ClientConn) HashMailClient {
	return &hashMailClient{cc}
}

func (c *hashMailClient) NewCipherBox(ctx context.Context, in *CipherBoxInit, opts ...grpc.CallOption) (*CipherInitResp, error) {
	out := new(CipherInitResp)
	err := c.cc.Invoke(ctx, "/poolrpc.HashMail/NewCipherBox", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hashMailClient) SendStream(ctx context.Context, opts ...grpc.CallOption) (HashMail_SendStreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &_HashMail_serviceDesc.Streams[0], "/poolrpc.HashMail/SendStream", opts...)
	if err != nil {
		return nil, err
	}
	x := &hashMailSendStreamClient{stream}
	return x, nil
}

type HashMail_SendStreamClient interface {
	Send(*CipherBox) error
	CloseAndRecv() (*CipherBoxDesc, error)
	grpc.ClientStream
}

type hashMailSendStreamClient struct {
	grpc.ClientStream
}

func (x *hashMailSendStreamClient) Send(m *CipherBox) error {
	return x.ClientStream.SendMsg(m)
}

func (x *hashMailSendStreamClient) CloseAndRecv() (*CipherBoxDesc, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(CipherBoxDesc)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *hashMailClient) RecvStream(ctx context.Context, in *CipherBoxDesc, opts ...grpc.CallOption) (HashMail_RecvStreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &_HashMail_serviceDesc.Streams[1], "/poolrpc.HashMail/RecvStream", opts...)
	if err != nil {
		return nil, err
	}
	x := &hashMailRecvStreamClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type HashMail_RecvStreamClient interface {
	Recv() (*CipherBox, error)
	grpc.ClientStream
}

type hashMailRecvStreamClient struct {
	grpc.ClientStream
}

func (x *hashMailRecvStreamClient) Recv() (*CipherBox, error) {
	m := new(CipherBox)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// HashMailServer is the server API for HashMail service.
type HashMailServer interface {
	//
	//NewCipherBox creates a new cipher box pipe/stream given a valid
	//authentication mechanism. If the authentication mechanism has been revoked,
	//or needs to be changed, then a CipherChallenge message is returned.
	//Otherwise the method will either be accepted or rejected.
	NewCipherBox(context.Context, *CipherBoxInit) (*CipherInitResp, error)
	//
	//SendStream opens up the write side of the passed CipherBox pipe. Writes
	//will be non-blocking up to the buffer size of the pipe. Beyond that writes
	//will block until completed.
	SendStream(HashMail_SendStreamServer) error
	//
	//RecvStream opens up the read side of the passed CipherBox pipe. This method
	//will block until a full message has been read as this is a message based
	//pipe/stream abstraction.
	RecvStream(*CipherBoxDesc, HashMail_RecvStreamServer) error
}

// UnimplementedHashMailServer can be embedded to have forward compatible implementations.
type UnimplementedHashMailServer struct {
}

func (*UnimplementedHashMailServer) NewCipherBox(ctx context.Context, req *CipherBoxInit) (*CipherInitResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NewCipherBox not implemented")
}
func (*UnimplementedHashMailServer) SendStream(srv HashMail_SendStreamServer) error {
	return status.Errorf(codes.Unimplemented, "method SendStream not implemented")
}
func (*UnimplementedHashMailServer) RecvStream(req *CipherBoxDesc, srv HashMail_RecvStreamServer) error {
	return status.Errorf(codes.Unimplemented, "method RecvStream not implemented")
}

func RegisterHashMailServer(s *grpc.Server, srv HashMailServer) {
	s.RegisterService(&_HashMail_serviceDesc, srv)
}

func _HashMail_NewCipherBox_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CipherBoxInit)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HashMailServer).NewCipherBox(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/poolrpc.HashMail/NewCipherBox",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HashMailServer).NewCipherBox(ctx, req.(*CipherBoxInit))
	}
	return interceptor(ctx, in, info, handler)
}

func _HashMail_SendStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(HashMailServer).SendStream(&hashMailSendStreamServer{stream})
}

type HashMail_SendStreamServer interface {
	SendAndClose(*CipherBoxDesc) error
	Recv() (*CipherBox, error)
	grpc.ServerStream
}

type hashMailSendStreamServer struct {
	grpc.ServerStream
}

func (x *hashMailSendStreamServer) SendAndClose(m *CipherBoxDesc) error {
	return x.ServerStream.SendMsg(m)
}

func (x *hashMailSendStreamServer) Recv() (*CipherBox, error) {
	m := new(CipherBox)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _HashMail_RecvStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(CipherBoxDesc)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(HashMailServer).RecvStream(m, &hashMailRecvStreamServer{stream})
}

type HashMail_RecvStreamServer interface {
	Send(*CipherBox) error
	grpc.ServerStream
}

type hashMailRecvStreamServer struct {
	grpc.ServerStream
}

func (x *hashMailRecvStreamServer) Send(m *CipherBox) error {
	return x.ServerStream.SendMsg(m)
}

var _HashMail_serviceDesc = grpc.ServiceDesc{
	ServiceName: "poolrpc.HashMail",
	HandlerType: (*HashMailServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "NewCipherBox",
			Handler:    _HashMail_NewCipherBox_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "SendStream",
			Handler:       _HashMail_SendStream_Handler,
			ClientStreams: true,
		},
		{
			StreamName:    "RecvStream",
			Handler:       _HashMail_RecvStream_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "hashmail.proto",
}
