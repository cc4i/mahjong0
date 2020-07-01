package utils

import (
	"math/rand"
	"time"
)

// Contains return true if slice[] include item
func Contains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}

	_, ok := set[item]
	return ok
}

// ContainsArray return true if slice[] include items[]
func ContainsArray(slice []string, items []string) bool {
	ret := 0
	for _, item := range items {
		if Contains(slice, item) {
			ret++
		}
	}
	if ret != 0 && len(items) == ret {
		return true
	}
	return false
}

// RandString return random string as per length 'n'
func RandString(n int) string {
	const letterBytes = "1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	if n < 0 {
		n = 0
	}
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.NewSource(time.Now().UnixNano()).Int63()%int64(len(letterBytes))]
	}
	return string(b)
}
