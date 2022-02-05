default: all

all: orion

orion: cmd/orion/orion.go cmd/orion/gemini.go cmd/orion/config.go
	go build -o $@ $^

cert:
	openssl genrsa -out orion.key 2048
	openssl req -x509 -nodes -days 3650 -key orion.key -out orion.crt

# Container recipies
docker:
	docker build . -t feldspaten.org/orion
podman:
	podman build . -t feldspaten.org/orion
