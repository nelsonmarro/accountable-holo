package main

import (
	"encoding/base64"
	"fmt"
)

func main() {
	password := "DarthMaul9199*"
	key := "fc0f9971-3a7e-4610-b20e-f3789b014f89"

	// XOR
	xorBytes := make([]byte, len(password))
	for i := 0; i < len(password); i++ {
		xorBytes[i] = password[i] ^ key[i%len(key)]
	}

	// Base64 encode
	fmt.Println(base64.StdEncoding.EncodeToString(xorBytes))
}
