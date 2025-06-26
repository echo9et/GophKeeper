package gcrypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	"GophKeeper.ru/internal/entities"
)

func EnecryptRecord(record *entities.Record) error {
	return nil
}

func DecryptRecord(record *entities.Record) error {
	return nil
}

// deriveNonce генерирует nonce на основе ключа и контекста
func deriveNonce(key, context string) []byte {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(context))
	sum := h.Sum(nil)
	return sum[:16]
}

// Encrypt шифруем данные
func Encrypt(data, key, context string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	nonce := deriveNonce(key, context)
	ciphertext := make([]byte, len(data))

	stream := cipher.NewCTR(block, nonce)
	stream.XORKeyStream(ciphertext, []byte(data))

	return hex.EncodeToString(ciphertext), nil
}

// Decrypt дешифруем данные
func Decrypt(cipherTextHex, key, context string) (string, error) {
	ciphertext, err := hex.DecodeString(cipherTextHex)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	nonce := deriveNonce(key, context)
	stream := cipher.NewCTR(block, nonce)

	plaintext := make([]byte, len(ciphertext))
	stream.XORKeyStream(plaintext, ciphertext)

	return string(plaintext), nil
}
