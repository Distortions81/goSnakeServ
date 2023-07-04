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

const helloString = "UxVS3Z0|9KT36BC-J$8rb7GJfC_XBB7&?atjv$Xvnj&c7OYIUg1sZ!p?*mpPzgETEpQlM|c!YrfU8WUG0l9#Xo6@0y#rD4kY8dZDFl*qf!k$!vXA-x?sXQE*S$MjuG^UGGA=jB3%#gs!BtqvtjWD22HAYSXyowrDAgwM$_9wkQ_4aRewC80fM|5-0Fi0NVxPhoT8YG#rO&m*3wTvyk+?E?HQ9&wsgXPq7Kr9$1-&MjWTIyAG&q_r9jx=Ltp--HO#Fx8cid$1!PzVnY%Lz|w9_%cqOGt9KTD|a8qYiNkUq$+F*vkhg$eAC!Rf+P3j20ii7OJvh1#HhaiZA$G23fY3G6zx+$@xee!jD5k8IEguUpzijM02p91yc_TOXe8-2aXxMei@83H?^eZ2*QMlW7m3LgHd1LbbFSIu$!pB0QSU?sZI2qLegZqV!KF*+sG6Lug26%UKo#U1bbn7d#2Vh$B#cYeQG?TLN1CF?-qtV6Q5LZ%ab5F9%04|#R!kXs=7q#b@#tCBWn=4AIeIvN$4+-fPOu#^E=3Kez8SI1_299myHzRE0oaYX#R!1i^gfoUQw!kxqKDd4$8FyqRkt8baBrj!eV+5@+H3^RG6buWpbimsxbylMy?%Z|d2p40iaXzLEn7gLR5uMMtklBY#7qDap@qPY76*q6!9hzgV-6xBeaA4=w|aM@39JaRmp5FfoUl#ii%|M$4+Ob9zAX&es!leNDKqun+b6iI=X4Pf+yeoI&qM1IahNufncWf6K^*2p^|WKq-qRs?qI0P+bIeU^Y!ZwIH9O1|GRmhEFZQpvn45!^c1PiJb^oyYheGVYr5|=*9_Q4pQ0P&A+_I&&k8_9!iI3h-AS!ESvTO4ryGqzgKPEd4rFg6wzJcVgQetZyWB7Oq-R=1&e|oUTRHJM8t1txp&0ZD2dL6p85#YTExXibbOnm^RB#RcMU?_e3g-A*iRd@YcrjRnBDtjXE@X&e+6|Z$HmvNGmWmSQc^=$7Eh-wCkj-0^8xZ01ZsfkW$?mRqi&lOKQ9^jsa@0&y2-rAU36$Z+U40&^wS!icMzb+k_EhFa$Ej5Txr+j==1y1L3?!Esv6-jQMCBv=9b=ncBLV&f=p@hPBQrMjRcc-ImNX+oNdSY#$KL3WvWr7CpTB-Lbv|ac%UXsUHxM*LcYfcyOkqvl-txJkV^yvF?Hs!%v?lEG1OS-W*GHxCS^8k#g03=jed8yrmm5vmxAN@w9!!euQE-39t^Z0B-7|que&Dm&bhS5pGpwzKycbRu-3N42j@+&vHRHPuiKRwd$X#CIHw0E#yAWvr?ih^nkXr9C_ikQeT0mHkVHlvmuhAdq@npqZ4OFYv=W!PX+epQgb$xRBwxoxwF$KhfR4OUUDru9mMjv9k%jnvFIX$?x*jse1Ps^Vto-n6k!g_BbHT_v@bCn^2Rkp7hEgtMW8KV&Qr_o5H38&fGbV$ifE!i58aRY++uF2Tpbkx^ctCx&RkvJzfR$kl5G&v1V*$@PH*ZK1Jw7NsaYUZE=Il5$Ef^pmu4Ie!BRmO$%BrQYLsdZ69y1Osuu6L0JLpgIHhyz!r@#qseKT1TV80flt*%q*77HKuL=dUh-VsxPe@F|GPmdNnH22t0s0z&DbH#ch?4xgQyuV4AJM%@g|s3vKo74K2I7nYw7n@HIF#7!3a*q2_bcN7p-4Mb_1R+8PycZ9t3oDKWR4dJt^U#Eprv8ko0jqdrBGD&$Hyhq&WE=DZryCAsJ4LUI1m=lzvEZ2%d-sz@vA#ox0RW_j8zAVdu=G?kdp*!+*v7oSjanS=5!Tl1yjVavCJ|qXjax61CCr2?aQb7&*c=jb3ZLPTvHJ-hEzN-xMNjO%d^1JlLskzz4A4FUc*zhel?up0fuwm?Vnb=+gZdA-RuT%?WM=4A+X_RMfy@_U@%fLzeYJegg^8ZqUci2DlqJYTHwe!GNt8F*qXdCD+cjWCgTPd@P9!i|4&0npTwD1coZ#C3&edIRqewAXyUSezEeb%z+YV!HzaOV?%b5oGJSPbcB*HY_Ac#!St42ohNQK#JMV&9O2eHB?6gmJS_Kgb1|8=U"

func generateSecret(player *playerData) []byte {
	// Get the current timestamp
	timestamp := time.Now().UTC().Unix()

	// Convert the timestamp to bytes
	var ID string = helloString
	if player != nil && player.ID != 0 {
		ID = base64.RawStdEncoding.EncodeToString([]byte(strconv.FormatUint(player.ID, 36)))
	}
	timestampBytes := []byte(fmt.Sprintf("%d,%d", timestamp, ID))

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

func checkSecret(player *playerData, input []byte) bool {

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
