package test

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

// StringsToInterfaceSlice converts strings to a slice of interfaces containing the strings.
func StringsToInterfaceSlice(strings ...string) []interface{} {
	args := make([]interface{}, len(strings))
	for i, str := range strings {
		args[i] = str
	}
	return args
}

// IntsToInterfaceSlice converts ints to a slice of interfaces containing the ints.
func IntsToInterfaceSlice(ints ...int) []interface{} {
	args := make([]interface{}, len(ints))
	for i, num := range ints {
		args[i] = num
	}
	return args
}

// RandomKey returns a key of the form test:<number>, where <number> is a random number. It is used for
// testing Redis data types using random keys.
func RandomKey() string {
	return fmt.Sprint("testkey" + strconv.Itoa(rand.Int()))
}
