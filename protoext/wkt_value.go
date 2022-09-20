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
	Value_message_fullname protoreflect.FullName = "google.protobuf.Value"
	// Value_Kind_oneof_name  protoreflect.Name     = "kind"
	// Value_NumberValue_field_number   protoreflect.FieldNumber = 2
	Value_NumberValue_field_fullname protoreflect.FullName = "google.protobuf.Value.number_value"
)

var wktValueCodec = NewElemTypeCodec(
	func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
		x := ((*structpb.Value)(ptr))
		switch v := x.GetKind().(type) {
		case *structpb.Value_NullValue:
			if v != nil {
				stream.WriteNil()
				return
			}
		case *structpb.Value_NumberValue:
			if v != nil {
				if math.IsNaN(v.NumberValue) || math.IsInf(v.NumberValue, 0) {
					stream.Error = fmt.Errorf("%s: invalid %v value", Value_NumberValue_field_fullname, v)
					return
				}
				stream.WriteFloat64(v.NumberValue)
				return
			}
		case *structpb.Value_StringValue:
			if v != nil {
				stream.WriteString(v.StringValue)
				return
			}
		case *structpb.Value_BoolValue:
			if v != nil {
				stream.WriteBool(v.BoolValue)
				return
			}
		case *structpb.Value_StructValue:
			if v != nil {
				stream.WriteVal(v.StructValue)
				return
			}
		case *structpb.Value_ListValue:
			if v != nil {
				stream.WriteVal(v.ListValue)
				return
			}
		}
		// TODO: 如果是在一个数组里不应该出现nil啊？咋办
		stream.WriteNil()
	},
	func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
		err := unmarshalValue(((*structpb.Value)(ptr)), iter)
		if err != nil {
			iter.ReportError("protobuf", fmt.Sprintf("%s: %v", Value_message_fullname, err))
		}
	},
)

func unmarshalValue(x *structpb.Value, iter *jsoniter.Iterator) error {
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
