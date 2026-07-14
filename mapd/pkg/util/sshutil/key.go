// Copyright (c) 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package sshutil

import (
	"crypto"
	"crypto/dsa"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"golang.org/x/crypto/ssh"
)

// LoadPrivateKey decodes unencrypted private key from PEM bytes.
func LoadPrivateKey(p []byte) (crypto.PrivateKey, error) {
	pk, err := ssh.ParseRawPrivateKey(p)
	if err != nil {
		return nil, err
	}
	if _, ok := pk.(*dsa.PrivateKey); ok {
		return nil, fmt.Errorf("DSA private key is not supported")
	}
	return pk, nil
}

// GeneratePrivateKey creates a RSA private Key of specified byte size.
func GeneratePrivateKey(bitSize int) (*rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return nil, err
	}

	err = privateKey.Validate()
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

// EncodePrivateKey encodes private Key to PEM format without encryption.
// It does not support OpenSSL encoding.
func EncodePrivateKey(privateKey crypto.PrivateKey) ([]byte, error) {
	var err error
	block := &pem.Block{Headers: nil}
	switch key := privateKey.(type) {
	case *rsa.PrivateKey:
		block.Type = "RSA PRIVATE KEY"
		block.Bytes = x509.MarshalPKCS1PrivateKey(key)
	case *ecdsa.PrivateKey:
		block.Type = "EC PRIVATE KEY"
		block.Bytes, err = x509.MarshalECPrivateKey(key)
	default:
		block.Type = "PRIVATE KEY"
		block.Bytes, err = x509.MarshalPKCS8PrivateKey(key)
	}
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(block), nil
}

// ExportPublicKey returns the raw public key from the private key.
func ExportPublicKey(privatekey crypto.PrivateKey) ([]byte, error) {
	signer, err := ssh.NewSignerFromKey(privatekey)
	if err != nil {
		return nil, err
	}
	return ssh.MarshalAuthorizedKey(signer.PublicKey()), nil
}
