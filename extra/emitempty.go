package extra

import (
	"unsafe"

	jsoniter "github.com/json-iterator/go"
	"github.com/modern-go/reflect2"
)

type EmitEmptyEncoder struct {
	jsoniter.ValEncoder
}

func (enc *EmitEmptyEncoder) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	enc.ValEncoder.Encode(ptr, stream)
}

func (enc *EmitEmptyEncoder) IsEmpty(ptr unsafe.Pointer) bool {
	return false
}

type EmitEmptyWithTypeExtension struct {
	jsoniter.DummyExtension
	Filter func(typ reflect2.Type) bool
}

func (e *EmitEmptyWithTypeExtension) DecorateEncoder(typ reflect2.Type, encoder jsoniter.ValEncoder) jsoniter.ValEncoder {
	if e.Filter == nil || e.Filter(typ) {
		return &EmitEmptyEncoder{ValEncoder: encoder}
	}
	return encoder
}

type EmitEmptyWithBindingExtension struct {
	jsoniter.DummyExtension
	Filter func(binding *jsoniter.Binding) bool
}

func (e *EmitEmptyWithBindingExtension) UpdateStructDescriptor(desc *jsoniter.StructDescriptor) {
	for _, binding := range desc.Fields {
		if binding.Encoder != nil {
			if e.Filter == nil || e.Filter(binding) {
				binding.Encoder = &EmitEmptyEncoder{ValEncoder: binding.Encoder}
			}
		}
	}
}
