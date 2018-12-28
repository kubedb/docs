package keystore

import (
	"crypto/sha1"
	"errors"
	"hash"
	"io"
)

const defaultCertificateType = "X509"

// ErrIo indicates i/o error
var ErrIo = errors.New("keystore: invalid keystore format")

// ErrIncorrectMagic indicates incorrect file magic
var ErrIncorrectMagic = errors.New("keystore: invalid keystore format")

// ErrIncorrectVersion indicates incorrect keystore version format
var ErrIncorrectVersion = errors.New("keystore: invalid keystore format")

// ErrIncorrectTag indicates incorrect keystore entry tag
var ErrIncorrectTag = errors.New("keystore: invalid keystore format")

// ErrIncorrectPrivateKey indicates incorrect private key entry content
var ErrIncorrectPrivateKey = errors.New("keystore: invalid private key format")

// ErrInvalidDigest indicates that keystore was tampered or password was incorrect
var ErrInvalidDigest = errors.New("keystore: invalid digest")

type keyStoreDecoder struct {
	r  io.Reader
	b  [bufSize]byte
	md hash.Hash
}

func (ksd *keyStoreDecoder) readUint16() (uint16, error) {
	const blockSize = 2
	_, err := io.ReadFull(ksd.r, ksd.b[:blockSize])
	if err != nil {
		return 0, ErrIo
	}
	_, err = ksd.md.Write(ksd.b[:blockSize])
	if err != nil {
		return 0, err
	}
	return order.Uint16(ksd.b[:blockSize]), nil
}

func (ksd *keyStoreDecoder) readUint32() (uint32, error) {
	const blockSize = 4
	_, err := io.ReadFull(ksd.r, ksd.b[:blockSize])
	if err != nil {
		return 0, ErrIo
	}
	_, err = ksd.md.Write(ksd.b[:blockSize])
	if err != nil {
		return 0, err
	}
	return order.Uint32(ksd.b[:blockSize]), nil
}

func (ksd *keyStoreDecoder) readUint64() (uint64, error) {
	const blockSize = 8
	_, err := io.ReadFull(ksd.r, ksd.b[:blockSize])
	if err != nil {
		return 0, ErrIo
	}
	_, err = ksd.md.Write(ksd.b[:blockSize])
	if err != nil {
		return 0, err
	}
	return order.Uint64(ksd.b[:blockSize]), nil
}

func (ksd *keyStoreDecoder) readBytes(num uint32) ([]byte, error) {
	var result []byte
	for lenToRead := num; lenToRead > 0; {
		blockSize := lenToRead
		if blockSize > bufSize {
			blockSize = bufSize
		}
		_, err := io.ReadFull(ksd.r, ksd.b[:blockSize])
		if err != nil {
			return result, ErrIo
		}
		result = append(result, ksd.b[:blockSize]...)
		lenToRead -= blockSize
	}
	_, err := ksd.md.Write(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (ksd *keyStoreDecoder) readString() (string, error) {
	strLen, err := ksd.readUint16()
	if err != nil {
		return "", err
	}
	bytes, err := ksd.readBytes(uint32(strLen))
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (ksd *keyStoreDecoder) readCertificate(version uint32) (*Certificate, error) {
	var certType string
	switch version {
	case version01:
		certType = defaultCertificateType
	case version02:
		readCertType, err := ksd.readString()
		if err != nil {
			return nil, err
		}
		certType = readCertType
	default:
		return nil, ErrIncorrectVersion
	}
	certLen, err := ksd.readUint32()
	if err != nil {
		return nil, err
	}
	certContent, err := ksd.readBytes(certLen)
	if err != nil {
		return nil, err
	}
	certificate := Certificate{
		Type:    certType,
		Content: certContent,
	}
	return &certificate, nil
}

func (ksd *keyStoreDecoder) readPrivateKeyEntry(version uint32, password []byte) (*PrivateKeyEntry, error) {
	creationDateTimeStamp, err := ksd.readUint64()
	if err != nil {
		return nil, err
	}
	privKeyLen, err := ksd.readUint32()
	if err != nil {
		return nil, err
	}
	encodedPrivateKeyContent, err := ksd.readBytes(privKeyLen)
	if err != nil {
		return nil, err
	}
	certCount, err := ksd.readUint32()
	if err != nil {
		return nil, err
	}
	var chain []Certificate
	for i := certCount; i > 0; i-- {
		cert, err := ksd.readCertificate(version)
		if err != nil {
			return nil, err
		}
		chain = append(chain, *cert)
	}
	plainPrivateKeyContent, err := recoverKey(encodedPrivateKeyContent, password)
	if err != nil {
		return nil, err
	}
	creationDateTime := millisecondsToTime(int64(creationDateTimeStamp))
	privateKeyEntry := PrivateKeyEntry{
		Entry: Entry{
			CreationDate: creationDateTime,
		},
		PrivKey:   plainPrivateKeyContent,
		CertChain: chain,
	}
	return &privateKeyEntry, nil
}

func (ksd *keyStoreDecoder) readTrustedCertificateEntry(version uint32) (*TrustedCertificateEntry, error) {
	creationDateTimeStamp, err := ksd.readUint64()
	if err != nil {
		return nil, err
	}
	cert, err := ksd.readCertificate(version)
	if err != nil {
		return nil, err
	}
	creationDateTime := millisecondsToTime(int64(creationDateTimeStamp))
	trustedCertificateEntry := TrustedCertificateEntry{
		Entry: Entry{
			CreationDate: creationDateTime,
		},
		Certificate: *cert,
	}
	return &trustedCertificateEntry, nil
}

func (ksd *keyStoreDecoder) readEntry(version uint32, password []byte) (string, interface{}, error) {
	tag, err := ksd.readUint32()
	if err != nil {
		return "", nil, err
	}
	alias, err := ksd.readString()
	if err != nil {
		return "", nil, err
	}
	switch tag {
	case privateKeyTag:
		entry, err := ksd.readPrivateKeyEntry(version, password)
		if err != nil {
			return "", nil, err
		}
		return alias, entry, nil
	case trustedCertificateTag:
		entry, err := ksd.readTrustedCertificateEntry(version)
		if err != nil {
			return "", nil, err
		}
		return alias, entry, nil
	}
	return "", nil, ErrIncorrectTag
}

// Decode reads keystore representation from r then decrypts and check signature using password
// It is strongly recommended to fill password slice with zero after usage
func Decode(r io.Reader, password []byte) (KeyStore, error) {
	ksd := keyStoreDecoder{
		r:  r,
		md: sha1.New(),
	}
	passwordBytes := passwordBytes(password)
	defer zeroing(passwordBytes)
	_, err := ksd.md.Write(passwordBytes)
	if err != nil {
		return nil, err
	}
	_, err = ksd.md.Write(whitenerMessage)
	if err != nil {
		return nil, err
	}

	readMagic, err := ksd.readUint32()
	if err != nil {
		return nil, err
	}
	if readMagic != magic {
		return nil, ErrIncorrectMagic
	}
	version, err := ksd.readUint32()
	if err != nil {
		return nil, err
	}
	count, err := ksd.readUint32()
	if err != nil {
		return nil, err
	}
	keyStore := KeyStore{}
	for entitiesCount := count; entitiesCount > 0; entitiesCount-- {
		alias, entry, err := ksd.readEntry(version, password)
		if err != nil {
			return nil, err
		}
		keyStore[alias] = entry
	}

	computedDigest := ksd.md.Sum(nil)
	actualDigest, err := ksd.readBytes(uint32(ksd.md.Size()))
	for i := 0; i < len(actualDigest); i++ {
		if actualDigest[i] != computedDigest[i] {
			return nil, ErrInvalidDigest
		}
	}

	return keyStore, nil
}
