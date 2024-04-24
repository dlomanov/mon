package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
)

type (
	Encryptor struct {
		publicKey *rsa.PublicKey
	}
	Decryptor struct {
		privateKey *rsa.PrivateKey
	}
)

func NewEncryptor(publicKeyPEM []byte) (*Encryptor, error) {
	publicKeyBlock, _ := pem.Decode(publicKeyPEM)
	publicKey, err := x509.ParsePKIXPublicKey(publicKeyBlock.Bytes)
	if err != nil {
		return nil, err
	}
	return &Encryptor{
		publicKey: publicKey.(*rsa.PublicKey),
	}, nil
}

func (enc *Encryptor) Encrypt(input []byte) ([]byte, error) {
	aesKey := make([]byte, 32) // AES 256
	if _, err := rand.Read(aesKey); err != nil {
		return nil, err
	}
	encAesKey, err := rsa.EncryptPKCS1v15(rand.Reader, enc.publicKey, aesKey)
	if err != nil {
		return nil, err
	}
	output, err := encryptAES(input, aesKey)
	if err != nil {
		return nil, err
	}
	return append(encAesKey, output...), nil
}

func encryptAES(input []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	output := make([]byte, len(input))
	stream := cipher.NewCTR(block, key[:block.BlockSize()])
	stream.XORKeyStream(output, input)
	return output, nil
}

func NewDecryptor(privateKeyPEM []byte) (*Decryptor, error) {
	privateKeyBlock, _ := pem.Decode(privateKeyPEM)
	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyBlock.Bytes)
	if err != nil {
		return nil, err
	}
	return &Decryptor{
		privateKey: privateKey,
	}, nil
}

func (dec *Decryptor) Decrypt(input []byte) ([]byte, error) {
	size := dec.privateKey.Size()
	encAesKey := input[:size]
	encPayload := input[size:]
	aesKey, err := rsa.DecryptPKCS1v15(rand.Reader, dec.privateKey, encAesKey)
	if err != nil {
		return nil, err
	}

	return decryptAES(encPayload, aesKey)
}

func decryptAES(input []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	output := make([]byte, len(input))
	stream := cipher.NewCTR(block, key[:block.BlockSize()])
	stream.XORKeyStream(output, input)
	return output, nil
}
