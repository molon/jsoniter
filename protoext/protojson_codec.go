package protoext

import (
	"fmt"
	"io"
	"unsafe"

	jsoniter "github.com/json-iterator/go"
	"github.com/modern-go/reflect2"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type protojsonEncoder struct {
	valueType reflect2.Type
}

func (enc *protojsonEncoder) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	// TODO: 这个opts有没有必要传递？
	data, err := protojson.Marshal(enc.valueType.PackEFace(ptr).(proto.Message))
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
