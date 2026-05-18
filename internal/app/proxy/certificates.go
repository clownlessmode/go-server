package proxy

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/AdguardTeam/gomitmproxy/mitm"
)

const (
	caPEMFile   = "rebellion-ca.pem"
	certPEMFile = "rebellion-ca-cert.pem"
	certCERFile = "rebellion-ca-cert.cer"
)

type certificateAuthority struct {
	cert *x509.Certificate
	key  *rsa.PrivateKey
}

func loadOrCreateCA(dir string) (*certificateAuthority, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create proxy cert dir: %w", err)
	}

	caPath := filepath.Join(dir, caPEMFile)
	if _, err := os.Stat(caPath); err == nil {
		proxyLog.Infof("ca loaded: path=%s", caPath)
		return loadCA(caPath)
	}

	proxyLog.Warnf("ca not found, generating new ca: dir=%s", dir)
	cert, key, err := mitm.NewAuthority("Rebellion MITM Proxy", "Rebellion", 10*365*24*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("create proxy ca: %w", err)
	}

	if err := writeCAFiles(dir, cert, key); err != nil {
		return nil, err
	}

	proxyLog.Successf("ca generated: cert=%s", filepath.Join(dir, certPEMFile))
	return &certificateAuthority{cert: cert, key: key}, nil
}

func loadCA(path string) (*certificateAuthority, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read proxy ca: %w", err)
	}

	var cert *x509.Certificate
	var key *rsa.PrivateKey
	for len(body) > 0 {
		var block *pem.Block
		block, body = pem.Decode(body)
		if block == nil {
			break
		}

		switch block.Type {
		case "CERTIFICATE":
			parsedCert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("parse proxy ca cert: %w", err)
			}
			cert = parsedCert
		case "RSA PRIVATE KEY":
			parsedKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("parse proxy ca key: %w", err)
			}
			key = parsedKey
		}
	}

	if cert == nil || key == nil {
		return nil, fmt.Errorf("proxy ca pem must contain certificate and rsa private key")
	}

	return &certificateAuthority{cert: cert, key: key}, nil
}

func writeCAFiles(dir string, cert *x509.Certificate, key *rsa.PrivateKey) error {
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})

	if err := os.WriteFile(filepath.Join(dir, caPEMFile), append(keyPEM, certPEM...), 0o600); err != nil {
		return fmt.Errorf("write proxy ca pem: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, certPEMFile), certPEM, 0o644); err != nil {
		return fmt.Errorf("write proxy ca cert pem: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, certCERFile), certPEM, 0o644); err != nil {
		return fmt.Errorf("write proxy ca cert cer: %w", err)
	}

	return nil
}
