default: all

all: orion

orion: cmd/orion/orion.go cmd/orion/gemini.go cmd/orion/config.go
	go build -o $@ $^

static: cmd/orion/orion.go cmd/orion/gemini.go cmd/orion/config.go
	CGO_ENABLED=0 GOARGS="-buildmode=pie" go build -o orion $^

cert:
	openssl genrsa -out orion.key 2048
	openssl req -x509 -nodes -days 3650 -key orion.key -out orion.crt

# Container recipies
docker:
	docker build . -t feldspaten.org/orion
podman:
	podman build . -t feldspaten.org/orion
