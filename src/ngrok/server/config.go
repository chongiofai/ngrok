package server

import (
	"fmt"
	"io/ioutil"
	"net"
	"ngrok/log"
	"strconv"

	"gopkg.in/yaml.v1"
)

const (
	defaultPath       string = "/etc/ngrok/server.yml"
	defaultHTTPAddr   string = ":80"
	defaultHTTPSAddr  string = ":443"
	defaultTunnelAddr string = ":4443"
)

type Configuration struct {
	HTTPAddr   string `yaml:"http_addr,omitempty"`
	HTTPSAddr  string `yaml:"https_addr,omitempty"`
	TunnelAddr string `yaml:"tunnel_addr,omitempty"`
	Hostname   string `yaml:"hostname,omitempty"`
	TlsCrt     string `yaml:"tls_crt,omitempty"`
	TlsKey     string `yaml:"tls_key,omitempty"`
	LogTo      string `yaml:"log_to,omitempty"`
	LogLevel   string `yaml:"log_level,omitempty"`
	Config     string `yaml:"-"`
}

func LoadConfiguration(opts *Options) (config *Configuration, err error) {
	configPath := opts.config
	if configPath == "" {
		configPath = defaultPath
	}

	log.Info("Reading configuration file %s", configPath)
	configBuf, err := ioutil.ReadFile(configPath)
	if err != nil {
		// failure to read a configuration file is onlyconfigconfig a fatal error if
		// the user specified one explicitly
		if opts.config != "" {
			err = fmt.Errorf("Failed to read configuration file %s: %v", configPath, err)
			return
		}
	}

	// deserialize/parse the config
	config = new(Configuration)
	if err = yaml.Unmarshal(configBuf, &config); err != nil {
		err = fmt.Errorf("Error parsing configuration file %s: %v", configPath, err)
		return
	}

	// // try to parse the old .ngrok format for backwards compatibility
	// matched := false
	// content := strings.TrimSpace(string(configBuf))
	// if matched, err = regexp.MatchString("^[0-9a-zA-Z_\\-!]+$", content); err != nil {
	// 	return
	// } else if matched {
	// 	config = &Configuration{AuthToken: content}
	// }

	// set configuration defaults
	if config.HTTPAddr == "" {
		config.HTTPAddr = defaultHTTPAddr
	}

	if config.HTTPSAddr == "" {
		config.HTTPSAddr = defaultHTTPSAddr
	}

	if config.TunnelAddr == "" {
		config.TunnelAddr = defaultTunnelAddr
	}

	// validate and normalize configuration
	if config.HTTPAddr, err = normalizeAddress(config.HTTPAddr, "HTTPAddr"); err != nil {
		return
	}

	if config.HTTPSAddr, err = normalizeAddress(config.HTTPSAddr, "HTTPSAddr"); err != nil {
		return
	}

	if config.TunnelAddr, err = normalizeAddress(config.TunnelAddr, "TunnelAddr"); err != nil {
		return
	}

	// override configuration with command-line options
	config.LogTo = opts.logto
	config.LogLevel = opts.loglevel
	config.Config = configPath

	return
}

func normalizeAddress(addr string, propName string) (string, error) {
	// normalize port to address
	if _, err := strconv.Atoi(addr); err == nil {
		addr = ":" + addr
	}

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return "", fmt.Errorf("Invalid address %s '%s': %s", propName, addr, err.Error())
	}

	if host == "" {
		host = "127.0.0.1"
	}

	return fmt.Sprintf("%s:%s", host, port), nil
}

func validateProtocol(proto, propName string) (err error) {
	switch proto {
	case "http", "https", "http+https", "tcp":
	default:
		err = fmt.Errorf("Invalid protocol for %s: %s", propName, proto)
	}

	return
}

// func SaveAuthToken(configPath, authtoken string) (err error) {
// 	// empty configuration by default for the case that we can't read it
// 	c := new(Configuration)

// 	// read the configuration
// 	oldConfigBytes, err := ioutil.ReadFile(configPath)
// 	if err == nil {
// 		// unmarshal if we successfully read the configuration file
// 		if err = yaml.Unmarshal(oldConfigBytes, c); err != nil {
// 			return
// 		}
// 	}

// 	// no need to save, the authtoken is already the correct value
// 	if c.AuthToken == authtoken {
// 		return
// 	}

// 	// update auth token
// 	c.AuthToken = authtoken

// 	// rewrite configuration
// 	newConfigBytes, err := yaml.Marshal(c)
// 	if err != nil {
// 		return
// 	}

// 	err = ioutil.WriteFile(configPath, newConfigBytes, 0600)
// 	return
// }
