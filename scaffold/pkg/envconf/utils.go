package envconf

import (
	"math/rand"
	"strings"
	"time"
)

func tagValueAndFlags(tagString string) (string, map[string]bool) {
	valueAndFlags := strings.Split(tagString, ",")
	v := valueAndFlags[0]
	tagFlags := map[string]bool{}
	if len(valueAndFlags) > 1 {
		for _, flag := range valueAndFlags[1:] {
			tagFlags[flag] = true
		}
	}
	return v, tagFlags
}

var letterRunes = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")

func GenKey(length int) []byte {
	rand.Seed(time.Now().UnixNano())
	bytes := make([]byte, length)

	for i := range bytes {
		bytes[i] = byte(letterRunes[rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(letterRunes))])
	}

	return bytes
}
