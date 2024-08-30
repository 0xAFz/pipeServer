package utils

import (
	"math/rand"
	"strings"
	"time"
)

func init() {
	rand.NewSource(time.Now().UnixNano())
}

const (
	letters = "abcdefghigklnmopqrsyuvwxyz"
)

func GenerateRandomPrivateID() string {
	var privateID strings.Builder

	for i := 0; i < 6; i++ {
		privateID.WriteByte(letters[rand.Intn(len(letters))])
	}

	return privateID.String()
}
