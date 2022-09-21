package protoext

import (
	"fmt"
	"reflect"
	"unsafe"

	jsoniter "github.com/json-iterator/go"
	"github.com/modern-go/reflect2"
)

type stringModeNumberEncoder struct {
	elemEncoder jsoniter.ValEncoder
}

func (encoder *stringModeNumberEncoder) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	stream.Write([]byte(`"`))
	encoder.elemEncoder.Encode(ptr, stream)
	stream.Write([]byte(`"`))
}

func (encoder *stringModeNumberEncoder) IsEmpty(ptr unsafe.Pointer) bool {
	return encoder.elemEncoder.IsEmpty(ptr)
}

type stringModeNumberDecoder struct {
	elemDecoder jsoniter.ValDecoder
}

func (decoder *stringModeNumberDecoder) Decode(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
	if iter.WhatIsNext() == jsoniter.StringValue {
		iter.NextToken()
		decoder.elemDecoder.Decode(ptr, iter)
		iter.NextToken()
		return
	}
	decoder.elemDecoder.Decode(ptr, iter)
}

type funcEncoder struct {
	fun         jsoniter.EncoderFunc
	isEmptyFunc func(ptr unsafe.Pointer) bool
}

func (encoder *funcEncoder) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	encoder.fun(ptr, stream)
}

func (encoder *funcEncoder) IsEmpty(ptr unsafe.Pointer) bool {
	if encoder.isEmptyFunc == nil {
		return false
	}
	return encoder.isEmptyFunc(ptr)
}

type funcDecoder struct {
	fun jsoniter.DecoderFunc
}

func (decoder *funcDecoder) Decode(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
	decoder.fun(ptr, iter)
}

type dynamicEncoder struct {
	valType reflect2.Type
}

func (encoder *dynamicEncoder) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	obj := encoder.valType.UnsafeIndirect(ptr)
	stream.WriteVal(obj)
}

func (encoder *dynamicEncoder) IsEmpty(ptr unsafe.Pointer) bool {
	return encoder.valType.UnsafeIndirect(ptr) == nil
}

func WrapElemEncoder(typ reflect2.Type, enc jsoniter.ValEncoder) jsoniter.ValEncoder {
	if typ.Kind() == reflect.Ptr {
		if typ.(reflect2.PtrType).Elem().Kind() == reflect.Struct {
			return &jsoniter.OptionalEncoder{
				ValueEncoder: enc,
			}
		}
		panic(fmt.Sprintf("WrapElemEncoder does not support type %v", typ))
	}
	return enc
}

func WrapElemDecoder(typ reflect2.Type, dec jsoniter.ValDecoder) jsoniter.ValDecoder {
	if typ.Kind() == reflect.Ptr {
		if typ.(reflect2.PtrType).Elem().Kind() == reflect.Struct {
			return &jsoniter.OptionalDecoder{
				ValueDecoder: dec,
			}
		}
		panic(fmt.Sprintf("WrapElemDecoder does not support type %v", typ))
	}
	return dec
}
