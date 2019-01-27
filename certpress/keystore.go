package certpress

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"strings"
	"time"

	keystore "github.com/pavel-v-chernykh/keystore-go"
)

type KeyStore struct {
	store      keystore.KeyStore
	passphrase string
}

type KeyStoreFiles struct {
	KeyURL     string
	CertURL    string
	CACertURLs []string
}

func NewKeyStore(opts KeyStoreFiles, passphrase string) (*KeyStore, error) {
	ks := keystore.KeyStore{}

	for _, caURL := range opts.CACertURLs {
		caCerts, err := fetchCACertFiles(caURL)
		if err != nil {
			return nil, err
		}

		for alias, cert := range caCerts {
			ks[alias] = &keystore.TrustedCertificateEntry{
				Entry:       keystore.Entry{CreationDate: time.Now()},
				Certificate: cert,
			}
		}
	}

	if opts.KeyURL != "" {
		priv, err := fetchKeyFile(opts.KeyURL)
		if err != nil {
			return nil, err
		}

		certs, err := fetchCertFiles(opts.CertURL)
		if err != nil {
			return nil, err
		}

		ks["certpress"] = &keystore.PrivateKeyEntry{
			Entry:     keystore.Entry{CreationDate: time.Now()},
			PrivKey:   priv,
			CertChain: certs,
		}
	}

	return &KeyStore{
		store:      ks,
		passphrase: passphrase,
	}, nil
}

func (ks *KeyStore) Encode(w io.Writer) error {
	return keystore.Encode(w, ks.store, []byte(ks.passphrase))
}

func fetchCACertFiles(fileURL string) (map[string]keystore.Certificate, error) {
	certs, err := fetchCertFiles(fileURL)
	if err != nil {
		return nil, err
	}

	aliasMap := map[string]keystore.Certificate{}
	for _, cert := range certs {
		parsed, err := x509.ParseCertificates(cert.Content)
		if err != nil {
			return nil, err
		}

		if len(parsed) < 1 {
			return nil, fmt.Errorf("Could not decode CA certificate")
		}

		for _, ca := range parsed {
			commonName := ca.Subject.CommonName
			if len(commonName) == 0 {
				return nil, fmt.Errorf("Missing cn in CA certificate subject: %v", ca.Subject)
			}

			alias := strings.Replace(strings.ToLower(commonName), " ", "", -1)
			aliasMap[alias] = cert
		}
	}

	return aliasMap, nil
}

func fetchCertFiles(fileURL string) ([]keystore.Certificate, error) {
	cbs, err := fetchPemfile(fileURL)
	if err != nil {
		return nil, err
	}

	var certs []keystore.Certificate
	for _, cb := range cbs {
		certs = append(certs, keystore.Certificate{
			Type:    "X509",
			Content: cb.Bytes,
		})
	}

	return certs, nil
}

func fetchKeyFile(fileURL string) ([]byte, error) {
	pkbs, err := fetchPemfile(fileURL)
	if err != nil {
		return nil, err
	}
	if len(pkbs) != 1 {
		return nil, fmt.Errorf("Failed to single PEM block from file %s", fileURL)
	}

	var pk interface{}
	pkb := pkbs[0]
	switch pkb.Type {
	case "PRIVATE KEY":
		return pkb.Bytes, nil
	case "RSA PRIVATE KEY":
		pk, err = x509.ParsePKCS1PrivateKey(pkb.Bytes)
	case "EC PRIVATE KEY":
		pk, err = x509.ParseECPrivateKey(pkb.Bytes)
	default:
		return nil, fmt.Errorf("Unsupported key type: %s", pkb.Type)
	}

	if err != nil {
		return nil, err
	}

	return convertPrivateKeyToPKCS8(pk)
}

func fetchPemfile(fileURL string) ([]*pem.Block, error) {
	raw, err := fetchBytes(fileURL)
	if err != nil {
		return nil, err
	}
	var (
		pemBlocks []*pem.Block
		current   *pem.Block
	)

	for {
		current, raw = pem.Decode(raw)
		if current == nil {
			if len(pemBlocks) > 0 {
				return pemBlocks, nil
			}

			return nil, fmt.Errorf("Failed to decode any PEM blocks from %s", fileURL)
		}
		pemBlocks = append(pemBlocks, current)
		if len(raw) == 0 {
			break
		}
	}
	return pemBlocks, nil
}
