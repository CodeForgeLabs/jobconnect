package hasher

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2Hasher implements domain.PasswordHasher with Argon2id.
type Argon2Hasher struct {
	time    uint32
	memory  uint32
	threads uint8
	keyLen  uint32
	saltLen int
}

// NewArgon2Hasher returns a hasher with sensible defaults.
func NewArgon2Hasher() *Argon2Hasher {
	return &Argon2Hasher{
		time:    1,
		memory:  64 * 1024, // 64 MiB
		threads: 2,
		keyLen:  32,
		saltLen: 16,
	}
}

// Hash hashes password with a random salt (format: argon2id$params$salt$hash).
func (h *Argon2Hasher) Hash(password string) (string, error) {
	salt := make([]byte, h.saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key := argon2.IDKey([]byte(password), salt, h.time, h.memory, h.threads, h.keyLen)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Key := base64.RawStdEncoding.EncodeToString(key)
	return fmt.Sprintf("argon2id$%d$%d$%d$%s$%s", h.time, h.memory, h.threads, b64Salt, b64Key), nil
}

// Verify checks password against stored hash.
func (h *Argon2Hasher) Verify(password, storedHash string) (bool, error) {
	parts := strings.Split(storedHash, "$")
	if len(parts) != 6 || parts[0] != "argon2id" {
		return false, errors.New("invalid hash format")
	}
	var time, memory uint32
	var threads uint8
	_, err := fmt.Sscanf(parts[1], "%d", &time)
	if err != nil {
		return false, err
	}
	_, err = fmt.Sscanf(parts[2], "%d", &memory)
	if err != nil {
		return false, err
	}
	_, err = fmt.Sscanf(parts[3], "%d", &threads)
	if err != nil {
		return false, err
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}
	expectedKey, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}
	key := argon2.IDKey([]byte(password), salt, time, memory, threads, uint32(len(expectedKey)))
	return subtle.ConstantTimeCompare(key, expectedKey) == 1, nil
}
