package protoext

import (
	"fmt"
	"io"
	"reflect"
	"unsafe"

	jsoniter "github.com/json-iterator/go"
	"github.com/modern-go/reflect2"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
)

var (
	Any_message_fullname protoreflect.FullName = "google.protobuf.Any"
)

type wktAnyEncoder struct {
	ext *ProtoExtension
}

func (c *wktAnyEncoder) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	m := ((*anypb.Any)(ptr))
	resolver := c.ext.GetResolver()

	// Resolve the type in order to unmarshal value field.
	emt, err := resolver.FindMessageByURL(m.GetTypeUrl())
	if err != nil {
		stream.Error = fmt.Errorf("%s: unable to resolve %q: %v", Any_message_fullname, m.GetTypeUrl(), err)
		return
	}

	em := emt.New().Interface()
	err = proto.UnmarshalOptions{
		AllowPartial: true, // never check required fields inside an Any
		Resolver:     resolver,
	}.Unmarshal(m.GetValue(), em)
	if err != nil {
		stream.Error = fmt.Errorf("%s: unable to unmarshal %q: %v", Any_message_fullname, m.GetTypeUrl(), err)
		return
	}

	// If type of value has custom JSON encoding, marshal out a field "value"
	// with corresponding custom JSON encoding of the embedded message as a
	// field.
	typ := reflect2.TypeOf(em)
	if typ.Kind() == reflect.Ptr {
		typ = typ.(reflect2.PtrType).Elem()
	}
	if IsWellKnownType(typ) {
		stream.WriteObjectStart()
		stream.WriteObjectField("@type")
		stream.WriteString(m.GetTypeUrl())
		stream.WriteMore()
		stream.WriteObjectField("value")
		stream.WriteVal(em)
		stream.WriteObjectEnd()
		return
	}

	// Else, marshal out the embedded message's fields in this Any object.
	subStream := stream.API().BorrowStream(nil)
	subStream.Attachment = stream.Attachment
	defer stream.API().ReturnStream(subStream)
	subStream.WriteVal(em)
	if subStream.Error != nil && subStream.Error != io.EOF {
		stream.Error = fmt.Errorf("%s: unable to marshal %q: %v", Any_message_fullname, m.GetTypeUrl(), subStream.Error)
		return
	}

	subIter := stream.API().BorrowIterator(subStream.Buffer())
	defer stream.API().ReturnIterator(subIter)

	stream.WriteObjectStart()
	stream.WriteObjectField("@type")
	stream.WriteString(m.GetTypeUrl())
	subIter.ReadObjectCB(func(iter *jsoniter.Iterator, field string) bool {
		stream.WriteMore()
		stream.WriteObjectField(field)
		stream.Write(iter.SkipAndReturnBytes())
		return true
	})
	stream.WriteObjectEnd()
}

func (c *wktAnyEncoder) IsEmpty(ptr unsafe.Pointer) bool {
	return false // this is for elem type , so does not need this
}

type wktAnyDecoder struct {
	ext *ProtoExtension
}

func (c *wktAnyDecoder) Decode(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
	m := ((*anypb.Any)(ptr))

	var typeUrl string
	var valueBytes []byte
	numField := 0

	subStream := iter.API().BorrowStream(nil)
	defer iter.API().ReturnStream(subStream)
	subStream.WriteObjectStart()
	iter.ReadMapCB(func(iter *jsoniter.Iterator, field string) bool {
		if field == "@type" {
			typeUrl = iter.ReadString()
			return true
		}
		value := iter.SkipAndReturnBytes()
		if field == "value" {
			valueBytes = value
		}
		subStream.WriteObjectField(field)
		subStream.Write(value)
		numField++
		return true
	})
	subStream.WriteObjectEnd()

	if typeUrl == "" {
		if numField > 0 {
			iter.ReportError("protobuf", fmt.Sprintf(`%s: missing "@type" field`, Any_message_fullname))
			return
		}
		// empty any object
		m.TypeUrl = typeUrl
		m.Value = nil
		return
	}

	resolver := c.ext.GetResolver()
	emt, err := resolver.FindMessageByURL(typeUrl)
	if err != nil {
		iter.ReportError("protobuf", fmt.Sprintf("%s: unable to resolve %q: %v", Any_message_fullname, typeUrl, err))
		return
	}
	em := emt.New().Interface()

	typ := reflect2.TypeOf(em)
	if typ.Kind() == reflect.Ptr {
		typ = typ.(reflect2.PtrType).Elem()
	}
	var subIter *jsoniter.Iterator
	if IsWellKnownType(typ) {
		subIter = iter.API().BorrowIterator(valueBytes)
	} else {
		subIter = iter.API().BorrowIterator(subStream.Buffer())
	}
	defer iter.API().ReturnIterator(subIter)
	subIter.ReadVal(em)
	if subIter.Error != nil && subIter.Error != io.EOF {
		iter.ReportError("protobuf", fmt.Sprintf("%s: unable to unmarshal %q: %v", Any_message_fullname, typeUrl, subIter.Error))
		return
	}

	b, err := proto.MarshalOptions{
		AllowPartial:  true, // No need to check required fields inside an Any.
		Deterministic: true,
	}.Marshal(em)
	if err != nil {
		iter.ReportError("protobuf", fmt.Sprintf("error in marshaling Any.value field: %v", err))
		return
	}

	m.TypeUrl = typeUrl
	m.Value = b
}

var wktAnyCodec = &ProtoCodec{
	EncoderCreator: func(e *ProtoExtension, typ reflect2.Type) jsoniter.ValEncoder {
		return WrapElemEncoder(typ, &wktAnyEncoder{ext: e}, nil)
	},
	DecoderCreator: func(e *ProtoExtension, typ reflect2.Type) jsoniter.ValDecoder {
		return WrapElemDecoder(typ, &wktAnyDecoder{ext: e}, nil)
	},
}
