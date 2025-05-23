package clickhouse

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// The defaults should be a safe configuration
const defaultMinTLSVersion = tls.VersionTLS12

// Uses the default MaxVersion from "crypto/tls"
const defaultMaxTLSVersion = 0

// TLSConfig is the interface used to configure a tcp client or server from a `Config`
type TLSConfig struct {
	// Enable TLS
	Enable bool `mapstructure:"enabled"`
	// Path to the CA cert. For a client this verifies the server certificate.
	CAFile string `mapstructure:"ca_file"`
	// Path to the TLS cert to use for TLS required connections. (optional)
	CertFile string `mapstructure:"cert_file"`
	// Path to the TLS key to use for TLS required connections. (optional)
	KeyFile string `mapstructure:"key_file"`
	// InsecureSkipVerify will enable TLS but not verify the certificate
	InsecureSkipVerify bool `mapstructure:"insecure_skip_verify"`
	// MinVersion sets the minimum TLS version that is acceptable.
	// If not set, TLS 1.2 will be used. (optional)
	MinVersion string `mapstructure:"min_version"`
	// MaxVersion sets the maximum TLS version that is acceptable.
	// If not set, refer to crypto/tls for defaults. (optional)
	MaxVersion string `mapstructure:"max_version"`
}

func (c *TLSConfig) LoadTLSConfig() (*tls.Config, error) {
	if !c.Enable {
		return nil, nil
	}
	var err error
	var certPool *x509.CertPool
	if c.CAFile != "" {
		certPool, err = c.loadCert(c.CAFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load CA CertPool: %w", err)
		}
	}
	if (c.CertFile == "" && c.KeyFile != "") || (c.CertFile != "" && c.KeyFile == "") {
		return nil, errors.New("for auth via TLS, either both certificate and key must be supplied, or neither")
	}
	var cert tls.Certificate
	if c.CertFile != "" && c.KeyFile != "" {
		cert, err = tls.LoadX509KeyPair(c.CertFile, c.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("could not load TLS client key/certificate from %s:%s: %s", c.KeyFile, c.CertFile, err)
		}
	}

	minVersion, err := convertVersion(c.MinVersion, defaultMinTLSVersion)
	if err != nil {
		return nil, fmt.Errorf("invalid TLS min_version: %w", err)
	}
	maxVersion, err := convertVersion(c.MaxVersion, defaultMaxTLSVersion)
	if err != nil {
		return nil, fmt.Errorf("invalid TLS max_version: %w", err)
	}
	return &tls.Config{
		RootCAs:            certPool,
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: c.InsecureSkipVerify, //nolint:gosec
		MinVersion:         minVersion,
		MaxVersion:         maxVersion,
	}, nil
}

func (c TLSConfig) loadCert(caPath string) (*x509.CertPool, error) {
	caPEM, err := os.ReadFile(filepath.Clean(caPath))
	if err != nil {
		return nil, fmt.Errorf("failed to load CA %s: %w", caPath, err)
	}
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caPEM) {
		return nil, fmt.Errorf("failed to parse CA %s", caPath)
	}
	return certPool, nil
}

func convertVersion(version string, defaultVersion uint16) (uint16, error) {
	if version == "" {
		return defaultVersion, nil
	}
	val, ok := tlsProtocolVersions[version]
	if !ok {
		return 0, fmt.Errorf("unsupported TLS version: %q", version)
	}
	return val, nil
}

var tlsProtocolVersions = map[string]uint16{
	"1.0": tls.VersionTLS10,
	"1.1": tls.VersionTLS11,
	"1.2": tls.VersionTLS12,
	"1.3": tls.VersionTLS13,
}
