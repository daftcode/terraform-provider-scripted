package scripted

import (
	"math/rand"
	"time"
)

const safeCharset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
	"0123456789"

const charset = safeCharset + "`~!@#$%^&*()_+[]{};':,./<>?\\|\""

var seededRand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func RandomStringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func RandomSafeString(length int) string {
	return RandomStringWithCharset(length, safeCharset)
}

func RandomString(length int) string {
	return RandomStringWithCharset(length, charset)
}
