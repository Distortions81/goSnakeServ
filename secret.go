package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

func generateSecret() []byte {
	// Get the current timestamp
	timestamp := time.Now().UTC().Unix()

	// Convert the timestamp to bytes
	timestampBytes := []byte(fmt.Sprintf("%d", timestamp))

	// Generate a random encryption key (16 bytes for AES-128)
	key := []byte("3c209f7155e3a50237bd95c9bb7d8125")

	// Create a new AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	// Create a new GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err)
	}

	// Generate a random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		panic(err)
	}

	// Encrypt the timestamp using AES-GCM
	encryptedTimestamp := gcm.Seal(nil, nonce, timestampBytes, nil)

	// Combine the nonce and encrypted timestamp
	obfuscatedTimestamp := append(nonce, encryptedTimestamp...)

	// Encode the obfuscated timestamp using base64
	encodedTimestamp := base64.StdEncoding.EncodeToString(obfuscatedTimestamp)

	return []byte(encodedTimestamp)
}

func checkSecret(input []byte) bool {

	// Decode the obfuscated timestamp from base64
	decodedTimestamp, err := base64.StdEncoding.DecodeString(string(input))
	if err != nil {
		panic(err)
	}

	inputLen := len(decodedTimestamp)
	if inputLen < 12 {
		return false
	}

	// Retrieve the nonce and encrypted timestamp
	nonceSize := 12 // GCM nonce size
	nonce := decodedTimestamp[:nonceSize]
	encryptedTimestamp := decodedTimestamp[nonceSize:]

	// Enter the secret key used for encryption
	key := []byte("3c209f7155e3a50237bd95c9bb7d8125")

	// Create a new AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	// Create a new GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err)
	}

	// Decrypt the encrypted timestamp using AES-GCM
	decryptedTimestamp, err := gcm.Open(nil, nonce, encryptedTimestamp, nil)
	if err != nil {
		panic(err)
	}

	// Convert the decrypted timestamp bytes to an integer
	timestamp, err := strconv.ParseInt(string(decryptedTimestamp), 10, 64)
	if err != nil {
		panic(err)
	}

	// Convert the timestamp to time.Time
	t := time.Unix(timestamp, 0)

	// Verify if the timestamp is within an acceptable range
	acceptableDuration := 5 * time.Minute
	currentTime := time.Now()
	diff := currentTime.Sub(t)
	if diff <= acceptableDuration || diff >= acceptableDuration {
		return true
	} else {
		return false
	}
}
