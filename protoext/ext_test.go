package protoext_test

import (
	"bytes"
	"encoding/json"
	"math"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	jsoniter "github.com/json-iterator/go"
	"github.com/json-iterator/go/extra"
	"github.com/json-iterator/go/protoext"
	testv1 "github.com/json-iterator/go/protoext/internal/gen/go/test/v1"
	"github.com/modern-go/reflect2"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var timeCase time.Time

func init() {
	timeCase, _ = time.Parse("2006-01-02 15:04:05.999", "2022-06-09 21:03:49.560")
	timeCase = timeCase.UTC()
}

func ProtoEqual(m, m2 proto.Message) bool {
	// proto.Equal cant handle any.Any which contains map
	// https://github.com/golang/protobuf/issues/455
	return cmp.Diff(m, m2, protocmp.Transform()) == ""
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

func commonCheckMarshalEqual(t *testing.T, cfg jsoniter.API, opts *protojson.MarshalOptions, m proto.Message) (string, string) {
	if opts == nil {
		opts = &protojson.MarshalOptions{}
	}

	var err error
	var jsnA, jsnB string

	jsnA, err = cfg.MarshalToString(m)
	assert.Nil(t, err)
	jsnB, err = pMarshalToStringWithOpts(*opts, m)
	assert.Nil(t, err)
	assert.Equal(t, jsnA, jsnB)
	// log.Println(jsnA)
	return jsnA, jsnB
}

func commonCheck(t *testing.T, cfg jsoniter.API, opts *protojson.MarshalOptions, m proto.Message) (string, string) {
	jsnA, jsnB := commonCheckMarshalEqual(t, cfg, opts, m)

	m2 := proto.Clone(m)
	err := cfg.UnmarshalFromString(jsnA, m2)
	assert.Nil(t, err)
	// TIPS: If you have operated on m, such as `Clone` `protojson.Marshal`, etc., you cant use assert.Equal(t,m,m2) to check equality
	assert.Equal(t, "", cmp.Diff(m, m2, protocmp.Transform()))
	// assert.True(t, proto.Equal(m, m2))

	m2 = proto.Clone(m)
	err = pUnmarshalFromString(jsnB, m2)
	assert.Nil(t, err)
	assert.Equal(t, "", cmp.Diff(m, m2, protocmp.Transform()))

	return jsnA, jsnB
}

func TestJsonName(t *testing.T) {
	var err error
	m2 := &testv1.All{}
	m := &testv1.All{
		SnakeCase:      "snakeCase✅",
		LowerCamelCase: "lowerCamelCase✅",
		UpwerCamelCase: "UpwerCamelCase✅",
	}

	// UseProtoNames: false
	cfg := jsoniter.Config{SortMapKeys: true}.Froze()
	cfg.RegisterExtension(&protoext.ProtoExtension{})
	commonCheck(t, cfg, nil, m)
	// fuzzy decode
	err = cfg.UnmarshalFromString(`{"snake_case":"snakeCase✅"}`, m2)
	assert.Nil(t, err)
	assert.Equal(t, "snakeCase✅", m2.SnakeCase)

	// UseProtoNames: true
	m.SnakeCase = "snake_case✅"
	cfg = jsoniter.Config{SortMapKeys: true}.Froze()
	cfg.RegisterExtension(&protoext.ProtoExtension{UseProtoNames: true})
	commonCheck(t, cfg, &protojson.MarshalOptions{UseProtoNames: true}, m)
	// fuzzy decode
	m2.Reset()
	err = cfg.UnmarshalFromString(`{"snakeCase":"snake_case✅"}`, m2)
	assert.Nil(t, err)
	assert.Equal(t, "snake_case✅", m2.SnakeCase)
}

func TestScalar(t *testing.T) {
	var err error
	cfg := jsoniter.Config{SortMapKeys: true}.Froze()
	cfg.RegisterExtension(&protoext.ProtoExtension{})

	// []byte decode???? // TODO:

	// nan
	nan := math.NaN()
	m := &testv1.Singular{
		I32:   int32(nan),
		I64:   int64(nan),
		U32:   uint32(nan),
		U64:   uint64(nan),
		Sfi64: int64(nan),
		F32:   float32(nan),
		F64:   float64(nan),
	}
	commonCheckMarshalEqual(t, cfg, nil, m)

	// inf
	inf := math.Inf(+1)
	m = &testv1.Singular{
		I32:   int32(inf),
		I64:   int64(inf),
		U32:   uint32(inf),
		U64:   uint64(inf),
		Sfi64: int64(inf),
		F32:   float32(inf),
		F64:   float64(inf),
	}
	commonCheck(t, cfg, nil, m)

	inf = math.Inf(-1)
	m = &testv1.Singular{
		I32:   int32(inf),
		I64:   int64(inf),
		U32:   uint32(inf),
		U64:   uint64(inf),
		Sfi64: int64(inf),
		F32:   float32(inf),
		F64:   float64(inf),
	}
	commonCheck(t, cfg, nil, m)

	// fuzzy decode float
	m = &testv1.Singular{}
	err = cfg.UnmarshalFromString(`{"f32":"123.1","f64":"234.5"}`, m)
	assert.Nil(t, err)
	assert.Equal(t, float32(123.1), m.F32)
	assert.Equal(t, float64(234.5), m.F64)

	// fuzzy decode all
	m = &testv1.Singular{}
	err = cfg.UnmarshalFromString(`{"e":"JSON_ENUM_SOME","s":100,"i32":"1","i64":2,"u32":"3","u64":4,"f32":"5","f64":"6","si32":"7","si64":8,"fi32":"9","fi64":10,"sfi32":"11","sfi64":12,"bl":"true"}`, m)
	assert.Nil(t, err)
	assert.True(t, ProtoEqual(&testv1.Singular{
		E:     testv1.JsonEnum_JSON_ENUM_SOME,
		S:     "100",
		I32:   1,
		I64:   2,
		U32:   3,
		U64:   4,
		F32:   5,
		F64:   6,
		Si32:  7,
		Si64:  8,
		Fi32:  9,
		Fi64:  10,
		Sfi32: 11,
		Sfi64: 12,
		Bl:    true,
	}, m))

	// nan wkt
	mm := &testv1.WKTs{
		F32: wrapperspb.Float(float32(math.NaN())),
		F64: wrapperspb.Double(math.NaN()),
	}
	commonCheckMarshalEqual(t, cfg, nil, mm)

	// inf
	mm = &testv1.WKTs{
		F32: wrapperspb.Float(float32(math.Inf(+1))),
		F64: wrapperspb.Double(math.Inf(-1)),
	}
	commonCheck(t, cfg, nil, mm)

	// fuzzy decode float
	mm = &testv1.WKTs{}
	err = cfg.UnmarshalFromString(`{"f32":"123.1","f64":"234.5"}`, mm)
	assert.Nil(t, err)
	assert.Equal(t, float32(123.1), mm.F32.GetValue())
	assert.Equal(t, float64(234.5), mm.F64.GetValue())

	vv := &wrapperspb.StringValue{Value: "abc\xff"}
	_, err = cfg.MarshalToString(vv)
	assert.Contains(t, err.Error(), "invalid UTF-8")

	commonCheck(t, cfg, nil, &wrapperspb.StringValue{Value: "\u0000\u0008\u2028\"\\/\b\f\n\r\t你好啊朋友"})

	cfg = jsoniter.Config{SortMapKeys: true}.Froze()
	cfg.RegisterExtension(&protoext.ProtoExtension{
		PermitInvalidUTF8: true,
	})
	jsn, err := cfg.MarshalToString(&wrapperspb.StringValue{Value: "abc\xff"})
	assert.Nil(t, err)
	assert.Equal(t, "\"abc\xff\"", jsn)
}

func TestEmitUnpopulated(t *testing.T) {
	lv, _ := structpb.NewList([]interface{}{
		nil,
		true,
		-1,
		1.5,
		"str",
		[]byte(nil),
		map[string]interface{}{
			"b": false,
		},
		[]interface{}{
			1, 2, 3, nil,
		},
	})
	m := &testv1.All{
		S: &testv1.Singular{
			E:    testv1.JsonEnum_JSON_ENUM_UNSPECIFIED,
			Si64: 0,
		},
		Wkt: &testv1.WKTs{
			T:    timestamppb.New(timeCase),
			D:    durationpb.New(36 * time.Second),
			I64:  wrapperspb.Int64(0), // protojson will not omit zero value, only omit zero pointer, we stay compatible,
			U64:  wrapperspb.UInt64(0),
			Ui32: wrapperspb.UInt32(0),
			I32:  wrapperspb.Int32(-2),
			Nu:   structpb.NullValue_NULL_VALUE,

			Em: &emptypb.Empty{},
			V:  structpb.NewNullValue(),
			Fm: &fieldmaskpb.FieldMask{},

			Lv: lv,
		},
	}

	cfg := jsoniter.Config{SortMapKeys: true}.Froze()
	cfg.RegisterExtension(&protoext.ProtoExtension{})
	commonCheck(t, cfg, &protojson.MarshalOptions{}, m)

	cfg = jsoniter.Config{SortMapKeys: true}.Froze()
	cfg.RegisterExtension(&protoext.ProtoExtension{EmitUnpopulated: true})
	commonCheck(t, cfg, &protojson.MarshalOptions{EmitUnpopulated: true}, m)
}

func TestWkt(t *testing.T) {
	type M struct {
		A    *anypb.Any              `json:"a,omitempty"`
		D    durationpb.Duration     `json:"d,omitempty"`
		T    timestamppb.Timestamp   `json:"t,omitempty"`
		St   *structpb.Struct        `json:"st,omitempty"`
		I32  wrapperspb.Int32Value   `json:"i32,omitempty"`
		Ui32 wrapperspb.UInt32Value  `json:"ui32,omitempty"`
		I64  wrapperspb.Int64Value   `json:"i64,omitempty"`
		U64  wrapperspb.UInt64Value  `json:"u64,omitempty"`
		F32  wrapperspb.FloatValue   `json:"f32,omitempty"`
		F64  wrapperspb.DoubleValue  `json:"f64,omitempty"`
		B    *wrapperspb.BoolValue   `json:"b,omitempty"`
		S    *wrapperspb.StringValue `json:"s,omitempty"`
		By   *wrapperspb.BytesValue  `json:"by,omitempty"`
		Fm   *fieldmaskpb.FieldMask  `json:"fm,omitempty"`
		Em   *emptypb.Empty          `json:"em,omitempty"`
		Nu   structpb.NullValue      `json:"nu,omitempty"`
	}
	m := &M{
		T:    *timestamppb.New(timeCase),
		D:    *durationpb.New(36 * time.Second),
		I64:  *wrapperspb.Int64(0),
		U64:  *wrapperspb.UInt64(0),
		Ui32: *wrapperspb.UInt32(0),
		I32:  *wrapperspb.Int32(-2),
		Em:   &emptypb.Empty{},
		Fm: &fieldmaskpb.FieldMask{
			Paths: []string{"f.display_name", "f.b.c"},
		},
	}

	cfg := jsoniter.Config{SortMapKeys: true}.Froze()
	cfg.RegisterExtension(&protoext.ProtoExtension{})
	jsn, err := cfg.MarshalToString(m)
	assert.Nil(t, err)
	assert.Equal(t, `{"d":"36s","t":"2022-06-09T21:03:49.560Z","i32":-2,"ui32":0,"i64":"0","u64":"0","f32":0,"f64":0,"fm":"f.displayName,f.b.c","em":{}}`, jsn)

	cfg = jsoniter.Config{SortMapKeys: true}.Froze()
	cfg.RegisterExtension(&protoext.ProtoExtension{})
	// because m is not proto.Message, if we want all emit empty instead with `EmitUnpopulated:true`, should register EmitEmptyExtension
	cfg.RegisterExtension(&extra.EmitEmptyExtension{})
	jsn, err = cfg.MarshalToString(m)
	assert.Nil(t, err)
	assert.Equal(t, `{"a":null,"d":"36s","t":"2022-06-09T21:03:49.560Z","st":null,"i32":-2,"ui32":0,"i64":"0","u64":"0","f32":0,"f64":0,"b":null,"s":null,"by":null,"fm":"f.displayName,f.b.c","em":{},"nu":null}`, jsn)
}

func TestUnmarshalExistWkt(t *testing.T) {
	cfg := jsoniter.Config{SortMapKeys: true}.Froze()
	cfg.RegisterExtension(&protoext.ProtoExtension{})

	var err error
	m := &testv1.All{
		Wkt: &testv1.WKTs{
			D: durationpb.New(30 * time.Second),
		},
	}
	origP := reflect2.PtrOf(m.Wkt.D)
	err = cfg.UnmarshalFromString(`{"wkt":{"d":"20s"}}`, m)
	assert.Nil(t, err)
	assert.Equal(t, 20*time.Second, m.Wkt.D.AsDuration())
	assert.Equal(t, origP, reflect2.PtrOf(m.Wkt.D))
}

func TestNullValue(t *testing.T) {
	var jsn string
	var err error
	var ok bool
	cfg := jsoniter.Config{SortMapKeys: true}.Froze()
	cfg.RegisterExtension(&protoext.ProtoExtension{})

	nu := structpb.NullValue_NULL_VALUE
	m := &testv1.All{
		OWkt: &testv1.OneOfWKT{
			OneOf: &testv1.OneOfWKT_Nu{
				Nu: structpb.NullValue_NULL_VALUE,
			},
		},
		OptWkt: &testv1.OptionalWKTs{
			Nu: &nu,
			V:  structpb.NewNullValue(),
		},
	}
	jsn, _ = commonCheck(t, cfg, nil, m)
	m2 := &testv1.All{}
	err = cfg.UnmarshalFromString(jsn, m2)
	assert.Nil(t, err)
	assert.True(t, ProtoEqual(m, m2))
	_, ok = m2.OptWkt.V.GetKind().(*structpb.Value_NullValue)
	assert.True(t, ok)
	assert.Equal(t, structpb.NullValue_NULL_VALUE, *(m2.OptWkt.Nu))
	_, ok = m2.OWkt.GetOneOf().(*testv1.OneOfWKT_Nu)
	assert.True(t, ok)

	m.OWkt.OneOf = &testv1.OneOfWKT_V{
		V: structpb.NewNullValue(),
	}
	jsn, _ = commonCheck(t, cfg, nil, m)
	m2 = &testv1.All{}
	err = cfg.UnmarshalFromString(jsn, m2)
	assert.Nil(t, err)
	assert.True(t, ProtoEqual(m, m2))
	_, ok = m2.OptWkt.V.GetKind().(*structpb.Value_NullValue)
	assert.True(t, ok)
	assert.Equal(t, structpb.NullValue_NULL_VALUE, *(m2.OptWkt.Nu))
	wktV, ok := m2.OWkt.GetOneOf().(*testv1.OneOfWKT_V)
	assert.True(t, ok)
	_, ok = wktV.V.GetKind().(*structpb.Value_NullValue)
	assert.True(t, ok)
}

func TestEnum(t *testing.T) {
	cfg := jsoniter.Config{SortMapKeys: true}.Froze()
	cfg.RegisterExtension(&protoext.ProtoExtension{})

	m := &testv1.All{}
	m.E = testv1.JsonEnum_JSON_ENUM_UNSPECIFIED
	m.O = &testv1.Optionals{
		E: &m.E,
	}
	commonCheck(t, cfg, nil, m)

	m.E = testv1.JsonEnum_JSON_ENUM_SOME
	commonCheck(t, cfg, nil, m)

	m.E = testv1.JsonEnum(2)
	commonCheck(t, cfg, nil, m)

	var err error
	var jsn, jsnA, jsnB string
	m2 := &testv1.All{}

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
	cfg = jsoniter.Config{SortMapKeys: true}.Froze()
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

func TestInteger64AsString(t *testing.T) {
	cfg := jsoniter.Config{SortMapKeys: true}.Froze()
	cfg.RegisterExtension(&protoext.ProtoExtension{})

	i64 := int64(-224123123123123123)
	u64 := uint64(22412312321312312)
	m := &testv1.All{
		R: &testv1.Repeated{
			I64: []int64{-12, -23},
			U64: []uint64{22, 33},
		},
		S: &testv1.Singular{
			I64: -123123123123123123,
			U64: 12312312321312312,
		},
		O: &testv1.Optionals{
			I64: &i64,
			U64: &u64,
		},
		OF: &testv1.OneOf{
			OneOf: &testv1.OneOf_I64{
				I64: -786,
			},
		},
		Wkt: &testv1.WKTs{
			I64: wrapperspb.Int64(-333),
			U64: wrapperspb.UInt64(0),
		},
		RWkt: &testv1.RepeatedWKTs{
			I64: []*wrapperspb.Int64Value{
				wrapperspb.Int64(-333), wrapperspb.Int64(444),
			},
			U64: []*wrapperspb.UInt64Value{
				wrapperspb.UInt64(555), wrapperspb.UInt64(666),
			},
		},
		OptWkt: &testv1.OptionalWKTs{
			I64: wrapperspb.Int64(-777),
			U64: wrapperspb.UInt64(888),
		},
		OWkt: &testv1.OneOfWKT{
			OneOf: &testv1.OneOfWKT_I64{
				I64: wrapperspb.Int64(-999),
			},
		},
		M: &testv1.Map{
			Str: map[int64]string{
				101010: "helloworld",
				202020: "hellogod",
			},
		},
	}
	jsn, err := cfg.MarshalToString(m)
	assert.Nil(t, err)
	assert.Equal(t, `{"r":{"i64":["-12","-23"],"u64":["22","33"]},"s":{"i64":"-123123123123123123","u64":"12312312321312312"},"oF":{"i64":"-786"},"oWkt":{"i64":"-999"},"wkt":{"i64":"-333","u64":"0"},"o":{"i64":"-224123123123123123","u64":"22412312321312312"},"rWkt":{"i64":["-333","444"],"u64":["555","666"]},"m":{"str":{"101010":"helloworld","202020":"hellogod"}},"optWkt":{"i64":"-777","u64":"888"}}`, jsn)
	commonCheck(t, cfg, nil, m)
	m.OF.OneOf = &testv1.OneOf_U64{
		U64: 890,
	}
	commonCheck(t, cfg, nil, m)

	cfg = jsoniter.Config{SortMapKeys: true}.Froze()
	cfg.RegisterExtension(&protoext.ProtoExtension{Encode64BitAsInteger: true})
	jsn, err = cfg.MarshalToString(m)
	assert.Nil(t, err)
	assert.Equal(t, `{"r":{"i64":[-12,-23],"u64":[22,33]},"s":{"i64":-123123123123123123,"u64":12312312321312312},"oF":{"u64":890},"oWkt":{"i64":-999},"wkt":{"i64":-333,"u64":0},"o":{"i64":-224123123123123123,"u64":22412312321312312},"rWkt":{"i64":[-333,444],"u64":[555,666]},"m":{"str":{"101010":"helloworld","202020":"hellogod"}},"optWkt":{"i64":-777,"u64":888}}`, jsn)

	// TIPS: protjson does not support Encode64BitAsInteger, so we does not need to check marshal result
	// but it support fuzzy unmarshal
	m2 := &testv1.All{}
	err = pUnmarshalFromString(jsn, m2)
	assert.Nil(t, err)
	assert.True(t, ProtoEqual(m, m2))

	// test map keys with 64bit
	mm := struct {
		M1 map[int64]uint64
		M2 map[uint64]int64
	}{
		M1: map[int64]uint64{-1: 10, -2: 20, -3: 30},
		M2: map[uint64]int64{1: -10, 2: -20, 3: -30},
	}
	jsn, err = cfg.MarshalToString(mm)
	assert.Nil(t, err)
	assert.Equal(t, `{"M1":{"-3":30,"-2":20,"-1":10},"M2":{"1":-10,"2":-20,"3":-30}}`, jsn)

	cfg = jsoniter.Config{SortMapKeys: true}.Froze()
	cfg.RegisterExtension(&protoext.ProtoExtension{})
	jsn, err = cfg.MarshalToString(mm)
	assert.Nil(t, err)
	assert.Equal(t, `{"M1":{"-3":"30","-2":"20","-1":"10"},"M2":{"1":"-10","2":"-20","3":"-30"}}`, jsn)
}

func TestSortMapKeys(t *testing.T) {
	cfg := jsoniter.Config{SortMapKeys: true}.Froze()
	cfg.RegisterExtension(&protoext.ProtoExtension{})

	m := &testv1.Map{
		Str: map[int64]string{-2: "a", -1: "b", -3: "c"},
		By:  map[bool][]byte{true: []byte(`a`), false: []byte(`b`)},
		Bo:  map[uint32]bool{10: false, 20: true, 188: true},
	}
	commonCheck(t, cfg, nil, m)

	cfg = jsoniter.Config{SortMapKeys: true}.Froze()
	cfg.RegisterExtension(&protoext.ProtoExtension{
		SortMapKeysAsString: true,
	})
	jsn, err := cfg.MarshalToString(m)
	assert.Nil(t, err)
	assert.Equal(t, `{"str":{"-1":"b","-2":"a","-3":"c"},"by":{"false":"Yg==","true":"YQ=="},"bo":{"10":false,"188":true,"20":true}}`, jsn)
}

func TestOneof(t *testing.T) {
	cfg := jsoniter.Config{SortMapKeys: true}.Froze()
	cfg.RegisterExtension(&protoext.ProtoExtension{})

	m := &testv1.All{}
	m.OF = &testv1.OneOf{
		OneOf: &testv1.OneOf_Bl{
			Bl: false,
		},
	}
	commonCheck(t, cfg, nil, m)
	m.OF.OneOf = &testv1.OneOf_STr{
		STr: "strOfOneof",
	}
	commonCheck(t, cfg, nil, m)

	// embedded test
	type InnerMM struct {
		*testv1.OneOf
		Name string  `json:"name"`
		F32  float32 `json:"f32"`
	}
	type MM struct {
		*InnerMM
		Age int   `json:"age"`
		I32 int32 `json:"i32,omitempty"` // test override
	}
	i32 := &testv1.OneOf_I32{
		I32: 100,
	}
	em := &MM{
		InnerMM: &InnerMM{
			OneOf: &testv1.OneOf{
				OneOf: i32,
			},
			Name: "nameA",
		},
		Age: 21,
	}
	jsn, err := cfg.MarshalToString(em.OneOf)
	assert.Nil(t, err)
	assert.Equal(t, `{"i32":100}`, jsn)
	jsn, err = cfg.MarshalToString(em.InnerMM)
	assert.Nil(t, err)
	assert.Equal(t, `{"i32":100,"name":"nameA","f32":0}`, jsn)
	jsn, err = cfg.MarshalToString(em)
	assert.Nil(t, err)
	assert.Equal(t, `{"name":"nameA","f32":0,"age":21}`, jsn)
	em.I32 = 300
	jsn, err = cfg.MarshalToString(em)
	assert.Nil(t, err)
	assert.Equal(t, `{"name":"nameA","f32":0,"age":21,"i32":300}`, jsn)

	err = cfg.UnmarshalFromString(`{"age":22}`, em)
	assert.Nil(t, err)
	assert.Equal(t, 22, em.Age)
	err = cfg.UnmarshalFromString(`{"f32":320}`, em)
	assert.Nil(t, err)
	assert.Equal(t, float32(320), em.F32)
	err = cfg.UnmarshalFromString(`{"extra":"extraS","i32":123}`, em)
	assert.Nil(t, err)
	assert.Equal(t, 22, em.Age)
	assert.Equal(t, float32(320), em.F32)
	assert.Equal(t, int32(123), em.I32)
	assert.Equal(t, "extraS", em.OneOf.Extra)
	assert.Equal(t, int32(100), em.OneOf.GetI32())
	err = cfg.UnmarshalFromString(`{"extra":"extraS2","i32":223}`, em.InnerMM)
	assert.Nil(t, err)
	assert.Equal(t, "extraS2", em.OneOf.Extra)
	assert.Equal(t, int32(223), em.OneOf.GetI32())

	// special wkt
	jsn, err = cfg.MarshalToString(structpb.NewStringValue("structpb.StrValue"))
	assert.Nil(t, err)
	assert.Equal(t, `"structpb.StrValue"`, jsn)
}

func TestAny(t *testing.T) {
	cfg := jsoniter.Config{SortMapKeys: true}.Froze()
	cfg.RegisterExtension(&protoext.ProtoExtension{})

	m := &testv1.All{
		Wkt:    &testv1.WKTs{},
		OptWkt: &testv1.OptionalWKTs{},
		RWkt:   &testv1.RepeatedWKTs{},
	}
	commonCheck(t, cfg, nil, m)
	m.Wkt.A = &anypb.Any{} // empty
	commonCheck(t, cfg, nil, m)

	m.Wkt.A, _ = anypb.New(wrapperspb.String("wrapStr"))
	m.OptWkt.A = m.Wkt.A
	m.RWkt.A = []*anypb.Any{m.Wkt.A}
	commonCheck(t, cfg, nil, m)
	m.Wkt.A, _ = anypb.New(&testv1.Message{Id: "idA"})
	m.OptWkt.A = m.Wkt.A
	m.RWkt.A = []*anypb.Any{m.Wkt.A}
	commonCheck(t, cfg, nil, m)
	s, _ := structpb.NewStruct(map[string]interface{}{
		"keyA": "valueA",
		"keyB": nil,
		"keyC": "valueC",
	})
	m.Wkt.A, _ = anypb.New(s)
	m.OptWkt.A = m.Wkt.A
	m.RWkt.A = []*anypb.Any{m.Wkt.A}
	commonCheck(t, cfg, nil, m)
	lv, _ := structpb.NewList([]interface{}{
		nil,
		true,
		-1,
		1.5,
		"str",
		[]byte(nil),
		map[string]interface{}{
			"b": false,
		},
		[]interface{}{
			1, 2, 3, nil,
		},
	})
	m.Wkt.A, _ = anypb.New(lv)
	m.OptWkt.A = m.Wkt.A
	m.RWkt.A = []*anypb.Any{m.Wkt.A}
	commonCheck(t, cfg, nil, m)

	// empty any
	err := cfg.UnmarshalFromString(`{"wkt":{"a":{}}}`, m)
	assert.Nil(t, err)
	assert.Equal(t, "", m.Wkt.A.GetTypeUrl())
	assert.Equal(t, 0, len(m.Wkt.A.GetValue()))

	// miss type
	err = cfg.UnmarshalFromString(`{"wkt":{"a":{"name":"s"}}}`, m)
	assert.Contains(t, err.Error(), "google.protobuf.Any: missing \"@type\" field")
}

func TestNilValues(t *testing.T) {
	cfg := jsoniter.Config{SortMapKeys: true}.Froze()
	cfg.RegisterExtension(&protoext.ProtoExtension{EmitUnpopulated: true})
	mOpts := &protojson.MarshalOptions{EmitUnpopulated: true}

	i32 := int32(-123)
	m := &testv1.Case{
		WktI32A:    nil,
		WktI32B:    wrapperspb.Int32(0),
		OptI32A:    nil, // protojson: be omitted even though EmitUnpopulated: true
		OptI32B:    &i32,
		OptWktI32A: nil, // protojson: be omitted even though EmitUnpopulated: true
		OptWktI32B: wrapperspb.Int32(0),
		RptWktI32: []*wrapperspb.Int32Value{
			wrapperspb.Int32(-1),
			wrapperspb.Int32(0),
			nil, // protojson: marshal to zero value instead with null
			wrapperspb.Int32(1),
		},
		MapWktI32: map[string]*wrapperspb.Int32Value{
			"a": nil,
			"b": wrapperspb.Int32(0),
		},

		B1:    nil, // protojson: marshal to "" instead with null
		B2:    []byte(`abc`),
		OptB1: nil, // protojson: be omitted even though EmitUnpopulated: true
		OptB2: []byte(`abc`),
		RptB:  [][]byte{[]byte(`ABC`), nil, []byte(``), []byte(`EFG`)},
		MapB:  map[string][]byte{"keyA": nil, "keyB": []byte(`HIJ`)},

		WktB1:    nil,
		WktB2:    wrapperspb.Bytes([]byte(`abc`)),
		OptWktB1: nil, // protojson: be omitted even though EmitUnpopulated: true
		OptWktB2: wrapperspb.Bytes([]byte(`abc`)),
		RptWktB: []*wrapperspb.BytesValue{
			wrapperspb.Bytes([]byte(`ABC`)),
			nil, // protojson: marshal to zero value instead with null
			wrapperspb.Bytes(nil),
			wrapperspb.Bytes([]byte(``)),
			wrapperspb.Bytes([]byte(`EFG`)),
		},
		MapWktB: map[string]*wrapperspb.BytesValue{
			"keyA": wrapperspb.Bytes(nil),
			"keyB": wrapperspb.Bytes([]byte(`HIJ`)),
		},

		RptMsg: []*testv1.Message{
			&testv1.Message{Id: "id1"},
			nil,
			&testv1.Message{Id: "id3"},
		},
		MapMsg: map[string]*testv1.Message{
			"msgA": &testv1.Message{Id: "ida"},
			"msgB": nil,
			"msgC": &testv1.Message{Id: "idc"},
		},
		MapEnum: map[string]testv1.JsonEnum{
			"enumA": testv1.JsonEnum_JSON_ENUM_SOME,
			"enumB": testv1.JsonEnum_JSON_ENUM_UNSPECIFIED,
		},
		MapWktU64: map[uint64]*wrapperspb.UInt64Value{
			1: wrapperspb.UInt64(123),
			2: wrapperspb.UInt64(223),
			3: nil,
		},
		WktV:  structpb.NewNullValue(),
		WktLv: (*(structpb.ListValue))(nil),
		WktS:  nil,
	}

	lv, err := structpb.NewList([]interface{}{"a", nil, "c"})
	assert.Nil(t, err)
	m.RptWktV = []*structpb.Value{
		structpb.NewBoolValue(true),
		// nil, // cant be nil, same with protojson
		structpb.NewListValue(lv),
		&structpb.Value{
			Kind: &structpb.Value_StructValue{}, // protojson marshal一个 nil struct value 为 {}
		},
	}

	s, err := structpb.NewStruct(map[string]interface{}{
		"keyA": "valueA",
		"keyB": nil,
		"keyC": "valueC",
	})
	assert.Nil(t, err)
	m.RptWktS = []*structpb.Struct{s, (*structpb.Struct)(nil)}
	m.RptWktLv = []*structpb.ListValue{nil, lv, nil}
	commonCheck(t, cfg, mOpts, m)
}

func TestOptionals(t *testing.T) {
	cfg := jsoniter.Config{SortMapKeys: true}.Froze()
	cfg.RegisterExtension(&protoext.ProtoExtension{})

	m := &testv1.Optionals{
		Bl:  proto.Bool(false),
		I32: proto.Int32(0),
		I64: proto.Int64(0),
		U32: proto.Uint32(0),
		U64: proto.Uint64(0),
		F32: proto.Float32(0),
		F64: proto.Float64(0),
		Id:  proto.String(""),
		By:  []byte{},
		E:   testv1.JsonEnum_JSON_ENUM_UNSPECIFIED.Enum(),
		S:   &testv1.Message{},
	}
	commonCheck(t, cfg, nil, m)

	mm := &testv1.OptionalWKTs{
		I64: wrapperspb.Int64(0),
		U64: wrapperspb.UInt64(0),
	}
	commonCheck(t, cfg, nil, mm)
}

// func TestCaseNull(t *testing.T) {
// 	var jsn string
// 	var err error
// 	cfg := jsoniter.Config{SortMapKeys: true}.Froze()
// 	cfg.RegisterExtension(&protoext.ProtoExtension{EmitUnpopulated: true})

// 	// var bs []byte
// 	// err = cfg.UnmarshalFromString(`"MTIz"`, &bs)
// 	// assert.Nil(t, err)
// 	// log.Printf("%s", string(bs))

// 	// a := "a"
// 	// strs := []string{a, "b"}
// 	// strsB := []*string{&a, nil}

// 	// strsC := []**string{&strsB[0], nil}
// 	// log.Printf("a => %p strs => %p strs[0] => %p strsB[0] => %p", &a, strs, &strs[0], &strsB[0])

// 	// m := struct {
// 	// 	Strs  []string
// 	// 	StrsB []*string
// 	// 	StrsC []**string
// 	// 	Strss [][]string
// 	// 	Bytes [][]byte
// 	// }{
// 	// 	Strs:  strs,
// 	// 	StrsB: strsB,
// 	// 	StrsC: strsC,
// 	// 	Strss: [][]string{[]string{"a"}, nil, []string{"c"}},
// 	// 	Bytes: [][]byte{[]byte(`a`), nil, []byte(`c`)},
// 	// }

// 	m := &testv1.CaseValue{
// 		V: structpb.NewBoolValue(false),
// 		// Strs: strs,
// 		// Nus: []structpb.NullValue{structpb.NullValue_NULL_VALUE, structpb.NullValue_NULL_VALUE},
// 		// Vs: []*structpb.Value{
// 		// 	structpb.NewNullValue(),
// 		// 	// nil,
// 		// 	// structpb.NewBoolValue(false),
// 		// 	&structpb.Value{
// 		// 		Kind: &structpb.Value_StructValue{}, // protojson marshal一个 nil struct value 为 {}
// 		// 	},
// 		// 	// &structpb.Value{
// 		// 	// 	Kind: (*structpb.Value_StructValue)(nil), // protojson marshal一个 nil struct value 为 {}
// 		// 	// },
// 		// },
// 	}
// 	// a, _ := anypb.New(wrapperspb.String("wrapStr"))
// 	// a, _ := anypb.New(&testv1.Message{Id: "idA"})
// 	// s, _ := structpb.NewStruct(map[string]interface{}{
// 	// 	"keyA": "valueA",
// 	// 	"keyB": nil,
// 	// 	"keyC": "valueC",
// 	// })
// 	// a, _ := anypb.New(s)
// 	// lv, _ := structpb.NewList([]interface{}{
// 	// 	nil,
// 	// 	true,
// 	// 	-1,
// 	// 	1.5,
// 	// 	"str",
// 	// 	[]byte(nil),
// 	// 	map[string]interface{}{
// 	// 		"b": false,
// 	// 	},
// 	// 	[]interface{}{
// 	// 		1, 2, 3, nil,
// 	// 	},
// 	// })
// 	// a, _ := anypb.New(lv)
// 	// m.A = a

// 	jsn, err = pMarshalToString(m)
// 	assert.Nil(t, err)
// 	log.Println(string(jsn))

// 	jsn, err = cfg.MarshalToString(m)
// 	assert.Nil(t, err)
// 	log.Println(string(jsn))

// 	bb, err := cfg.MarshalIndent(m, "", "    ")
// 	assert.Nil(t, err)
// 	log.Println(string(bb))

// 	m2 := proto.Clone(m)
// 	err = cfg.UnmarshalFromString(jsn, m2)
// 	assert.Nil(t, err)
// 	assert.True(t, ProtoEqual(m, m2))
// 	log.Printf("%+v", base64.StdEncoding.EncodeToString(m.GetA().GetValue()))
// 	log.Printf("%+v", base64.StdEncoding.EncodeToString(m2.(*testv1.CaseValue).GetA().GetValue()))
// 	log.Printf("%s", cmp.Diff(m, m2, protocmp.Transform()))

// 	m2 = proto.Clone(m)
// 	err = pUnmarshalFromString(jsn, m2)
// 	assert.Nil(t, err)
// 	assert.True(t, ProtoEqual(m, m2))
// 	log.Printf("%+v", base64.StdEncoding.EncodeToString(m.GetA().GetValue()))
// 	log.Printf("%+v", base64.StdEncoding.EncodeToString(m2.(*testv1.CaseValue).GetA().GetValue()))
// 	log.Printf("%s", cmp.Diff(m, m2, protocmp.Transform()))

// 	log.Println("----")
// 	cfg.MarshalToString(structpb.NewBoolValue(false))
// }
