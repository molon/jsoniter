package protoext_test

import (
	"math/rand"
	"testing"

	gofuzz "github.com/google/gofuzz"
	jsoniter "github.com/json-iterator/go"
	"github.com/json-iterator/go/protoext"
	testv1 "github.com/json-iterator/go/protoext/internal/gen/go/test/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

func BenchmarkWrite(b *testing.B) {
	f := appendFuzzFuncs(gofuzz.New())
	var ms []*testv1.All
	for i := 0; i < 10000; i++ {
		var all testv1.All
		f.Fuzz(&all)
		ms = append(ms, &all)
	}

	b.ReportAllocs()
	b.ResetTimer()
	b.Run("protojson", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			m := ms[rand.Intn(len(ms))]
			_, _ = protojson.Marshal(m)
		}
	})

	cfg := jsoniter.Config{SortMapKeys: true, DisallowUnknownFields: true}.Froze()
	cfg.RegisterExtension(&protoext.ProtoExtension{})
	b.Run("jsoniter", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			m := ms[rand.Intn(len(ms))]
			_, _ = cfg.Marshal(m)
		}
	})

	fcfg := jsoniter.Config{SortMapKeys: false, DisallowUnknownFields: false}.Froze()
	fcfg.RegisterExtension(&protoext.ProtoExtension{PermitInvalidUTF8: true})
	b.Run("jsoniter-fast", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			m := ms[rand.Intn(len(ms))]
			_, _ = fcfg.Marshal(m)
		}
	})
}

func BenchmarkRead(b *testing.B) {
	f := appendFuzzFuncs(gofuzz.New())
	var buffers [][]byte
	for i := 0; i < 10000; i++ {
		var all testv1.All
		f.Fuzz(&all)
		buffer, _ := protojson.Marshal(&all)
		buffers = append(buffers, buffer)
	}

	var all testv1.All
	b.ReportAllocs()
	b.ResetTimer()
	b.Run("protojson", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buffer := buffers[rand.Intn(len(buffers))]
			_ = protojson.Unmarshal(buffer, &all)
		}
	})

	cfg := jsoniter.Config{SortMapKeys: true, DisallowUnknownFields: true}.Froze()
	cfg.RegisterExtension(&protoext.ProtoExtension{})
	b.Run("jsoniter", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buffer := buffers[rand.Intn(len(buffers))]
			_ = cfg.Unmarshal(buffer, &all)
		}
	})

	cfg = jsoniter.Config{SortMapKeys: false, DisallowUnknownFields: false}.Froze()
	cfg.RegisterExtension(&protoext.ProtoExtension{PermitInvalidUTF8: true})
	b.Run("jsoniter-fast", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buffer := buffers[rand.Intn(len(buffers))]
			_ = cfg.Unmarshal(buffer, &all)
		}
	})

	cfg = jsoniter.Config{SortMapKeys: true, DisallowUnknownFields: true}.Froze()
	cfg.RegisterExtension(&protoext.ProtoExtension{DisableFuzzyDecode: true})
	b.Run("jsoniter-nofuzzydecode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buffer := buffers[rand.Intn(len(buffers))]
			_ = cfg.Unmarshal(buffer, &all)
		}
	})

	cfg = jsoniter.Config{SortMapKeys: false, DisallowUnknownFields: false}.Froze()
	cfg.RegisterExtension(&protoext.ProtoExtension{PermitInvalidUTF8: true, DisableFuzzyDecode: true})
	b.Run("jsoniter-fast-nofuzzydecode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buffer := buffers[rand.Intn(len(buffers))]
			_ = cfg.Unmarshal(buffer, &all)
		}
	})

	cfg = jsoniter.Config{SortMapKeys: false, DisallowUnknownFields: false}.Froze()
	b.Run("jsoniter-noprotoext", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buffer := buffers[rand.Intn(len(buffers))]
			_ = cfg.Unmarshal(buffer, &all)
		}
	})
}
