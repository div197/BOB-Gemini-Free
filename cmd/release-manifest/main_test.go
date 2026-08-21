package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"testing"
)

func TestVerifyPublicKeyMatch(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	encoded := base64.StdEncoding.EncodeToString(publicKey)
	if err := verifyPublicKeyMatch(privateKey, encoded); err != nil {
		t.Fatalf("matching key rejected: %v", err)
	}

	otherPublic, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate other key: %v", err)
	}
	if err := verifyPublicKeyMatch(privateKey, base64.StdEncoding.EncodeToString(otherPublic)); err == nil {
		t.Fatal("mismatched public key was accepted")
	}
}

func TestDecodePublicKeyAcceptsHexadecimal(t *testing.T) {
	publicKey, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	const hexChars = "0123456789abcdef"
	hexEncoded := make([]byte, len(publicKey)*2)
	for i, value := range publicKey {
		hexEncoded[i*2] = hexChars[value>>4]
		hexEncoded[i*2+1] = hexChars[value&0x0f]
	}
	decoded, err := decodePublicKey(string(hexEncoded))
	if err != nil {
		t.Fatalf("decode hexadecimal public key: %v", err)
	}
	if string(decoded) != string(publicKey) {
		t.Fatal("decoded hexadecimal public key differs")
	}
}
