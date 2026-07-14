// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cryptoutil

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

type AESGCMCipher struct {
	block cipher.Block
}

func (c AESGCMCipher) Encrypt(b []byte) []byte {
	// gcm or Galois/Counter Mode, is a mode of operation
	// for symmetric key cryptographic block ciphers
	// - https://en.wikipedia.org/wiki/Galois/Counter_Mode
	gcm, err := cipher.NewGCM(c.block)
	// if any error generating new GCM
	// handle them
	if err != nil {
		panic(err)
	}
	// creates a new byte array the size of the nonce
	// which must be passed to Seal
	nonce := make([]byte, gcm.NonceSize())
	// populates our nonce with a cryptographically secure
	// random sequence
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err)
	}
	// here we encrypt our text using the Seal function
	// Seal encrypts and authenticates plaintext, authenticates the
	// additional data and appends the result to dst, returning the updated
	// slice. The nonce must be NonceSize() bytes long and unique for all
	// time, for a given key.
	return gcm.Seal(nonce, nonce, b, nil)
}

func (c AESGCMCipher) Decrypt(b []byte) []byte {
	gcm, err := cipher.NewGCM(c.block)
	if err != nil {
		panic(err)
	}
	nonceSize := gcm.NonceSize()
	if len(b) < nonceSize {
		panic(err)
	}
	nonce, ciphertext := b[:nonceSize], b[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err)
	}
	return plaintext
}

func NewAESGCMCipher(key []byte) (*AESGCMCipher, error) {
	cipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return &AESGCMCipher{block: cipher}, nil
}
