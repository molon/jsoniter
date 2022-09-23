package protoext

import (
	"encoding/base64"
	"io"
	"math"
	"reflect"
	"strings"
	"unsafe"

	jsoniter "github.com/json-iterator/go"
	"github.com/modern-go/reflect2"
)

func decorateEncoderForScalar(typ reflect2.Type, enc jsoniter.ValEncoder) jsoniter.ValEncoder {
	var bitSize int
	switch typ.Kind() {
	case reflect.Float32:
		bitSize = 32
	case reflect.Float64:
		bitSize = 64
	}

	if bitSize <= 0 {
		return enc
	}

	return &funcEncoder{
		fun: func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
			var n float64
			if bitSize == 32 {
				n = float64(*((*float32)(ptr)))
			} else {
				n = *((*float64)(ptr))
			}
			switch {
			case math.IsNaN(n):
				stream.WriteString("NaN")
			case math.IsInf(n, +1):
				stream.WriteString("Infinity")
			case math.IsInf(n, -1):
				stream.WriteString("-Infinity")
			default:
				enc.Encode(ptr, stream)
			}
		},
		isEmptyFunc: func(ptr unsafe.Pointer) bool {
			return enc.IsEmpty(ptr)
		},
	}
}

func decorateDecoderForScalar(typ reflect2.Type, dec jsoniter.ValDecoder) jsoniter.ValDecoder {
	// []byte
	if typ.Kind() == reflect.Slice && typ.(reflect2.SliceType).Elem().Kind() == reflect.Uint8 {
		return &funcDecoder{
			fun: func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
				if iter.WhatIsNext() == jsoniter.StringValue {
					s := iter.ReadString()
					// copy from protobuf-go
					enc := base64.StdEncoding
					if strings.ContainsAny(s, "-_") {
						enc = base64.URLEncoding
					}
					if len(s)%4 != 0 {
						enc = enc.WithPadding(base64.NoPadding)
					}

					dst, err := enc.DecodeString(s)
					if err != nil {
						iter.ReportError("decode base64", err.Error())
					} else {
						typ.UnsafeSet(ptr, unsafe.Pointer(&dst))
					}
					return
				}
				dec.Decode(ptr, iter)
			},
		}
	}

	// TODO: bool/null fuzzy??

	var bitSize int
	switch typ.Kind() {
	case reflect.Float32:
		bitSize = 32
	case reflect.Float64:
		bitSize = 64
	case reflect.Bool,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64:
		// fuzzy decode
		if !typ.Implements(protoEnumType) {
			return &stringModeNumberDecoder{dec}
		}
	case reflect.String:
		return &funcDecoder{
			fun: func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
				valueType := iter.WhatIsNext()
				switch valueType {
				case jsoniter.NumberValue:
					*((*string)(ptr)) = string(iter.ReadNumber())
				case jsoniter.NilValue:
					iter.Skip()
					*((*string)(ptr)) = ""
				default:
					dec.Decode(ptr, iter)
				}
			},
		}
	}

	if bitSize <= 0 {
		return dec
	}

	return &funcDecoder{
		fun: func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
			if iter.WhatIsNext() == jsoniter.StringValue {
				str := iter.ReadString()
				switch str {
				case "NaN":
					if bitSize == 32 {
						*((*float32)(ptr)) = float32(math.NaN())
					} else {
						*((*float64)(ptr)) = math.NaN()
					}
				case "Infinity":
					if bitSize == 32 {
						*((*float32)(ptr)) = float32(math.Inf(+1))
					} else {
						*((*float64)(ptr)) = math.Inf(+1)
					}
				case "-Infinity":
					if bitSize == 32 {
						*((*float32)(ptr)) = float32(math.Inf(-1))
					} else {
						*((*float64)(ptr)) = math.Inf(-1)
					}
				default:
					// fuzzy decode
					subIter := iter.Pool().BorrowIterator([]byte(str))
					defer iter.API().ReturnIterator(subIter)
					dec.Decode(ptr, subIter)
					if subIter.Error != nil && subIter.Error != io.EOF && iter.Error == nil {
						iter.Error = subIter.Error
					}
				}
				return
			}
			dec.Decode(ptr, iter)
		},
	}
}
