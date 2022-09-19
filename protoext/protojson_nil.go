package protoext

import (
	"reflect"
	"unsafe"

	jsoniter "github.com/json-iterator/go"
	"github.com/modern-go/reflect2"
)

// https://github.com/golang/protobuf/issues/1487

func decorateEncoderOfNilCollection(typ reflect2.Type, encoder jsoniter.ValEncoder) jsoniter.ValEncoder {
	// - marshal nil []byte to ""
	// - marshal nil slice to []
	// - marshal nil map to {}
	switch typ.Kind() {
	case reflect.Slice, reflect.Map:
		return &funcEncoder{
			fun: func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
				if *((*unsafe.Pointer)(ptr)) == nil {
					if typ.Kind() == reflect.Slice {
						if typ.(reflect2.SliceType).Elem().Kind() == reflect.Uint8 {
							stream.Write([]byte{'"', '"'})
							return
						}
						stream.WriteEmptyArray()
					} else if typ.Kind() == reflect.Map {
						stream.WriteEmptyObject()
					}
					return
				}
				encoder.Encode(ptr, stream)
			},
			isEmptyFunc: func(ptr unsafe.Pointer) bool {
				return encoder.IsEmpty(ptr)
			},
		}
	}
	return nil
}

// - marshal []type{a,null,c} to [a,zero,c]
// - marshal map[string]type to {"a":"valueA",b:zero,c:"valueC"}
func noNullElemEncoderForCollection(elemType reflect2.Type, encoder jsoniter.ValEncoder) jsoniter.ValEncoder {
	return &funcEncoder{
		fun: func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
			if *(*unsafe.Pointer)(ptr) == nil {
				if elemType.Kind() != reflect.Ptr {
					ptr = elemType.UnsafeNew()
				} else {
					elemType := elemType.(reflect2.PtrType).Elem()
					if elemType.Kind() != reflect.Ptr { // just only check once
						ptr = elemType.UnsafeNew()
					} else {
						encoder.Encode(ptr, stream)
						return
					}
				}
				ptr = unsafe.Pointer(&ptr)
			}
			encoder.Encode(ptr, stream)
		},
		isEmptyFunc: func(ptr unsafe.Pointer) bool {
			return encoder.IsEmpty(ptr)
		},
	}
}

func (e *ProtoExtension) UpdateMapEncoderConstructor(v *jsoniter.MapEncoderConstructor) {
	v.ElemEncoder = noNullElemEncoderForCollection(v.MapType.Elem(), v.ElemEncoder)
}

func (e *ProtoExtension) UpdateSliceEncoderConstructor(v *jsoniter.SliceEncoderConstructor) {
	v.ElemEncoder = noNullElemEncoderForCollection(v.SliceType.Elem(), v.ElemEncoder)
}

func (e *ProtoExtension) UpdateArrayEncoderConstructor(v *jsoniter.ArrayEncoderConstructor) {
	v.ElemEncoder = noNullElemEncoderForCollection(v.ArrayType.Elem(), v.ElemEncoder)
}
