package protoext

import (
	"reflect"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/json-iterator/go/extra"
	"github.com/modern-go/reflect2"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type ProtoExtension struct {
	jsoniter.DummyExtension

	EmitUnpopulated bool
	UseEnumNumbers  bool
	UseProtoNames   bool
	// TODO: 目前还未考虑 Unmarshal 里对 FindExtensionByName 的逻辑相关处理
	Resolver interface {
		protoregistry.MessageTypeResolver
		protoregistry.ExtensionTypeResolver
	}

	Encode64BitAsInteger bool
}

func (e *ProtoExtension) CreateEncoder(typ reflect2.Type) jsoniter.ValEncoder {
	if enc := e.createProtoEncoder(typ); enc != nil {
		return enc
	}
	if enc := e.createProtoEnumEncoder(typ); enc != nil {
		return enc
	}
	return nil
}

func (e *ProtoExtension) CreateDecoder(typ reflect2.Type) jsoniter.ValDecoder {
	if dec := e.createProtoDecoder(typ); dec != nil {
		return dec
	}
	if dec := e.createProtoEnumDecoder(typ); dec != nil {
		return dec
	}
	return nil
}

// Handle 64BitInteger as string
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
	if enc := decorateEncoderForNilCollection(typ, encoder); enc != nil {
		return enc
	}

	if e.Encode64BitAsInteger {
		return encoder
	}
	// https://developers.google.com/protocol-buffers/docs/proto3 int64, fixed64, uint64 should be string
	// https://github.com/protocolbuffers/protobuf-go/blob/e62d8edb7570c986a51e541c161a0c93bbaf9253/encoding/protojson/encode.go#L274-L277
	// https://github.com/protocolbuffers/protobuf-go/pull/14
	// https://github.com/golang/protobuf/issues/1414
	if typ.Kind() == reflect.Int64 || typ.Kind() == reflect.Uint64 ||
		wellKnown64BitIntegerTypes[typ] {
		return &stringModeNumberEncoder{encoder}
	}
	return encoder
}

func (e *ProtoExtension) DecorateDecoder(typ reflect2.Type, decoder jsoniter.ValDecoder) jsoniter.ValDecoder {
	// fuzzy decode, so we dont check Encode64BitAsInteger
	if typ.Kind() == reflect.Int64 || typ.Kind() == reflect.Uint64 ||
		wellKnown64BitIntegerTypes[typ] {
		return &stringModeNumberDecoder{decoder}
	}
	return decoder
}

// Handle EmitUnpopulated and UseProtoNames
func (e *ProtoExtension) UpdateStructDescriptor(desc *jsoniter.StructDescriptor) {
	for _, binding := range desc.Fields {
		if len(binding.FromNames) <= 0 { // simple check should exported
			continue
		}

		// Because oneof wrapper does not satisfy proto.Message, we can only check with tag instead of protoreflect here
		tag, hastag := binding.Field.Tag().Lookup("protobuf")
		if !hastag {
			continue
		}

		if e.EmitUnpopulated {
			binding.Encoder = &extra.EmitEmptyEncoder{binding.Encoder}
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
