package extra

import (
	"unsafe"

	jsoniter "github.com/json-iterator/go"
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

type EmitEmptyExtension struct {
	jsoniter.DummyExtension
	Filter func(binding *jsoniter.Binding) bool
}

func (e *EmitEmptyExtension) UpdateStructDescriptor(desc *jsoniter.StructDescriptor) {
	for _, binding := range desc.Fields {
		if binding.Encoder != nil {
			if e.Filter == nil || e.Filter(binding) {
				binding.Encoder = &EmitEmptyEncoder{ValEncoder: binding.Encoder}
			}
		}
	}
}
