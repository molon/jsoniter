package protoext

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"unsafe"

	jsoniter "github.com/json-iterator/go"
	"github.com/json-iterator/go/extra"
	"github.com/modern-go/reflect2"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var protoMessageType = reflect2.TypeOfPtr((*proto.Message)(nil)).Elem()

func (e *ProtoExtension) UpdateStructDescriptorConstructor(c *jsoniter.StructDescriptorConstructor) {
	newBindings := make([]*jsoniter.Binding, 0, len(c.Bindings))
	defer func() {
		c.Bindings = newBindings
	}()

	var pb proto.Message
	var pbReflect protoreflect.Message
	for _, binding := range c.Bindings {
		field := binding.Field

		if field.Type().Kind() == reflect.Interface {
			oneofsTag, hasOneofsTag := field.Tag().Lookup("protobuf_oneof")
			if hasOneofsTag && reflect2.PtrTo(c.Type).Implements(protoMessageType) {
				if pb == nil {
					pb = c.Type.New().(proto.Message)
					pbReflect = pb.ProtoReflect()
				}
				fieldType := field.Type()
				fieldPtr := field.UnsafeGet(reflect2.PtrOf(pb))
				od := pbReflect.Descriptor().Oneofs().ByName(protoreflect.Name(oneofsTag))
				if !od.IsSynthetic() { // ignore optional
					fds := od.Fields()
					for j := 0; j < fds.Len(); j++ {
						fd := fds.Get(j)
						value := pbReflect.NewField(fd)
						pbReflect.Set(fd, value)

						fTyp := reflect2.TypeOf(fieldType.UnsafeIndirect(fieldPtr))
						if fTyp.Kind() == reflect.Ptr {
							wrapPtrType := fTyp.(*reflect2.UnsafePtrType)
							if wrapPtrType.Elem().Kind() == reflect.Struct {
								structDescriptor := c.DescribeStructFunc(wrapPtrType.Elem())
								for _, b := range structDescriptor.Fields {
									b.Levels = append([]int{binding.Levels[0], j}, b.Levels...)
									omitempty := b.Encoder.(*jsoniter.StructFieldEncoder).OmitEmpty
									b.Encoder = &protoOneofWrapperEncoder{b.Field.Name(), wrapPtrType, b.Encoder}
									b.Encoder = &jsoniter.StructFieldEncoder{field, b.Encoder, omitempty}
									b.Decoder = &protoOneofWrapperDecoder{b.Field.Name(), field.Type(), wrapPtrType, wrapPtrType.Elem(), b.Decoder}
									b.Decoder = &jsoniter.StructFieldDecoder{field, b.Decoder}
									c.EmbeddedBindings = append(c.EmbeddedBindings, b)
								}
								continue
							}
						}
					}
					continue
				}
			}
		}

		newBindings = append(newBindings, binding)

		if len(binding.FromNames) <= 0 { // simple check should exported
			continue
		}

		// Because oneof wrapper does not satisfy proto.Message, we can only check with tag instead of protoreflect here
		tag, hastag := binding.Field.Tag().Lookup("protobuf")
		if !hastag {
			continue
		}

		var name, jsonName string
		tagParts := strings.Split(tag, ",")
		for _, part := range tagParts {
			colons := strings.SplitN(part, "=", 2)
			if len(colons) == 2 {
				switch strings.TrimSpace(colons[0]) {
				case "name":
					name = strings.TrimSpace(colons[1])
				case "json":
					jsonName = strings.TrimSpace(colons[1])
				}
				continue
			}
			if strings.TrimSpace(part) == "oneof" {
				if reflect2.PtrTo(c.Type).Implements(protoMessageType) {
					if pb == nil {
						pb = c.Type.New().(proto.Message)
						pbReflect = pb.ProtoReflect()
					}
					od := pbReflect.Descriptor().Fields().ByName(protoreflect.Name(name))
					if od != nil {
						oneof := od.ContainingOneof()
						// IsSynthetic OneOf (optional keyword)
						if oneof != nil && oneof.IsSynthetic() {
							binding.Encoder = &extra.ImmunityEmitEmptyEncoder{
								&protoOptionalEncoder{binding.Encoder},
							}
						}
					}
				}
			}
		}
		if jsonName == "" {
			jsonName = name
		}
		if name != "" {
			if e.UseProtoNames {
				binding.FromNames = []string{name}
				// fuzzy
				if jsonName != name {
					binding.FromNames = append(binding.FromNames, jsonName)
				}
				binding.ToNames = []string{name}
			} else {
				binding.FromNames = []string{jsonName}
				// fuzzy
				if name != jsonName {
					binding.FromNames = append(binding.FromNames, name)
				}
				binding.ToNames = []string{jsonName}
			}
		}
	}
}

type protoOptionalEncoder struct {
	jsoniter.ValEncoder
}

func (enc *protoOptionalEncoder) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	enc.ValEncoder.Encode(ptr, stream)
}

func (enc *protoOptionalEncoder) IsEmpty(ptr unsafe.Pointer) bool {
	return enc.ValEncoder.IsEmpty(ptr)
}

func (enc *protoOptionalEncoder) IsEmbeddedPtrNil(ptr unsafe.Pointer) bool {
	if *((*unsafe.Pointer)(ptr)) == nil {
		return true
	}
	isEmbeddedPtrNil, converted := enc.ValEncoder.(jsoniter.IsEmbeddedPtrNil)
	if !converted {
		return false
	}
	return isEmbeddedPtrNil.IsEmbeddedPtrNil(ptr)
}

type protoOneofWrapperEncoder struct {
	innerFieldName string
	valuePtrType   reflect2.Type
	valueEncoder   jsoniter.ValEncoder
}

func (encoder *protoOneofWrapperEncoder) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
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
	isEmbeddedPtrNil, converted := encoder.valueEncoder.(jsoniter.IsEmbeddedPtrNil)
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
	valueDecoder   jsoniter.ValDecoder
}

func (decoder *protoOneofWrapperDecoder) Decode(fieldPtr unsafe.Pointer, iter *jsoniter.Iterator) {
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
