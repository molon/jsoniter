package protoext

import (
	"fmt"
	"strconv"
	"sync"
	"unsafe"

	jsoniter "github.com/json-iterator/go"
	"github.com/modern-go/reflect2"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoimpl"
)

var protoEnumType = reflect2.TypeOfPtr((*protoreflect.Enum)(nil)).Elem()

type protoEnumNameEncoder struct {
	valueType reflect2.Type
	once      sync.Once
	enumDesc  protoreflect.EnumDescriptor
}

// Full name for google.protobuf.NullValue.
const (
	NullValue_enum_fullname = "google.protobuf.NullValue"
)

func (enc *protoEnumNameEncoder) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	x := enc.valueType.UnsafeIndirect(ptr).(protoreflect.Enum)
	enc.once.Do(func() {
		enc.enumDesc = x.Descriptor()
	})
	if enc.enumDesc.FullName() == NullValue_enum_fullname {
		stream.WriteNil()
		return
	}
	stream.WriteString(protoimpl.X.EnumStringOf(enc.enumDesc, x.Number()))
}

func (enc *protoEnumNameEncoder) IsEmpty(ptr unsafe.Pointer) bool {
	return *((*protoreflect.EnumNumber)(ptr)) == 0
}

type protoEnumDecoder struct {
	valueType    reflect2.Type
	once         sync.Once
	enumValDescs protoreflect.EnumValueDescriptors
}

func (dec *protoEnumDecoder) Decode(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
	valueType := iter.WhatIsNext()
	switch valueType {
	case jsoniter.NumberValue:
		num := iter.ReadInt32()
		*((*protoreflect.EnumNumber)(ptr)) = protoreflect.EnumNumber(num)
	case jsoniter.StringValue:
		name := iter.ReadString()
		dec.once.Do(func() {
			x := dec.valueType.UnsafeIndirect(ptr).(protoreflect.Enum)
			dec.enumValDescs = x.Descriptor().Values()
		})
		ev := dec.enumValDescs.ByName(protoreflect.Name(name))
		if ev != nil {
			*((*protoreflect.EnumNumber)(ptr)) = ev.Number()
		} else {
			// is "num"?
			num, err := strconv.ParseInt(name, 10, 32)
			if err == nil {
				*((*protoreflect.EnumNumber)(ptr)) = protoreflect.EnumNumber(num)
			} else {
				iter.ReportError("DecodeProtoEnum", fmt.Sprintf(
					"error decode from string for type %s",
					dec.valueType,
				))
			}
		}
	case jsoniter.NilValue:
		iter.Skip()
		*((*protoreflect.EnumNumber)(ptr)) = 0
	default:
		iter.ReportError("DecodeProtoEnum", fmt.Sprintf(
			"error decode for type %s",
			dec.valueType,
		))
	}
}
