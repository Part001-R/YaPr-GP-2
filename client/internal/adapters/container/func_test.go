package container

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Проверка шифровки и расшифровки.
func TestEncryptDecrypt(t *testing.T) {

	key := [32]byte{}
	if _, err := rand.Read(key[:]); err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	srcData := []byte("Данные для ширования.")

	// Шифровка..
	encrypted, err := encrypt(srcData, key)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	// Расшифровка.
	decrData, err := decrypt(encrypted, key)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	assert.Equal(t, srcData, decrData, "нет соответствия значений")

}
