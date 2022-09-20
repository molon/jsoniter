package protoext

import (
	"fmt"
	"io"
	"reflect"
	"unsafe"

	jsoniter "github.com/json-iterator/go"
	"github.com/modern-go/reflect2"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type protojsonEncoder struct {
	valueType   reflect2.Type
	marshalOpts protojson.MarshalOptions
}

func (enc *protojsonEncoder) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	// TODO: indent?
	data, err := enc.marshalOpts.Marshal(enc.valueType.PackEFace(ptr).(proto.Message))
	if err != nil {
		stream.Error = fmt.Errorf("error calling protojson.Marshal for type %s: %w", reflect2.PtrTo(enc.valueType), err)
		return
	}
	_, stream.Error = stream.Write(data)
}

func (enc *protojsonEncoder) IsEmpty(ptr unsafe.Pointer) bool {
	// protojson will not omit zero value, only omit zero pointer, we stay compatible,
	return false
}

type protojsonDecoder struct {
	valueType reflect2.Type
}

func (dec *protojsonDecoder) Decode(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
	bytes := iter.SkipAndReturnBytes()
	if iter.Error != nil && iter.Error != io.EOF {
		return
	}

	err := protojson.Unmarshal(bytes, dec.valueType.PackEFace(ptr).(proto.Message))
	if err != nil {
		iter.ReportError("protobuf", fmt.Sprintf(
			"error calling protojson.Unmarshal for type %s: %s",
			reflect2.PtrTo(dec.valueType), err,
		))
	}
}

func (e *ProtoExtension) createProtoMessageEncoder(typ reflect2.Type) (xret jsoniter.ValEncoder) {
	if v, ok := ProtoMessageCodecs[typ]; ok {
		defer func() {
			if xret != nil && typ.Kind() == reflect.Ptr {
				xret = &jsoniter.OptionalEncoder{
					ValueEncoder: xret,
				}
			}
		}()
		var codec *Codec
		if v != nil {
			switch vv := v.(type) {
			case CodecCreator:
				codec = vv(e)
			case *Codec:
				codec = vv
			default:
				panic(fmt.Sprintf("invalid ProtoMessageCodecs value: %v:%#v", typ, v))
			}
		}
		if codec != nil && codec.Encoder != nil {
			return codec.Encoder
		}
		// If not specified, use protojson for processing
		return &protojsonEncoder{
			valueType: typ,
			marshalOpts: protojson.MarshalOptions{
				EmitUnpopulated: e.EmitUnpopulated,
				UseEnumNumbers:  e.UseEnumNumbers,
				UseProtoNames:   e.UseProtoNames,
				Resolver:        e.Resolver,
			},
		}
	}
	return nil
}

func (e *ProtoExtension) createProtoMessageDecoder(typ reflect2.Type) (xret jsoniter.ValDecoder) {
	if v, ok := ProtoMessageCodecs[typ]; ok {
		defer func() {
			if xret != nil && typ.Kind() == reflect.Ptr {
				xret = &jsoniter.OptionalDecoder{
					ValueDecoder: xret,
				}
			}
		}()
		var codec *Codec
		if v != nil {
			switch vv := v.(type) {
			case CodecCreator:
				codec = vv(e)
			case *Codec:
				codec = vv
			default:
				panic(fmt.Sprintf("invalid ProtoMessageCodecs value: %v:%#v", typ, v))
			}
		}
		if codec != nil && codec.Decoder != nil {
			return codec.Decoder
		}
		// If not specified, use protojson for processing
		return &protojsonDecoder{
			valueType: typ,
		}
	}
	return nil
}
