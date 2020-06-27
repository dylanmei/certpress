package main

import (
	"fmt"
	"os"

	"github.com/dylanmei/certpress/certpress"
)

func main() {
	if version() {
		os.Exit(0)
	}

	opts := parseOptions(os.Args[1:])
	if len(opts) == 0 {
		fmt.Printf("Nothing to do")
	}

	for _, opt := range opts {
		ks, err := certpress.NewKeyStore(opt.Files, opt.Password)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create KeyStore '%s': %v\n", opt.Name, err)
			os.Exit(1)
		}

		if err = writeKeyStore(ks, opt.Name+".jks"); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to save '%s' KeyStore: %v\n", opt.Name, err)
			os.Exit(1)
		}

		fmt.Printf("Created '%s' KeyStore\n", opt.Name)
	}
}

func writeKeyStore(ks *certpress.KeyStore, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}

	defer file.Close()
	return ks.Encode(file)
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
