package shared

import (
	"time"

	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"

	"github.com/golang/glog"
	"gopkg.in/mgo.v2"
)

const (
	dialMongodbTimeout = 10 * time.Second
	syncMongodbTimeout = 1 * time.Minute
)

// MongoSessionOpts represents options for a Mongo session
type MongoSessionOpts struct {
	URI                   string
	TLSCertificateFile    string
	TLSPrivateKeyFile     string
	TLSCaFile             string
	TLSHostnameValidation bool
	UserName              string
	AuthMechanism         string
}

// MongoSession creates a Mongo session
func MongoSession(opts MongoSessionOpts) *mgo.Session {
	dialInfo, err := mgo.ParseURL(opts.URI)
	if err != nil {
		glog.Errorf("Cannot connect to server using url %s: %s", opts.URI, err)
		return nil
	}

	dialInfo.Direct = true // Force direct connection
	dialInfo.Timeout = dialMongodbTimeout
	if opts.UserName != "" {
		dialInfo.Username = opts.UserName
	}

	err = opts.configureDialInfoIfRequired(dialInfo)
	if err != nil {
		glog.Errorf("%s", err)
		return nil
	}

	session, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		glog.Errorf("Cannot connect to server using url %s: %s", opts.URI, err)
		return nil
	}
	session.SetMode(mgo.Eventual, true)
	session.SetSyncTimeout(syncMongodbTimeout)
	session.SetSocketTimeout(0)
	return session
}

func (opts MongoSessionOpts) configureDialInfoIfRequired(dialInfo *mgo.DialInfo) error {
	if opts.AuthMechanism != "" {
		dialInfo.Mechanism = opts.AuthMechanism
	}
	if len(opts.TLSCertificateFile) > 0 {
		certificates, err := LoadKeyPairFrom(opts.TLSCertificateFile, opts.TLSPrivateKeyFile)
		if err != nil {
			return fmt.Errorf("Cannot load key pair from '%s' and '%s' to connect to server '%s'. Got: %v", opts.TLSCertificateFile, opts.TLSPrivateKeyFile, opts.URI, err)
		}
		config := &tls.Config{
			Certificates:       []tls.Certificate{certificates},
			InsecureSkipVerify: !opts.TLSHostnameValidation,
		}
		if len(opts.TLSCaFile) > 0 {
			ca, err := LoadCertificatesFrom(opts.TLSCaFile)
			if err != nil {
				return fmt.Errorf("Couldn't load client CAs from %s. Got: %s", opts.TLSCaFile, err)
			}
			config.RootCAs = ca
		}
		dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
			conn, err := tls.Dial("tcp", addr.String(), config)
			if err != nil {
				glog.Infof("Could not connect to %v. Got: %v", addr, err)
				return nil, err
			}
			if config.InsecureSkipVerify {
				err = enrichWithOwnChecks(conn, config)
				if err != nil {
					glog.Infof("Could not disable hostname validation. Got: %v", err)
				}
			}
			return conn, err
		}
	}
	return nil
}

func enrichWithOwnChecks(conn *tls.Conn, tlsConfig *tls.Config) error {
	var err error
	if err = conn.Handshake(); err != nil {
		conn.Close()
		return err
	}

	opts := x509.VerifyOptions{
		Roots:         tlsConfig.RootCAs,
		CurrentTime:   time.Now(),
		DNSName:       "",
		Intermediates: x509.NewCertPool(),
	}

	certs := conn.ConnectionState().PeerCertificates
	for i, cert := range certs {
		if i == 0 {
			continue
		}
		opts.Intermediates.AddCert(cert)
	}

	_, err = certs[0].Verify(opts)
	if err != nil {
		conn.Close()
		return err
	}

	return nil
}
