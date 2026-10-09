package crypto

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateSecureTokenAndHash(t *testing.T) {
	first, err := GenerateSecureToken()
	require.NoError(t, err)
	second, err := GenerateSecureToken()
	require.NoError(t, err)
	require.NotEmpty(t, first)
	require.NotEqual(t, first, second)
	require.Len(t, HashToken(first), 64)
	require.Equal(t, HashToken(first), HashToken(first))
	require.NotEqual(t, first, HashToken(first))
}
