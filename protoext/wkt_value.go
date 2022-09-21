package protoext

import (
	"errors"
	"fmt"
	"math"
	"unsafe"

	jsoniter "github.com/json-iterator/go"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	Value_message_fullname           protoreflect.FullName = "google.protobuf.Value"
	Value_NumberValue_field_fullname protoreflect.FullName = "google.protobuf.Value.number_value"
)

var wktValueCodec = NewElemTypeCodec(
	func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
		x := ((*structpb.Value)(ptr))
		err := marshalWktValue(x, stream)
		if err != nil {
			stream.Error = fmt.Errorf("%s: %w", Value_message_fullname, err)
			return
		}
	},
	func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
		err := unmarshalWktValue(((*structpb.Value)(ptr)), iter)
		if err != nil {
			iter.ReportError("protobuf", fmt.Sprintf("%s: %v", Value_message_fullname, err))
		}
	},
)

func marshalWktValue(x *structpb.Value, stream *jsoniter.Stream) error {
	switch v := x.GetKind().(type) {
	case *structpb.Value_NullValue:
		if v != nil {
			stream.WriteNil()
			return nil
		}
	case *structpb.Value_NumberValue:
		if v != nil {
			if math.IsNaN(v.NumberValue) || math.IsInf(v.NumberValue, 0) {
				return fmt.Errorf("%s: invalid %v value", Value_NumberValue_field_fullname, v)
			}
			stream.WriteFloat64(v.NumberValue)
			return nil
		}
	case *structpb.Value_StringValue:
		if v != nil {
			stream.WriteString(v.StringValue)
			return nil
		}
	case *structpb.Value_BoolValue:
		if v != nil {
			stream.WriteBool(v.BoolValue)
			return nil
		}
	case *structpb.Value_StructValue:
		if v != nil {
			stream.WriteVal(v.StructValue)
			return nil
		}
	case *structpb.Value_ListValue:
		if v != nil {
			stream.WriteVal(v.ListValue)
			return nil
		}
	}
	// TODO: 如果是在一个数组里不应该出现nil啊？咋办
	stream.WriteNil()
	return nil
}

func unmarshalWktValue(x *structpb.Value, iter *jsoniter.Iterator) error {
	valueType := iter.WhatIsNext()
	switch valueType {
	case jsoniter.NilValue:
		x.Kind = &structpb.Value_NullValue{
			NullValue: structpb.NullValue_NULL_VALUE,
		}
	case jsoniter.BoolValue:
		x.Kind = &structpb.Value_BoolValue{
			BoolValue: iter.ReadBool(),
		}
	case jsoniter.NumberValue:
		x.Kind = &structpb.Value_NumberValue{
			NumberValue: iter.ReadFloat64(),
		}
	case jsoniter.StringValue:
		x.Kind = &structpb.Value_StringValue{
			StringValue: iter.ReadString(),
		}
	case jsoniter.ObjectValue:
		v := &structpb.Struct{}
		iter.ReadVal(v)
		x.Kind = &structpb.Value_StructValue{
			StructValue: v,
		}
	case jsoniter.ArrayValue:
		v := &structpb.ListValue{}
		iter.ReadVal(v)
		x.Kind = &structpb.Value_ListValue{
			ListValue: v,
		}
	default:
		return errors.New("not number or string or object")
	}
	return nil
}
