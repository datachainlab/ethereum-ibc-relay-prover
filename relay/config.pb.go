// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: relayer/provers/ethereum_light_client/config/config.proto

package relay

import (
	fmt "fmt"
	_ "github.com/cosmos/cosmos-sdk/codec/types"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type ProverConfig struct {
	BeaconEndpoint       string            `protobuf:"bytes,1,opt,name=beacon_endpoint,json=beaconEndpoint,proto3" json:"beacon_endpoint,omitempty"`
	Network              string            `protobuf:"bytes,2,opt,name=network,proto3" json:"network,omitempty"`
	TrustingPeriod       string            `protobuf:"bytes,3,opt,name=trusting_period,json=trustingPeriod,proto3" json:"trusting_period,omitempty"`
	MaxClockDrift        string            `protobuf:"bytes,4,opt,name=max_clock_drift,json=maxClockDrift,proto3" json:"max_clock_drift,omitempty"`
	RefreshThresholdRate *Fraction         `protobuf:"bytes,5,opt,name=refresh_threshold_rate,json=refreshThresholdRate,proto3" json:"refresh_threshold_rate,omitempty"`
	MinimalForkSched     map[string]uint64 `protobuf:"bytes,6,rep,name=minimal_fork_sched,json=minimalForkSched,proto3" json:"minimal_fork_sched,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
}

func (m *ProverConfig) Reset()         { *m = ProverConfig{} }
func (m *ProverConfig) String() string { return proto.CompactTextString(m) }
func (*ProverConfig) ProtoMessage()    {}
func (*ProverConfig) Descriptor() ([]byte, []int) {
	return fileDescriptor_f862fff68264353b, []int{0}
}
func (m *ProverConfig) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *ProverConfig) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_ProverConfig.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *ProverConfig) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ProverConfig.Merge(m, src)
}
func (m *ProverConfig) XXX_Size() int {
	return m.Size()
}
func (m *ProverConfig) XXX_DiscardUnknown() {
	xxx_messageInfo_ProverConfig.DiscardUnknown(m)
}

var xxx_messageInfo_ProverConfig proto.InternalMessageInfo

type Fraction struct {
	Numerator   uint64 `protobuf:"varint,1,opt,name=numerator,proto3" json:"numerator,omitempty"`
	Denominator uint64 `protobuf:"varint,2,opt,name=denominator,proto3" json:"denominator,omitempty"`
}

func (m *Fraction) Reset()         { *m = Fraction{} }
func (m *Fraction) String() string { return proto.CompactTextString(m) }
func (*Fraction) ProtoMessage()    {}
func (*Fraction) Descriptor() ([]byte, []int) {
	return fileDescriptor_f862fff68264353b, []int{1}
}
func (m *Fraction) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Fraction) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Fraction.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Fraction) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Fraction.Merge(m, src)
}
func (m *Fraction) XXX_Size() int {
	return m.Size()
}
func (m *Fraction) XXX_DiscardUnknown() {
	xxx_messageInfo_Fraction.DiscardUnknown(m)
}

var xxx_messageInfo_Fraction proto.InternalMessageInfo

func init() {
	proto.RegisterType((*ProverConfig)(nil), "relayer.provers.ethereum_light_client.config.ProverConfig")
	proto.RegisterMapType((map[string]uint64)(nil), "relayer.provers.ethereum_light_client.config.ProverConfig.MinimalForkSchedEntry")
	proto.RegisterType((*Fraction)(nil), "relayer.provers.ethereum_light_client.config.Fraction")
}

func init() {
	proto.RegisterFile("relayer/provers/ethereum_light_client/config/config.proto", fileDescriptor_f862fff68264353b)
}

var fileDescriptor_f862fff68264353b = []byte{
	// 472 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x92, 0xcf, 0x6f, 0xd3, 0x30,
	0x14, 0xc7, 0x9b, 0xb5, 0x1b, 0xcc, 0x05, 0x36, 0x59, 0x05, 0x85, 0x0a, 0x45, 0xd5, 0x0e, 0xd0,
	0x03, 0x4d, 0xa4, 0x21, 0xf1, 0xeb, 0x48, 0xd9, 0x0e, 0x48, 0x48, 0x55, 0x80, 0x0b, 0x17, 0xcb,
	0x71, 0x5e, 0x13, 0x2b, 0x8e, 0x5f, 0xe5, 0x3a, 0x63, 0xbd, 0xf0, 0x37, 0xf0, 0x67, 0xed, 0xb8,
	0x0b, 0x12, 0x47, 0x68, 0xff, 0x11, 0x14, 0x27, 0x85, 0x69, 0xda, 0x65, 0x27, 0xfb, 0x7d, 0xfc,
	0xbe, 0x9f, 0x28, 0xcf, 0x26, 0x6f, 0x0c, 0x28, 0xbe, 0x02, 0x13, 0x2d, 0x0c, 0x9e, 0x81, 0x59,
	0x46, 0x60, 0x73, 0x30, 0x50, 0x95, 0x4c, 0xc9, 0x2c, 0xb7, 0x4c, 0x28, 0x09, 0xda, 0x46, 0x02,
	0xf5, 0x5c, 0x66, 0xed, 0x12, 0x2e, 0x0c, 0x5a, 0xa4, 0xcf, 0xdb, 0x68, 0xd8, 0x46, 0xc3, 0x1b,
	0xa3, 0x61, 0x93, 0x19, 0x0e, 0x32, 0xcc, 0xd0, 0x05, 0xa3, 0x7a, 0xd7, 0x38, 0x86, 0x8f, 0x33,
	0xc4, 0x4c, 0x41, 0xe4, 0xaa, 0xa4, 0x9a, 0x47, 0x5c, 0xaf, 0x9a, 0xa3, 0xa3, 0x9f, 0x5d, 0x72,
	0x6f, 0xe6, 0xcc, 0x53, 0x67, 0xa0, 0xcf, 0xc8, 0x41, 0x02, 0x5c, 0xa0, 0x66, 0xa0, 0xd3, 0x05,
	0x4a, 0x6d, 0x7d, 0x6f, 0xe4, 0x8d, 0xf7, 0xe3, 0x07, 0x0d, 0x3e, 0x69, 0x29, 0xf5, 0xc9, 0x1d,
	0x0d, 0xf6, 0x1b, 0x9a, 0xc2, 0xdf, 0x71, 0x0d, 0xdb, 0xb2, 0x56, 0x58, 0x53, 0x2d, 0xad, 0xd4,
	0x19, 0x5b, 0x80, 0x91, 0x98, 0xfa, 0xdd, 0x46, 0xb1, 0xc5, 0x33, 0x47, 0xe9, 0x53, 0x72, 0x50,
	0xf2, 0x73, 0x26, 0x14, 0x8a, 0x82, 0xa5, 0x46, 0xce, 0xad, 0xdf, 0x73, 0x8d, 0xf7, 0x4b, 0x7e,
	0x3e, 0xad, 0xe9, 0xfb, 0x1a, 0x52, 0x45, 0x1e, 0x19, 0x98, 0x1b, 0x58, 0xe6, 0xcc, 0xe6, 0xf5,
	0x82, 0x2a, 0x65, 0x86, 0x5b, 0xf0, 0x77, 0x47, 0xde, 0xb8, 0x7f, 0xfc, 0x32, 0xbc, 0xcd, 0x90,
	0xc2, 0x53, 0xc3, 0x85, 0x95, 0xa8, 0xe3, 0x41, 0x6b, 0xfd, 0xbc, 0x95, 0xc6, 0xdc, 0x02, 0xfd,
	0x4e, 0x68, 0x29, 0xb5, 0x2c, 0xb9, 0x62, 0x73, 0x34, 0x05, 0x5b, 0x8a, 0x1c, 0x52, 0x7f, 0x6f,
	0xd4, 0x1d, 0xf7, 0x8f, 0x67, 0xb7, 0xfb, 0xd2, 0xd5, 0xc9, 0x86, 0x1f, 0x1b, 0xe9, 0x29, 0x9a,
	0xe2, 0x53, 0xad, 0x3c, 0xd1, 0xd6, 0xac, 0xe2, 0xc3, 0xf2, 0x1a, 0x1e, 0x4e, 0xc9, 0xc3, 0x1b,
	0x5b, 0xe9, 0x21, 0xe9, 0x16, 0xb0, 0x6a, 0xaf, 0xa3, 0xde, 0xd2, 0x01, 0xd9, 0x3d, 0xe3, 0xaa,
	0x02, 0x77, 0x03, 0xbd, 0xb8, 0x29, 0xde, 0xee, 0xbc, 0xf6, 0x8e, 0x3e, 0x90, 0xbb, 0xdb, 0xdf,
	0xa4, 0x4f, 0xc8, 0xbe, 0xae, 0x4a, 0x30, 0xdc, 0xa2, 0x71, 0xe9, 0x5e, 0xfc, 0x1f, 0xd0, 0x11,
	0xe9, 0xa7, 0xa0, 0xb1, 0x94, 0xda, 0x9d, 0x37, 0xa6, 0xab, 0xe8, 0xdd, 0x97, 0x8b, 0x3f, 0x41,
	0xe7, 0x62, 0x1d, 0x78, 0x97, 0xeb, 0xc0, 0xfb, 0xbd, 0x0e, 0xbc, 0x1f, 0x9b, 0xa0, 0x73, 0xb9,
	0x09, 0x3a, 0xbf, 0x36, 0x41, 0xe7, 0xeb, 0xab, 0x4c, 0xda, 0xbc, 0x4a, 0x42, 0x81, 0x65, 0x94,
	0x72, 0xcb, 0x45, 0xce, 0xa5, 0x56, 0x3c, 0xf9, 0xf7, 0xc6, 0x27, 0x32, 0x11, 0x13, 0x37, 0xb6,
	0x49, 0x33, 0xb4, 0xc8, 0x15, 0xc9, 0x9e, 0x7b, 0x81, 0x2f, 0xfe, 0x06, 0x00, 0x00, 0xff, 0xff,
	0x38, 0x14, 0xdb, 0xb6, 0x1d, 0x03, 0x00, 0x00,
}

func (m *ProverConfig) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ProverConfig) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *ProverConfig) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.MinimalForkSched) > 0 {
		for k := range m.MinimalForkSched {
			v := m.MinimalForkSched[k]
			baseI := i
			i = encodeVarintConfig(dAtA, i, uint64(v))
			i--
			dAtA[i] = 0x10
			i -= len(k)
			copy(dAtA[i:], k)
			i = encodeVarintConfig(dAtA, i, uint64(len(k)))
			i--
			dAtA[i] = 0xa
			i = encodeVarintConfig(dAtA, i, uint64(baseI-i))
			i--
			dAtA[i] = 0x32
		}
	}
	if m.RefreshThresholdRate != nil {
		{
			size, err := m.RefreshThresholdRate.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintConfig(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x2a
	}
	if len(m.MaxClockDrift) > 0 {
		i -= len(m.MaxClockDrift)
		copy(dAtA[i:], m.MaxClockDrift)
		i = encodeVarintConfig(dAtA, i, uint64(len(m.MaxClockDrift)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.TrustingPeriod) > 0 {
		i -= len(m.TrustingPeriod)
		copy(dAtA[i:], m.TrustingPeriod)
		i = encodeVarintConfig(dAtA, i, uint64(len(m.TrustingPeriod)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Network) > 0 {
		i -= len(m.Network)
		copy(dAtA[i:], m.Network)
		i = encodeVarintConfig(dAtA, i, uint64(len(m.Network)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.BeaconEndpoint) > 0 {
		i -= len(m.BeaconEndpoint)
		copy(dAtA[i:], m.BeaconEndpoint)
		i = encodeVarintConfig(dAtA, i, uint64(len(m.BeaconEndpoint)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *Fraction) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Fraction) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Fraction) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Denominator != 0 {
		i = encodeVarintConfig(dAtA, i, uint64(m.Denominator))
		i--
		dAtA[i] = 0x10
	}
	if m.Numerator != 0 {
		i = encodeVarintConfig(dAtA, i, uint64(m.Numerator))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func encodeVarintConfig(dAtA []byte, offset int, v uint64) int {
	offset -= sovConfig(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *ProverConfig) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.BeaconEndpoint)
	if l > 0 {
		n += 1 + l + sovConfig(uint64(l))
	}
	l = len(m.Network)
	if l > 0 {
		n += 1 + l + sovConfig(uint64(l))
	}
	l = len(m.TrustingPeriod)
	if l > 0 {
		n += 1 + l + sovConfig(uint64(l))
	}
	l = len(m.MaxClockDrift)
	if l > 0 {
		n += 1 + l + sovConfig(uint64(l))
	}
	if m.RefreshThresholdRate != nil {
		l = m.RefreshThresholdRate.Size()
		n += 1 + l + sovConfig(uint64(l))
	}
	if len(m.MinimalForkSched) > 0 {
		for k, v := range m.MinimalForkSched {
			_ = k
			_ = v
			mapEntrySize := 1 + len(k) + sovConfig(uint64(len(k))) + 1 + sovConfig(uint64(v))
			n += mapEntrySize + 1 + sovConfig(uint64(mapEntrySize))
		}
	}
	return n
}

func (m *Fraction) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Numerator != 0 {
		n += 1 + sovConfig(uint64(m.Numerator))
	}
	if m.Denominator != 0 {
		n += 1 + sovConfig(uint64(m.Denominator))
	}
	return n
}

func sovConfig(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozConfig(x uint64) (n int) {
	return sovConfig(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *ProverConfig) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowConfig
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: ProverConfig: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ProverConfig: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BeaconEndpoint", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowConfig
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthConfig
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthConfig
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.BeaconEndpoint = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Network", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowConfig
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthConfig
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthConfig
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Network = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TrustingPeriod", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowConfig
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthConfig
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthConfig
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.TrustingPeriod = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxClockDrift", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowConfig
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthConfig
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthConfig
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.MaxClockDrift = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field RefreshThresholdRate", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowConfig
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthConfig
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthConfig
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.RefreshThresholdRate == nil {
				m.RefreshThresholdRate = &Fraction{}
			}
			if err := m.RefreshThresholdRate.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MinimalForkSched", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowConfig
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthConfig
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthConfig
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.MinimalForkSched == nil {
				m.MinimalForkSched = make(map[string]uint64)
			}
			var mapkey string
			var mapvalue uint64
			for iNdEx < postIndex {
				entryPreIndex := iNdEx
				var wire uint64
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowConfig
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					wire |= uint64(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				fieldNum := int32(wire >> 3)
				if fieldNum == 1 {
					var stringLenmapkey uint64
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowConfig
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						stringLenmapkey |= uint64(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					intStringLenmapkey := int(stringLenmapkey)
					if intStringLenmapkey < 0 {
						return ErrInvalidLengthConfig
					}
					postStringIndexmapkey := iNdEx + intStringLenmapkey
					if postStringIndexmapkey < 0 {
						return ErrInvalidLengthConfig
					}
					if postStringIndexmapkey > l {
						return io.ErrUnexpectedEOF
					}
					mapkey = string(dAtA[iNdEx:postStringIndexmapkey])
					iNdEx = postStringIndexmapkey
				} else if fieldNum == 2 {
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowConfig
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						mapvalue |= uint64(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
				} else {
					iNdEx = entryPreIndex
					skippy, err := skipConfig(dAtA[iNdEx:])
					if err != nil {
						return err
					}
					if (skippy < 0) || (iNdEx+skippy) < 0 {
						return ErrInvalidLengthConfig
					}
					if (iNdEx + skippy) > postIndex {
						return io.ErrUnexpectedEOF
					}
					iNdEx += skippy
				}
			}
			m.MinimalForkSched[mapkey] = mapvalue
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipConfig(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthConfig
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *Fraction) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowConfig
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Fraction: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Fraction: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Numerator", wireType)
			}
			m.Numerator = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowConfig
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Numerator |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Denominator", wireType)
			}
			m.Denominator = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowConfig
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Denominator |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipConfig(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthConfig
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipConfig(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowConfig
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowConfig
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowConfig
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthConfig
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupConfig
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthConfig
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthConfig        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowConfig          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupConfig = fmt.Errorf("proto: unexpected end of group")
)
