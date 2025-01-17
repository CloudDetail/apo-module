package clickhouse

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type ClickhouseConfig struct {
	Authentication   Authentication `mapstructure:",squash"`
	Endpoints        []string       `mapstructure:"endpoint"`
	Compression      string         `mapstructure:"compression"`
	Cluster          string         `mapstructure:"cluster"`
	Table            string         `mapstructure:"table"`
	MaxExecutionTime int            `mapstructure:"max_execution_time"`
	DialTimeout      time.Duration  `mapstructure:"dial_timeout"`
	MaxOpenConns     int            `mapstructure:"max_open_conns"`
	MaxIdleConns     int            `mapstructure:"max_idle_conns"`
	ConnMaxLifetime  time.Duration  `mapstructure:"conn_max_life_time"`
	BlockBufferSize  uint8          `mapstructure:"block_buffer_size"`

	BufferNumLayers int `mapstructure:"buffer_num_layers"`
	BufferMinTime   int `mapstructure:"buffer_min_time"`
	BufferMaxTime   int `mapstructure:"buffer_max_time"`
	BufferMinRows   int `mapstructure:"buffer_min_rows"`
	BufferMaxRows   int `mapstructure:"buffer_max_rows"`
	BufferMinBytes  int `mapstructure:"buffer_min_bytes"`
	BufferMaxBytes  int `mapstructure:"buffer_max_bytes"`
}

func newConn(cfg *ClickhouseConfig) (driver.Conn, error) {
	compression, err := compressionMethod(cfg.Compression)
	if err != nil {
		return nil, err
	}
	opt := &clickhouse.Options{
		Addr: cfg.Endpoints,
		DialContext: func(ctx context.Context, addr string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, "tcp", addr)
		},
		Settings: clickhouse.Settings{
			"max_execution_time": cfg.MaxExecutionTime,
		},
		Compression: &clickhouse.Compression{
			Method: compression,
		},
		DialTimeout:      cfg.DialTimeout,
		MaxOpenConns:     cfg.MaxOpenConns,
		MaxIdleConns:     cfg.MaxIdleConns,
		ConnMaxLifetime:  cfg.ConnMaxLifetime,
		ConnOpenStrategy: clickhouse.ConnOpenInOrder,
		BlockBufferSize:  cfg.BlockBufferSize,
	}
	if err = cfg.Authentication.ConfigureAuthentication(opt); err != nil {
		err = fmt.Errorf("configure authentication failed, err: %w", err)
		return nil, err
	}
	conn, err := clickhouse.Open(opt)
	if err != nil {
		err = fmt.Errorf("sql open failed, err: %w", err)
		return nil, err
	}
	if err = conn.Ping(context.Background()); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			log.Printf("connect clickhouse with Exception:\n\t[%d]: %s\n%s", exception.Code, exception.Message, exception.StackTrace)
		} else {
			log.Printf("failed to connect with clickhouse: %v", err)
		}
		return nil, err
	}
	return conn, nil
}

func compressionMethod(compression string) (clickhouse.CompressionMethod, error) {
	switch compression {
	case "none":
		return clickhouse.CompressionNone, nil
	case "gzip":
		return clickhouse.CompressionGZIP, nil
	case "lz4":
		return clickhouse.CompressionLZ4, nil
	case "zstd":
		return clickhouse.CompressionZSTD, nil
	case "deflate":
		return clickhouse.CompressionDeflate, nil
	case "br":
		return clickhouse.CompressionBrotli, nil
	default:
		return clickhouse.CompressionNone, fmt.Errorf("producer.compression should be one of 'none', 'gzip', 'deflate', 'lz4', 'br' or 'zstd'. configured value %v", compression)
	}
}
