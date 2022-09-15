package protoext_test

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/json-iterator/go/extra"
	"github.com/json-iterator/go/protoext"
	testv1 "github.com/json-iterator/go/protoext/internal/gen/go/test/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var timeCase time.Time

func init() {
	timeCase, _ = time.Parse("2006-01-02 15:04:05.999", "2022-06-09 21:03:49.560")
	timeCase = timeCase.UTC()
}

func pMarshalToStringWithOpts(opts protojson.MarshalOptions, m proto.Message) (string, error) {
	by, err := opts.Marshal(m)
	if err != nil {
		return "", err
	}
	// https://github.com/golang/protobuf/issues/1121
	var out bytes.Buffer
	err = json.Compact(&out, by)
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

func pMarshalToString(m proto.Message) (string, error) {
	return pMarshalToStringWithOpts(protojson.MarshalOptions{}, m)
}

func pUnmarshalFromString(s string, m proto.Message) error {
	return protojson.Unmarshal([]byte(s), m)
}

func TestCompareStdAndProto(t *testing.T) {
	type MM struct {
		I64   int64    `json:"i64"`
		I64S  []int64  `json:"i64S"`
		I64P  *int64   `json:"i64P"`
		I64PS []*int64 `json:"i64PS"`
	}
	s := MM{
		I64:  1502878518952376288,
		I64S: []int64{1502878518952376289, 1502878518952376290, 1502878518952376291},
	}
	s.I64P = &s.I64
	s.I64PS = []*int64{&s.I64, nil}

	bb, err := json.Marshal(s)
	jsn := string(bb)
	assert.Nil(t, err)
	assert.Equal(t, `{"i64":1502878518952376288,"i64S":[1502878518952376289,1502878518952376290,1502878518952376291],"i64P":1502878518952376288,"i64PS":[1502878518952376288,null]}`, jsn)

	cfg := jsoniter.Config{}.Froze()
	cfg.RegisterExtension(&protoext.ProtoExtension{})

	jsn, err = cfg.MarshalToString(s)
	assert.Nil(t, err)
	assert.Equal(t, `{"i64":1502878518952376288,"i64S":[1502878518952376289,1502878518952376290,1502878518952376291],"i64P":1502878518952376288,"i64PS":[1502878518952376288,null]}`, jsn)

	// TODO: protojson will stringify some scalar type
	/*
		case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Uint64Kind,
			protoreflect.Sfixed64Kind, protoreflect.Fixed64Kind:
			// 64-bit integers are written out as JSON string.
			e.WriteString(val.String())
	*/

	m := &testv1.Repeated{
		By: [][]byte{nil, []byte(`bytesA`)},
	}
	bb, err = json.Marshal(m)
	jsn = string(bb)
	assert.Nil(t, err)
	assert.Equal(t, `{"by":[null,"Ynl0ZXNB"]}`, jsn)

	jsn, err = cfg.MarshalToString(m)
	assert.Nil(t, err)
	assert.Equal(t, `{"by":[null,"Ynl0ZXNB"]}`, jsn)

	// TODO: nil at array, protojson will not returns `null``
	jsn, err = pMarshalToString(m)
	assert.Nil(t, err)
	assert.Equal(t, `{"by":["","Ynl0ZXNB"]}`, jsn)

	mm := &testv1.RepeatedWKTs{
		I64: []*wrapperspb.Int64Value{wrapperspb.Int64(123), wrapperspb.Int64(533), nil},
		Nu:  []structpb.NullValue{structpb.NullValue_NULL_VALUE, structpb.NewNullValue().GetNullValue()},
	}
	jsn, err = cfg.MarshalToString(mm)
	assert.Nil(t, err)
	assert.Equal(t, `{"i64":["123","533",null],"nu":[null,null]}`, jsn)

	jsn, err = pMarshalToString(mm)
	assert.Nil(t, err)
	assert.Equal(t, `{"i64":["123","533","0"],"nu":[null,null]}`, jsn)
	mm.Reset()
	err = pUnmarshalFromString(jsn, mm)
	assert.Nil(t, err)

	mm.Reset()
	err = pUnmarshalFromString(`{"i64":["123",533,"0"],"nu":[null,null]}`, mm)
	assert.Nil(t, err)

	// TODO: 会出错，protojson 不会解析数组和map里的null
	mm.Reset()
	err = pUnmarshalFromString(`{"i64":["123",533,"0",null],"nu":[null,null]}`, mm)
	assert.Nil(t, err)
}

func TestJsonName(t *testing.T) {
	var err error
	var jsnA, jsnB string
	m2 := &testv1.All{}
	m := &testv1.All{
		SnakeCase:      "snakeCase✅",
		LowerCamelCase: "lowerCamelCase✅",
		UpwerCamelCase: "UpwerCamelCase✅",
	}

	cfg := jsoniter.Config{}.Froze()
	cfg.RegisterExtension(&protoext.ProtoExtension{})
	jsnA, err = cfg.MarshalToString(m)
	assert.Nil(t, err)
	jsnB, err = pMarshalToString(m)
	assert.Nil(t, err)
	assert.Equal(t, jsnA, jsnB)

	m2.Reset()
	err = cfg.UnmarshalFromString(jsnA, m2)
	assert.Nil(t, err)
	assert.True(t, proto.Equal(m, m2))

	// fuzze decode
	m2.Reset()
	err = cfg.UnmarshalFromString(`{"snake_case":"snakeCase✅"}`, m2)
	assert.Nil(t, err)
	assert.Equal(t, "snakeCase✅", m2.SnakeCase)

	// UseProtoNames: true
	m.SnakeCase = "snake_case✅"

	cfg = jsoniter.Config{}.Froze()
	cfg.RegisterExtension(&protoext.ProtoExtension{UseProtoNames: true})
	jsnA, err = cfg.MarshalToString(m)
	assert.Nil(t, err)
	jsnB, err = pMarshalToStringWithOpts(protojson.MarshalOptions{UseProtoNames: true}, m)
	assert.Nil(t, err)
	assert.Equal(t, jsnA, jsnB)

	m2.Reset()
	err = cfg.UnmarshalFromString(jsnA, m2)
	assert.Nil(t, err)
	assert.True(t, proto.Equal(m, m2))

	// fuzze decode
	m2.Reset()
	err = cfg.UnmarshalFromString(`{"snakeCase":"snake_case✅"}`, m2)
	assert.Nil(t, err)
	assert.Equal(t, "snake_case✅", m2.SnakeCase)
}

func TestEmitUnpopulated(t *testing.T) {
	var err error
	var jsnA, jsnB string
	m2 := &testv1.All{}
	m := &testv1.All{
		Wkt: &testv1.WKTs{},
	}
	m.Wkt.T = timestamppb.New(timeCase)
	m.Wkt.D = durationpb.New(36 * time.Second)

	cfg := jsoniter.Config{}.Froze()
	cfg.RegisterExtension(&protoext.ProtoExtension{})
	cfg.RegisterExtension(&extra.EmitEmptyWithBindingExtension{Filter: protoext.ProtoEmitUnpopulated})
	jsnA, err = cfg.MarshalToString(m)
	assert.Nil(t, err)
	jsnB, err = pMarshalToStringWithOpts(protojson.MarshalOptions{EmitUnpopulated: true}, m)
	assert.Nil(t, err)
	assert.Equal(t, jsnA, jsnB)

	m2.Reset()
	err = cfg.UnmarshalFromString(jsnA, m2)
	assert.Nil(t, err)
	// TODO: 考虑内部不使用 reflect 方法去设置
	// assert.Equal(t, m, m2)
	assert.True(t, proto.Equal(m, m2))
}

func TestWellKnownTypes(t *testing.T) {
	// TODO: 如果是 any ，那 protojson 的 opts 如何传递进去呢？
	var err error
	var jsnA, jsnB string
	m2 := &testv1.All{}
	m := &testv1.All{
		Wkt: &testv1.WKTs{
			T:  timestamppb.Now(),
			D:  durationpb.New(36 * time.Second),
			Nu: structpb.NullValue_NULL_VALUE,
			// ...
		},
	}
	cfg := jsoniter.Config{}.Froze()
	cfg.RegisterExtension(&protoext.ProtoExtension{})

	jsnA, err = cfg.MarshalToString(m)
	assert.Nil(t, err)
	jsnB, err = pMarshalToString(m)
	assert.Nil(t, err)
	assert.Equal(t, jsnA, jsnB)

	m2.Reset()
	err = cfg.UnmarshalFromString(jsnA, m2)
	assert.Nil(t, err)
	// TODO: 考虑内部不使用 reflect 方法去设置
	assert.True(t, proto.Equal(m, m2))
}

func TestEnum(t *testing.T) {
	cfg := jsoniter.Config{}.Froze()
	cfg.RegisterExtension(&protoext.ProtoExtension{})

	var err error
	var jsnA, jsnB string
	m2 := &testv1.All{}
	m := &testv1.All{}
	m.E = testv1.JsonEnum_JSON_ENUM_UNSPECIFIED
	m.O = &testv1.Optionals{
		E: &m.E,
	}
	jsnA, err = cfg.MarshalToString(m)
	assert.Nil(t, err)
	jsnB, err = pMarshalToString(m)
	assert.Nil(t, err)
	assert.Equal(t, jsnA, jsnB)

	m.E = testv1.JsonEnum_JSON_ENUM_SOME
	jsn, err := cfg.MarshalToString(m)
	assert.Nil(t, err)
	assert.Equal(t, `{"e":"JSON_ENUM_SOME","o":{"e":"JSON_ENUM_SOME"}}`, jsn)

	m2.Reset()
	err = cfg.UnmarshalFromString(jsn, m2)
	assert.Nil(t, err)
	assert.True(t, proto.Equal(m, m2))

	// test fuzzy decode enum
	m2.Reset()
	err = cfg.UnmarshalFromString(`{"e":1,"o":{"e":"JSON_ENUM_SOME"}}`, m2)
	assert.Nil(t, err)
	assert.Equal(t, testv1.JsonEnum_JSON_ENUM_SOME, m2.E)
	assert.Equal(t, testv1.JsonEnum_JSON_ENUM_SOME, *m2.O.E)

	m2.Reset()
	err = cfg.UnmarshalFromString(`{"e":null,"o":{"e":"JSON_ENUM_SOME"}}`, m2)
	assert.Nil(t, err)
	assert.Equal(t, testv1.JsonEnum_JSON_ENUM_UNSPECIFIED, m2.E)
	assert.Equal(t, testv1.JsonEnum_JSON_ENUM_SOME, *m2.O.E)

	m2.Reset()
	err = cfg.UnmarshalFromString(`{"e":"1","o":{"e":null}}`, m2)
	assert.Nil(t, err)
	assert.Equal(t, testv1.JsonEnum_JSON_ENUM_SOME, m2.E)
	assert.Nil(t, m2.O.E)

	m2.Reset()
	err = protojson.UnmarshalOptions{}.Unmarshal([]byte(`{"e":"JSON_ENUM_SOME","o":{"e":null}}`), m2)
	assert.Nil(t, err)
	assert.Equal(t, testv1.JsonEnum_JSON_ENUM_SOME, m2.E)
	assert.Nil(t, m2.O.E)

	m = &testv1.All{
		R: &testv1.Repeated{},
	}
	m.R.E = []testv1.JsonEnum{testv1.JsonEnum_JSON_ENUM_SOME, testv1.JsonEnum_JSON_ENUM_UNSPECIFIED}
	jsn, err = cfg.MarshalToString(m)
	assert.Nil(t, err)
	assert.Equal(t, `{"r":{"e":["JSON_ENUM_SOME","JSON_ENUM_UNSPECIFIED"]}}`, jsn)

	// UseEnumNumbers: true
	cfg = jsoniter.Config{}.Froze()
	cfg.RegisterExtension(&protoext.ProtoExtension{UseEnumNumbers: true})

	m = &testv1.All{
		E: testv1.JsonEnum_JSON_ENUM_SOME,
		R: &testv1.Repeated{E: []testv1.JsonEnum{testv1.JsonEnum_JSON_ENUM_SOME, testv1.JsonEnum_JSON_ENUM_UNSPECIFIED}},
	}
	m.O = &testv1.Optionals{E: &m.E}
	jsnA, err = cfg.MarshalToString(m)
	assert.Nil(t, err)
	jsnB, err = pMarshalToStringWithOpts(protojson.MarshalOptions{UseEnumNumbers: true}, m)
	assert.Nil(t, err)
	assert.Equal(t, jsnA, jsnB)
}

func TestOneof(t *testing.T) {
	cfg := jsoniter.Config{}.Froze()
	cfg.RegisterExtension(&protoext.ProtoExtension{})

	var err error
	var jsnA, jsnB string
	m2 := &testv1.All{}
	m := &testv1.All{}

	m.OF = &testv1.OneOf{
		// OneOf: &testv1.OneOf_STr{
		// 	STr: "strOfOneof",
		// },
		OneOf: &testv1.OneOf_Bl{
			Bl: false,
		},
	}
	jsnA, err = cfg.MarshalToString(m)
	assert.Nil(t, err)
	jsnB, err = pMarshalToString(m)
	assert.Nil(t, err)
	assert.Equal(t, jsnA, jsnB)

	m2.Reset()
	err = cfg.UnmarshalFromString(jsnA, m2)
	assert.Nil(t, err)
	// TODO: 考虑内部不使用 reflect 方法去设置
	assert.True(t, proto.Equal(m, m2))

	// 	cfg := jsoniter.Config{}.Froze()
	// 	cfg.RegisterExtension(&protoext.ProtoExtension{})
	// 	// cfg.RegisterExtension(&protoext.EmitEmptyWithTypeExtension{})

	// 	fakeOneOfStr := "fakeOneOf"

	// 	type MM struct {
	// 		*testv1.OneOf
	// 		OneOf_  string  `json:"oneOf_,omitempty"`
	// 		OneOf_y string  `json:"oneOf_Y,omitempty"`
	// 		F32     float32 `json:"f32"`
	// 	}

	// 	i32 := &testv1.OneOf_I32{
	// 		I32: 3,
	// 	}
	// 	m := &MM{
	// 		OneOf: &testv1.OneOf{
	// 			OneOf:  &fakeOneOfStr,
	// 			OneOf_: i32,
	// 		},
	// 		OneOf_: "OutOneOf_x",
	// 	}
	// 	// m.OneOf.OneOf_

	// 	// log.Printf("%p %p", &m.OneOf.OneOf_, i32)
	// 	// log.Printf("%v", m.OneOf_)
	// 	// log.Printf("%v", m.OneOf)

	// 	jsn, err := cfg.MarshalToString(m)
	// 	assert.Nil(t, err)
	// 	assert.Equal(t, `{"OneOf":"fakeOneOf","i32":3,"oneOf_":"OutOneOf_x","f32":0}`, jsn)

	// 	// m = &MM{}
	// 	// err = cfg.UnmarshalFromString(`{"OneOf":"fakeOneOf","i32":3,"oneOf_":"OutOneOf_x","f32":0.5}`, m)
	// 	// assert.Nil(t, err)
	// 	// log.Printf("%+v", m.OneOf)
	// 	m = &MM{}
	// 	// TODO: 需要测试本来not nil，然后 unmarshal 成nil
	// 	err = cfg.UnmarshalFromString(`{"OneOf":"fakeOneOf","i32":3,"oneOf_":"OutOneOf_x","f32":0.5}`, m)
	// 	assert.Nil(t, err)
	// 	log.Printf("%#v", m.OneOf)

	// 	err = cfg.UnmarshalFromString(`{"OneOf":"fakeOneOf","i32":3,"oneOf_":"OutOneOf_x","f32":0.5}`, m)
	// 	assert.Nil(t, err)
	// 	log.Printf("%#v", m.OneOf)

	// 	// jsn, err = cfg.MarshalToString(m)
	// 	// assert.Nil(t, err)
	// 	// assert.Equal(t, `{"OneOf":"fakeOneOf","i32":3,"oneOf_":"OutOneOf_x","f32":0}`, jsn)

	// // jsn, err = pMarshalToString(m)
	// // assert.Nil(t, err)
	// // // protojson will only handle internal of testv1.OneOf
	// // assert.Equal(t, `{"OneOf":"fakeOneOf","i32":3}`, jsn)
}
