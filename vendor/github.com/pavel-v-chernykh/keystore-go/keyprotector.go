package keystore

import (
	"crypto/sha1"
	"crypto/x509/pkix"
	"encoding/asn1"
	"errors"
	"io"
)

const saltLen = 20

var supportedPrivateKeyAlgorithmOid = asn1.ObjectIdentifier([]int{1, 3, 6, 1, 4, 1, 42, 2, 17, 1, 1})

// ErrUnsupportedPrivateKeyAlgorithm indicates unsupported private key algorithm
var ErrUnsupportedPrivateKeyAlgorithm = errors.New("keystore: unsupported private key algorithm")

// ErrUnrecoverablePrivateKey indicates unrecoverable private key content (often means wrong password usage)
var ErrUnrecoverablePrivateKey = errors.New("keystore: unrecoverable private key")

type keyInfo struct {
	Algo       pkix.AlgorithmIdentifier
	PrivateKey []byte
}

func recoverKey(encodedKey []byte, password []byte) ([]byte, error) {
	var keyInfo keyInfo
	asn1Rest, err := asn1.Unmarshal(encodedKey, &keyInfo)
	if err != nil || len(asn1Rest) > 0 {
		return nil, ErrIncorrectPrivateKey
	}
	if !keyInfo.Algo.Algorithm.Equal(supportedPrivateKeyAlgorithmOid) {
		return nil, ErrUnsupportedPrivateKeyAlgorithm
	}

	md := sha1.New()
	passwordBytes := passwordBytes(password)
	defer zeroing(passwordBytes)
	salt := make([]byte, saltLen)
	copy(salt, keyInfo.PrivateKey)
	encrKeyLen := len(keyInfo.PrivateKey) - saltLen - md.Size()
	numRounds := encrKeyLen / md.Size()

	if encrKeyLen%md.Size() != 0 {
		numRounds++
	}

	encrKey := make([]byte, encrKeyLen)
	copy(encrKey, keyInfo.PrivateKey[saltLen:])

	xorKey := make([]byte, encrKeyLen)

	digest := salt
	for i, xorOffset := 0, 0; i < numRounds; i++ {
		_, err := md.Write(passwordBytes)
		if err != nil {
			return nil, ErrUnrecoverablePrivateKey
		}
		_, err = md.Write(digest)
		if err != nil {
			return nil, ErrUnrecoverablePrivateKey
		}
		digest = md.Sum(nil)
		md.Reset()
		copy(xorKey[xorOffset:], digest)
		xorOffset += md.Size()
	}

	plainKey := make([]byte, encrKeyLen)
	for i := 0; i < len(plainKey); i++ {
		plainKey[i] = encrKey[i] ^ xorKey[i]
	}

	_, err = md.Write(passwordBytes)
	if err != nil {
		return nil, ErrUnrecoverablePrivateKey
	}
	_, err = md.Write(plainKey)
	if err != nil {
		return nil, ErrUnrecoverablePrivateKey
	}
	digest = md.Sum(nil)
	md.Reset()

	digestOffset := saltLen + encrKeyLen
	for i := 0; i < len(digest); i++ {
		if digest[i] != keyInfo.PrivateKey[digestOffset+i] {
			return nil, ErrUnrecoverablePrivateKey
		}
	}

	return plainKey, nil
}

func protectKey(rand io.Reader, plainKey []byte, password []byte) ([]byte, error) {
	md := sha1.New()
	passwdBytes := passwordBytes(password)
	defer zeroing(passwdBytes)
	plainKeyLen := len(plainKey)
	numRounds := plainKeyLen / md.Size()

	if plainKeyLen%md.Size() != 0 {
		numRounds++
	}

	salt := make([]byte, saltLen)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}

	xorKey := make([]byte, plainKeyLen)

	digest := salt
	for i, xorOffset := 0, 0; i < numRounds; i++ {
		_, err = md.Write(passwdBytes)
		if err != nil {
			return nil, err
		}
		_, err = md.Write(digest)
		if err != nil {
			return nil, err
		}
		digest = md.Sum(nil)
		md.Reset()
		copy(xorKey[xorOffset:], digest)
		xorOffset += md.Size()
	}

	tmpKey := make([]byte, plainKeyLen)
	for i := 0; i < plainKeyLen; i++ {
		tmpKey[i] = plainKey[i] ^ xorKey[i]
	}

	encrKey := make([]byte, saltLen+plainKeyLen+md.Size())
	encrKeyOffset := 0
	copy(encrKey[encrKeyOffset:], salt)
	encrKeyOffset += saltLen
	copy(encrKey[encrKeyOffset:], tmpKey)
	encrKeyOffset += plainKeyLen

	_, err = md.Write(passwdBytes)
	if err != nil {
		return nil, err
	}
	_, err = md.Write(plainKey)
	if err != nil {
		return nil, err
	}
	digest = md.Sum(nil)
	md.Reset()
	copy(encrKey[encrKeyOffset:], digest)
	keyInfo := keyInfo{
		Algo: pkix.AlgorithmIdentifier{
			Algorithm: supportedPrivateKeyAlgorithmOid,
			Parameters: asn1.RawValue{Tag: 5},
		},
		PrivateKey: encrKey,
	}
	encodedKey, err := asn1.Marshal(keyInfo)
	if err != nil {
		return nil, err
	}
	return encodedKey, nil
}
