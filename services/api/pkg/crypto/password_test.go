package crypto_test

import (
	"github.com/mmtaee/ocserv-users-management/api/pkg/crypto"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func setup() *crypto.CustomPassword {
	err := os.Setenv("JWT_SECRET", "my-secret-key")
	if err != nil {
		return nil
	}
	return crypto.NewCustomPassword()
}

func TestCreatePasswordDefaultSaltLength(t *testing.T) {
	cp := setup()
	result := cp.CreatePassword("mypassword")

	assert.NotEmpty(t, result.Salt)
	assert.Equal(t, 6, len(result.Salt))
	assert.NotEmpty(t, result.Hash)
}

func TestCreatePasswordCustomSaltLength(t *testing.T) {
	cp := setup()
	result := cp.CreatePassword("mypassword", 10)

	assert.Equal(t, 10, len(result.Salt))
}

func TestCheckPasswordCorrectPassword(t *testing.T) {
	cp := setup()
	data := cp.CreatePassword("securepass")

	match := cp.CheckPassword("securepass", data.Hash, data.Salt)
	assert.True(t, match)
}

func TestCheckPasswordWrongPassword(t *testing.T) {
	cp := setup()
	data := cp.CreatePassword("securepass")

	match := cp.CheckPassword("wrongpass", data.Hash, data.Salt)
	assert.False(t, match)
}

func TestCheckPasswordWrongSalt(t *testing.T) {
	cp := setup()
	data := cp.CreatePassword("securepass")

	match := cp.CheckPassword("securepass", data.Hash, "badSalt")
	assert.False(t, match)
}
