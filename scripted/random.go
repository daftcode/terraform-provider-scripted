package scripted

import (
	"math/rand"
	"time"
)

const alphaCharset = "abcdefghijklmnopqrstuvwxyz" + "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const alnumCharset = alphaCharset + "0123456789"

const charset = alnumCharset + "`~!@#$%^&*()_+[]{};':,./<>?\\|\""

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
	return RandomStringWithCharset(1, alphaCharset) + RandomStringWithCharset(length-1, alnumCharset)
}

func RandomString(length int) string {
	return RandomStringWithCharset(length, charset)
}
