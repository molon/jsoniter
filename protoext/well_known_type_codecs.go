package protoext

import (
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

	reflect2.TypeOf((*timestamppb.Timestamp)(nil)): nil,

	reflect2.TypeOf((*durationpb.Duration)(nil)): nil,

	reflect2.TypeOf((*wrapperspb.BoolValue)(nil)):  nil,
	reflect2.TypeOf((*wrapperspb.Int32Value)(nil)): nil,

	reflect2.TypeOf((*wrapperspb.Int64Value)(nil)): NewPtrTypeCodec(
		reflect2.TypeOfPtr((*wrapperspb.Int64Value)(nil)),
		func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
			stream.WriteInt64(((*wrapperspb.Int64Value)(ptr)).GetValue())
		},
		func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
			(*wrapperspb.Int64Value)(ptr).Value = iter.ReadInt64()
		},
	),
	reflect2.TypeOf((*wrapperspb.UInt32Value)(nil)): nil,
	reflect2.TypeOf((*wrapperspb.UInt64Value)(nil)): NewPtrTypeCodec(
		reflect2.TypeOfPtr((*wrapperspb.UInt64Value)(nil)),
		func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
			stream.WriteUint64(((*wrapperspb.UInt64Value)(ptr)).GetValue())
		},
		func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
			(*wrapperspb.UInt64Value)(ptr).Value = iter.ReadUint64()
		},
	),
	reflect2.TypeOf((*wrapperspb.FloatValue)(nil)):  nil,
	reflect2.TypeOf((*wrapperspb.DoubleValue)(nil)): nil,
	reflect2.TypeOf((*wrapperspb.StringValue)(nil)): nil,
	reflect2.TypeOf((*wrapperspb.BytesValue)(nil)):  nil,
	reflect2.TypeOf((*structpb.Struct)(nil)):        nil,
	reflect2.TypeOf((*structpb.ListValue)(nil)):     nil,
	reflect2.TypeOf((*structpb.Value)(nil)):         nil,
	reflect2.TypeOf((*fieldmaskpb.FieldMask)(nil)):  nil,
	reflect2.TypeOf((*emptypb.Empty)(nil)):          nil,
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
