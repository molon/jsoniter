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
	ptrType reflect2.Type
}

func (enc *protojsonEncoder) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
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

func (enc *protojsonEncoder) IsEmpty(ptr unsafe.Pointer) bool {
	return *((*unsafe.Pointer)(ptr)) == nil
}

type protojsonDecoder struct {
	ptrType  reflect2.Type
	elemType reflect2.Type
}

func (dec *protojsonDecoder) Decode(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
	if iter.ReadNil() {
		*((*unsafe.Pointer)(ptr)) = nil
	} else {
		bytes := iter.SkipAndReturnBytes()
		if iter.Error != nil && iter.Error != io.EOF {
			return
		}

		if *((*unsafe.Pointer)(ptr)) == nil {
			elem := dec.elemType.UnsafeNew()
			err := protojson.Unmarshal(bytes, dec.elemType.PackEFace(elem).(proto.Message))
			if err != nil {
				iter.ReportError("protojson.Unmarshal", fmt.Sprintf(
					"errorr calling protojson.Unmarshal for type %s: %s",
					dec.ptrType, err,
				))
			}
			*((*unsafe.Pointer)(ptr)) = elem
		} else {
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
