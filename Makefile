.PHONY: default server client deps fmt clean all release-all assets client-assets server-assets contributors
export GOPATH:=$(shell pwd)

BUILDTAGS=debug
DOMAIN=ngrokd.ngrok.com
default: all

signature:
	[ -d tmp ] || mkdir tmp
	openssl genrsa -out tmp/root.key 2048
	openssl req -x509 -new -nodes -key tmp/root.key -subj "/CN=$(DOMAIN)" -days 365 -out tmp/root.pem
	openssl genrsa -out tmp/device.key 2048
	openssl req -new -key tmp/device.key -subj "/CN=$(DOMAIN)" -out tmp/device.csr
	openssl x509 -req -in tmp/device.csr -CA tmp/root.pem -CAkey tmp/root.key -CAcreateserial -out tmp/device.crt -days 365
	cp tmp/root.key assets/client/tls/ngrokroot.key
	cp tmp/root.pem assets/client/tls/ngrokroot.crt
	cp tmp/device.crt assets/client/tls//snakeoil.crt
	cp tmp/device.crt assets/server/tls/snakeoil.crt
	cp tmp/device.key assets/server/tls/snakeoil.key

deps: assets
	go get -tags '$(BUILDTAGS)' -d -v ngrok/...

server: fmt deps 
	go install -tags '$(BUILDTAGS)' ngrok/main/ngrokd

client: fmt deps 
	go install -tags '$(BUILDTAGS)' ngrok/main/ngrok

fmt:
	go fmt ngrok/...

log:
	[ -d /var/log/ngrok ] || mkdir -p /var/log/ngrok

config-server: log
	[ -d /etc/ngrok/tls ] || mkdir -p /etc/ngrok/tls
	[ -e /etc/ngrok/server.yml ] || cp etc/server.yml /etc/ngrok/server.yml
	cp assets/server/tls/snakeoil.crt /etc/ngrok/tls/server.crt
	cp assets/server/tls/snakeoil.key /etc/ngrok/tls/server.key

config-client: log
	[ -d /etc/ngrok/tls ] || mkdir -p /etc/ngrok/tls
	[ -e /etc/ngrok/client.yml ] || cp etc/client.yml /etc/ngrok/client.yml
	cp assets/client/tls/ngrokroot.crt /etc/ngrok/tls/client.crt

link-server:
	ln -sf $(shell pwd)/bin/ngrokd /usr/bin/ngrokd

link-client:
	ln -sf $(shell pwd)/bin/ngrokd /usr/bin/ngrokd

service-server: link-server config-server
	cp service/ngrokd.service /etc/systemd/system/ngrokd.service
	systemctl daemon-reload

service-client: link-client config-client
	cp service/ngrok.service /etc/systemd/system/ngrok.service
	systemctl daemon-reload

assets: client-assets server-assets

bin/go-bindata:
	GOOS="" GOARCH="" go get github.com/jteeuwen/go-bindata/go-bindata

client-assets: bin/go-bindata
	bin/go-bindata -nomemcopy -pkg=assets -tags=$(BUILDTAGS) \
		-debug=$(if $(findstring debug,$(BUILDTAGS)),true,false) \
		-o=src/ngrok/client/assets/assets_$(BUILDTAGS).go \
		assets/client/...

server-assets: bin/go-bindata
	bin/go-bindata -nomemcopy -pkg=assets -tags=$(BUILDTAGS) \
		-debug=$(if $(findstring debug,$(BUILDTAGS)),true,false) \
		-o=src/ngrok/server/assets/assets_$(BUILDTAGS).go \
		assets/server/...

release-client: BUILDTAGS=release
release-client: client

release-server: BUILDTAGS=release
release-server: server

release-all: fmt release-client release-server

all: fmt client server services config

clean:
	go clean -i -r ngrok/...
	rm -rf src/ngrok/client/assets/ src/ngrok/server/assets/

contributors:
	echo "Contributors to ngrok, both large and small:\n" > CONTRIBUTORS
	git log --raw | grep "^Author: " | sort | uniq | cut -d ' ' -f2- | sed 's/^/- /' | cut -d '<' -f1 >> CONTRIBUTORS

install: server client
config-all: config-server config-client
service-all: service-server service-client
install-all: install config-all service-all