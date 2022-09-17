package protoext

import (
	"reflect"

	jsoniter "github.com/json-iterator/go"
	"github.com/modern-go/reflect2"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// TODO: genid.Value_message_fullname 相关可能也需要特殊处理
type ProtoExtension struct {
	jsoniter.DummyExtension

	UseEnumNumbers       bool
	UseProtoNames        bool
	Encode64BitAsInteger bool
}

func (e *ProtoExtension) CreateEncoder(typ reflect2.Type) jsoniter.ValEncoder {
	if codec, ok := WellKnownTypeCodecs[typ]; ok {
		if codec != nil && codec.Encoder != nil {
			return codec.Encoder
		}
		// If not specified, use protojson for processing
		return &protojsonEncoder{
			valueType: typ,
		}
	}

	if !e.UseEnumNumbers {
		if typ.Implements(protoEnumType) && typ.Kind() != reflect.Ptr {
			// TODO: 如果直接是 interface 呢？
			return &protoEnumNameEncoder{
				valueType: typ,
			}
		}
	}

	return nil
}

func (e *ProtoExtension) CreateDecoder(typ reflect2.Type) jsoniter.ValDecoder {
	if codec, ok := WellKnownTypeCodecs[typ]; ok {
		if codec != nil && codec.Decoder != nil {
			return codec.Decoder
		}
		// If not specified, use protojson for processing
		return &protojsonDecoder{
			valueType: typ,
		}
	}

	// we want fuzzy decode, so does not need to check e.UseEnumNumbers
	if typ.Implements(protoEnumType) {
		if typ.Kind() != reflect.Ptr {
			return &protoEnumDecoder{
				valueType: typ,
			}
		}

		if decoder := createDecoderOfNullValueEnumPtr(typ); decoder != nil {
			return decoder
		}
	}

	return nil
}

var wellKnown64BitIntegerTypes = map[reflect2.Type]bool{
	reflect2.TypeOfPtr((*wrapperspb.Int64Value)(nil)).Elem():  true,
	reflect2.TypeOfPtr((*wrapperspb.UInt64Value)(nil)).Elem(): true,
}

func (e *ProtoExtension) CreateMapKeyEncoder(typ reflect2.Type) jsoniter.ValEncoder {
	if e.Encode64BitAsInteger {
		return nil
	}
	if typ.Kind() == reflect.Int64 || typ.Kind() == reflect.Uint64 {
		// avoid quote it repeatedly
		return &dynamicEncoder{typ}
	}
	return nil
}

func (e *ProtoExtension) DecorateEncoder(typ reflect2.Type, encoder jsoniter.ValEncoder) jsoniter.ValEncoder {
	if e.Encode64BitAsInteger {
		return encoder
	}
	// https://developers.google.com/protocol-buffers/docs/proto3 int64, fixed64, uint64 should be string
	// https://github.com/protocolbuffers/protobuf-go/blob/e62d8edb7570c986a51e541c161a0c93bbaf9253/encoding/protojson/encode.go#L274-L277
	// https://github.com/protocolbuffers/protobuf-go/pull/14
	// https://github.com/golang/protobuf/issues/1414
	if typ.Kind() == reflect.Int64 || typ.Kind() == reflect.Uint64 {
		return &stringModeNumberEncoder{encoder}
	}
	if wellKnown64BitIntegerTypes[typ] {
		return &stringModeNumberEncoder{encoder}
	}
	return encoder
}

func (e *ProtoExtension) DecorateDecoder(typ reflect2.Type, decoder jsoniter.ValDecoder) jsoniter.ValDecoder {
	// fuzzy decode, so we dont check Encode64BitAsInteger
	if typ.Kind() == reflect.Int64 || typ.Kind() == reflect.Uint64 {
		return &stringModeNumberDecoder{decoder}
	}
	if wellKnown64BitIntegerTypes[typ] {
		return &stringModeNumberDecoder{decoder}
	}
	return decoder
}
