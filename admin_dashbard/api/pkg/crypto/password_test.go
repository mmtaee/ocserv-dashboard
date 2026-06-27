package crypto_test

import (
	"testing"

	"github.com/mmtaee/ocserv-dashboard/api/pkg/crypto"
	"github.com/mmtaee/ocserv-dashboard/core/pkg/config"
	"github.com/stretchr/testify/assert"
)

func setup(t *testing.T) *crypto.CustomPassword {
	t.Helper()

	t.Setenv("SECRET_KEY", "my-secret-key")
	config.Init(false, "", 0)

	return crypto.NewCustomPassword()
}

func TestCreatePasswordDefaultSaltLength(t *testing.T) {
	cp := setup(t)
	result := cp.CreatePassword("mypassword")

	assert.NotEmpty(t, result.Salt)
	assert.Equal(t, 6, len(result.Salt))
	assert.NotEmpty(t, result.Hash)
}

func TestCreatePasswordCustomSaltLength(t *testing.T) {
	cp := setup(t)
	result := cp.CreatePassword("mypassword", 10)

	assert.Equal(t, 10, len(result.Salt))
}

func TestCheckPasswordCorrectPassword(t *testing.T) {
	cp := setup(t)
	data := cp.CreatePassword("securepass")

	match := cp.CheckPassword("securepass", data.Hash, data.Salt)
	assert.True(t, match)
}

func TestCheckPasswordWrongPassword(t *testing.T) {
	cp := setup(t)
	data := cp.CreatePassword("securepass")

	match := cp.CheckPassword("wrongpass", data.Hash, data.Salt)
	assert.False(t, match)
}

func TestCheckPasswordWrongSalt(t *testing.T) {
	cp := setup(t)
	data := cp.CreatePassword("securepass")

	match := cp.CheckPassword("securepass", data.Hash, "badSalt")
	assert.False(t, match)
}
