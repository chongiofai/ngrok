package server

import (
	"crypto/tls"
	"fmt"
	"math/rand"
	"ngrok/conn"
	log "ngrok/log"
	"ngrok/msg"
	"ngrok/util"
	"os"
	"runtime/debug"
	"time"
)

const (
	registryCacheSize uint64        = 1024 * 1024 // 1 MB
	connReadTimeout   time.Duration = 10 * time.Second
)

// GLOBALS
var (
	tunnelRegistry  *TunnelRegistry
	controlRegistry *ControlRegistry

	// XXX: kill these global variables - they're only used in tunnel.go for constructing forwarding URLs
	opts      *Options
	config    *Configuration
	listeners map[string]*conn.Listener
)

func NewProxy(pxyConn conn.Conn, regPxy *msg.RegProxy) {
	// fail gracefully if the proxy connection fails to register
	defer func() {
		if r := recover(); r != nil {
			pxyConn.Warn("Failed with error: %v", r)
			pxyConn.Close()
		}
	}()

	// set logging prefix
	pxyConn.SetType("pxy")

	// look up the control connection for this proxy
	pxyConn.Info("Registering new proxy for %s", regPxy.ClientId)
	ctl := controlRegistry.Get(regPxy.ClientId)

	if ctl == nil {
		panic("No client found for identifier: " + regPxy.ClientId)
	}

	ctl.RegisterProxy(pxyConn)
}

// Listen for incoming control and proxy connections
// We listen for incoming control and proxy connections on the same port
// for ease of deployment. The hope is that by running on port 443, using
// TLS and running all connections over the same port, we can bust through
// restrictive firewalls.
func tunnelListener(config *Configuration, tlsConfig *tls.Config) {
	// listen for incoming connections
	listener, err := conn.Listen(config.TunnelAddr, "tun", tlsConfig)
	if err != nil {
		panic(err)
	}

	log.Info("Listening for control and proxy connections on %s", listener.Addr.String())
	for c := range listener.Conns {
		go func(tunnelConn conn.Conn) {
			// don't crash on panics
			defer func() {
				if r := recover(); r != nil {
					tunnelConn.Info("tunnelListener failed with error %v: %s", r, debug.Stack())
				}
			}()

			tunnelConn.SetReadDeadline(time.Now().Add(connReadTimeout))
			var rawMsg msg.Message
			if rawMsg, err = msg.ReadMsg(tunnelConn); err != nil {
				tunnelConn.Warn("Failed to read message: %v", err)
				tunnelConn.Close()
				return
			}

			// don't timeout after the initial read, tunnel heartbeating will kill
			// dead connections
			tunnelConn.SetReadDeadline(time.Time{})

			switch m := rawMsg.(type) {
			case *msg.Auth:
				NewControl(tunnelConn, config, m)

			case *msg.RegProxy:
				NewProxy(tunnelConn, m)

			default:
				tunnelConn.Close()
			}
		}(c)
	}
}

func Main() {
	// parse options
	opts = parseArgs()
	config, err := LoadConfiguration(opts)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// init logging
	log.LogTo(config.LogTo, config.LogLevel)

	// seed random number generator
	seed, err := util.RandomSeed()
	if err != nil {
		panic(err)
	}
	rand.Seed(seed)

	// init tunnel/control registry
	registryCacheFile := os.Getenv("REGISTRY_CACHE_FILE")
	tunnelRegistry = NewTunnelRegistry(registryCacheSize, registryCacheFile)
	controlRegistry = NewControlRegistry()

	// start listeners
	listeners = make(map[string]*conn.Listener)

	// load tls configuration
	tlsConfig, err := LoadTLSConfig(config.TlsCrt, config.TlsKey)
	if err != nil {
		panic(err)
	}

	// listen for http
	if config.HTTPAddr != "" {
		listeners["http"] = startHttpListener(config.HTTPAddr, nil)
	}

	// listen for https
	if config.HTTPSAddr != "" {
		listeners["https"] = startHttpListener(config.HTTPSAddr, tlsConfig)
	}

	// ngrok clients
	tunnelListener(config, tlsConfig)
}
