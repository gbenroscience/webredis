package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

const ModeCFB = 0
const ModeCBC = 1

// Kryptik When encrypting using this api, make sure your encryption and decryption modes always match!
// So if you encrypted using ModeCFB, alo use ModeCFB when decrypting
type Kryptik struct {
	Key  string
	Mode int //CBC or CFB
}

// NewKryptik creates a new Kryptik pointer. If an invalid mode is specified, returns a nil pointer
func NewKryptik(key string, mode int) (*Kryptik, error) {
	if mode != ModeCFB && mode != ModeCBC {
		return nil, errors.New("invalid AES mode specified... specify mode=0 for CFB and mode=1 for CBC")
	}
	return &Kryptik{
		Key:  key,
		Mode: mode,
	}, nil
}

// DefaultKryptik creates a new Kryptik pointer with CBC mode enabled
func DefaultKryptik(key string, ivText string) *Kryptik {
	return &Kryptik{
		Key:  key,
		Mode: ModeCBC,
	}
}

func (k Kryptik) encryptCFB(message string) (encmess string, err error) {
	plainText := []byte(message)

	block, err := aes.NewCipher([]byte(k.Key))
	if err != nil {
		return
	}
	if len(k.Key) != 32 {
		return "", errors.New("the key must be 32 bytes long")
	}

	//IV needs to be unique, but doesn't have to be secure.
	//It's common to put it at the beginning of the ciphertext.
	cipherText := make([]byte, aes.BlockSize+len(plainText))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], plainText)

	//returns to base64 encoded string
	encmess = base64.RawURLEncoding.EncodeToString(cipherText)
	return
}

func (k Kryptik) decryptCFB(encrypted string) (decrypted string, err error) {
	cipherText, err := base64.RawURLEncoding.DecodeString(encrypted)
	if err != nil {
		return
	}

	block, err := aes.NewCipher([]byte(k.Key))
	if err != nil {
		return
	}

	if len(cipherText) < aes.BlockSize {
		err = errors.New("the block size of the ciphertext is too short")
		return
	}

	//IV needs to be unique, but doesn't have to be secure.
	//It's common to put it at the beginning of the ciphertext.
	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(cipherText, cipherText)

	decrypted = string(cipherText)
	return
}

func (k Kryptik) encryptCBC(message string) (encmess string, err error) {
	plainText := []byte(message)

	plainTextWithPadding := pkcs5Padding(plainText, aes.BlockSize)

	block, err := aes.NewCipher([]byte(k.Key))
	if err != nil {
		return
	}
	if len(k.Key) != 32 {
		return "", errors.New("the key must be 32 bytes long")
	}

	//IV needs to be unique, but doesn't have to be secure.
	//It's common to put it at the beginning of the ciphertext.
	cipherText := make([]byte, aes.BlockSize+len(plainTextWithPadding))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCBCEncrypter(block, iv)
	stream.CryptBlocks(cipherText[aes.BlockSize:], plainTextWithPadding)

	//returns to base64 encoded string
	encmess = base64.RawURLEncoding.EncodeToString(cipherText)
	return
}

func (k Kryptik) decryptCBC(encrypted string) (decrypted string, err error) {
	cipherText, err := base64.RawURLEncoding.DecodeString(encrypted)
	if err != nil {
		return
	}

	block, err := aes.NewCipher([]byte(k.Key))
	if err != nil {
		return
	}

	if len(cipherText) < aes.BlockSize {
		err = errors.New("the block size of the ciphertext is too short")
		return
	}

	//IV needs to be unique, but doesn't have to be secure.
	//It's common to put it at the beginning of the ciphertext.
	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCBCDecrypter(block, iv)
	// XORKeyStream can work in-place if the two arguments are the same.
	stream.CryptBlocks(cipherText, cipherText)

	decrypted = string(pkcs5UnPadding(cipherText))
	return
}

func (k Kryptik) Encrypt(input string) (decrypted string, err error) {
	if k.Mode == ModeCFB {
		return k.encryptCFB(input)
	} else if k.Mode == ModeCBC {
		return k.encryptCBC(input)
	}
	return "", errors.New("invalid encryption mode specified")
}
func (k Kryptik) Decrypt(encrypted string) (decrypted string, err error) {
	if k.Mode == ModeCFB {
		return k.decryptCFB(encrypted)
	} else if k.Mode == ModeCBC {
		return k.decryptCBC(encrypted)
	}
	return "", errors.New("invalid decryption mode specified")
}

func pkcs5Padding(ciphertext []byte, blockSize int) []byte {
	padding := (blockSize - len(ciphertext)%blockSize)
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}
func pkcs5UnPadding(encrypt []byte) []byte {
	padding := encrypt[len(encrypt)-1]
	return encrypt[:len(encrypt)-int(padding)]
}
