package protoext

import (
	"reflect"
	"sync"
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

type lazyValue struct {
	once    sync.Once
	ret     interface{}
	creator func() interface{}
}

func newLazyValue(creator func() interface{}) *lazyValue {
	return &lazyValue{
		creator: creator,
	}
}

func (v *lazyValue) Get() interface{} {
	v.once.Do(func() {
		v.ret = v.creator()
	})
	return v.ret
}

var lazyPtrWithZeroValueMap sync.Map

// - marshal []type{a,null,c} to [a,zero,c]
// - marshal map[string]type to {"a":"valueA",b:zero,c:"valueC"}
func noNullElemEncoderForCollection(valueType reflect2.Type, encoder jsoniter.ValEncoder) jsoniter.ValEncoder {
	return &funcEncoder{
		fun: func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
			if valueType.Kind() == reflect.Ptr {
				if *(*unsafe.Pointer)(ptr) == nil {
					v, _ := lazyPtrWithZeroValueMap.LoadOrStore(valueType, newLazyValue(func() interface{} {
						ptrType := valueType.(reflect2.PtrType)
						elemType := ptrType.Elem()
						elemPtr := elemType.UnsafeNew()

						// record first
						newPtr := ptrType.UnsafeNew()
						*(*unsafe.Pointer)(newPtr) = elemPtr

						for elemType.Kind() == reflect.Ptr {
							ptrType = elemType.(reflect2.PtrType)
							elemType = ptrType.Elem()
							newElemPtr := elemType.UnsafeNew()
							*(*unsafe.Pointer)(elemPtr) = newElemPtr
							elemPtr = newElemPtr
						}
						return newPtr
					}))
					ptr = v.(*lazyValue).Get().(unsafe.Pointer)
				}
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
