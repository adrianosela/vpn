package main

import "errors"

// Config represents data required to
// configure the VPN server
type Config struct {
	ListenTCPPort int
	MaxTunnels    int
}

func (c *Config) validate() error {
	if c.ListenTCPPort < 0 || c.ListenTCPPort > 65535 {
		return errors.New("invalid tcp port")
	}
	if c.MaxTunnels < 0 {
		return errors.New("invalid max tunnels number")
	}
	return nil
}
