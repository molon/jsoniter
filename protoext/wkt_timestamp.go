package protoext

import (
	"fmt"
	"strings"
	"time"
	"unsafe"

	jsoniter "github.com/json-iterator/go"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	maxTimestampSeconds                              = 253402300799
	minTimestampSeconds                              = -62135596800
	Timestamp_message_fullname protoreflect.FullName = "google.protobuf.Timestamp"
)

var wktTimestampCodec = NewElemTypeCodec(
	func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
		s, err := marshalTimestamp(((*timestamppb.Timestamp)(ptr)))
		if err != nil {
			stream.Error = err
			return
		}
		stream.WriteString(s)
	},
	func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
		s := iter.ReadString()
		if err := unmarshalTimestamp(s, (*timestamppb.Timestamp)(ptr)); err != nil {
			iter.ReportError("protobuf", err.Error())
			return
		}
	},
)

func marshalTimestamp(m *timestamppb.Timestamp) (string, error) {
	secs := m.Seconds
	nanos := int64(m.Nanos)
	if secs < minTimestampSeconds || secs > maxTimestampSeconds {
		return "", fmt.Errorf("%s: seconds out of range %v", Timestamp_message_fullname, secs)
	}
	if nanos < 0 || nanos > secondsInNanos {
		return "", fmt.Errorf("%s: nanos out of range %v", Timestamp_message_fullname, nanos)
	}
	// Uses RFC 3339, where generated output will be Z-normalized and uses 0, 3,
	// 6 or 9 fractional digits.
	t := time.Unix(secs, nanos).UTC()
	x := t.Format("2006-01-02T15:04:05.000000000")
	x = strings.TrimSuffix(x, "000")
	x = strings.TrimSuffix(x, "000")
	x = strings.TrimSuffix(x, ".000")
	return x + "Z", nil
}

func unmarshalTimestamp(s string, m *timestamppb.Timestamp) error {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return fmt.Errorf("invalid %v value %v: %w", Timestamp_message_fullname, s, err)
	}
	// Validate seconds.
	secs := t.Unix()
	if secs < minTimestampSeconds || secs > maxTimestampSeconds {
		return fmt.Errorf("%v value out of range: %v", Timestamp_message_fullname, s)
	}
	// Validate subseconds.
	i := strings.LastIndexByte(s, '.')  // start of subsecond field
	j := strings.LastIndexAny(s, "Z-+") // start of timezone field
	if i >= 0 && j >= i && j-i > len(".999999999") {
		return fmt.Errorf("invalid %v value %v", Timestamp_message_fullname, s)
	}

	m.Seconds = secs
	m.Nanos = int32(t.Nanosecond())
	return nil
}
