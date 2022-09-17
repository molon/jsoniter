package protoext_test

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	gofuzz "github.com/google/gofuzz"
	jsoniter "github.com/json-iterator/go"
	"github.com/json-iterator/go/protoext"
	testv1 "github.com/json-iterator/go/protoext/internal/gen/go/test/v1"
	"github.com/json-iterator/go/protoext/internal/gen/go/test/v1/testv1fuzz"
	"github.com/srikrsna/goprotofuzz"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/structpb"
)

func fuzzNullValue(x *structpb.NullValue, f gofuzz.Continue) {
	*x = structpb.NullValue_NULL_VALUE
}

func appendFuzzFuncs(f *gofuzz.Fuzzer) *gofuzz.Fuzzer {
	return f.Funcs(testv1fuzz.FuzzFuncs()...).Funcs(goprotofuzz.FuzzWKT[:]...).Funcs(fuzzNullValue)
}

func FuzzReadWrite(f *testing.F) {
	cfg := jsoniter.Config{}.Froze()
	cfg.RegisterExtension(&protoext.ProtoExtension{})

	f.Add([]byte("0"))
	f.Add([]byte("7"))
	f.Add([]byte("8"))
	f.Add([]byte("9"))
	f.Add([]byte(""))
	f.Add([]byte("\u007f\x922"))
	f.Fuzz(func(t *testing.T, data []byte) {
		f := appendFuzzFuncs(gofuzz.NewFromGoFuzz(data))
		var before testv1.All
		f.Fuzz(&before)

		jsn, err := cfg.Marshal(&before)
		if err != nil {
			t.Fatal("marshal error: ", err)
		}
		if !json.Valid(jsn) {
			t.Fatal("invalid json: ", string(jsn))
		}

		var after testv1.All
		err = cfg.Unmarshal(jsn, &after)
		if err != nil {
			t.Fatal("unmarshal error: ", err)
		}

		if diff := cmp.Diff(&before, &after, protocmp.Transform()); diff != "" {
			t.Errorf("before and after did not match:\n %s", diff)
			t.Error(string(jsn))
		}
	})
}

func FuzzReadFromProtoJson(f *testing.F) {
	cfg := jsoniter.Config{}.Froze()
	cfg.RegisterExtension(&protoext.ProtoExtension{})

	f.Add([]byte("0"))
	f.Fuzz(func(t *testing.T, data []byte) {
		f := appendFuzzFuncs(gofuzz.NewFromGoFuzz(data))
		var before testv1.All
		f.Fuzz(&before)
		jsonData, err := protojson.Marshal(&before)
		if err != nil {
			t.Fatal("marshal error: ", err)
		}
		if !json.Valid(jsonData) {
			t.Fatal("invalid json: ", string(jsonData))
		}
		var after testv1.All
		err = cfg.Unmarshal(jsonData, &after)
		if err != nil {
			t.Fatal("unmarshal error: ", err)
		}
		if diff := cmp.Diff(&before, &after, protocmp.Transform()); diff != "" {
			t.Error("before and after did not match", diff)
			t.Error(string(jsonData))
		}
	})
}

func FuzzWriteToProtoJson(f *testing.F) {
	cfg := jsoniter.Config{}.Froze()
	cfg.RegisterExtension(&protoext.ProtoExtension{})

	f.Fuzz(func(t *testing.T, data []byte) {
		f := appendFuzzFuncs(gofuzz.NewFromGoFuzz(data))
		var before testv1.All
		f.Fuzz(&before)

		jsn, err := cfg.Marshal(&before)
		if err != nil {
			t.Fatal("marshal error: ", err)
		}

		if !json.Valid(jsn) {
			t.Fatal("invalid json: ", string(jsn))
		}
		var after testv1.All
		if err := protojson.Unmarshal(jsn, &after); err != nil {
			t.Fatal(err, string(jsn))
		}
		if diff := cmp.Diff(&before, &after, protocmp.Transform()); diff != "" {
			t.Error("before and after did not match", diff)
			t.Error(string(jsn))
		}
	})
}
