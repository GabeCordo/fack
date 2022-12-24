package fack

import (
	"bytes"
	"math/rand"
	"net/http"
	"time"
)

// integer constants
const (
	maxGeneratedStringLength = 32
	lowerASCIIBound          = 97
	upperASCIIBound          = 122
)

// string constants
const (
	EmptyString = ""
	StringSpace = " "
)

func RandInteger(min int, max int) int {
	rand.Seed(time.Now().UTC().UnixNano())
	return min + rand.Intn(max-min)
}

func RandInteger64(min int64, max int64) int64 {
	rand.Seed(time.Now().UTC().UnixNano())
	return min + rand.Int63n(max-min)
}

func GenerateRandomString(seed int) string {
	buffer := new(bytes.Buffer)
	for i := 0; i < maxGeneratedStringLength; i++ {
		char := RandInteger(lowerASCIIBound, upperASCIIBound)
		buffer.WriteString(string(char))
	}
	return buffer.String()
}

func GenerateNonce() int64 {
	return time.Now().Unix() * RandInteger64(4, 9)
}

func GetInternetProtocol(r *http.Request) (*Address, error) {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return NewAddress(forwarded)
	}
	return NewAddress(r.RemoteAddr)
}

func IsUsingJSONContent(r *http.Request) bool {
	content := r.Header.Get("Content-Type")
	return content == "application/json"
}
