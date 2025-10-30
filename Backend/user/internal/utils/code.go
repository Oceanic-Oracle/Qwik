package utils

import (
	"fmt"
	"math/rand"
	"time"
)

var code = rand.New(rand.NewSource(time.Now().UnixNano()))

func GenVerifyCode() string {
	code := code.Intn(1_000_000)

	return fmt.Sprintf("%06d", code)
}
