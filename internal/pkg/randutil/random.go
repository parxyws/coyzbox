package randutil

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func GenerateRandomInteger(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("length must be positive")
	}

	minVar := int64(intPow(10, length-1))
	maxVar := int64(intPow(10, length) - 1)

	rangeSize := big.NewInt(maxVar - minVar + 1)
	randomNum, err := rand.Int(rand.Reader, rangeSize)
	if err != nil {
		return "", err
	}

	otp := randomNum.Int64() + minVar
	return fmt.Sprintf("%0*d", length, otp), nil
}

func GenerateRandomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		b[i] = charset[num.Int64()]
	}
	return string(b), nil
}

func intPow(base, exp int) int {
	result := 1
	for i := 0; i < exp; i++ {
		result *= base
	}
	return result
}
