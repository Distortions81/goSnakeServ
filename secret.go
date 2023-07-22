package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"time"
)

const hello = 2210441626
const crypt = "x5XGXt8cvJo+v1fcJxh+EeIEWQc8hncABp3e9z0jIPA="

var key []byte

func init() {
	// Encryption key
	key, _ = base64.StdEncoding.DecodeString(crypt)
}

func generateSecret(player *playerData) []byte {
	// Get the current timestamp
	timestamp := time.Now().UTC().Unix()

	// Convert the timestamp to bytes
	var ID []byte = uint32ToByteArray(hello)
	if player != nil && player.id != 0 {
		ID = uint32ToByteArray(player.id)
	}

	var payload []byte = uint32ToByteArray(uint32(timestamp))
	payload = append(payload, ID...)

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

	if len(decryptedData) < 8 {
		doLog(true, "Decrypted data is invalid.")
		return false
	}
	decodedTimestamp := byteArrayToUint32(decryptedData[0:])
	playerID := byteArrayToUint32(decryptedData[4:])

	if player == nil {
		if playerID != hello {
			doLog(true, "hello decode failed")
			return false
		}
	} else {
		if playerID != player.id {
			doLog(true, "playerid incorrect")
			return false
		}
	}

	// Convert the decrypted timestamp bytes to an integer

	// Convert the timestamp to time.Time
	t := time.Unix(int64(decodedTimestamp), 0)

	// Verify if the timestamp is within an acceptable range
	acceptableDuration := 10 * time.Second
	currentTime := time.Now()
	diff := currentTime.Sub(t)
	if diff <= acceptableDuration || diff >= acceptableDuration {
		return true
	} else {
		doLog(true, "secret expired")
		return false
	}
}
