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
var WellKnownTypes = map[reflect2.Type]bool{
	reflect2.TypeOfPtr((*anypb.Any)(nil)).Elem():              true,
	reflect2.TypeOfPtr((*timestamppb.Timestamp)(nil)).Elem():  true,
	reflect2.TypeOfPtr((*durationpb.Duration)(nil)).Elem():    true,
	reflect2.TypeOfPtr((*wrapperspb.BoolValue)(nil)).Elem():   true,
	reflect2.TypeOfPtr((*wrapperspb.Int32Value)(nil)).Elem():  true,
	reflect2.TypeOfPtr((*wrapperspb.Int64Value)(nil)).Elem():  true,
	reflect2.TypeOfPtr((*wrapperspb.UInt32Value)(nil)).Elem(): true,
	reflect2.TypeOfPtr((*wrapperspb.UInt64Value)(nil)).Elem(): true,
	reflect2.TypeOfPtr((*wrapperspb.FloatValue)(nil)).Elem():  true,
	reflect2.TypeOfPtr((*wrapperspb.DoubleValue)(nil)).Elem(): true,
	reflect2.TypeOfPtr((*wrapperspb.StringValue)(nil)).Elem(): true,
	reflect2.TypeOfPtr((*wrapperspb.BytesValue)(nil)).Elem():  true,
	reflect2.TypeOfPtr((*structpb.Struct)(nil)).Elem():        true,
	reflect2.TypeOfPtr((*structpb.ListValue)(nil)).Elem():     true,
	reflect2.TypeOfPtr((*structpb.Value)(nil)).Elem():         true,
	reflect2.TypeOfPtr((*fieldmaskpb.FieldMask)(nil)).Elem():  true,
	reflect2.TypeOfPtr((*emptypb.Empty)(nil)).Elem():          true,
}

func IsWellKnownType(typ reflect2.Type) bool {
	return WellKnownTypes[typ]
}

var ProtoMessageCodecs = map[reflect2.Type]*Codec{
	reflect2.TypeOfPtr((*anypb.Any)(nil)).Elem(): wktAnyCodec,

	reflect2.TypeOfPtr((*timestamppb.Timestamp)(nil)).Elem(): wktTimestampCodec,
	reflect2.TypeOfPtr((*durationpb.Duration)(nil)).Elem():   wktDurationCodec,

	reflect2.TypeOfPtr((*wrapperspb.BoolValue)(nil)).Elem(): NewElemTypeCodec(
		func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
			stream.WriteBool(((*wrapperspb.BoolValue)(ptr)).GetValue())
		},
		func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
			(*wrapperspb.BoolValue)(ptr).Value = iter.ReadBool()
		},
	),
	reflect2.TypeOfPtr((*wrapperspb.Int32Value)(nil)).Elem(): NewElemTypeCodec(
		func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
			stream.WriteInt32(((*wrapperspb.Int32Value)(ptr)).GetValue())
		},
		func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
			(*wrapperspb.Int32Value)(ptr).Value = iter.ReadInt32()
		},
	),
	// "NaN" "Infinity" "-Infinity" handle??? // TODO:
	reflect2.TypeOfPtr((*wrapperspb.Int64Value)(nil)).Elem(): NewElemTypeCodec(
		func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
			stream.WriteInt64(((*wrapperspb.Int64Value)(ptr)).GetValue())
		},
		func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
			(*wrapperspb.Int64Value)(ptr).Value = iter.ReadInt64()
		},
	),
	reflect2.TypeOfPtr((*wrapperspb.UInt32Value)(nil)).Elem(): NewElemTypeCodec(
		func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
			stream.WriteUint32(((*wrapperspb.UInt32Value)(ptr)).GetValue())
		},
		func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
			(*wrapperspb.UInt32Value)(ptr).Value = iter.ReadUint32()
		},
	),
	reflect2.TypeOfPtr((*wrapperspb.UInt64Value)(nil)).Elem(): NewElemTypeCodec(
		func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
			stream.WriteUint64(((*wrapperspb.UInt64Value)(ptr)).GetValue())
		},
		func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
			(*wrapperspb.UInt64Value)(ptr).Value = iter.ReadUint64()
		},
	),
	reflect2.TypeOfPtr((*wrapperspb.FloatValue)(nil)).Elem(): NewElemTypeCodec(
		func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
			stream.WriteFloat32(((*wrapperspb.FloatValue)(ptr)).GetValue())
		},
		func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
			(*wrapperspb.FloatValue)(ptr).Value = iter.ReadFloat32()
		},
	),
	reflect2.TypeOfPtr((*wrapperspb.DoubleValue)(nil)).Elem(): NewElemTypeCodec(
		func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
			// TODO: stream.WriteFloat64Lossy ???
			stream.WriteFloat64(((*wrapperspb.DoubleValue)(ptr)).GetValue())
		},
		func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
			(*wrapperspb.DoubleValue)(ptr).Value = iter.ReadFloat64()
		},
	),
	reflect2.TypeOfPtr((*wrapperspb.StringValue)(nil)).Elem(): NewElemTypeCodec(
		func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
			stream.WriteString(((*wrapperspb.StringValue)(ptr)).GetValue())
		},
		func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
			(*wrapperspb.StringValue)(ptr).Value = iter.ReadString()
		},
	),
	reflect2.TypeOfPtr((*wrapperspb.BytesValue)(nil)).Elem(): NewElemTypeCodec(
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
				iter.ReportError("protobuf", fmt.Sprintf("google.protobuf.BytesValue: %v", err))
				return
			}
			(*wrapperspb.BytesValue)(ptr).Value = b
		},
	),

	// TODO: 这三个本身就实现了 json.Marshaler 所以这里设置了也没意义，暂时很无奈
	reflect2.TypeOfPtr((*structpb.Struct)(nil)).Elem():    nil,
	reflect2.TypeOfPtr((*structpb.ListValue)(nil)).Elem(): nil,
	reflect2.TypeOfPtr((*structpb.Value)(nil)).Elem():     wktValueCodec,

	reflect2.TypeOfPtr((*fieldmaskpb.FieldMask)(nil)).Elem(): wktFieldmaskCodec,
	reflect2.TypeOfPtr((*emptypb.Empty)(nil)).Elem(): NewElemTypeCodec(
		func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
			stream.WriteObjectStart()
			stream.WriteObjectEnd()
		},
		func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
			iter.Skip()
		},
	),
}
