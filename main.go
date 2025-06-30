package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"syscall/js"
)

// generateKeys creates an RSA key pair.
// It returns a map containing the public and private keys in PEM format.
func generateKeys(this js.Value, args []js.Value) interface{} {
	// Generate a new private key with a key size of 2048 bits.
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return map[string]interface{}{"error": "Failed to generate private key: " + err.Error()}
	}

	// Extract the public key from the private key.
	publicKey := &privateKey.PublicKey

	// Encode the private key to the PEM format.
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// Encode the public key to the PKIX, ASN.1 DER format, then to PEM.
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return map[string]interface{}{"error": "Failed to marshal public key: " + err.Error()}
	}
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	// Return the keys as a JavaScript object.
	return map[string]interface{}{
		"privateKey": string(privateKeyPEM),
		"publicKey":  string(publicKeyPEM),
	}
}

// encrypt encrypts a message using the RSA public key.
// It takes the public key (PEM format) and plaintext as string arguments.
// It returns the ciphertext as a base64 encoded string.
func encrypt(this js.Value, args []js.Value) interface{} {
	if len(args) != 2 {
		return map[string]interface{}{"error": "Invalid arguments for encrypt function"}
	}
	publicKeyPEM := args[0].String()
	plaintext := args[1].String()

	// Decode the public key from PEM format.
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return map[string]interface{}{"error": "Failed to decode PEM block containing public key"}
	}

	// Parse the public key.
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return map[string]interface{}{"error": "Failed to parse public key: " + err.Error()}
	}
	publicKey, ok := pub.(*rsa.PublicKey)
	if !ok {
		return map[string]interface{}{"error": "Not an RSA public key"}
	}

	// Encrypt the plaintext message using RSA-OAEP.
	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, []byte(plaintext), nil)
	if err != nil {
		return map[string]interface{}{"error": "Encryption failed: " + err.Error()}
	}

	// Return the ciphertext as a base64 encoded string.
	return map[string]interface{}{
		"ciphertext": base64.StdEncoding.EncodeToString(ciphertext),
	}
}

// decrypt decrypts a message using the RSA private key.
// It takes the private key (PEM format) and the base64-encoded ciphertext as string arguments.
// It returns the original plaintext.
func decrypt(this js.Value, args []js.Value) interface{} {
	if len(args) != 2 {
		return map[string]interface{}{"error": "Invalid arguments for decrypt function"}
	}
	privateKeyPEM := args[0].String()
	ciphertextB64 := args[1].String()

	// Decode the private key from PEM format.
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return map[string]interface{}{"error": "Failed to decode PEM block containing private key"}
	}

	// Parse the private key.
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return map[string]interface{}{"error": "Failed to parse private key: " + err.Error()}
	}

	// Decode the base64 ciphertext.
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return map[string]interface{}{"error": "Failed to decode ciphertext: " + err.Error()}
	}

	// Decrypt the ciphertext using RSA-OAEP.
	plaintext, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, ciphertext, nil)
	if err != nil {
		return map[string]interface{}{"error": "Decryption failed: " + err.Error()}
	}

	return map[string]interface{}{
		"plaintext": string(plaintext),
	}
}

func main() {
	// Create a channel to keep the Go program running.
	c := make(chan struct{}, 0)

	// Register the Go functions to be callable from JavaScript.
	js.Global().Set("generateKeys", js.FuncOf(generateKeys))
	js.Global().Set("encrypt", js.FuncOf(encrypt))
	js.Global().Set("decrypt", js.FuncOf(decrypt))

	// Wait forever.
	<-c
}
