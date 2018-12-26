package keystore

import (
	"time"
)

// KeyStore is a mapping of alias to pointer to PrivateKeyEntry or TrustedCertificateEntry
type KeyStore map[string]interface{}

// Certificate describes type of certificate
type Certificate struct {
	Type    string
	Content []byte
}

// Entry is a basis of entries types supported by keystore
type Entry struct {
	CreationDate time.Time
}

// PrivateKeyEntry is an entry for private keys and associated certificates
type PrivateKeyEntry struct {
	Entry
	PrivKey   []byte
	CertChain []Certificate
}

// TrustedCertificateEntry is an entry for certificates only
type TrustedCertificateEntry struct {
	Entry
	Certificate Certificate
}
