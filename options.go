package main

import (
	"strings"

	"github.com/dylanmei/certpress/certpress"
)

type Options struct {
	Name     string
	Password string
	Files    certpress.KeyStoreFiles
}

func parseOptions(args []string) []Options {
	opts := make(map[string]*Options)
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

		var opt *Options
		opt = opts[name]
		if opt == nil {
			opt = &Options{
				Name: name,
				//Password: "certpress",
			}

			opts[name] = opt
		}

		if prop == "password" {
			opt.Password = value
		}
		if prop == "certificate" {
			opt.Files.CertURL = value
		}
		if prop == "key" {
			opt.Files.KeyURL = value
		}
		if prop == "certificate-authority" {
			opt.Files.CACertURLs = []string{value}
		}
	}

	list := []Options{}
	for _, opt := range opts {
		list = append(list, *opt)
	}

	return list
}
