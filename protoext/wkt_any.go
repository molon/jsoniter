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
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/known/anypb"
)

var (
	Any_message_fullname protoreflect.FullName = "google.protobuf.Any"
	// TODO: 改为可设置？
	// TODO: 目前还未考虑 Unmarshal 里对 FindExtensionByName 的逻辑相关处理
	Resolver = protoregistry.GlobalTypes
)

var wktAnyCodec = NewElemTypeCodec(
	func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
		m := ((*anypb.Any)(ptr))

		// Resolve the type in order to unmarshal value field.
		emt, err := Resolver.FindMessageByURL(m.GetTypeUrl())
		if err != nil {
			stream.Error = fmt.Errorf("%s: unable to resolve %q: %v", Any_message_fullname, m.GetTypeUrl(), err)
			return
		}

		em := emt.New().Interface()
		err = proto.UnmarshalOptions{
			AllowPartial: true, // never check required fields inside an Any
			Resolver:     Resolver,
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

		// // Else, marshal out the embedded message's fields in this Any object.
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
	},
	nil, // TODO: 暂且借用 protojson 的 Unmarshal 方法，但是得注意 ProtoExtension 里得提供 Resolver 的设置
)
