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
	cfg := jsoniter.Config{}.Froze()
	cfg.RegisterExtension(&protoext.ProtoExtension{})

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
	b.Run("jsoniter", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			m := ms[rand.Intn(len(ms))]
			_, _ = cfg.Marshal(m)
		}
	})
}

func BenchmarkRead(b *testing.B) {
	cfg := jsoniter.Config{}.Froze()
	cfg.RegisterExtension(&protoext.ProtoExtension{})

	f := appendFuzzFuncs(gofuzz.New())
	var buffers [][]byte
	for i := 0; i < 10000; i++ {
		var all testv1.All
		f.Fuzz(&all)
		buffer, _ := protojson.Marshal(&all)
		// TODO: slow if use this, dont know why now
		// buffer, _ := cfg.Marshal(&all)
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
	b.Run("jsoniter", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buffer := buffers[rand.Intn(len(buffers))]
			_ = cfg.Unmarshal(buffer, &all)
		}
	})
}
