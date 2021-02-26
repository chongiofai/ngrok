package client

import (
	"flag"
	"fmt"
	"ngrok/version"
	"os"
	"time"
)

const usage1 string = `Usage: %s [OPTIONS] <local port or address>
Options:
`
const usage2 string = `
Examples:
	ngrok 80
	ngrok -subdomain=example 8080
	ngrok -proto=tcp 22
	ngrok -hostname="example.com" -httpauth="user:password" 10.0.0.1


Advanced usage: ngrok [OPTIONS] <command> [command args] [...]
Commands:
	ngrok start [tunnel] [...]    Start tunnels by name from config file
	ngork start-all               Start all tunnels defined in config file
	ngrok list                    List tunnel names from config file
	ngrok help                    Print help
	ngrok version                 Print ngrok version

Examples:
	ngrok start www api blog pubsub
	ngrok -log=stdout -config=ngrok.yml start ssh
	ngrok start-all
	ngrok version

`
const (
	defaultServerAddr      string        = "ngrokd.ngrok.com:443"
	defaultInspectAddr     string        = "127.0.0.1:4040"
	pingInterval           time.Duration = 20 * time.Second
	maxPongLatency         time.Duration = 15 * time.Second
	updateCheckInterval    time.Duration = 6 * time.Hour
	defaultConfigPath      string        = "/etc/ngrok/client.yml"
	defaultLogTo           string        = "none"
	defaultLogLevel        string        = "WARNING"
	defaultAuthToken       string        = ""
	defaultHttpAuth        string        = ""
	defaultHostname        string        = ""
	defaultProtocol        string        = "http+https"
	defaultSubdomain       string        = ""
	defaultHttpRequestPath string        = "/"
)

var (
	defaultRootCrtPaths = []string{"assets/client/tls/ngrokroot.crt", "assets/client/tls/snakeoilca.crt"}
	defaultRootCrtPath  = defaultRootCrtPaths[0]
)

type Options struct {
	configPath      string
	logTo           string
	logLevel        string
	authToken       string
	httpAuth        string
	hostname        string
	protocol        string
	subdomain       string
	httpRequestPath string
	rootCrtPath     string
	command         string
	args            []string
}

func ParseArgs() (opts *Options, err error) {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usage1, os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, usage2)
	}

	configPath := flag.String(
		"config-path",
		defaultConfigPath,
		"Path to ngrok configuration file.")

	logTo := flag.String(
		"log",
		defaultLogTo,
		"Write log messages to this file. 'stdout' and 'none' have special meanings")

	logLevel := flag.String(
		"log-level",
		defaultLogLevel,
		"The level of messages to log. One of: DEBUG, INFO, WARNING, ERROR")

	authToken := flag.String(
		"authtoken",
		defaultAuthToken,
		"Authentication token for identifying")

	httpAuth := flag.String(
		"httpauth",
		defaultHttpAuth,
		"username:password HTTP basic auth creds protecting the public tunnel endpoint")

	subdomain := flag.String(
		"subdomain",
		defaultSubdomain,
		"Request a custom subdomain from the ngrok server. (HTTP only)")

	httpRequestPath := flag.String(
		"requestPath",
		defaultHttpRequestPath,
		"HTTP Request a custom request path from the ngrok server. (HTTP only)")

	hostname := flag.String(
		"hostname",
		defaultHostname,
		"Request a custom hostname from the ngrok server. (HTTP only) (requires CNAME of your DNS)")

	protocol := flag.String(
		"proto",
		defaultProtocol,
		"The protocol of the traffic over the tunnel {'http', 'https', 'tcp'}")

	rootCrtPath := flag.String(
		"root-crt-paths",
		defaultRootCrtPath,
		"")

	flag.Parse()

	opts = &Options{
		configPath:      *configPath,
		logTo:           *logTo,
		logLevel:        *logLevel,
		authToken:       *authToken,
		httpAuth:        *httpAuth,
		hostname:        *hostname,
		protocol:        *protocol,
		subdomain:       *subdomain,
		httpRequestPath: *httpRequestPath,
		rootCrtPath:     *rootCrtPath,
		command:         flag.Arg(0),
	}

	switch opts.command {
	case "list":
		opts.args = flag.Args()[1:]
	case "start":
		opts.args = flag.Args()[1:]
	case "start-all":
		opts.args = flag.Args()[1:]
	case "version":
		fmt.Println(version.MajorMinor())
		os.Exit(0)
	case "help":
		flag.Usage()
		os.Exit(0)
	case "":
		err = fmt.Errorf("Error: Specify a local port to tunnel to, or " +
			"an ngrok command.\n\nExample: To expose port 80, run " +
			"'ngrok 80'")
		return

	default:
		if len(flag.Args()) > 1 {
			err = fmt.Errorf("You may only specify one port to tunnel to on the command line, got %d: %v",
				len(flag.Args()),
				flag.Args())
			return
		}

		opts.command = "default"
		opts.args = flag.Args()
	}

	return
}
