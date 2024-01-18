package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	b64 "encoding/base64"
	"io"
)

func EncryptMessage(key []byte, message string) string {
	byteMsg := []byte(message)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "srv_error1_enc"
	}

	cipherText := make([]byte, aes.BlockSize+len(byteMsg))
	iv := cipherText[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return "srv_error2_enc"
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], byteMsg)

	return base64.StdEncoding.EncodeToString(cipherText)
}

func DecryptMessage(key []byte, message string) string {
	cipherText, err := base64.StdEncoding.DecodeString(message)
	if err != nil {
		return "srv_error1_dec"
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "srv_error2_dec"
	}

	if len(cipherText) < aes.BlockSize {
		return "srv_error2_dec"
	}

	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherText, cipherText)

	return string(cipherText)
}

func ToBase64(key []byte, plaintext string) string {
	return b64.StdEncoding.EncodeToString([]byte(plaintext))
}

func FromBase64(key []byte, ct string) string {
	b, _ := b64.StdEncoding.DecodeString(ct)
	return string(b)
}
