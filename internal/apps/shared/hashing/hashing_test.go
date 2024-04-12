package hashing_test

import (
	"testing"

	"github.com/dlomanov/mon/internal/apps/shared/hashing"
)

func BenchmarkComputeBase64URLHash(b *testing.B) {
	const key = "91368e80-3324-4d67-b608-4d2be7934958"
	const value = "acbf05b3-a963-4c1b-9096-963c32ca5a19"

	for i := 0; i < b.N; i++ {
		_ = hashing.ComputeBase64URLHash(key, []byte(value))
	}
}
