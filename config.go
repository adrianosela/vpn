package main

// Config represents data required to
// configure the VPN server
type Config struct {
	ListenTCPPort int
	MaxTunnels    int
}

func (c *Config) validate() error {
	// TODO
	return nil
}
