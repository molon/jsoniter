
// generated by go run gen.go; DO NOT EDIT
// github.com/protocolbuffers/protobuf-go
// commit: 784c4825545540dc41a1dc85715d3251903bc8ce
// date: 2022-05-24T20:05:50Z

// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package protojson_test

import (
	"math"
	"strings"
	"testing"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	pb3 "github.com/json-iterator/go/protoext/internal/protojson/textpb3"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	jsoniter "github.com/json-iterator/go"
	"github.com/json-iterator/go/protoext"
)

func TestUnmarshal(t *testing.T) {
	tests := []struct {
		desc         string
		umo          protojson.UnmarshalOptions
		inputMessage proto.Message
		inputText    string
		wantMessage  proto.Message
		wantErr      string // Expected error substring.
		skip         bool
	}{{
		desc:         "proto3 scalars set to zero values",
		inputMessage: &pb3.Scalars{},
		inputText: `{
  "sBool": false,
  "sInt32": 0,
  "sInt64": 0,
  "sUint32": 0,
  "sUint64": 0,
  "sSint32": 0,
  "sSint64": 0,
  "sFixed32": 0,
  "sFixed64": 0,
  "sSfixed32": 0,
  "sSfixed64": 0,
  "sFloat": 0,
  "sDouble": 0,
  "sBytes": "",
  "sString": ""
}`,
		wantMessage: &pb3.Scalars{},
	}, {
		desc:         "proto3 optional set to zero values",
		inputMessage: &pb3.Proto3Optional{},
		inputText: `{
  "optBool": false,
  "optInt32": 0,
  "optInt64": 0,
  "optUint32": 0,
  "optUint64": 0,
  "optFloat": 0,
  "optDouble": 0,
  "optString": "",
  "optBytes": "",
  "optEnum": "ZERO",
  "optMessage": {}
}`,
		wantMessage: &pb3.Proto3Optional{
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
	}, {
		desc:         "proto3 scalars set to null",
		inputMessage: &pb3.Scalars{},
		inputText: `{
  "sBool": null,
  "sInt32": null,
  "sInt64": null,
  "sUint32": null,
  "sUint64": null,
  "sSint32": null,
  "sSint64": null,
  "sFixed32": null,
  "sFixed64": null,
  "sSfixed32": null,
  "sSfixed64": null,
  "sFloat": null,
  "sDouble": null,
  "sBytes": null,
  "sString": null
}`,
		wantMessage: &pb3.Scalars{},
	}, {
		desc:         "boolean",
		inputMessage: &pb3.Scalars{},
		inputText:    `{"sBool": true}`,
		wantMessage: &pb3.Scalars{
			SBool: true,
		},
	}, {
		desc:         "not boolean",
		inputMessage: &pb3.Scalars{},
		inputText:    `{"sBool": "true"}`,
		wantErr:      `invalid value for bool type: "true"`,
	}, {
		desc:         "float and double",
		inputMessage: &pb3.Scalars{},
		inputText: `{
  "sFloat": 1.234,
  "sDouble": 5.678
}`,
		wantMessage: &pb3.Scalars{
			SFloat:  1.234,
			SDouble: 5.678,
		},
	}, {
		desc:         "float and double in string",
		inputMessage: &pb3.Scalars{},
		inputText: `{
  "sFloat": "1.234",
  "sDouble": "5.678"
}`,
		wantMessage: &pb3.Scalars{
			SFloat:  1.234,
			SDouble: 5.678,
		},
	}, {
		desc:         "float and double in E notation",
		inputMessage: &pb3.Scalars{},
		inputText: `{
  "sFloat": 12.34E-1,
  "sDouble": 5.678e4
}`,
		wantMessage: &pb3.Scalars{
			SFloat:  1.234,
			SDouble: 56780,
		},
	}, {
		desc:         "float and double in string E notation",
		inputMessage: &pb3.Scalars{},
		inputText: `{
  "sFloat": "12.34E-1",
  "sDouble": "5.678e4"
}`,
		wantMessage: &pb3.Scalars{
			SFloat:  1.234,
			SDouble: 56780,
		},
	}, {
		desc:         "float exceeds limit",
		inputMessage: &pb3.Scalars{},
		inputText:    `{"sFloat": 3.4e39}`,
		wantErr:      `invalid value for float type: 3.4e39`,
	}, {
		desc:         "float in string exceeds limit",
		inputMessage: &pb3.Scalars{},
		inputText:    `{"sFloat": "-3.4e39"}`,
		wantErr:      `invalid value for float type: "-3.4e39"`,
	}, {
		desc:         "double exceeds limit",
		inputMessage: &pb3.Scalars{},
		inputText:    `{"sDouble": -1.79e+309}`,
		wantErr:      `invalid value for double type: -1.79e+309`,
	}, {
		desc:         "double in string exceeds limit",
		inputMessage: &pb3.Scalars{},
		inputText:    `{"sDouble": "1.79e+309"}`,
		wantErr:      `invalid value for double type: "1.79e+309"`,
	}, {
		desc:         "infinites",
		inputMessage: &pb3.Scalars{},
		inputText:    `{"sFloat": "Infinity", "sDouble": "-Infinity"}`,
		wantMessage: &pb3.Scalars{
			SFloat:  float32(math.Inf(+1)),
			SDouble: math.Inf(-1),
		},
	}, {
		desc:         "float string with leading space",
		inputMessage: &pb3.Scalars{},
		inputText:    `{"sFloat": " 1.234"}`,
		wantErr:      `invalid value for float type: " 1.234"`,
	}, {
		desc:         "double string with trailing space",
		inputMessage: &pb3.Scalars{},
		inputText:    `{"sDouble": "5.678 "}`,
		wantErr:      `invalid value for double type: "5.678 "`,
	}, {
		desc:         "not float",
		inputMessage: &pb3.Scalars{},
		inputText:    `{"sFloat": true}`,
		wantErr:      `invalid value for float type: true`,
	}, {
		desc:         "not double",
		inputMessage: &pb3.Scalars{},
		inputText:    `{"sDouble": "not a number"}`,
		wantErr:      `invalid value for double type: "not a number"`,
	}, {
		desc:         "integers",
		inputMessage: &pb3.Scalars{},
		inputText: `{
  "sInt32": 1234,
  "sInt64": -1234,
  "sUint32": 1e2,
  "sUint64": 100E-2,
  "sSint32": 1.0,
  "sSint64": -1.0,
  "sFixed32": 1.234e+5,
  "sFixed64": 1200E-2,
  "sSfixed32": -1.234e05,
  "sSfixed64": -1200e-02
}`,
		wantMessage: &pb3.Scalars{
			SInt32:    1234,
			SInt64:    -1234,
			SUint32:   100,
			SUint64:   1,
			SSint32:   1,
			SSint64:   -1,
			SFixed32:  123400,
			SFixed64:  12,
			SSfixed32: -123400,
			SSfixed64: -12,
		},
	}, {
		desc:         "integers in string",
		inputMessage: &pb3.Scalars{},
		inputText: `{
  "sInt32": "1234",
  "sInt64": "-1234",
  "sUint32": "1e2",
  "sUint64": "100E-2",
  "sSint32": "1.0",
  "sSint64": "-1.0",
  "sFixed32": "1.234e+5",
  "sFixed64": "1200E-2",
  "sSfixed32": "-1.234e05",
  "sSfixed64": "-1200e-02"
}`,
		wantMessage: &pb3.Scalars{
			SInt32:    1234,
			SInt64:    -1234,
			SUint32:   100,
			SUint64:   1,
			SSint32:   1,
			SSint64:   -1,
			SFixed32:  123400,
			SFixed64:  12,
			SSfixed32: -123400,
			SSfixed64: -12,
		},
	}, {
		desc:         "integers in escaped string",
		inputMessage: &pb3.Scalars{},
		inputText:    `{"sInt32": "\u0031\u0032"}`,
		wantMessage: &pb3.Scalars{
			SInt32: 12,
		},
	}, {
		desc:         "integer string with leading space",
		inputMessage: &pb3.Scalars{},
		inputText:    `{"sInt32": " 1234"}`,
		wantErr:      `invalid value for int32 type: " 1234"`,
	}, {
		desc:         "integer string with trailing space",
		inputMessage: &pb3.Scalars{},
		inputText:    `{"sUint32": "1e2 "}`,
		wantErr:      `invalid value for uint32 type: "1e2 "`,
	}, {
		desc:         "number is not an integer",
		inputMessage: &pb3.Scalars{},
		inputText:    `{"sInt32": 1.001}`,
		wantErr:      `invalid value for int32 type: 1.001`,
	}, {
		desc:         "32-bit int exceeds limit",
		inputMessage: &pb3.Scalars{},
		inputText:    `{"sInt32": 2e10}`,
		wantErr:      `invalid value for int32 type: 2e10`,
	}, {
		desc:         "64-bit int exceeds limit",
		inputMessage: &pb3.Scalars{},
		inputText:    `{"sSfixed64": -9e19}`,
		wantErr:      `invalid value for sfixed64 type: -9e19`,
	}, {
		desc:         "not integer",
		inputMessage: &pb3.Scalars{},
		inputText:    `{"sInt32": "not a number"}`,
		wantErr:      `invalid value for int32 type: "not a number"`,
	}, {
		desc:         "not unsigned integer",
		inputMessage: &pb3.Scalars{},
		inputText:    `{"sUint32": "not a number"}`,
		wantErr:      `invalid value for uint32 type: "not a number"`,
	}, {
		desc:         "number is not an unsigned integer",
		inputMessage: &pb3.Scalars{},
		inputText:    `{"sUint32": -1}`,
		wantErr:      `invalid value for uint32 type: -1`,
	}, {
		desc:         "string with invalid UTF-8",
		inputMessage: &pb3.Scalars{},
		inputText:    "{\"sString\": \"\xff\"}",
		wantErr:      `(line 1:13): invalid UTF-8 in string`,
	}, {
		desc:         "bytes",
		inputMessage: &pb3.Scalars{},
		inputText:    `{"sBytes": "aGVsbG8gd29ybGQ"}`,
		wantMessage: &pb3.Scalars{
			SBytes: []byte("hello world"),
		},
	}, {
		desc:         "bytes padded",
		inputMessage: &pb3.Scalars{},
		inputText:    `{"sBytes": "aGVsbG8gd29ybGQ="}`,
		wantMessage: &pb3.Scalars{
			SBytes: []byte("hello world"),
		},
	}, {
		desc:         "not bytes",
		inputMessage: &pb3.Scalars{},
		inputText:    `{"sBytes": true}`,
		wantErr:      `invalid value for bytes type: true`,
	}, {
		desc:         "proto3 enum",
		inputMessage: &pb3.Enums{},
		inputText: `{
  "sEnum": "ONE",
  "sNestedEnum": "DIEZ"
}`,
		wantMessage: &pb3.Enums{
			SEnum:       pb3.Enum_ONE,
			SNestedEnum: pb3.Enums_DIEZ,
		},
	}, {
		desc:         "enum numeric value",
		inputMessage: &pb3.Enums{},
		inputText: `{
  "sEnum": 2,
  "sNestedEnum": 2
}`,
		wantMessage: &pb3.Enums{
			SEnum:       pb3.Enum_TWO,
			SNestedEnum: pb3.Enums_DOS,
		},
	}, {
		desc:         "enum unnamed numeric value",
		inputMessage: &pb3.Enums{},
		inputText: `{
  "sEnum": 101,
  "sNestedEnum": -101
}`,
		wantMessage: &pb3.Enums{
			SEnum:       101,
			SNestedEnum: -101,
		},
	}, {
		desc:         "enum set to number string",
		inputMessage: &pb3.Enums{},
		inputText: `{
  "sEnum": "1"
}`,
		wantErr: `invalid value for enum type: "1"`,
	}, {
		desc:         "enum set to invalid named",
		inputMessage: &pb3.Enums{},
		inputText: `{
  "sEnum": "UNNAMED"
}`,
		wantErr: `invalid value for enum type: "UNNAMED"`,
	}, {
		desc:         "enum set to not enum",
		inputMessage: &pb3.Enums{},
		inputText: `{
  "sEnum": true
}`,
		wantErr: `invalid value for enum type: true`,
	}, {
		desc:         "enum set to JSON null",
		inputMessage: &pb3.Enums{},
		inputText: `{
  "sEnum": null
}`,
		wantMessage: &pb3.Enums{},
	}, {
		desc:         "proto name",
		inputMessage: &pb3.JSONNames{},
		inputText: `{
  "s_string": "proto name used"
}`,
		wantMessage: &pb3.JSONNames{
			SString: "proto name used",
		},
	}, {
		desc:         "json_name",
		inputMessage: &pb3.JSONNames{},
		inputText: `{
  "foo_bar": "json_name used"
}`,
		wantMessage: &pb3.JSONNames{
			SString: "json_name used",
		},
	}, {
		desc:         "camelCase name",
		inputMessage: &pb3.JSONNames{},
		inputText: `{
  "sString": "camelcase used"
}`,
		wantErr: `unknown field "sString"`,
	}, {
		desc:         "proto name and json_name",
		inputMessage: &pb3.JSONNames{},
		inputText: `{
  "foo_bar": "json_name used",
  "s_string": "proto name used"
}`,
		wantErr: `(line 3:3): duplicate field "s_string"`,
	}, {
		desc:         "duplicate field names",
		inputMessage: &pb3.JSONNames{},
		inputText: `{
  "foo_bar": "one",
  "foo_bar": "two",
}`,
		wantErr: `(line 3:3): duplicate field "foo_bar"`,
	}, {
		desc:         "proto3 nested message not set",
		inputMessage: &pb3.Nests{},
		inputText:    "{}",
		wantMessage:  &pb3.Nests{},
	}, {
		desc:         "proto3 nested message set to null",
		inputMessage: &pb3.Nests{},
		inputText:    `{"sNested": null}`,
		wantMessage:  &pb3.Nests{},
	}, {
		desc:         "proto3 nested message set to empty",
		inputMessage: &pb3.Nests{},
		inputText:    `{"sNested": {}}`,
		wantMessage: &pb3.Nests{
			SNested: &pb3.Nested{},
		},
	}, {
		desc:         "proto3 nested message",
		inputMessage: &pb3.Nests{},
		inputText: `{
  "sNested": {
    "sString": "nested message",
    "sNested": {
      "sString": "another nested message"
    }
  }
}`,
		wantMessage: &pb3.Nests{
			SNested: &pb3.Nested{
				SString: "nested message",
				SNested: &pb3.Nested{
					SString: "another nested message",
				},
			},
		},
	}, {
		desc:         "message set to non-message",
		inputMessage: &pb3.Nests{},
		inputText:    `"not valid"`,
		wantErr:      `unexpected token "not valid"`,
	}, {
		desc:         "nested message set to non-message",
		inputMessage: &pb3.Nests{},
		inputText:    `{"sNested": true}`,
		wantErr:      `(line 1:13): unexpected token true`,
	}, {
		desc:         "oneof not set",
		inputMessage: &pb3.Oneofs{},
		inputText:    "{}",
		wantMessage:  &pb3.Oneofs{},
	}, {
		desc:         "oneof set to empty string",
		inputMessage: &pb3.Oneofs{},
		inputText:    `{"oneofString": ""}`,
		wantMessage: &pb3.Oneofs{
			Union: &pb3.Oneofs_OneofString{},
		},
	}, {
		desc:         "oneof set to string",
		inputMessage: &pb3.Oneofs{},
		inputText:    `{"oneofString": "hello"}`,
		wantMessage: &pb3.Oneofs{
			Union: &pb3.Oneofs_OneofString{
				OneofString: "hello",
			},
		},
	}, {
		desc:         "oneof set to enum",
		inputMessage: &pb3.Oneofs{},
		inputText:    `{"oneofEnum": "ZERO"}`,
		wantMessage: &pb3.Oneofs{
			Union: &pb3.Oneofs_OneofEnum{
				OneofEnum: pb3.Enum_ZERO,
			},
		},
	}, {
		desc:         "oneof set to empty message",
		inputMessage: &pb3.Oneofs{},
		inputText:    `{"oneofNested": {}}`,
		wantMessage: &pb3.Oneofs{
			Union: &pb3.Oneofs_OneofNested{
				OneofNested: &pb3.Nested{},
			},
		},
	}, {
		desc:         "oneof set to message",
		inputMessage: &pb3.Oneofs{},
		inputText: `{
  "oneofNested": {
    "sString": "nested message"
  }
}`,
		wantMessage: &pb3.Oneofs{
			Union: &pb3.Oneofs_OneofNested{
				OneofNested: &pb3.Nested{
					SString: "nested message",
				},
			},
		},
	}, {
		desc:         "oneof set to more than one field",
		inputMessage: &pb3.Oneofs{},
		inputText: `{
  "oneofEnum": "ZERO",
  "oneofString": "hello"
}`,
		wantErr: `(line 3:3): error parsing "oneofString", oneof pb3.Oneofs.union is already set`,
	}, {
		desc:         "oneof set to null and value",
		inputMessage: &pb3.Oneofs{},
		inputText: `{
  "oneofEnum": "ZERO",
  "oneofString": null
}`,
		wantMessage: &pb3.Oneofs{
			Union: &pb3.Oneofs_OneofEnum{
				OneofEnum: pb3.Enum_ZERO,
			},
		},
	}, {
		desc:         "map fields 1",
		inputMessage: &pb3.Maps{},
		inputText: `{
  "int32ToStr": {
    "-101": "-101",
	"0"   : "zero",
	"255" : "0xff"
  },
  "boolToUint32": {
    "false": 101,
	"true" : "42"
  }
}`,
		wantMessage: &pb3.Maps{
			Int32ToStr: map[int32]string{
				-101: "-101",
				0xff: "0xff",
				0:    "zero",
			},
			BoolToUint32: map[bool]uint32{
				true:  42,
				false: 101,
			},
		},
	}, {
		desc:         "map fields 2",
		inputMessage: &pb3.Maps{},
		inputText: `{
  "uint64ToEnum": {
    "1" : "ONE",
	"2" : 2,
	"10": 101
  }
}`,
		wantMessage: &pb3.Maps{
			Uint64ToEnum: map[uint64]pb3.Enum{
				1:  pb3.Enum_ONE,
				2:  pb3.Enum_TWO,
				10: 101,
			},
		},
	}, {
		desc:         "map fields 3",
		inputMessage: &pb3.Maps{},
		inputText: `{
  "strToNested": {
    "nested_one": {
	  "sString": "nested in a map"
    },
    "nested_two": {}
  }
}`,
		wantMessage: &pb3.Maps{
			StrToNested: map[string]*pb3.Nested{
				"nested_one": {
					SString: "nested in a map",
				},
				"nested_two": {},
			},
		},
	}, {
		desc:         "map fields 4",
		inputMessage: &pb3.Maps{},
		inputText: `{
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
		wantMessage: &pb3.Maps{
			StrToOneofs: map[string]*pb3.Oneofs{
				"string": {
					Union: &pb3.Oneofs_OneofString{
						OneofString: "hello",
					},
				},
				"nested": {
					Union: &pb3.Oneofs_OneofNested{
						OneofNested: &pb3.Nested{
							SString: "nested oneof in map field value",
						},
					},
				},
			},
		},
	}, {
		desc:         "map contains duplicate keys",
		inputMessage: &pb3.Maps{},
		inputText: `{
  "int32ToStr": {
    "0": "cero",
    "0": "zero"
  }
}
`,
		wantErr: `(line 4:5): duplicate map key "0"`,
	}, {
		desc:         "map key empty string",
		inputMessage: &pb3.Maps{},
		inputText: `{
  "strToNested": {
    "": {}
  }
}`,
		wantMessage: &pb3.Maps{
			StrToNested: map[string]*pb3.Nested{
				"": {},
			},
		},
	}, {
		desc:         "map contains invalid key 1",
		inputMessage: &pb3.Maps{},
		inputText: `{
  "int32ToStr": {
    "invalid": "cero"
  }
}`,
		wantErr: `invalid value for int32 key: "invalid"`,
	}, {
		desc:         "map contains invalid key 2",
		inputMessage: &pb3.Maps{},
		inputText: `{
  "int32ToStr": {
    "1.02": "float"
  }
}`,
		wantErr: `invalid value for int32 key: "1.02"`,
	}, {
		desc:         "map contains invalid key 3",
		inputMessage: &pb3.Maps{},
		inputText: `{
  "int32ToStr": {
    "2147483648": "exceeds 32-bit integer max limit"
  }
}`,
		wantErr: `invalid value for int32 key: "2147483648"`,
	}, {
		desc:         "map contains invalid key 4",
		inputMessage: &pb3.Maps{},
		inputText: `{
  "uint64ToEnum": {
    "-1": 0
  }
}`,
		wantErr: `invalid value for uint64 key: "-1"`,
	}, {
		desc:         "map contains invalid value",
		inputMessage: &pb3.Maps{},
		inputText: `{
  "int32ToStr": {
    "101": true
}`,
		wantErr: `invalid value for string type: true`,
	}, {
		desc:         "map contains null for scalar value",
		inputMessage: &pb3.Maps{},
		inputText: `{
  "int32ToStr": {
    "101": null
}`,
		wantErr: `invalid value for string type: null`,
	}, {
		desc:         "map contains null for message value",
		inputMessage: &pb3.Maps{},
		inputText: `{
  "strToNested": {
    "hello": null
  }
}`,
		wantErr: `unexpected token null`,
	}, {
		desc:         "map contains contains message value with invalid UTF8",
		inputMessage: &pb3.Maps{},
		inputText: `{
  "strToNested": {
    "hello": {
      "sString": "` + "abc\xff" + `"
	}
  }
}`,
		wantErr: `invalid UTF-8`,
	}, {
		desc:         "map key contains invalid UTF8",
		inputMessage: &pb3.Maps{},
		inputText: `{
  "strToNested": {
    "` + "abc\xff" + `": {}
  }
}`,
		wantErr: `invalid UTF-8`,
	}, {
		desc:         "Empty",
		inputMessage: &emptypb.Empty{},
		inputText:    `{}`,
		wantMessage:  &emptypb.Empty{},
	}, {
		desc:         "Empty contains unknown",
		inputMessage: &emptypb.Empty{},
		inputText:    `{"unknown": null}`,
		wantErr:      `unknown field "unknown"`,
	}, {
		desc:         "BoolValue false",
		inputMessage: &wrapperspb.BoolValue{},
		inputText:    `false`,
		wantMessage:  &wrapperspb.BoolValue{},
	}, {
		desc:         "BoolValue true",
		inputMessage: &wrapperspb.BoolValue{},
		inputText:    `true`,
		wantMessage:  &wrapperspb.BoolValue{Value: true},
	}, {
		desc:         "BoolValue invalid value",
		inputMessage: &wrapperspb.BoolValue{},
		inputText:    `{}`,
		wantErr:      `invalid value for bool type: {`,
	}, {
		desc:         "Int32Value",
		inputMessage: &wrapperspb.Int32Value{},
		inputText:    `42`,
		wantMessage:  &wrapperspb.Int32Value{Value: 42},
	}, {
		desc:         "Int32Value in JSON string",
		inputMessage: &wrapperspb.Int32Value{},
		inputText:    `"1.23e3"`,
		wantMessage:  &wrapperspb.Int32Value{Value: 1230},
	}, {
		desc:         "Int64Value",
		inputMessage: &wrapperspb.Int64Value{},
		inputText:    `"42"`,
		wantMessage:  &wrapperspb.Int64Value{Value: 42},
	}, {
		desc:         "UInt32Value",
		inputMessage: &wrapperspb.UInt32Value{},
		inputText:    `42`,
		wantMessage:  &wrapperspb.UInt32Value{Value: 42},
	}, {
		desc:         "UInt64Value",
		inputMessage: &wrapperspb.UInt64Value{},
		inputText:    `"42"`,
		wantMessage:  &wrapperspb.UInt64Value{Value: 42},
	}, {
		desc:         "FloatValue",
		inputMessage: &wrapperspb.FloatValue{},
		inputText:    `1.02`,
		wantMessage:  &wrapperspb.FloatValue{Value: 1.02},
	}, {
		desc:         "FloatValue exceeds max limit",
		inputMessage: &wrapperspb.FloatValue{},
		inputText:    `1.23e+40`,
		wantErr:      `invalid value for float type: 1.23e+40`,
	}, {
		desc:         "FloatValue Infinity",
		inputMessage: &wrapperspb.FloatValue{},
		inputText:    `"-Infinity"`,
		wantMessage:  &wrapperspb.FloatValue{Value: float32(math.Inf(-1))},
	}, {
		desc:         "DoubleValue",
		inputMessage: &wrapperspb.DoubleValue{},
		inputText:    `1.02`,
		wantMessage:  &wrapperspb.DoubleValue{Value: 1.02},
	}, {
		desc:         "DoubleValue Infinity",
		inputMessage: &wrapperspb.DoubleValue{},
		inputText:    `"Infinity"`,
		wantMessage:  &wrapperspb.DoubleValue{Value: math.Inf(+1)},
	}, {
		desc:         "StringValue empty",
		inputMessage: &wrapperspb.StringValue{},
		inputText:    `""`,
		wantMessage:  &wrapperspb.StringValue{},
	}, {
		desc:         "StringValue",
		inputMessage: &wrapperspb.StringValue{},
		inputText:    `"谷歌"`,
		wantMessage:  &wrapperspb.StringValue{Value: "谷歌"},
	}, {
		desc:         "StringValue with invalid UTF8 error",
		inputMessage: &wrapperspb.StringValue{},
		inputText:    "\"abc\xff\"",
		wantErr:      `invalid UTF-8`,
	}, {
		desc:         "BytesValue",
		inputMessage: &wrapperspb.BytesValue{},
		inputText:    `"aGVsbG8="`,
		wantMessage:  &wrapperspb.BytesValue{Value: []byte("hello")},
	}, {
		desc:         "Value null",
		inputMessage: &structpb.Value{},
		inputText:    `null`,
		wantMessage:  &structpb.Value{Kind: &structpb.Value_NullValue{}},
	}, {
		desc:         "Value bool",
		inputMessage: &structpb.Value{},
		inputText:    `false`,
		wantMessage:  &structpb.Value{Kind: &structpb.Value_BoolValue{}},
	}, {
		desc:         "Value number",
		inputMessage: &structpb.Value{},
		inputText:    `1.02`,
		wantMessage:  &structpb.Value{Kind: &structpb.Value_NumberValue{1.02}},
	}, {
		desc:         "Value string",
		inputMessage: &structpb.Value{},
		inputText:    `"hello"`,
		wantMessage:  &structpb.Value{Kind: &structpb.Value_StringValue{"hello"}},
	}, {
		desc:         "Value string with invalid UTF8",
		inputMessage: &structpb.Value{},
		inputText:    "\"\xff\"",
		wantErr:      `invalid UTF-8`,
	}, {
		desc:         "Value empty struct",
		inputMessage: &structpb.Value{},
		inputText:    `{}`,
		wantMessage: &structpb.Value{
			Kind: &structpb.Value_StructValue{
				&structpb.Struct{Fields: map[string]*structpb.Value{}},
			},
		},
	}, {
		desc:         "Value struct",
		inputMessage: &structpb.Value{},
		inputText: `{
  "string": "hello",
  "number": 123,
  "null": null,
  "bool": false,
  "struct": {
    "string": "world"
  },
  "list": []
}`,
		wantMessage: &structpb.Value{
			Kind: &structpb.Value_StructValue{
				&structpb.Struct{
					Fields: map[string]*structpb.Value{
						"string": {Kind: &structpb.Value_StringValue{"hello"}},
						"number": {Kind: &structpb.Value_NumberValue{123}},
						"null":   {Kind: &structpb.Value_NullValue{}},
						"bool":   {Kind: &structpb.Value_BoolValue{false}},
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
							Kind: &structpb.Value_ListValue{&structpb.ListValue{}},
						},
					},
				},
			},
		},
	}, {
		desc:         "Value struct with invalid UTF8 string",
		inputMessage: &structpb.Value{},
		inputText:    "{\"string\": \"abc\xff\"}",
		wantErr:      `invalid UTF-8`,
	}, {
		desc:         "Value empty list",
		inputMessage: &structpb.Value{},
		inputText:    `[]`,
		wantMessage: &structpb.Value{
			Kind: &structpb.Value_ListValue{
				&structpb.ListValue{Values: []*structpb.Value{}},
			},
		},
	}, {
		desc:         "Value list",
		inputMessage: &structpb.Value{},
		inputText: `[
  "string",
  123,
  null,
  true,
  {},
  [
    "string",
	1.23,
	null,
	false
  ]
]`,
		wantMessage: &structpb.Value{
			Kind: &structpb.Value_ListValue{
				&structpb.ListValue{
					Values: []*structpb.Value{
						{Kind: &structpb.Value_StringValue{"string"}},
						{Kind: &structpb.Value_NumberValue{123}},
						{Kind: &structpb.Value_NullValue{}},
						{Kind: &structpb.Value_BoolValue{true}},
						{Kind: &structpb.Value_StructValue{&structpb.Struct{}}},
						{
							Kind: &structpb.Value_ListValue{
								&structpb.ListValue{
									Values: []*structpb.Value{
										{Kind: &structpb.Value_StringValue{"string"}},
										{Kind: &structpb.Value_NumberValue{1.23}},
										{Kind: &structpb.Value_NullValue{}},
										{Kind: &structpb.Value_BoolValue{false}},
									},
								},
							},
						},
					},
				},
			},
		},
	}, {
		desc:         "Value list with invalid UTF8 string",
		inputMessage: &structpb.Value{},
		inputText:    "[\"abc\xff\"]",
		wantErr:      `invalid UTF-8`,
	}, {
		desc:         "Duration empty string",
		inputMessage: &durationpb.Duration{},
		inputText:    `""`,
		wantErr:      `invalid google.protobuf.Duration value ""`,
	}, {
		desc:         "Duration with secs",
		inputMessage: &durationpb.Duration{},
		inputText:    `"3s"`,
		wantMessage:  &durationpb.Duration{Seconds: 3},
	}, {
		desc:         "Duration with escaped unicode",
		inputMessage: &durationpb.Duration{},
		inputText:    `"\u0033s"`,
		wantMessage:  &durationpb.Duration{Seconds: 3},
	}, {
		desc:         "Duration with -secs",
		inputMessage: &durationpb.Duration{},
		inputText:    `"-3s"`,
		wantMessage:  &durationpb.Duration{Seconds: -3},
	}, {
		desc:         "Duration with plus sign",
		inputMessage: &durationpb.Duration{},
		inputText:    `"+3s"`,
		wantMessage:  &durationpb.Duration{Seconds: 3},
	}, {
		desc:         "Duration with nanos",
		inputMessage: &durationpb.Duration{},
		inputText:    `"0.001s"`,
		wantMessage:  &durationpb.Duration{Nanos: 1e6},
	}, {
		desc:         "Duration with -nanos",
		inputMessage: &durationpb.Duration{},
		inputText:    `"-0.001s"`,
		wantMessage:  &durationpb.Duration{Nanos: -1e6},
	}, {
		desc:         "Duration with -nanos",
		inputMessage: &durationpb.Duration{},
		inputText:    `"-.001s"`,
		wantMessage:  &durationpb.Duration{Nanos: -1e6},
	}, {
		desc:         "Duration with +nanos",
		inputMessage: &durationpb.Duration{},
		inputText:    `"+.001s"`,
		wantMessage:  &durationpb.Duration{Nanos: 1e6},
	}, {
		desc:         "Duration with -secs -nanos",
		inputMessage: &durationpb.Duration{},
		inputText:    `"-123.000000450s"`,
		wantMessage:  &durationpb.Duration{Seconds: -123, Nanos: -450},
	}, {
		desc:         "Duration with large secs",
		inputMessage: &durationpb.Duration{},
		inputText:    `"10000000000.000000001s"`,
		wantMessage:  &durationpb.Duration{Seconds: 1e10, Nanos: 1},
	}, {
		desc:         "Duration with decimal without fractional",
		inputMessage: &durationpb.Duration{},
		inputText:    `"3.s"`,
		wantMessage:  &durationpb.Duration{Seconds: 3},
	}, {
		desc:         "Duration with decimal without integer",
		inputMessage: &durationpb.Duration{},
		inputText:    `"0.5s"`,
		wantMessage:  &durationpb.Duration{Nanos: 5e8},
	}, {
		desc:         "Duration max value",
		inputMessage: &durationpb.Duration{},
		inputText:    `"315576000000.999999999s"`,
		wantMessage:  &durationpb.Duration{Seconds: 315576000000, Nanos: 999999999},
	}, {
		desc:         "Duration min value",
		inputMessage: &durationpb.Duration{},
		inputText:    `"-315576000000.999999999s"`,
		wantMessage:  &durationpb.Duration{Seconds: -315576000000, Nanos: -999999999},
	}, {
		desc:         "Duration with +secs out of range",
		inputMessage: &durationpb.Duration{},
		inputText:    `"315576000001s"`,
		wantErr:      `google.protobuf.Duration value out of range: "315576000001s"`,
	}, {
		desc:         "Duration with -secs out of range",
		inputMessage: &durationpb.Duration{},
		inputText:    `"-315576000001s"`,
		wantErr:      `google.protobuf.Duration value out of range: "-315576000001s"`,
	}, {
		desc:         "Duration with nanos beyond 9 digits",
		inputMessage: &durationpb.Duration{},
		inputText:    `"0.1000000000s"`,
		wantErr:      `invalid google.protobuf.Duration value "0.1000000000s"`,
	}, {
		desc:         "Duration without suffix s",
		inputMessage: &durationpb.Duration{},
		inputText:    `"123"`,
		wantErr:      `invalid google.protobuf.Duration value "123"`,
	}, {
		desc:         "Duration invalid signed fraction",
		inputMessage: &durationpb.Duration{},
		inputText:    `"123.+123s"`,
		wantErr:      `invalid google.protobuf.Duration value "123.+123s"`,
	}, {
		desc:         "Duration invalid multiple .",
		inputMessage: &durationpb.Duration{},
		inputText:    `"123.123.s"`,
		wantErr:      `invalid google.protobuf.Duration value "123.123.s"`,
	}, {
		desc:         "Duration invalid integer",
		inputMessage: &durationpb.Duration{},
		inputText:    `"01s"`,
		wantErr:      `invalid google.protobuf.Duration value "01s"`,
	}, {
		desc:         "Timestamp zero",
		inputMessage: &timestamppb.Timestamp{},
		inputText:    `"1970-01-01T00:00:00Z"`,
		wantMessage:  &timestamppb.Timestamp{},
	}, {
		desc:         "Timestamp with tz adjustment",
		inputMessage: &timestamppb.Timestamp{},
		inputText:    `"1970-01-01T00:00:00+01:00"`,
		wantMessage:  &timestamppb.Timestamp{Seconds: -3600},
	}, {
		desc:         "Timestamp UTC",
		inputMessage: &timestamppb.Timestamp{},
		inputText:    `"2019-03-19T23:03:21Z"`,
		wantMessage:  &timestamppb.Timestamp{Seconds: 1553036601},
	}, {
		desc:         "Timestamp with escaped unicode",
		inputMessage: &timestamppb.Timestamp{},
		inputText:    `"2019-0\u0033-19T23:03:21Z"`,
		wantMessage:  &timestamppb.Timestamp{Seconds: 1553036601},
	}, {
		desc:         "Timestamp with nanos",
		inputMessage: &timestamppb.Timestamp{},
		inputText:    `"2019-03-19T23:03:21.000000001Z"`,
		wantMessage:  &timestamppb.Timestamp{Seconds: 1553036601, Nanos: 1},
	}, {
		desc:         "Timestamp max value",
		inputMessage: &timestamppb.Timestamp{},
		inputText:    `"9999-12-31T23:59:59.999999999Z"`,
		wantMessage:  &timestamppb.Timestamp{Seconds: 253402300799, Nanos: 999999999},
	}, {
		desc:         "Timestamp above max value",
		inputMessage: &timestamppb.Timestamp{},
		inputText:    `"9999-12-31T23:59:59-01:00"`,
		wantErr:      `google.protobuf.Timestamp value out of range: "9999-12-31T23:59:59-01:00"`,
	}, {
		desc:         "Timestamp min value",
		inputMessage: &timestamppb.Timestamp{},
		inputText:    `"0001-01-01T00:00:00Z"`,
		wantMessage:  &timestamppb.Timestamp{Seconds: -62135596800},
	}, {
		desc:         "Timestamp below min value",
		inputMessage: &timestamppb.Timestamp{},
		inputText:    `"0001-01-01T00:00:00+01:00"`,
		wantErr:      `google.protobuf.Timestamp value out of range: "0001-01-01T00:00:00+01:00"`,
	}, {
		desc:         "Timestamp with nanos beyond 9 digits",
		inputMessage: &timestamppb.Timestamp{},
		inputText:    `"1970-01-01T00:00:00.0000000001Z"`,
		wantErr:      `invalid google.protobuf.Timestamp value`,
	}, {
		desc:         "FieldMask empty",
		inputMessage: &fieldmaskpb.FieldMask{},
		inputText:    `""`,
		wantMessage:  &fieldmaskpb.FieldMask{Paths: []string{}},
	}, {
		desc:         "FieldMask",
		inputMessage: &fieldmaskpb.FieldMask{},
		inputText:    `"foo,fooBar,foo.barQux,Foo"`,
		wantMessage: &fieldmaskpb.FieldMask{
			Paths: []string{
				"foo",
				"foo_bar",
				"foo.bar_qux",
				"_foo",
			},
		},
	}, {
		desc:         "FieldMask empty path 1",
		inputMessage: &fieldmaskpb.FieldMask{},
		inputText:    `"foo,"`,
		wantErr:      `google.protobuf.FieldMask.paths contains invalid path: ""`,
	}, {
		desc:         "FieldMask empty path 2",
		inputMessage: &fieldmaskpb.FieldMask{},
		inputText:    `"foo,  ,bar"`,
		wantErr:      `google.protobuf.FieldMask.paths contains invalid path: "  "`,
	}, {
		desc:         "FieldMask invalid char 1",
		inputMessage: &fieldmaskpb.FieldMask{},
		inputText:    `"foo_bar"`,
		wantErr:      `google.protobuf.FieldMask.paths contains invalid path: "foo_bar"`,
	}, {
		desc:         "FieldMask invalid char 2",
		inputMessage: &fieldmaskpb.FieldMask{},
		inputText:    `"foo@bar"`,
		wantErr:      `google.protobuf.FieldMask.paths contains invalid path: "foo@bar"`,
	}, {
		desc:         "Any empty",
		inputMessage: &anypb.Any{},
		inputText:    `{}`,
		wantMessage:  &anypb.Any{},
	}, {
		desc:         "Any with BoolValue",
		inputMessage: &anypb.Any{},
		inputText: `{
  "@type": "type.googleapis.com/google.protobuf.BoolValue",
  "value": true
}`,
		wantMessage: func() proto.Message {
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
	}, {
		desc:         "Any with Empty",
		inputMessage: &anypb.Any{},
		inputText: `{
  "value": {},
  "@type": "type.googleapis.com/google.protobuf.Empty"
}`,
		wantMessage: &anypb.Any{
			TypeUrl: "type.googleapis.com/google.protobuf.Empty",
		},
	}, {
		desc:         "Any with missing Empty",
		inputMessage: &anypb.Any{},
		inputText: `{
  "@type": "type.googleapis.com/google.protobuf.Empty"
}`,
		wantErr: `(line 3:1): missing "value" field`,
	}, {
		desc:         "Any with StringValue containing invalid UTF8",
		inputMessage: &anypb.Any{},
		inputText: `{
  "@type": "google.protobuf.StringValue",
  "value": "` + "abc\xff" + `"
}`,
		wantErr: `(line 3:12): invalid UTF-8`,
	}, {
		desc:         "Any with Int64Value",
		inputMessage: &anypb.Any{},
		inputText: `{
  "@type": "google.protobuf.Int64Value",
  "value": "42"
}`,
		wantMessage: func() proto.Message {
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
	}, {
		desc:         "Any with invalid Int64Value",
		inputMessage: &anypb.Any{},
		inputText: `{
  "@type": "google.protobuf.Int64Value",
  "value": "forty-two"
}`,
		wantErr: `(line 3:12): invalid value for int64 type: "forty-two"`,
	}, {
		desc:         "Any with invalid UInt64Value",
		inputMessage: &anypb.Any{},
		inputText: `{
  "@type": "google.protobuf.UInt64Value",
  "value": -42
}`,
		wantErr: `(line 3:12): invalid value for uint64 type: -42`,
	}, {
		desc:         "Any with Duration",
		inputMessage: &anypb.Any{},
		inputText: `{
  "@type": "type.googleapis.com/google.protobuf.Duration",
  "value": "0s"
}`,
		wantMessage: func() proto.Message {
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
	}, {
		desc:         "Any with Value of StringValue",
		inputMessage: &anypb.Any{},
		inputText: `{
  "@type": "google.protobuf.Value",
  "value": "` + "abc\xff" + `"
}`,
		wantErr: `(line 3:12): invalid UTF-8`,
	}, {
		desc:         "Any with Value of NullValue",
		inputMessage: &anypb.Any{},
		inputText: `{
  "@type": "google.protobuf.Value",
  "value": null
}`,
		wantMessage: func() proto.Message {
			m := &structpb.Value{Kind: &structpb.Value_NullValue{}}
			b, err := proto.MarshalOptions{Deterministic: true}.Marshal(m)
			if err != nil {
				t.Fatalf("error in binary marshaling message for Any.value: %v", err)
			}
			return &anypb.Any{
				TypeUrl: "google.protobuf.Value",
				Value:   b,
			}
		}(),
	}, {
		desc:         "Any with Struct",
		inputMessage: &anypb.Any{},
		inputText: `{
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
		wantMessage: func() proto.Message {
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
	}, {
		desc:         "Any with missing @type",
		umo:          protojson.UnmarshalOptions{},
		inputMessage: &anypb.Any{},
		inputText: `{
  "value": {}
}`,
		wantErr: `(line 1:1): missing "@type" field`,
	}, {
		desc:         "Any with empty @type",
		inputMessage: &anypb.Any{},
		inputText: `{
  "@type": ""
}`,
		wantErr: `(line 2:12): @type field contains empty value`,
	}, {
		desc:         "Any with duplicate value",
		inputMessage: &anypb.Any{},
		inputText: `{
  "@type": "google.protobuf.StringValue",
  "value": "hello",
  "value": "world"
}`,
		wantErr: `(line 4:3): duplicate "value" field`,
	}, {
		desc:         "DiscardUnknown: regular messages",
		umo:          protojson.UnmarshalOptions{DiscardUnknown: true},
		inputMessage: &pb3.Nests{},
		inputText: `{
  "sNested": {
    "unknown": {
      "foo": 1,
	  "bar": [1, 2, 3]
    }
  },
  "unknown": "not known"
}`,
		wantMessage: &pb3.Nests{SNested: &pb3.Nested{}},
	}, {
		desc:         "DiscardUnknown: map",
		umo:          protojson.UnmarshalOptions{DiscardUnknown: true},
		inputMessage: &pb3.Maps{},
		inputText: `{
  "strToNested": {
    "nested_one": {
	  "unknown": "what you see is not"
    }
  }
}`,
		wantMessage: &pb3.Maps{
			StrToNested: map[string]*pb3.Nested{
				"nested_one": {},
			},
		},
	}, {
		desc:         "DiscardUnknown: Empty",
		umo:          protojson.UnmarshalOptions{DiscardUnknown: true},
		inputMessage: &emptypb.Empty{},
		inputText:    `{"unknown": "something"}`,
		wantMessage:  &emptypb.Empty{},
	}, {
		desc:         "DiscardUnknown: Any without type",
		umo:          protojson.UnmarshalOptions{DiscardUnknown: true},
		inputMessage: &anypb.Any{},
		inputText: `{
  "value": {"foo": "bar"},
  "unknown": true
}`,
		wantMessage: &anypb.Any{},
	}, {
		desc: "DiscardUnknown: Any with Empty",
		umo: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
		inputMessage: &anypb.Any{},
		inputText: `{
  "@type": "type.googleapis.com/google.protobuf.Empty",
  "value": {"unknown": 47}
}`,
		wantMessage: &anypb.Any{
			TypeUrl: "type.googleapis.com/google.protobuf.Empty",
		},
	}, }

	for _, tt := range tests {
		tt := tt
		if tt.skip {
			continue
		}
		t.Run(tt.desc, func(t *testing.T) {
			cfg := jsoniter.Config{SortMapKeys: true}.Froze()
			cfg.RegisterExtension(&protoext.ProtoExtension{})
			err := cfg.Unmarshal([]byte(tt.inputText), tt.inputMessage)
			if err != nil {
				if tt.wantErr == "" {
					t.Errorf("Unmarshal() got unexpected error: %v", err)
				} else if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("Unmarshal() error got %q, want %q", err, tt.wantErr)
				}
				return
			}
			if tt.wantErr != "" {
				t.Errorf("Unmarshal() got nil error, want error %q", tt.wantErr)
			}
			if tt.wantMessage != nil && !proto.Equal(tt.inputMessage, tt.wantMessage) {
				t.Errorf("Unmarshal()\n<got>\n%v\n<want>\n%v\n", tt.inputMessage, tt.wantMessage)
			}
		})
	}
}
