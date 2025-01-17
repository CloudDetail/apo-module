package config

import (
	"log"
	"net/http"
	"net/url"
)

type CenterServerConfig struct {
	Address      string `mapstructure:"address"`
	ProxyAddress string `mapstructure:"proxy_address"`
}

func createHttpClient(proxyAddress string) *http.Client {
	transport := &http.Transport{}
	if len(proxyAddress) > 0 {
		proxyURL, err := url.Parse(proxyAddress)
		if err != nil {
			log.Printf("warning: no proxy will be used since cannot parse the proxy address %v", err)
		} else {
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}
	return &http.Client{
		Transport: transport,
	}
}
