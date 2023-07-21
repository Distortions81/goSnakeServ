package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const helloStr = "w15r7Ju25b8C8OcDaNUnc6mT7AVn6TnSfBPYiV8cz0="
const crypt = "x5XGXt8cvJo+v1fcJxh+EeIEWQc8hncABp3e9z0jIPA="

var key []byte
var hello []byte

func init() {
	// Encryption key
	key, _ = base64.StdEncoding.DecodeString(crypt)
	hello, _ = base64.StdEncoding.DecodeString(helloStr)
}

func generateSecret(player *playerData) []byte {
	// Get the current timestamp
	timestamp := time.Now().UTC().Unix()

	// Convert the timestamp to bytes

	ID := hello
	if player != nil && player.id != 0 {
		ID = []byte(strconv.FormatUint(uint64(player.id), 18))
	}

	timeStampString := []byte(strconv.FormatInt(timestamp, 27))
	payload := []byte(fmt.Sprintf("%v,%v", string(timeStampString), string(ID)))

	// Create a new AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil
	}

	// Create a new GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil
	}

	// Generate a random nonce
	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		return nil
	}

	// Encrypt the timestamp using AES-GCM
	encryptedData := gcm.Seal(nil, nonce, payload, nil)

	// Combine the nonce and encrypted timestamp
	obfuscatedTimestamp := append(nonce, encryptedData...)

	return []byte(obfuscatedTimestamp)
}

func checkSecret(player *playerData, input []byte) bool {

	inputLen := len(input)
	if inputLen < 12 {
		doLog(true, "input too short")
		return false
	}

	// Retrieve the nonce and encrypted timestamp
	nonceSize := 12 // GCM nonce size
	nonce := input[:nonceSize]
	encryptedTimestamp := input[nonceSize:]

	// Create a new AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		doLog(true, "new cipher failed")
		return false
	}

	// Create a new GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		doLog(true, "new gcm failed")
		return false
	}

	// Decrypt the encrypted timestamp using AES-GCM
	decryptedData, err := gcm.Open(nil, nonce, encryptedTimestamp, nil)
	if err != nil {
		doLog(true, "gcm open failed")
		return false
	}

	parts := strings.Split(string(decryptedData), ",")
	if len(parts) != 2 {
		doLog(true, "split failed")
		return false
	}
	decodedTimestamp := []byte(parts[0])
	playerID := parts[1]

	if player == nil {
		if playerID != string(hello) {
			doLog(true, "hello decode failed")
			return false
		}
	} else {
		ID := strconv.FormatUint(uint64(player.id), 18)
		if playerID != ID {
			doLog(true, "playerid incorrect")
			//return false
		}
	}

	// Convert the decrypted timestamp bytes to an integer
	timestamp, err := strconv.ParseInt(string(decodedTimestamp), 27, 64)
	if err != nil {
		doLog(true, "timestamp parse failed")
		return false
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
		doLog(true, "secret expired")
		return false
	}
}
