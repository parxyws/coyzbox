package helper

import (
	"github.com/parxyws/cozybox/internal/pkg/randutil"
)

func GenerateRandomInteger(length int) (string, error) {
	return randutil.GenerateRandomInteger(length)
}

func GenerateRandomString(length int) (string, error) {
	return randutil.GenerateRandomString(length)
}

func GenerateRandomStringSync(length int) string {
	str, _ := randutil.GenerateRandomString(length)
	return str
}
