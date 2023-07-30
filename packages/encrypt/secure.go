package encrypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// This key is used to gen
var SecretKey string = "Jdh1e4pTjbjFxgP5+V6E93aA+W5MCxOgLDi5ZS8dhL4="

// EncryptAES cifra la stringa data utilizzando l'algoritmo AES con la chiave fornita.
func EncryptAES(key, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Aggiungiamo padding per rendere la lunghezza dei dati multipla della dimensione del blocco AES.
	paddedData := addPadding(data, block.BlockSize())

	// Generiamo un vettore di inizializzazione casuale.
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	// Cifriamo i dati.
	mode := cipher.NewCBCEncrypter(block, iv)
	encryptedData := make([]byte, len(paddedData))
	mode.CryptBlocks(encryptedData, paddedData)

	// Concateniamo il vettore di inizializzazione ai dati cifrati.
	finalData := append(iv, encryptedData...)
	return finalData, nil
}

// DecryptAES decifra i dati cifrati forniti utilizzando l'algoritmo AES con la chiave data.
func DecryptAES(key, encryptedData []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Verifichiamo che i dati cifrati siano sufficientemente lunghi.
	if len(encryptedData) < aes.BlockSize {
		return nil, errors.New("dati cifrati non validi")
	}

	// Estraiamo il vettore di inizializzazione dai dati cifrati.
	iv := encryptedData[:aes.BlockSize]
	encryptedData = encryptedData[aes.BlockSize:]

	// Decifriamo i dati.
	mode := cipher.NewCBCDecrypter(block, iv)
	decryptedData := make([]byte, len(encryptedData))
	mode.CryptBlocks(decryptedData, encryptedData)

	// Rimuoviamo il padding per ottenere i dati originali.
	decryptedData, err = removePadding(decryptedData)
	if err != nil {
		return nil, err
	}

	return decryptedData, nil
}

// Funzione di padding.
func addPadding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// Funzione di rimozione del padding.
func removePadding(data []byte) ([]byte, error) {
	length := len(data)
	unpadding := int(data[length-1])
	if unpadding > length {
		return nil, errors.New("padding non valido")
	}
	return data[:(length - unpadding)], nil
}

func generateRandomKey(keyLength int) (string, error) {
	key := make([]byte, keyLength)
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}

	// Converti il byte slice in una stringa base64
	encodedKey := base64.StdEncoding.EncodeToString(key)
	return encodedKey, nil
}

func genKey() string {
	keyLength := 32 // Usa 16, 24 o 32 per AES-128, AES-192, o AES-256
	secretKey, err := generateRandomKey(keyLength)
	if err != nil {
		panic(err)
	}

	return secretKey
}
