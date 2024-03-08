package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"time"
)

const (
	crypt        = "x5XGXt8cvJo+v1fcJxh+EeIEWQc8hncABp3e9z0jIPA="
	hello        = 2210441626 + protoVersion
	timeStampMod = 1333
	nonceSize    = 12
)

/* We reuse these for speed */
var (
	key        []byte
	helloBytes []byte
	gcm        cipher.AEAD
	block      cipher.Block
	nonce      []byte
)

func init() {

	var err error
	/* Encryption key*/
	key, _ = base64.StdEncoding.DecodeString(crypt)

	/* Create a new AES cipher block*/
	block, err = aes.NewCipher(key)
	if err != nil {
		return
	}

	/* Create a new GCM mode*/
	gcm, err = cipher.NewGCM(block)
	if err != nil {
		return
	}

	helloBytes = uint32ToByteArray(hello)

	newNonce()
}

func newNonce() {
	/* Generate a random nonce*/
	nonce = make([]byte, gcm.NonceSize())
	_, err := rand.Read(nonce)
	if err != nil {
		return
	}
}

func generateSecret(player *playerData) []byte {
	/* Get the current timestamp*/
	timestamp := uint32(time.Now().UTC().Unix() + ((timeStampMod) * protoVersion))

	/* Convert the timestamp to bytes*/
	var ID []byte
	if player != nil && player.id != 0 {
		ID = uint32ToByteArray(player.id)
	} else {
		ID = helloBytes
	}

	payload := append(uint32ToByteArray(timestamp), ID...)

	newNonce()

	/* Encrypt the timestamp using AES-GCM*/
	encryptedData := gcm.Seal(nil, nonce, payload, nil)

	/* Combine the nonce and encrypted timestamp*/
	obfuscatedTimestamp := append(nonce, encryptedData...)

	return []byte(obfuscatedTimestamp)
}

func checkSecret(player *playerData, input []byte) bool {

	inputLen := len(input)
	if inputLen < nonceSize {
		doLog(true, "input too short")
		return false
	}

	/* Retrieve the nonce and encrypted timestamp*/
	nonceSize := nonceSize /* GCM nonce size*/
	nonce := input[:nonceSize]
	encryptedTimestamp := input[nonceSize:]

	/* Decrypt the encrypted timestamp using AES-GCM*/
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

	if player == nil || (player != nil && player.id == 0) {
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

	/* Convert the decrypted timestamp bytes to an integer*/

	/* Convert the timestamp to time.Time*/
	t := time.Unix(int64(decodedTimestamp)-((timeStampMod)*protoVersion), 0)

	/* Verify if the timestamp is within an acceptable range*/
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
