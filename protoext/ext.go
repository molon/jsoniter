package protoext

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"unsafe"

	jsoniter "github.com/json-iterator/go"
	"github.com/modern-go/reflect2"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type protoEncoder struct {
	ptrType reflect2.Type
}

func (enc *protoEncoder) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	if *((*unsafe.Pointer)(ptr)) == nil {
		stream.WriteNil()
		return
	}
	data, err := protojson.Marshal(enc.ptrType.UnsafeIndirect(ptr).(proto.Message))
	if err != nil {
		stream.Error = fmt.Errorf("error calling protojson.Marshal for type %s: %w", enc.ptrType, err)
		return
	}
	_, stream.Error = stream.Write(data)
}

func (enc *protoEncoder) IsEmpty(ptr unsafe.Pointer) bool {
	return *((*unsafe.Pointer)(ptr)) == nil
}

type protoDecoder struct {
	ptrType  reflect2.Type
	elemType reflect2.Type
}

func (dec *protoDecoder) Decode(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
	if iter.ReadNil() {
		*((*unsafe.Pointer)(ptr)) = nil
	} else {
		bytes := iter.SkipAndReturnBytes()
		if iter.Error != nil && iter.Error != io.EOF {
			return
		}

		if *((*unsafe.Pointer)(ptr)) == nil {
			//pointer to null, we have to allocate memory to hold the value
			// newPtr := dec.elemType.UnsafeNew()
			// err := protojson.Unmarshal(bytes, dec.ptrType.UnsafeIndirect(unsafe.Pointer(&newPtr)).(proto.Message))
			m := dec.ptrType.UnsafeIndirect(ptr).(proto.Message).ProtoReflect().New().Interface()
			err := protojson.Unmarshal(bytes, m)
			if err != nil {
				iter.ReportError("protojson.Unmarshal", fmt.Sprintf(
					"errorr calling protojson.Unmarshal for type %s: %s",
					dec.ptrType, err,
				))
			}
			// *((*unsafe.Pointer)(ptr)) = newPtr
			*((*unsafe.Pointer)(ptr)) = reflect2.PtrOf(m)
		} else {
			//reuse existing instance
			err := protojson.Unmarshal(bytes, dec.ptrType.UnsafeIndirect(ptr).(proto.Message))
			if err != nil {
				iter.ReportError("protojson.Unmarshal", fmt.Sprintf(
					"error calling protojson.Unmarshal for type %s: %s",
					dec.ptrType, err,
				))
			}
		}
	}
}

// TODO: 会需要一个 64位数字 应该以字符串表示的可选要求，但是要注意默认的 protojson 就一定为如此，所以我们需要针对此情况做一定的反向处理
// TODO: genid.Value_message_fullname 相关可能也需要特殊处理
type ProtoExtension struct {
	jsoniter.DummyExtension

	UseEnumNumbers bool
	UseProtoNames  bool
}

func (e *ProtoExtension) UpdateStructDescriptor(desc *jsoniter.StructDescriptor) {
	for _, binding := range desc.Fields {
		if len(binding.FromNames) <= 0 { // simple check should exported
			continue
		}

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

func (e *ProtoExtension) CreateEncoder(typ reflect2.Type) jsoniter.ValEncoder {
	if _, ok := wellKnownPtrTypes[typ]; ok {
		return &protoEncoder{
			ptrType: typ,
		}
	}

	if !e.UseEnumNumbers {
		if typ.Implements(protoEnumType) {
			// TODO: 如果直接是 interface 呢？
			if typ.Kind() == reflect.Ptr {
				return &jsoniter.OptionalEncoder{
					ValueEncoder: &protoEnumNameEncoder{
						elemType: typ.(reflect2.PtrType).Elem(),
					},
				}
			}
			return &protoEnumNameEncoder{
				elemType: typ,
			}
		}
	}

	return nil
}

func (e *ProtoExtension) CreateDecoder(typ reflect2.Type) jsoniter.ValDecoder {
	if _, ok := wellKnownPtrTypes[typ]; ok {
		return &protoDecoder{
			ptrType:  typ,
			elemType: typ.(reflect2.PtrType).Elem(),
		}
	}

	// we want fuzzy decode, so does not need to check e.UseEnumNumbers
	if typ.Implements(protoEnumType) {
		if typ.Kind() == reflect.Ptr {
			elem := typ.(reflect2.PtrType).Elem()
			return &jsoniter.OptionalDecoder{
				ValueType: elem,
				ValueDecoder: &protoEnumDecoder{
					elemType: elem,
				},
			}
		}
		return &protoEnumDecoder{
			elemType: typ,
		}
	}
	return nil
}
