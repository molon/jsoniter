// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package protojson_test

import (
	"bytes"
	"encoding/json"
	"math"
	"testing"

	"github.com/google/go-cmp/cmp"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	jsoniter "github.com/json-iterator/go"
	"github.com/json-iterator/go/protoext"
	pb3 "github.com/json-iterator/go/protoext/internal/protojson/textpb3"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestMarshal(t *testing.T) {
	tests := []struct {
		desc    string
		mo      protojson.MarshalOptions
		input   proto.Message
		want    string
		wantErr bool // TODO: Verify error message substring.
		skip    bool
	}{{
		desc:  "proto3 scalars not set",
		input: &pb3.Scalars{},
		want:  "{}",
	}, {
		desc:  "proto3 optional not set",
		input: &pb3.Proto3Optional{},
		want:  "{}",
	}, {
		desc: "proto3 optional set to zero values",
		input: &pb3.Proto3Optional{
			OptBool:    proto.Bool(false),
			OptInt32:   proto.Int32(0),
			OptInt64:   proto.Int64(0),
			OptUint32:  proto.Uint32(0),
			OptUint64:  proto.Uint64(0),
			OptFloat:   proto.Float32(0),
			OptDouble:  proto.Float64(0),
			OptString:  proto.String(""),
			OptBytes:   []byte{},
			OptEnum:    pb3.Enum_ZERO.Enum(),
			OptMessage: &pb3.Nested{},
		},
		want: `{
  "optBool": false,
  "optInt32": 0,
  "optInt64": "0",
  "optUint32": 0,
  "optUint64": "0",
  "optFloat": 0,
  "optDouble": 0,
  "optString": "",
  "optBytes": "",
  "optEnum": "ZERO",
  "optMessage": {}
}`,
	}, {
		desc: "string",
		input: &pb3.Scalars{
			SString: "谷歌",
		},
		want: `{
  "sString": "谷歌"
}`,
	}, {
		desc: "string with invalid UTF8",
		input: &pb3.Scalars{
			SString: "abc\xff",
		},
		wantErr: true,
	}, {
		desc: "float nan",
		input: &pb3.Scalars{
			SFloat: float32(math.NaN()),
		},
		want: `{
  "sFloat": "NaN"
}`,
	}, {
		desc: "float positive infinity",
		input: &pb3.Scalars{
			SFloat: float32(math.Inf(1)),
		},
		want: `{
  "sFloat": "Infinity"
}`,
	}, {
		desc: "float negative infinity",
		input: &pb3.Scalars{
			SFloat: float32(math.Inf(-1)),
		},
		want: `{
  "sFloat": "-Infinity"
}`,
	}, {
		desc: "double nan",
		input: &pb3.Scalars{
			SDouble: math.NaN(),
		},
		want: `{
  "sDouble": "NaN"
}`,
	}, {
		desc: "double positive infinity",
		input: &pb3.Scalars{
			SDouble: math.Inf(1),
		},
		want: `{
  "sDouble": "Infinity"
}`,
	}, {
		desc: "double negative infinity",
		input: &pb3.Scalars{
			SDouble: math.Inf(-1),
		},
		want: `{
  "sDouble": "-Infinity"
}`,
	}, {
		desc:  "proto3 enum not set",
		input: &pb3.Enums{},
		want:  "{}",
	}, {
		desc: "proto3 enum set to zero value",
		input: &pb3.Enums{
			SEnum:       pb3.Enum_ZERO,
			SNestedEnum: pb3.Enums_CERO,
		},
		want: "{}",
	}, {
		desc: "proto3 enum",
		input: &pb3.Enums{
			SEnum:       pb3.Enum_ONE,
			SNestedEnum: pb3.Enums_UNO,
		},
		want: `{
  "sEnum": "ONE",
  "sNestedEnum": "UNO"
}`,
	}, {
		desc: "proto3 enum set to numeric values",
		input: &pb3.Enums{
			SEnum:       2,
			SNestedEnum: 2,
		},
		want: `{
  "sEnum": "TWO",
  "sNestedEnum": "DOS"
}`,
	}, {
		desc: "proto3 enum set to unnamed numeric values",
		input: &pb3.Enums{
			SEnum:       -47,
			SNestedEnum: 47,
		},
		want: `{
  "sEnum": -47,
  "sNestedEnum": 47
}`,
	}, {
		desc:  "proto3 nested message not set",
		input: &pb3.Nests{},
		want:  "{}",
	}, {
		desc: "proto3 nested message set to empty",
		input: &pb3.Nests{
			SNested: &pb3.Nested{},
		},
		want: `{
  "sNested": {}
}`,
	}, {
		desc: "proto3 nested message",
		input: &pb3.Nests{
			SNested: &pb3.Nested{
				SString: "nested message",
				SNested: &pb3.Nested{
					SString: "another nested message",
				},
			},
		},
		want: `{
  "sNested": {
    "sString": "nested message",
    "sNested": {
      "sString": "another nested message"
    }
  }
}`,
	}, {
		desc:  "oneof not set",
		input: &pb3.Oneofs{},
		want:  "{}",
	}, {
		desc: "oneof set to empty string",
		input: &pb3.Oneofs{
			Union: &pb3.Oneofs_OneofString{},
		},
		want: `{
  "oneofString": ""
}`,
	}, {
		desc: "oneof set to string",
		input: &pb3.Oneofs{
			Union: &pb3.Oneofs_OneofString{
				OneofString: "hello",
			},
		},
		want: `{
  "oneofString": "hello"
}`,
	}, {
		desc: "oneof set to enum",
		input: &pb3.Oneofs{
			Union: &pb3.Oneofs_OneofEnum{
				OneofEnum: pb3.Enum_ZERO,
			},
		},
		want: `{
  "oneofEnum": "ZERO"
}`,
	}, {
		desc: "oneof set to empty message",
		input: &pb3.Oneofs{
			Union: &pb3.Oneofs_OneofNested{
				OneofNested: &pb3.Nested{},
			},
		},
		want: `{
  "oneofNested": {}
}`,
	}, {
		desc: "oneof set to message",
		input: &pb3.Oneofs{
			Union: &pb3.Oneofs_OneofNested{
				OneofNested: &pb3.Nested{
					SString: "nested message",
				},
			},
		},
		want: `{
  "oneofNested": {
    "sString": "nested message"
  }
}`,
	}, {
		desc:  "map fields not set",
		input: &pb3.Maps{},
		want:  "{}",
	}, {
		desc: "map fields set to empty",
		input: &pb3.Maps{
			Int32ToStr:   map[int32]string{},
			BoolToUint32: map[bool]uint32{},
			Uint64ToEnum: map[uint64]pb3.Enum{},
			StrToNested:  map[string]*pb3.Nested{},
			StrToOneofs:  map[string]*pb3.Oneofs{},
		},
		want: "{}",
	}, {
		desc: "map fields 1",
		input: &pb3.Maps{
			BoolToUint32: map[bool]uint32{
				true:  42,
				false: 101,
			},
		},
		want: `{
  "boolToUint32": {
    "false": 101,
    "true": 42
  }
}`,
	}, {
		desc: "map fields 2",
		input: &pb3.Maps{
			Int32ToStr: map[int32]string{
				-101: "-101",
				0xff: "0xff",
				0:    "zero",
			},
		},
		want: `{
  "int32ToStr": {
    "-101": "-101",
    "0": "zero",
    "255": "0xff"
  }
}`,
	}, {
		desc: "map fields 3",
		input: &pb3.Maps{
			Uint64ToEnum: map[uint64]pb3.Enum{
				1:  pb3.Enum_ONE,
				2:  pb3.Enum_TWO,
				10: pb3.Enum_TEN,
				47: 47,
			},
		},
		want: `{
  "uint64ToEnum": {
    "1": "ONE",
    "2": "TWO",
    "10": "TEN",
    "47": 47
  }
}`,
	}, {
		desc: "map fields 4",
		input: &pb3.Maps{
			StrToNested: map[string]*pb3.Nested{
				"nested": &pb3.Nested{
					SString: "nested in a map",
				},
			},
		},
		want: `{
  "strToNested": {
    "nested": {
      "sString": "nested in a map"
    }
  }
}`,
	}, {
		desc: "map fields 5",
		input: &pb3.Maps{
			StrToOneofs: map[string]*pb3.Oneofs{
				"string": &pb3.Oneofs{
					Union: &pb3.Oneofs_OneofString{
						OneofString: "hello",
					},
				},
				"nested": &pb3.Oneofs{
					Union: &pb3.Oneofs_OneofNested{
						OneofNested: &pb3.Nested{
							SString: "nested oneof in map field value",
						},
					},
				},
			},
		},
		want: `{
  "strToOneofs": {
    "nested": {
      "oneofNested": {
        "sString": "nested oneof in map field value"
      }
    },
    "string": {
      "oneofString": "hello"
    }
  }
}`,
	}, {
		desc: "map field contains nil value",
		input: &pb3.Maps{
			StrToNested: map[string]*pb3.Nested{
				"nil": nil,
			},
		},
		want: `{
  "strToNested": {
    "nil": {}
  }
}`,
	}, {
		desc: "json_name",
		input: &pb3.JSONNames{
			SString: "json_name",
		},
		want: `{
  "foo_bar": "json_name"
}`,
	}, {
		desc:  "BoolValue empty",
		input: &wrapperspb.BoolValue{},
		want:  `false`,
	}, {
		desc:  "BoolValue",
		input: &wrapperspb.BoolValue{Value: true},
		want:  `true`,
	}, {
		desc:  "Int32Value empty",
		input: &wrapperspb.Int32Value{},
		want:  `0`,
	}, {
		desc:  "Int32Value",
		input: &wrapperspb.Int32Value{Value: 42},
		want:  `42`,
	}, {
		desc:  "Int64Value",
		input: &wrapperspb.Int64Value{Value: 42},
		want:  `"42"`,
	}, {
		desc:  "UInt32Value",
		input: &wrapperspb.UInt32Value{Value: 42},
		want:  `42`,
	}, {
		desc:  "UInt64Value",
		input: &wrapperspb.UInt64Value{Value: 42},
		want:  `"42"`,
	}, {
		desc:  "FloatValue",
		input: &wrapperspb.FloatValue{Value: 1.02},
		want:  `1.02`,
	}, {
		desc:  "FloatValue Infinity",
		input: &wrapperspb.FloatValue{Value: float32(math.Inf(-1))},
		want:  `"-Infinity"`,
	}, {
		desc:  "DoubleValue",
		input: &wrapperspb.DoubleValue{Value: 1.02},
		want:  `1.02`,
	}, {
		desc:  "DoubleValue NaN",
		input: &wrapperspb.DoubleValue{Value: math.NaN()},
		want:  `"NaN"`,
	}, {
		desc:  "StringValue empty",
		input: &wrapperspb.StringValue{},
		want:  `""`,
	}, {
		desc:  "StringValue",
		input: &wrapperspb.StringValue{Value: "谷歌"},
		want:  `"谷歌"`,
	}, {
		desc:    "StringValue with invalid UTF8 error",
		input:   &wrapperspb.StringValue{Value: "abc\xff"},
		wantErr: true,
	}, {
		desc:  "BytesValue",
		input: &wrapperspb.BytesValue{Value: []byte("hello")},
		want:  `"aGVsbG8="`,
	}, {
		desc:  "Empty",
		input: &emptypb.Empty{},
		want:  `{}`,
	}, {
		desc:    "Value empty",
		input:   &structpb.Value{},
		wantErr: true,
	}, {
		desc:  "Value contains NullValue",
		input: &structpb.Value{Kind: &structpb.Value_NullValue{}},
		want:  `null`,
	}, {
		desc:  "Value contains BoolValue",
		input: &structpb.Value{Kind: &structpb.Value_BoolValue{}},
		want:  `false`,
	}, {
		desc:  "Value contains NumberValue",
		input: &structpb.Value{Kind: &structpb.Value_NumberValue{1.02}},
		want:  `1.02`,
	}, {
		desc:  "Value contains StringValue",
		input: &structpb.Value{Kind: &structpb.Value_StringValue{"hello"}},
		want:  `"hello"`,
	}, {
		desc:    "Value contains StringValue with invalid UTF8",
		input:   &structpb.Value{Kind: &structpb.Value_StringValue{"\xff"}},
		wantErr: true,
	}, {
		desc: "Value contains Struct",
		input: &structpb.Value{
			Kind: &structpb.Value_StructValue{
				&structpb.Struct{
					Fields: map[string]*structpb.Value{
						"null":   {Kind: &structpb.Value_NullValue{}},
						"number": {Kind: &structpb.Value_NumberValue{}},
						"string": {Kind: &structpb.Value_StringValue{}},
						"struct": {Kind: &structpb.Value_StructValue{}},
						"list":   {Kind: &structpb.Value_ListValue{}},
						"bool":   {Kind: &structpb.Value_BoolValue{}},
					},
				},
			},
		},
		want: `{
  "bool": false,
  "list": [],
  "null": null,
  "number": 0,
  "string": "",
  "struct": {}
}`,
	}, {
		desc: "Value contains ListValue",
		input: &structpb.Value{
			Kind: &structpb.Value_ListValue{
				&structpb.ListValue{
					Values: []*structpb.Value{
						{Kind: &structpb.Value_BoolValue{}},
						{Kind: &structpb.Value_NullValue{}},
						{Kind: &structpb.Value_NumberValue{}},
						{Kind: &structpb.Value_StringValue{}},
						{Kind: &structpb.Value_StructValue{}},
						{Kind: &structpb.Value_ListValue{}},
					},
				},
			},
		},
		want: `[
  false,
  null,
  0,
  "",
  {},
  []
]`,
	}, {
		desc:    "Value with NaN",
		input:   structpb.NewNumberValue(math.NaN()),
		wantErr: true,
	}, {
		desc:    "Value with -Inf",
		input:   structpb.NewNumberValue(math.Inf(-1)),
		wantErr: true,
	}, {
		desc:    "Value with +Inf",
		input:   structpb.NewNumberValue(math.Inf(+1)),
		wantErr: true,
	}, {
		desc:  "Struct with nil map",
		input: &structpb.Struct{},
		want:  `{}`,
	}, {
		desc: "Struct with empty map",
		input: &structpb.Struct{
			Fields: map[string]*structpb.Value{},
		},
		want: `{}`,
	}, {
		desc: "Struct",
		input: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"bool":   {Kind: &structpb.Value_BoolValue{true}},
				"null":   {Kind: &structpb.Value_NullValue{}},
				"number": {Kind: &structpb.Value_NumberValue{3.1415}},
				"string": {Kind: &structpb.Value_StringValue{"hello"}},
				"struct": {
					Kind: &structpb.Value_StructValue{
						&structpb.Struct{
							Fields: map[string]*structpb.Value{
								"string": {Kind: &structpb.Value_StringValue{"world"}},
							},
						},
					},
				},
				"list": {
					Kind: &structpb.Value_ListValue{
						&structpb.ListValue{
							Values: []*structpb.Value{
								{Kind: &structpb.Value_BoolValue{}},
								{Kind: &structpb.Value_NullValue{}},
								{Kind: &structpb.Value_NumberValue{}},
							},
						},
					},
				},
			},
		},
		want: `{
  "bool": true,
  "list": [
    false,
    null,
    0
  ],
  "null": null,
  "number": 3.1415,
  "string": "hello",
  "struct": {
    "string": "world"
  }
}`,
	}, {
		desc: "Struct message with invalid UTF8 string",
		input: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"string": {Kind: &structpb.Value_StringValue{"\xff"}},
			},
		},
		wantErr: true,
	}, {
		desc:  "ListValue with nil values",
		input: &structpb.ListValue{},
		want:  `[]`,
	}, {
		desc: "ListValue with empty values",
		input: &structpb.ListValue{
			Values: []*structpb.Value{},
		},
		want: `[]`,
	}, {
		desc: "ListValue",
		input: &structpb.ListValue{
			Values: []*structpb.Value{
				{Kind: &structpb.Value_BoolValue{true}},
				{Kind: &structpb.Value_NullValue{}},
				{Kind: &structpb.Value_NumberValue{3.1415}},
				{Kind: &structpb.Value_StringValue{"hello"}},
				{
					Kind: &structpb.Value_ListValue{
						&structpb.ListValue{
							Values: []*structpb.Value{
								{Kind: &structpb.Value_BoolValue{}},
								{Kind: &structpb.Value_NullValue{}},
								{Kind: &structpb.Value_NumberValue{}},
							},
						},
					},
				},
				{
					Kind: &structpb.Value_StructValue{
						&structpb.Struct{
							Fields: map[string]*structpb.Value{
								"string": {Kind: &structpb.Value_StringValue{"world"}},
							},
						},
					},
				},
			},
		},
		want: `[
  true,
  null,
  3.1415,
  "hello",
  [
    false,
    null,
    0
  ],
  {
    "string": "world"
  }
]`,
	}, {
		desc: "ListValue with invalid UTF8 string",
		input: &structpb.ListValue{
			Values: []*structpb.Value{
				{Kind: &structpb.Value_StringValue{"\xff"}},
			},
		},
		wantErr: true,
	}, {
		desc:  "Duration empty",
		input: &durationpb.Duration{},
		want:  `"0s"`,
	}, {
		desc:  "Duration with secs",
		input: &durationpb.Duration{Seconds: 3},
		want:  `"3s"`,
	}, {
		desc:  "Duration with -secs",
		input: &durationpb.Duration{Seconds: -3},
		want:  `"-3s"`,
	}, {
		desc:  "Duration with nanos",
		input: &durationpb.Duration{Nanos: 1e6},
		want:  `"0.001s"`,
	}, {
		desc:  "Duration with -nanos",
		input: &durationpb.Duration{Nanos: -1e6},
		want:  `"-0.001s"`,
	}, {
		desc:  "Duration with large secs",
		input: &durationpb.Duration{Seconds: 1e10, Nanos: 1},
		want:  `"10000000000.000000001s"`,
	}, {
		desc:  "Duration with 6-digit nanos",
		input: &durationpb.Duration{Nanos: 1e4},
		want:  `"0.000010s"`,
	}, {
		desc:  "Duration with 3-digit nanos",
		input: &durationpb.Duration{Nanos: 1e6},
		want:  `"0.001s"`,
	}, {
		desc:  "Duration with -secs -nanos",
		input: &durationpb.Duration{Seconds: -123, Nanos: -450},
		want:  `"-123.000000450s"`,
	}, {
		desc:  "Duration max value",
		input: &durationpb.Duration{Seconds: 315576000000, Nanos: 999999999},
		want:  `"315576000000.999999999s"`,
	}, {
		desc:  "Duration min value",
		input: &durationpb.Duration{Seconds: -315576000000, Nanos: -999999999},
		want:  `"-315576000000.999999999s"`,
	}, {
		desc:    "Duration with +secs -nanos",
		input:   &durationpb.Duration{Seconds: 1, Nanos: -1},
		wantErr: true,
	}, {
		desc:    "Duration with -secs +nanos",
		input:   &durationpb.Duration{Seconds: -1, Nanos: 1},
		wantErr: true,
	}, {
		desc:    "Duration with +secs out of range",
		input:   &durationpb.Duration{Seconds: 315576000001},
		wantErr: true,
	}, {
		desc:    "Duration with -secs out of range",
		input:   &durationpb.Duration{Seconds: -315576000001},
		wantErr: true,
	}, {
		desc:    "Duration with +nanos out of range",
		input:   &durationpb.Duration{Seconds: 0, Nanos: 1e9},
		wantErr: true,
	}, {
		desc:    "Duration with -nanos out of range",
		input:   &durationpb.Duration{Seconds: 0, Nanos: -1e9},
		wantErr: true,
	}, {
		desc:  "Timestamp zero",
		input: &timestamppb.Timestamp{},
		want:  `"1970-01-01T00:00:00Z"`,
	}, {
		desc:  "Timestamp",
		input: &timestamppb.Timestamp{Seconds: 1553036601},
		want:  `"2019-03-19T23:03:21Z"`,
	}, {
		desc:  "Timestamp with nanos",
		input: &timestamppb.Timestamp{Seconds: 1553036601, Nanos: 1},
		want:  `"2019-03-19T23:03:21.000000001Z"`,
	}, {
		desc:  "Timestamp with 6-digit nanos",
		input: &timestamppb.Timestamp{Nanos: 1e3},
		want:  `"1970-01-01T00:00:00.000001Z"`,
	}, {
		desc:  "Timestamp with 3-digit nanos",
		input: &timestamppb.Timestamp{Nanos: 1e7},
		want:  `"1970-01-01T00:00:00.010Z"`,
	}, {
		desc:  "Timestamp max value",
		input: &timestamppb.Timestamp{Seconds: 253402300799, Nanos: 999999999},
		want:  `"9999-12-31T23:59:59.999999999Z"`,
	}, {
		desc:  "Timestamp min value",
		input: &timestamppb.Timestamp{Seconds: -62135596800},
		want:  `"0001-01-01T00:00:00Z"`,
	}, {
		desc:    "Timestamp with +secs out of range",
		input:   &timestamppb.Timestamp{Seconds: 253402300800},
		wantErr: true,
	}, {
		desc:    "Timestamp with -secs out of range",
		input:   &timestamppb.Timestamp{Seconds: -62135596801},
		wantErr: true,
	}, {
		desc:    "Timestamp with -nanos",
		input:   &timestamppb.Timestamp{Nanos: -1},
		wantErr: true,
	}, {
		desc:    "Timestamp with +nanos out of range",
		input:   &timestamppb.Timestamp{Nanos: 1e9},
		wantErr: true,
	}, {
		desc:  "FieldMask empty",
		input: &fieldmaskpb.FieldMask{},
		want:  `""`,
	}, {
		desc: "FieldMask",
		input: &fieldmaskpb.FieldMask{
			Paths: []string{
				"foo",
				"foo_bar",
				"foo.bar_qux",
				"_foo",
			},
		},
		want: `"foo,fooBar,foo.barQux,Foo"`,
	}, {
		desc: "FieldMask empty string path",
		input: &fieldmaskpb.FieldMask{
			Paths: []string{""},
		},
		wantErr: true,
	}, {
		desc: "FieldMask path contains spaces only",
		input: &fieldmaskpb.FieldMask{
			Paths: []string{"  "},
		},
		wantErr: true,
	}, {
		desc: "FieldMask irreversible error 1",
		input: &fieldmaskpb.FieldMask{
			Paths: []string{"foo_"},
		},
		wantErr: true,
	}, {
		desc: "FieldMask irreversible error 2",
		input: &fieldmaskpb.FieldMask{
			Paths: []string{"foo__bar"},
		},
		wantErr: true,
	}, {
		desc: "FieldMask invalid char",
		input: &fieldmaskpb.FieldMask{
			Paths: []string{"foo@bar"},
		},
		wantErr: true,
	}, {
		desc:  "Any empty",
		input: &anypb.Any{},
		want:  `{}`,
	}, {
		desc: "Any with EmitUnpopulated",
		mo: protojson.MarshalOptions{
			EmitUnpopulated: true,
		},
		input: func() proto.Message {
			return &anypb.Any{
				TypeUrl: string(new(pb3.Scalars).ProtoReflect().Descriptor().FullName()),
			}
		}(),
		want: `{
  "@type": "pb3.Scalars",
  "sBool": false,
  "sInt32": 0,
  "sInt64": "0",
  "sUint32": 0,
  "sUint64": "0",
  "sSint32": 0,
  "sSint64": "0",
  "sFixed32": 0,
  "sFixed64": "0",
  "sSfixed32": 0,
  "sSfixed64": "0",
  "sFloat": 0,
  "sDouble": 0,
  "sBytes": "",
  "sString": ""
}`,
	}, {
		desc: "Any with BoolValue",
		input: func() proto.Message {
			m := &wrapperspb.BoolValue{Value: true}
			b, err := proto.MarshalOptions{Deterministic: true}.Marshal(m)
			if err != nil {
				t.Fatalf("error in binary marshaling message for Any.value: %v", err)
			}
			return &anypb.Any{
				TypeUrl: "type.googleapis.com/google.protobuf.BoolValue",
				Value:   b,
			}
		}(),
		want: `{
  "@type": "type.googleapis.com/google.protobuf.BoolValue",
  "value": true
}`,
	}, {
		desc: "Any with Empty",
		input: func() proto.Message {
			m := &emptypb.Empty{}
			b, err := proto.MarshalOptions{Deterministic: true}.Marshal(m)
			if err != nil {
				t.Fatalf("error in binary marshaling message for Any.value: %v", err)
			}
			return &anypb.Any{
				TypeUrl: "type.googleapis.com/google.protobuf.Empty",
				Value:   b,
			}
		}(),
		want: `{
  "@type": "type.googleapis.com/google.protobuf.Empty",
  "value": {}
}`,
	}, {
		desc: "Any with StringValue containing invalid UTF8",
		input: func() proto.Message {
			m := &wrapperspb.StringValue{Value: "abcd"}
			b, err := proto.MarshalOptions{Deterministic: true}.Marshal(m)
			if err != nil {
				t.Fatalf("error in binary marshaling message for Any.value: %v", err)
			}
			return &anypb.Any{
				TypeUrl: "google.protobuf.StringValue",
				Value:   bytes.Replace(b, []byte("abcd"), []byte("abc\xff"), -1),
			}
		}(),
		wantErr: true,
	}, {
		desc: "Any with Int64Value",
		input: func() proto.Message {
			m := &wrapperspb.Int64Value{Value: 42}
			b, err := proto.MarshalOptions{Deterministic: true}.Marshal(m)
			if err != nil {
				t.Fatalf("error in binary marshaling message for Any.value: %v", err)
			}
			return &anypb.Any{
				TypeUrl: "google.protobuf.Int64Value",
				Value:   b,
			}
		}(),
		want: `{
  "@type": "google.protobuf.Int64Value",
  "value": "42"
}`,
	}, {
		desc: "Any with Duration",
		input: func() proto.Message {
			m := &durationpb.Duration{}
			b, err := proto.MarshalOptions{Deterministic: true}.Marshal(m)
			if err != nil {
				t.Fatalf("error in binary marshaling message for Any.value: %v", err)
			}
			return &anypb.Any{
				TypeUrl: "type.googleapis.com/google.protobuf.Duration",
				Value:   b,
			}
		}(),
		want: `{
  "@type": "type.googleapis.com/google.protobuf.Duration",
  "value": "0s"
}`,
	}, {
		desc: "Any with empty Value",
		input: func() proto.Message {
			m := &structpb.Value{}
			b, err := proto.Marshal(m)
			if err != nil {
				t.Fatalf("error in binary marshaling message for Any.value: %v", err)
			}
			return &anypb.Any{
				TypeUrl: "type.googleapis.com/google.protobuf.Value",
				Value:   b,
			}
		}(),
		wantErr: true,
	}, {
		desc: "Any with Value of StringValue",
		input: func() proto.Message {
			m := &structpb.Value{Kind: &structpb.Value_StringValue{"abcd"}}
			b, err := proto.MarshalOptions{Deterministic: true}.Marshal(m)
			if err != nil {
				t.Fatalf("error in binary marshaling message for Any.value: %v", err)
			}
			return &anypb.Any{
				TypeUrl: "type.googleapis.com/google.protobuf.Value",
				Value:   bytes.Replace(b, []byte("abcd"), []byte("abc\xff"), -1),
			}
		}(),
		wantErr: true,
	}, {
		desc: "Any with Value of NullValue",
		input: func() proto.Message {
			m := &structpb.Value{Kind: &structpb.Value_NullValue{}}
			b, err := proto.MarshalOptions{Deterministic: true}.Marshal(m)
			if err != nil {
				t.Fatalf("error in binary marshaling message for Any.value: %v", err)
			}
			return &anypb.Any{
				TypeUrl: "type.googleapis.com/google.protobuf.Value",
				Value:   b,
			}
		}(),
		want: `{
  "@type": "type.googleapis.com/google.protobuf.Value",
  "value": null
}`,
	}, {
		desc: "Any with Struct",
		input: func() proto.Message {
			m := &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"bool":   {Kind: &structpb.Value_BoolValue{true}},
					"null":   {Kind: &structpb.Value_NullValue{}},
					"string": {Kind: &structpb.Value_StringValue{"hello"}},
					"struct": {
						Kind: &structpb.Value_StructValue{
							&structpb.Struct{
								Fields: map[string]*structpb.Value{
									"string": {Kind: &structpb.Value_StringValue{"world"}},
								},
							},
						},
					},
				},
			}
			b, err := proto.MarshalOptions{Deterministic: true}.Marshal(m)
			if err != nil {
				t.Fatalf("error in binary marshaling message for Any.value: %v", err)
			}
			return &anypb.Any{
				TypeUrl: "google.protobuf.Struct",
				Value:   b,
			}
		}(),
		want: `{
  "@type": "google.protobuf.Struct",
  "value": {
    "bool": true,
    "null": null,
    "string": "hello",
    "struct": {
      "string": "world"
    }
  }
}`,
	}, {
		desc: "Any with missing type_url",
		input: func() proto.Message {
			m := &wrapperspb.BoolValue{Value: true}
			b, err := proto.MarshalOptions{Deterministic: true}.Marshal(m)
			if err != nil {
				t.Fatalf("error in binary marshaling message for Any.value: %v", err)
			}
			return &anypb.Any{
				Value: b,
			}
		}(),
		wantErr: true,
	}, {
		desc:  "EmitUnpopulated: proto3 scalars",
		mo:    protojson.MarshalOptions{EmitUnpopulated: true},
		input: &pb3.Scalars{},
		want: `{
  "sBool": false,
  "sInt32": 0,
  "sInt64": "0",
  "sUint32": 0,
  "sUint64": "0",
  "sSint32": 0,
  "sSint64": "0",
  "sFixed32": 0,
  "sFixed64": "0",
  "sSfixed32": 0,
  "sSfixed64": "0",
  "sFloat": 0,
  "sDouble": 0,
  "sBytes": "",
  "sString": ""
}`,
	}, {
		desc:  "EmitUnpopulated: proto3 enum",
		mo:    protojson.MarshalOptions{EmitUnpopulated: true},
		input: &pb3.Enums{},
		want: `{
  "sEnum": "ZERO",
  "sNestedEnum": "CERO"
}`,
	}, {
		desc:  "EmitUnpopulated: proto3 message field",
		mo:    protojson.MarshalOptions{EmitUnpopulated: true},
		input: &pb3.Nests{},
		want: `{
  "sNested": null
}`,
	}, {
		desc: "EmitUnpopulated: proto3 empty message field",
		mo:   protojson.MarshalOptions{EmitUnpopulated: true},
		input: &pb3.Nests{
			SNested: &pb3.Nested{},
		},
		want: `{
  "sNested": {
    "sString": "",
    "sNested": null
  }
}`,
	}, {
		desc:  "EmitUnpopulated: map fields",
		mo:    protojson.MarshalOptions{EmitUnpopulated: true},
		input: &pb3.Maps{},
		want: `{
  "int32ToStr": {},
  "boolToUint32": {},
  "uint64ToEnum": {},
  "strToNested": {},
  "strToOneofs": {}
}`,
	}, {
		desc: "EmitUnpopulated: map containing empty message",
		mo:   protojson.MarshalOptions{EmitUnpopulated: true},
		input: &pb3.Maps{
			StrToNested: map[string]*pb3.Nested{
				"nested": &pb3.Nested{},
			},
			StrToOneofs: map[string]*pb3.Oneofs{
				"nested": &pb3.Oneofs{},
			},
		},
		want: `{
  "int32ToStr": {},
  "boolToUint32": {},
  "uint64ToEnum": {},
  "strToNested": {
    "nested": {
      "sString": "",
      "sNested": null
    }
  },
  "strToOneofs": {
    "nested": {}
  }
}`,
	}, {
		desc:  "EmitUnpopulated: oneof fields",
		mo:    protojson.MarshalOptions{EmitUnpopulated: true},
		input: &pb3.Oneofs{},
		want:  `{}`,
	}, {
		desc: "UseEnumNumbers in map field",
		mo:   protojson.MarshalOptions{UseEnumNumbers: true},
		input: &pb3.Maps{
			Uint64ToEnum: map[uint64]pb3.Enum{
				1:  pb3.Enum_ONE,
				2:  pb3.Enum_TWO,
				10: pb3.Enum_TEN,
				47: 47,
			},
		},
		want: `{
  "uint64ToEnum": {
    "1": 1,
    "2": 2,
    "10": 10,
    "47": 47
  }
}`,
	}}

	for _, tt := range tests {
		tt := tt
		if tt.skip {
			continue
		}
		t.Run(tt.desc, func(t *testing.T) {
			cfg := jsoniter.Config{SortMapKeys: true}.Froze()
			cfg.RegisterExtension(&protoext.ProtoExtension{
				UseEnumNumbers:  tt.mo.UseEnumNumbers,
				UseProtoNames:   tt.mo.UseProtoNames,
				EmitUnpopulated: tt.mo.EmitUnpopulated,
				Resolver:        tt.mo.Resolver,
			})
			b, err := cfg.MarshalIndent(tt.input, "", "  ")
			if err != nil && !tt.wantErr {
				t.Errorf("MarshalIndent() returned error: %v\n", err)
			}
			if err == nil && tt.wantErr {
				t.Errorf("MarshalIndent() got nil error, want error\n")
			}
			got := string(b)
			if compact(t, got) != compact(t, tt.want) {
				t.Errorf("MarshalIndent()\n<want>\n%v\n<got>\n%v\n", tt.want, got)
				if diff := cmp.Diff(tt.want, got); diff != "" {
					t.Errorf("MarshalIndent() diff -want +got\n%v\n", diff)
				}
			}
		})
	}
}

func compact(t *testing.T, want string) string {
	if want == "" {
		return want
	}
	var out bytes.Buffer
	err := json.Compact(&out, []byte(want))
	if err != nil {
		t.Errorf("Compact returned error: %v\n", err)
		return want
	}
	return out.String()
}
