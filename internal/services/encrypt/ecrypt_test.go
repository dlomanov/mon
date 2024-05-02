package encrypt_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"strings"
	"testing"

	"github.com/dlomanov/mon/internal/services/encrypt"
	"github.com/stretchr/testify/require"
)

func TestEncrypt(t *testing.T) {
	publicKey, privateKey := createKeys(t)
	enc, err := encrypt.NewEncryptor(publicKey)
	require.NoError(t, err, "failed to create encryptor")
	dec, err := encrypt.NewDecryptor(privateKey)
	require.NoError(t, err, "failed to create decryptor")

	msg := []byte(strings.Repeat("test", 1000))
	encMsg, err := enc.Encrypt(msg)
	require.NoError(t, err, "failed to encrypt message")
	require.NotEqual(t, msg, encMsg)
	decMsg, err := dec.Decrypt(encMsg)
	require.NoError(t, err, "failed to decrypt message")
	require.Equal(t, msg, decMsg)
}

func createKeys(t *testing.T) (public []byte, private []byte) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	require.NoError(t, err)
	publicKey := &privateKey.PublicKey
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	require.NoError(t, err)
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	})
	return publicKeyPEM, privateKeyPEM
}
