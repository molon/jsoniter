package protoext

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
	"strings"
	"unicode/utf8"
	"unsafe"

	jsoniter "github.com/json-iterator/go"
	"github.com/modern-go/reflect2"
)

func (e *ProtoExtension) updateMapEncoderConstructorForScalar(v *jsoniter.MapEncoderConstructor) {
	// handle 64bit integer key, avoid quote it repeatedly
	if e.Encode64BitAsInteger {
		return
	}
	typ := v.MapType.Key()
	switch typ.Kind() {
	case reflect.Int64, reflect.Uint64:
		v.KeyEncoder = &dynamicEncoder{v.MapType.Key()}
	}
}

func (e *ProtoExtension) decorateEncoderForScalar(typ reflect2.Type, enc jsoniter.ValEncoder) jsoniter.ValEncoder {
	var bitSize int
	switch typ.Kind() {
	case reflect.String:
		if e.PermitInvalidUTF8 {
			return enc
		}
		return &protoStringEncoder{}
	case reflect.Int64, reflect.Uint64:
		// https://developers.google.com/protocol-buffers/docs/proto3 int64, fixed64, uint64 should be string
		// https://github.com/protocolbuffers/protobuf-go/blob/e62d8edb7570c986a51e541c161a0c93bbaf9253/encoding/protojson/encode.go#L274-L277
		// https://github.com/protocolbuffers/protobuf-go/pull/14
		// https://github.com/golang/protobuf/issues/1414
		if e.Encode64BitAsInteger {
			return enc
		}
		return &stringModeNumberEncoder{enc}
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
				stream.WriteRaw(`"NaN"`)
			case math.IsInf(n, +1):
				stream.WriteRaw(`"Infinity"`)
			case math.IsInf(n, -1):
				stream.WriteRaw(`"-Infinity"`)
			default:
				enc.Encode(ptr, stream)
			}
		},
		isEmptyFunc: func(ptr unsafe.Pointer) bool {
			return enc.IsEmpty(ptr)
		},
	}
}

func (e *ProtoExtension) decorateDecoderForScalar(typ reflect2.Type, dec jsoniter.ValDecoder) jsoniter.ValDecoder {
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
					if iter.Error == nil {
						if !e.PermitInvalidUTF8 {
							if !utf8.ValidString(*((*string)(ptr))) {
								iter.Error = errInvalidUTF8
							}
						}
					}
				}
			},
		}
	case reflect.Float32:
		bitSize = 32
	case reflect.Float64:
		bitSize = 64
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
					subIter.Attachment = iter.Attachment
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

type protoStringEncoder struct{}

func (encoder *protoStringEncoder) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	str := *((*string)(ptr))
	buf, err := ParseString(str)
	if err != nil {
		stream.Error = fmt.Errorf("ProtoStringEncoder: %w", err)
		return
	}
	stream.Write(buf)
}

func (encoder *protoStringEncoder) IsEmpty(ptr unsafe.Pointer) bool {
	return *((*string)(ptr)) == ""
}

func ParseString(s string) ([]byte, error) {
	valLen := len(s)

	buf := []byte{'"'}
	// write string, the fast path, without utf8 and escape support
	i := 0
	for ; i < valLen; i++ {
		c := s[i]
		if c < utf8.RuneSelf && safeSet[c] {
			buf = append(buf, c)
		} else {
			break
		}
	}
	if i == valLen {
		buf = append(buf, '"')
		return buf, nil
	}
	return appendStringSlowPath(buf, i, s, valLen)
}

var errInvalidUTF8 = errors.New("invalid UTF-8")

func appendStringSlowPath(buf []byte, i int, s string, valLen int) ([]byte, error) {
	start := i
	// for the remaining parts, we process them char by char
	for i < valLen {
		if b := s[i]; b < utf8.RuneSelf {
			if safeSet[b] {
				i++
				continue
			}
			if start < i {
				buf = append(buf, s[start:i]...)
			}
			switch b {
			case '\\', '"':
				buf = append(buf, '\\', b)
			case '\b':
				buf = append(buf, '\\', 'b')
			case '\f':
				buf = append(buf, '\\', 'f')
			case '\n':
				buf = append(buf, '\\', 'n')
			case '\r':
				buf = append(buf, '\\', 'r')
			case '\t':
				buf = append(buf, '\\', 't')
			default:
				buf = append(buf, `\u00`...)
				buf = append(buf, hex[b>>4], hex[b&0xF])
			}
			i++
			start = i
			continue
		}
		c, size := utf8.DecodeRuneInString(s[i:])
		if c == utf8.RuneError && size == 1 {
			return buf, errInvalidUTF8
		}
		i += size
	}
	if start < len(s) {
		buf = append(buf, s[start:]...)
	}
	buf = append(buf, '"')
	return buf, nil
}

var safeSet = [utf8.RuneSelf]bool{
	' ':      true,
	'!':      true,
	'"':      false,
	'#':      true,
	'$':      true,
	'%':      true,
	'&':      true,
	'\'':     true,
	'(':      true,
	')':      true,
	'*':      true,
	'+':      true,
	',':      true,
	'-':      true,
	'.':      true,
	'/':      true,
	'0':      true,
	'1':      true,
	'2':      true,
	'3':      true,
	'4':      true,
	'5':      true,
	'6':      true,
	'7':      true,
	'8':      true,
	'9':      true,
	':':      true,
	';':      true,
	'<':      true,
	'=':      true,
	'>':      true,
	'?':      true,
	'@':      true,
	'A':      true,
	'B':      true,
	'C':      true,
	'D':      true,
	'E':      true,
	'F':      true,
	'G':      true,
	'H':      true,
	'I':      true,
	'J':      true,
	'K':      true,
	'L':      true,
	'M':      true,
	'N':      true,
	'O':      true,
	'P':      true,
	'Q':      true,
	'R':      true,
	'S':      true,
	'T':      true,
	'U':      true,
	'V':      true,
	'W':      true,
	'X':      true,
	'Y':      true,
	'Z':      true,
	'[':      true,
	'\\':     false,
	']':      true,
	'^':      true,
	'_':      true,
	'`':      true,
	'a':      true,
	'b':      true,
	'c':      true,
	'd':      true,
	'e':      true,
	'f':      true,
	'g':      true,
	'h':      true,
	'i':      true,
	'j':      true,
	'k':      true,
	'l':      true,
	'm':      true,
	'n':      true,
	'o':      true,
	'p':      true,
	'q':      true,
	'r':      true,
	's':      true,
	't':      true,
	'u':      true,
	'v':      true,
	'w':      true,
	'x':      true,
	'y':      true,
	'z':      true,
	'{':      true,
	'|':      true,
	'}':      true,
	'~':      true,
	'\u007f': true,
}

var hex = "0123456789abcdef"
