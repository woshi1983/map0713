// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package resolver

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
)

var (
	regexpName = regexp.MustCompile("^[a-z]([-a-z0-9]*[a-z0-9])?$")
)

func validateName(s string) bool {
	return regexpName.MatchString(s)
}

func standardizeString(s string) string {
	return strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(s, "_", "-"), "/", "-"), " ", "-"))
}

type Image struct {
	Registry   string
	Project    string
	Repository string
	Tag        string
}

func (i *Image) String() string {
	var ss []string
	if i.Registry != "" {
		ss = append(ss, i.Registry)
	}
	if i.Project != "" {
		ss = append(ss, i.Project)
	}
	if i.Repository != "" {
		ss = append(ss, i.Repository)
	}
	s := filepath.Join(ss...)
	if i.Tag != "" {
		s += ":" + i.Tag
	}
	return s
}

func (i *Image) IsQualified() bool {
	return i.Registry != "" && i.Project != "" && i.Repository != ""
}

func ParseImage(s string) (*Image, error) {
	if strings.Count(s, ":") > 1 {
		return nil, fmt.Errorf("invalid image name: %s", s)
	}
	splitter := strings.Index(s, ":")
	if splitter == 0 {
		return nil, fmt.Errorf("invalid image name: %s", s)
	}
	var (
		ss        []string
		name, tag string
	)
	if splitter < 0 {
		name = s
	} else {
		name = s[:splitter]
		tag = s[splitter+1:]
	}
LOOP:
	for {
		switch name {
		case ".":
			break LOOP
		case "/":
			return nil, fmt.Errorf("invalid image name: %s", s)
		default:
			str := filepath.Base(name)
			if str == "." || str == "/" {
				return nil, fmt.Errorf("invalid image name: %s", s)
			}
			name = filepath.Dir(name)
			ss = append([]string{str}, ss...)
		}
	}

	switch len(ss) {
	case 0:
		return nil, fmt.Errorf("invalid image name: %s", s)
	case 1:
		return &Image{Repository: ss[0], Tag: tag}, nil
	case 2:
		return &Image{Project: ss[0], Repository: ss[1], Tag: tag}, nil
	}
	return &Image{Registry: ss[0], Project: ss[1], Repository: strings.Join(ss[2:], "/"), Tag: tag}, nil
}

func validateLabels(labels map[string]string, matches map[string]string) bool {
	if matches == nil {
		return true
	}
	if labels == nil {
		return false
	}
	for k, v := range matches {
		vv, ok := labels[k]
		if !ok || vv != v {
			return false
		}
	}
	return true
}

func convertX509Certificate(crt *x509.Certificate) *model.X509Certificate {
	obj := &model.X509Certificate{
		Raw:                   string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: crt.Raw})),
		Sig:                   crt.Signature,
		SigAlg:                crt.SignatureAlgorithm.String(),
		Ver:                   int64(crt.Version),
		Sn:                    crt.SerialNumber.String(),
		Iss:                   crt.Issuer.String(),
		Sub:                   crt.Subject.String(),
		NotBefore:             crt.NotBefore,
		NotAfter:              crt.NotAfter,
		BasicConstraintsValid: crt.BasicConstraintsValid,
		IsCa:                  crt.IsCA,
		DNSNames:              crt.DNSNames,
	}
	switch crt.KeyUsage {
	case x509.KeyUsageDigitalSignature:
		obj.KeyUsage = "digital signature"
	case x509.KeyUsageContentCommitment:
		obj.KeyUsage = "content commitment"
	case x509.KeyUsageKeyEncipherment:
		obj.KeyUsage = "key encipherment"
	case x509.KeyUsageDataEncipherment:
		obj.KeyUsage = "data encipherment"
	case x509.KeyUsageKeyAgreement:
		obj.KeyUsage = "key agreement"
	case x509.KeyUsageCertSign:
		obj.KeyUsage = "cert sign"
	case x509.KeyUsageCRLSign:
		obj.KeyUsage = "crl sign"
	case x509.KeyUsageEncipherOnly:
		obj.KeyUsage = "encipher only"
	case x509.KeyUsageDecipherOnly:
		obj.KeyUsage = "decipher only"
	default:
		obj.KeyUsage = "<none>"
	}
	for _, ext := range crt.Extensions {
		obj.Extensions = append(obj.Extensions, &model.X509CertificateExtension{
			ID:    ext.Id.String(),
			Value: ext.Value,
		})
	}
	return obj
}
