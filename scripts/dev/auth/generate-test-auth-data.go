package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/vedhavyas/go-subkey/v2"
	"github.com/vedhavyas/go-subkey/v2/ed25519"
	"github.com/vedhavyas/go-subkey/v2/sr25519"
)

const (
	domain = "kyc.qa.grid.tf"
)

// Generate test auth data for development use
func main() {
	message := createSignMessage()
	// if no arg provided, generate random keys
	krSr25519, krEd25519, err := loadKeys()
	if err != nil {
		panic(err)
	}
	msg := []byte(message)
	sigSr25519, err := krSr25519.Sign(msg)
	if err != nil {
		panic(err)
	}
	sigEd25519, err := krEd25519.Sign(msg)
	if err != nil {
		panic(err)
	}
	messageString := hex.EncodeToString([]byte(message))
	fmt.Println("______________________")
	fmt.Println("Auth Data")
	fmt.Println("______________________")
	fmt.Println("** SR25519 **")
	//fmt.Println("Public key sr25519: ", hex.EncodeToString(krSr25519.Public()))
	fmt.Println("SS58Address sr25519: ", krSr25519.SS58Address(42))
	fmt.Println("Challenge hex: ", hex.EncodeToString([]byte(message)))
	fmt.Println("Signature sr25519: ", hex.EncodeToString(sigSr25519))
	fmt.Println("______________________")
	fmt.Println("** ED25519 **")
	// fmt.Println("Public key ed25519: ", hex.EncodeToString(krEd25519.Public()))
	fmt.Println("SS58Address ed25519: ", krEd25519.SS58Address(42))
	fmt.Println("Challenge hex: ", hex.EncodeToString([]byte(message)))
	fmt.Println("Signature ed25519: ", hex.EncodeToString(sigEd25519))
	fmt.Println("______________________")
	bytes, err := hex.DecodeString(messageString)
	if err != nil {
		panic(err)
	}
	fmt.Println("challenge string (plain text): ", string(bytes))
}

func createSignMessage() string {
	// return a message with the domain and the current timestamp in hex
	message := fmt.Sprintf("%s:%d", domain, time.Now().Unix())
	fmt.Println("message: ", message)
	return message
}

func loadKeys() (subkey.KeyPair, subkey.KeyPair, error) {
	if len(os.Args) < 2 {
		krSr25519, err := sr25519.Scheme{}.Generate()
		if err != nil {
			return nil, nil, err
		}
		krEd25519, err := ed25519.Scheme{}.Generate()
		if err != nil {
			return nil, nil, err
		}
		return krSr25519, krEd25519, nil
	} else {
		krSr25519, err := sr25519.Scheme{}.FromPhrase(os.Args[1], "")
		if err != nil {
			return nil, nil, err
		}
		krEd25519, err := ed25519.Scheme{}.FromPhrase(os.Args[1], "")
		if err != nil {
			return nil, nil, err
		}
		return krSr25519, krEd25519, nil
	}
}
