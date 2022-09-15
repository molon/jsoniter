package jsoniter

import (
	"fmt"
	"io"
	"reflect"
	"unsafe"

	"github.com/modern-go/reflect2"
)

type protoOneofWrapperEncoder struct {
	innerFieldName string
	valuePtrType   reflect2.Type
	valueEncoder   ValEncoder
}

func (encoder *protoOneofWrapperEncoder) Encode(ptr unsafe.Pointer, stream *Stream) {
	if *((*unsafe.Pointer)(ptr)) == nil {
		stream.WriteNil()
		return
	}
	val := reflect2.IFaceToEFace(ptr)
	if reflect2.TypeOf(val).RType() != encoder.valuePtrType.RType() {
		stream.WriteNil()
		return
	}
	encoder.valueEncoder.Encode(reflect2.PtrOf(val), stream)
	if stream.Error != nil && stream.Error != io.EOF {
		stream.Error = fmt.Errorf("%s: %s", encoder.innerFieldName, stream.Error.Error())
	}
}

func (encoder *protoOneofWrapperEncoder) IsEmpty(ptr unsafe.Pointer) bool {
	if *((*unsafe.Pointer)(ptr)) == nil {
		return true
	}
	val := reflect2.IFaceToEFace(ptr)
	if reflect2.TypeOf(val).RType() != encoder.valuePtrType.RType() {
		return true
	}
	return encoder.valueEncoder.IsEmpty(reflect2.PtrOf(val))
}

func (encoder *protoOneofWrapperEncoder) IsEmbeddedPtrNil(ptr unsafe.Pointer) bool {
	if *((*unsafe.Pointer)(ptr)) == nil {
		return true
	}
	val := reflect2.IFaceToEFace(ptr)
	if reflect2.TypeOf(val).RType() != encoder.valuePtrType.RType() {
		return true
	}
	isEmbeddedPtrNil, converted := encoder.valueEncoder.(IsEmbeddedPtrNil)
	if !converted {
		return false
	}
	return isEmbeddedPtrNil.IsEmbeddedPtrNil(reflect2.PtrOf(val))
}

type protoOneofWrapperDecoder struct {
	innerFieldName string
	fieldType      reflect2.Type
	valuePtrType   reflect2.Type
	valueElemType  reflect2.Type
	valueDecoder   ValDecoder
}

func (decoder *protoOneofWrapperDecoder) Decode(fieldPtr unsafe.Pointer, iter *Iterator) {
	if iter.ReadNil() {
		decoder.fieldType.UnsafeSet(fieldPtr, decoder.fieldType.UnsafeNew())
		return
	}

	var elem interface{}

	// reuse it if type match
	if *((*unsafe.Pointer)(fieldPtr)) != nil {
		elem = reflect2.IFaceToEFace(fieldPtr)
		if reflect2.TypeOf(elem).RType() != decoder.valuePtrType.RType() {
			elem = nil
		}
	}
	if elem == nil {
		elem = decoder.valueElemType.New()
	}

	decoder.valueDecoder.Decode(reflect2.PtrOf(elem), iter)
	if iter.Error != nil && iter.Error != io.EOF {
		iter.Error = fmt.Errorf("%s: %s", decoder.innerFieldName, iter.Error.Error())
		return
	}

	rval := reflect.ValueOf(decoder.fieldType.PackEFace(fieldPtr))
	rval.Elem().Set(reflect.ValueOf(elem))
}
