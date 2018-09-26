package main

import (
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"software.sslmate.com/src/go-pkcs12"
)

type CertSpec struct {
	Name                 string
	Key                  string
	Certificate          string
	CertificateAuthority string
	Password             string
}

func main() {
	if version() {
		os.Exit(0)
	}

	specs := parseSpecs(os.Args[1:])
	if len(specs) == 0 {
		fmt.Printf("Nothing to do")
	}

	for _, spec := range specs {
		bytes, err := createCertificate(spec.Certificate, spec.Key, spec.CertificateAuthority, spec.Password)
		if err != nil {
			fmt.Printf("Failed to create '%s' PKCS12 certificate: %v\n", spec.Name, err)
			os.Exit(1)
		}

		if err := ioutil.WriteFile(spec.Name+".pkcs12", bytes, 0644); err != nil {
			fmt.Printf("Failed to write '%s' PKCS12 certificate: %v\n", spec.Name, err)
			os.Exit(1)
		}

		fmt.Printf("Created '%s' PKCS12 certificate\n", spec.Name)
	}
}

func createCertificate(certificateURL, keyURL, caCertificateURL, password string) ([]byte, error) {
	keyBytes, err := fetch(keyURL)
	if err != nil {
		return nil, fmt.Errorf("ERROR downloading Key %s: %v", keyURL, err)
	}

	fmt.Printf("Fetched %d bytes from key '%s'\n", len(keyBytes), keyURL)

	certificateBytes, err := fetch(certificateURL)
	if err != nil {
		return nil, fmt.Errorf("ERROR downloading Certificate %s: %v", certificateURL, err)
	}

	fmt.Printf("Fetched %d bytes from certificate '%s'\n", len(certificateBytes), certificateURL)

	caCertificateBytes, err := fetch(caCertificateURL)
	if err != nil {
		return nil, fmt.Errorf("ERROR downloading CA Certificate %s: %v", caCertificateURL, err)
	}

	fmt.Printf("Fetched %d bytes from CA certificate '%s'\n", len(caCertificateBytes), caCertificateURL)

	return encodeBytes(certificateBytes, keyBytes, caCertificateBytes, password)
}

func encodeBytes(certificateBytes, keyBytes, caCertificateBytes []byte, password string) ([]byte, error) {
	var err error

	keyBlock, _ := pem.Decode(keyBytes)
	if keyBlock == nil {
		return nil, errors.New("error decoding private key")
	}

	var privateKey interface{}
	switch keyBlock.Type {
	case "RSA PRIVATE KEY":
		privateKey, err = x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
		if err != nil {
			return nil, errors.New("Could not parse RSA PRIVATE KEY")
		}
	case "EC PRIVATE KEY":
		privateKey, err = x509.ParseECPrivateKey(keyBlock.Bytes)
		if err != nil {
			return nil, errors.New("Could not parse EC PRIVATE KEY")
		}
	default:
		return nil, fmt.Errorf("Unsupported key type: %v", keyBlock.Type)
	}

	certBlock, _ := pem.Decode(certificateBytes)
	if certBlock == nil {
		return nil, errors.New("Cound not decode certificate")
	}

	certificate, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("Could not parse certificate: %v", err)
	}

	caCertBlock, _ := pem.Decode(caCertificateBytes)
	if caCertBlock == nil {
		return nil, errors.New("Could not decode CA certificate")
	}

	caCertificate, err := x509.ParseCertificate(caCertBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("Could not parse ca certificate: %v", err)
	}

	return pkcs12.Encode(rand.Reader, privateKey, certificate, []*x509.Certificate{caCertificate}, password)
}

func parseSpecs(args []string) []CertSpec {
	specs := make(map[string]*CertSpec)
	for _, arg := range args {
		fvp := strings.Split(arg, "=")
		if len(fvp) != 2 {
			continue
		}

		field := fvp[0]
		value := fvp[1]

		npp := strings.Split(field, ".")
		if len(npp) != 2 {
			npp = []string{"certpress", npp[0]}
		}

		name := strings.TrimLeft(npp[0], "-")
		prop := strings.TrimLeft(npp[1], "-")

		var spec *CertSpec
		spec = specs[name]
		if spec == nil {
			spec = &CertSpec{
				Name:     name,
				Password: "certpress",
			}

			specs[name] = spec
		}

		if prop == "certificate" {
			spec.Certificate = value
		}
		if prop == "key" {
			spec.Key = value
		}
		if prop == "certificate-authority" {
			spec.CertificateAuthority = value
		}
		if prop == "password" {
			spec.Password = value
		}
	}

	list := []CertSpec{}
	for _, spec := range specs {
		list = append(list, *spec)
	}

	return list
}

func version() bool {
	for _, arg := range os.Args[1:] {
		if arg == "-version" {
			PrintVersion(os.Stdout)
			return true
		}
	}

	return false
}
