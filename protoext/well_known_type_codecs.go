package protoext

import (
	"encoding/base64"
	"fmt"
	"strings"
	"unsafe"

	jsoniter "github.com/json-iterator/go"
	"github.com/modern-go/reflect2"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/structpb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// https://github.com/protocolbuffers/protobuf-go/blob/master/encoding/protojson/well_known_types.go
var WellKnownTypeCodecs = map[reflect2.Type]*Codec{
	reflect2.TypeOf((*anypb.Any)(nil)): nil,

	reflect2.TypeOf((*timestamppb.Timestamp)(nil)): timestampCodec,
	reflect2.TypeOf((*durationpb.Duration)(nil)):   durationCodec,

	reflect2.TypeOf((*wrapperspb.BoolValue)(nil)): NewPtrTypeCodec(
		reflect2.TypeOfPtr((*wrapperspb.BoolValue)(nil)),
		func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
			stream.WriteBool(((*wrapperspb.BoolValue)(ptr)).GetValue())
		},
		func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
			(*wrapperspb.BoolValue)(ptr).Value = iter.ReadBool()
		},
	),
	reflect2.TypeOf((*wrapperspb.Int32Value)(nil)): NewPtrTypeCodec(
		reflect2.TypeOfPtr((*wrapperspb.Int32Value)(nil)),
		func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
			stream.WriteInt32(((*wrapperspb.Int32Value)(ptr)).GetValue())
		},
		func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
			(*wrapperspb.Int32Value)(ptr).Value = iter.ReadInt32()
		},
	),
	reflect2.TypeOf((*wrapperspb.Int64Value)(nil)): NewPtrTypeCodec(
		reflect2.TypeOfPtr((*wrapperspb.Int64Value)(nil)),
		func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
			stream.WriteInt64(((*wrapperspb.Int64Value)(ptr)).GetValue())
		},
		func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
			(*wrapperspb.Int64Value)(ptr).Value = iter.ReadInt64()
		},
	),
	reflect2.TypeOf((*wrapperspb.UInt32Value)(nil)): NewPtrTypeCodec(
		reflect2.TypeOfPtr((*wrapperspb.UInt32Value)(nil)),
		func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
			stream.WriteUint32(((*wrapperspb.UInt32Value)(ptr)).GetValue())
		},
		func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
			(*wrapperspb.UInt32Value)(ptr).Value = iter.ReadUint32()
		},
	),
	reflect2.TypeOf((*wrapperspb.UInt64Value)(nil)): NewPtrTypeCodec(
		reflect2.TypeOfPtr((*wrapperspb.UInt64Value)(nil)),
		func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
			stream.WriteUint64(((*wrapperspb.UInt64Value)(ptr)).GetValue())
		},
		func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
			(*wrapperspb.UInt64Value)(ptr).Value = iter.ReadUint64()
		},
	),
	reflect2.TypeOf((*wrapperspb.FloatValue)(nil)): NewPtrTypeCodec(
		reflect2.TypeOfPtr((*wrapperspb.FloatValue)(nil)),
		func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
			stream.WriteFloat32(((*wrapperspb.FloatValue)(ptr)).GetValue())
		},
		func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
			(*wrapperspb.FloatValue)(ptr).Value = iter.ReadFloat32()
		},
	),
	reflect2.TypeOf((*wrapperspb.DoubleValue)(nil)): NewPtrTypeCodec(
		reflect2.TypeOfPtr((*wrapperspb.DoubleValue)(nil)),
		func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
			// TODO: stream.WriteFloat64Lossy ???
			stream.WriteFloat64(((*wrapperspb.DoubleValue)(ptr)).GetValue())
		},
		func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
			(*wrapperspb.DoubleValue)(ptr).Value = iter.ReadFloat64()
		},
	),
	reflect2.TypeOf((*wrapperspb.StringValue)(nil)): NewPtrTypeCodec(
		reflect2.TypeOfPtr((*wrapperspb.StringValue)(nil)),
		func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
			stream.WriteString(((*wrapperspb.StringValue)(ptr)).GetValue())
		},
		func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
			(*wrapperspb.StringValue)(ptr).Value = iter.ReadString()
		},
	),
	reflect2.TypeOf((*wrapperspb.BytesValue)(nil)): NewPtrTypeCodec(
		reflect2.TypeOfPtr((*wrapperspb.BytesValue)(nil)),
		func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
			stream.WriteString(
				base64.StdEncoding.EncodeToString(((*wrapperspb.BytesValue)(ptr)).GetValue()),
			)
		},
		func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
			s := iter.ReadString()
			enc := base64.StdEncoding
			if strings.ContainsAny(s, "-_") {
				enc = base64.URLEncoding
			}
			if len(s)%4 != 0 {
				enc = enc.WithPadding(base64.NoPadding)
			}
			b, err := enc.DecodeString(s)
			if err != nil {
				iter.ReportError("decode", fmt.Sprintf("google.protobuf.BytesValue: %v", err))
				return
			}
			(*wrapperspb.BytesValue)(ptr).Value = b
		},
	),
	reflect2.TypeOf((*structpb.Struct)(nil)):       nil,
	reflect2.TypeOf((*structpb.ListValue)(nil)):    nil,
	reflect2.TypeOf((*structpb.Value)(nil)):        nil,
	reflect2.TypeOf((*fieldmaskpb.FieldMask)(nil)): nil,
	reflect2.TypeOf((*emptypb.Empty)(nil)): NewPtrTypeCodec(
		reflect2.TypeOfPtr((*emptypb.Empty)(nil)),
		func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
			stream.WriteObjectStart()
			stream.WriteObjectEnd()
		},
		func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
			iter.Skip()
		},
	),
}

type Codec struct {
	Encoder jsoniter.ValEncoder
	Decoder jsoniter.ValDecoder
}

func NewPtrTypeCodec(typ reflect2.PtrType, encodeFunc jsoniter.EncoderFunc, decodeFunc jsoniter.DecoderFunc) *Codec {
	c := &Codec{}
	c.Encoder = &jsoniter.OptionalEncoder{
		ValueEncoder: &funcEncoder{
			fun: encodeFunc,
		},
	}
	c.Decoder = &jsoniter.OptionalDecoder{
		ValueType: typ.Elem(),
		ValueDecoder: &funcDecoder{
			fun: decodeFunc,
		},
	}
	return c
}
