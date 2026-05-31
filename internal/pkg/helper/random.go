package helper

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// GenerateRandomInteger GenerateOTP generates a random n-digit OTP
func GenerateRandomInteger(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("OTP length must be positive")
	}

	// Calculate the range: 10^(length-1) to 10^length - 1
	// For 6 digits: 100000 to 999999
	minVar := intPow(10, length-1)
	maxVar := intPow(10, length) - 1

	// Generate random number in range [0, max-min]
	rangeSize := big.NewInt(int64(maxVar - minVar + 1))
	randomNum, err := rand.Int(rand.Reader, rangeSize)
	if err != nil {
		return "", err
	}

	// Add min to get number in desired range
	otp := randomNum.Int64() + int64(minVar)

	// Format with leading zeros if needed
	return fmt.Sprintf("%0*d", length, otp), nil
}

// intPow calculates base^exp for integers
func intPow(base, exp int) int {
	result := 1
	for i := 0; i < exp; i++ {
		result *= base
	}
	return result
}

// GenerateRandomString generates a random alphanumeric string
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

// GenerateRandomStringSync generates a random alphanumeric string synchronously
// (ignoring errors, fallback to a pseudo-random or panic if absolutely needed,
// though rand.Int rarely fails)
func GenerateRandomStringSync(length int) string {
	str, _ := GenerateRandomString(length)
	return str
}
