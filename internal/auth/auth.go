package auth

import "github.com/alexedwards/argon2id"

func HashPassword(password string) (string, error) {
	argon2idParams := &argon2id.Params{
		Memory:      64 * 1024, // 64 MB
		Iterations:  3,
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	}
	return argon2id.CreateHash(password, argon2idParams)
}

func CheckPasswordHash(password, hash string) (bool, error) {
	return argon2id.ComparePasswordAndHash(password, hash)
}
