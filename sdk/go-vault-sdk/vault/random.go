package vault

import (
	"crypto/rand"
	"math/big"
)

// passwordAlphabet omits 0/O/1/l/I to avoid visual confusion.
const passwordAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789"

const lowercaseAlphaNumAlphabet = "abcdefghijklmnopqrstuvwxyz0123456789"

func newRandomBytes(n int) ([]byte, error) {
	if n <= 0 {
		return nil, nil
	}
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}

func randomString(n int, alphabet string) (string, error) {
	if n <= 0 || alphabet == "" {
		return "", nil
	}
	out := make([]byte, n)
	max := big.NewInt(int64(len(alphabet)))
	for i := 0; i < n; i++ {
		idx, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		out[i] = alphabet[idx.Int64()]
	}
	return string(out), nil
}

// randomLowerAlphaNum returns n random lowercase alphanumeric chars.
func randomLowerAlphaNum(n int) (string, error) {
	return randomString(n, lowercaseAlphaNumAlphabet)
}
