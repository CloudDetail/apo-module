package clickhouse

import (
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
)

type Authentication struct {
	PlainText *PlainTextConfig `mapstructure:",squash"`
	TLS       *TLSConfig       `mapstructure:"tls"`
}

type PlainTextConfig struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
}

func (config *Authentication) ConfigureAuthentication(opts *clickhouse.Options) error {
	if len(config.PlainText.Database) == 0 {
		return fmt.Errorf("database is required")
	}
	if config.PlainText != nil {
		if err := config.PlainText.ConfigurePlaintext(opts); err != nil {
			return err
		}
	}
	if config.TLS != nil {
		if err := configureTLS(config.TLS, opts); err != nil {
			return err
		}
	}
	return nil
}

func (plainTextConfig *PlainTextConfig) ConfigurePlaintext(opts *clickhouse.Options) error {
	opts.Auth.Username = plainTextConfig.Username
	opts.Auth.Password = plainTextConfig.Password
	opts.Auth.Database = plainTextConfig.Database
	return nil
}

func configureTLS(config *TLSConfig, opts *clickhouse.Options) error {
	tlsConfig, err := config.LoadTLSConfig()
	if err != nil {
		return fmt.Errorf("error loading tls config: %w", err)
	}
	if tlsConfig != nil && tlsConfig.InsecureSkipVerify {
		opts.TLS = tlsConfig
	}
	return nil
}
