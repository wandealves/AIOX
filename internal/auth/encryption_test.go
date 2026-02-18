package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptor(t *testing.T) {
	key := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	enc, err := NewEncryptor(key)
	require.NoError(t, err)

	t.Run("encrypt and decrypt", func(t *testing.T) {
		plaintext := "This is a secret system prompt"
		ciphertext, err := enc.Encrypt(plaintext)
		require.NoError(t, err)
		assert.NotEqual(t, plaintext, ciphertext)

		decrypted, err := enc.Decrypt(ciphertext)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("different encryptions produce different ciphertexts", func(t *testing.T) {
		plaintext := "Same text encrypted twice"
		ct1, _ := enc.Encrypt(plaintext)
		ct2, _ := enc.Encrypt(plaintext)
		assert.NotEqual(t, ct1, ct2, "random nonce should produce different ciphertexts")
	})

	t.Run("invalid key length", func(t *testing.T) {
		_, err := NewEncryptor("short")
		assert.Error(t, err)
	})

	t.Run("invalid ciphertext", func(t *testing.T) {
		_, err := enc.Decrypt("invalid-hex")
		assert.Error(t, err)
	})

	t.Run("tampered ciphertext", func(t *testing.T) {
		ct, _ := enc.Encrypt("test")
		// Tamper with the last byte
		tampered := ct[:len(ct)-2] + "00"
		_, err := enc.Decrypt(tampered)
		assert.Error(t, err)
	})
}
