package protoext

import (
	"unsafe"

	jsoniter "github.com/json-iterator/go"
)

type stringModeNumberEncoder struct {
	elemEncoder jsoniter.ValEncoder
}

func (encoder *stringModeNumberEncoder) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	if *((*unsafe.Pointer)(ptr)) == nil {
		stream.WriteNil()
	} else {
		stream.Write([]byte(`"`))
		encoder.elemEncoder.Encode(ptr, stream)
		stream.Write([]byte(`"`))
	}
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
