package goche

import (
	"context"
	"testing"
)

func BenchmarkLFUSet(b *testing.B) {
	c := New(LFU)

	for i := 0; i < b.N; i++ {
		c.Set(context.Background(), "key", "value")
	}
}

func BenchmarkLFUGet(b *testing.B) {
	c := New(LFU)

	for i := 0; i < b.N; i++ {
		c.Get(context.Background(), "key")
	}
}
