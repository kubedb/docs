package keystore

import (
	"encoding/binary"
	"time"
)

const magic uint32 = 0xfeedfeed
const (
	version01 uint32 = 1
	version02 uint32 = 2
)
const (
	privateKeyTag         uint32 = 1
	trustedCertificateTag uint32 = 2
)
const bufSize = 1024

var order = binary.BigEndian

var whitenerMessage = []byte("Mighty Aphrodite")

func passwordBytes(password []byte) []byte {
	passwdBytes := make([]byte, 0, len(password)*2)
	for _, b := range password {
		passwdBytes = append(passwdBytes, 0, b)
	}
	return passwdBytes
}

func zeroing(s []byte) {
	for i := 0; i < len(s); i++ {
		s[i] = 0
	}
}

func millisecondsToTime(ms int64) time.Time {
	return time.Unix(0, ms*int64(time.Millisecond))
}

func timeToMilliseconds(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}
