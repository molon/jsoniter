package protoext

import (
	"io"
	"unsafe"

	jsoniter "github.com/json-iterator/go"
	"google.golang.org/protobuf/types/known/structpb"
)

var wktStructCodec = (&ProtoCodec{}).
	SetElemEncodeFunc(func(e *ProtoExtension, ptr unsafe.Pointer, stream *jsoniter.Stream) {
		x := ((*structpb.Struct)(ptr))
		if x.Fields == nil {
			stream.WriteEmptyObject()
			return
		}
		stream.WriteVal(x.Fields)
	}).
	SetElemDecodeFunc(func(e *ProtoExtension, ptr unsafe.Pointer, iter *jsoniter.Iterator) {
		x := ((*structpb.Struct)(ptr))

		fields := map[string]*structpb.Value{}
		iter.ReadMapCB(func(iter *jsoniter.Iterator, field string) bool {
			v := &structpb.Value{}
			iter.ReadVal(v)
			if iter.Error != nil && iter.Error != io.EOF {
				return false
			}
			fields[field] = v
			return true
		})

		x.Fields = fields
	})
