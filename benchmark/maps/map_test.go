package maps

import (
	"fmt"
	"sync"
	"testing"
)

var (
	m     = make(map[string]string)
	syncM = sync.Map{}
	keys  = []string{"key1", "key2", "key3", "key4", "key5", "key6", "key7", "key8", "key9", "key10"}
)

func init() {
	for i, key := range keys {
		value := fmt.Sprintf("value%d", i)
		m[key] = value
		syncM.Store(key, value)
	}
}

//go test -bench=.
// goos: darwin
// goarch: amd64
// pkg: go-workspace/benchmark/maps
// Benchmark/map-12                70013574                17.1 ns/op
// Benchmark/syncmap-12            32989068                34.7 ns/op
// PASS
// ok      go-workspace/benchmark/maps     3.679s
func Benchmark(b *testing.B) {
	b.Run("map", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := keys[i%len(keys)]
			_, _ = m[key]
		}
	})

	b.Run("syncmap", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := keys[i%len(keys)]
			_, _ = syncM.Load(key)
		}
	})
}
