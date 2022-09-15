package protoext

import jsoniter "github.com/json-iterator/go"

// Used by EmitEmptyWithBindingExtension
func ProtoEmitUnpopulated(binding *jsoniter.Binding) bool {
	_, ok := binding.Field.Tag().Lookup("protobuf")
	return ok
}
