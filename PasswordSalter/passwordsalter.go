package main

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
)

const randombytes = 8

var password string
var salt string

func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func main() {
	flag.StringVar(&password, "p", "", "Password to be salted")
	flag.StringVar(&salt, "s", "", "Input Salt (generated if not provided)")
	flag.Parse()

	if password == "" {
		log.Fatal("Error: Password must not be empty")
	}

	if salt == "" {
		bytes, err := generateRandomBytes(randombytes)
		if err != nil {
			log.Fatal("Error generating random bytes", err)
		}
		dst := make([]byte, hex.EncodedLen(len(bytes)))
		hex.Encode(dst, bytes)
		salt = string(dst)
	}

	h := sha1.New()
	h.Write([]byte(salt))
	bs := h.Sum(nil)
	saltstring := hex.EncodeToString(bs)

	combined := string(saltstring) + password

	h = sha1.New()
	h.Write([]byte(combined))
	bs = h.Sum(nil)

	passwordstring := hex.EncodeToString(bs)

	fmt.Println("Hashed password:")
	fmt.Println(passwordstring)
	fmt.Println("Used salt:")
	fmt.Println(salt)
}
