package hashing_test

import (
	"github.com/dlomanov/mon/internal/infra/services/hashing"
	"testing"

	"github.com/stretchr/testify/require"
)

func BenchmarkComputeBase64URLHash(b *testing.B) {
	const key = "91368e80-3324-4d67-b608-4d2be7934958"
	const value = "acbf05b3-a963-4c1b-9096-963c32ca5a19"

	for i := 0; i < b.N; i++ {
		_ = hashing.ComputeBase64URLHash(key, []byte(value))
	}
}

func TestComputeBase64URLHash(t *testing.T) {
	const key = "91368e80-3324-4d67-b608-4d2be7934958"
	const value = "acbf05b3-a963-4c1b-9096-963c32ca5a19"

	hashs := make([]string, 100)
	for i := 0; i < len(hashs); i++ {
		hash := hashing.ComputeBase64URLHash(key, []byte(value))
		require.NotEmpty(t, hash)
		hashs[i] = hash
	}

	for i := 1; i < len(hashs); i++ {
		require.Equal(t, hashs[0], hashs[i])
	}
}
