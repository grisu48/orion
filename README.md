# orion

![Github workflow badge](https://github.com/grisu48/orion/workflows/orion/badge.svg)

**orion is still in development** *However: It works. Feedback, Issues and Pull request are very welcome.*

orion is a minimalistic gemini server written in go with the goal of being easy-to-use and to have minimal requirements as well as a small footprint.

Requirements: `go >= 1.14`

## Usage

Running `orion` is as simple as

    ./orion -config orion.conf

`orion` requires three things to work properly

* A valid configuration file (see [orion.conf](orion.conf))
* A TLS certificate and key file (see below)
* Your awesome gemini content (See `ContentDir` in the `orion.conf`)

A example TLS certificate and key file is required. See the [Create self-signed certificate](#create-self-signed-certificates) section below.

The recommended way of running `orion` is as a `podman` container (See [below](#run-a-podman-docker-container)).

### Pre-build binaries

Pre-build binaries for Linux are available on the [releases page](https://github.com/grisu48/orion/releases).

### Run a podman/docker container

Assuming you have your configuration files in `/srv/orion/conf` and your data directory in `/srv/orion/data`:

    docker run -d --name orion -v /srv/orion/conf:/conf -v /srv/orion/data:/data -p 1965:1965 grisu48/orion
    podman run -d --name orion -v /srv/orion/conf:/conf -v /srv/orion/data:/data -p 1965:1965 --memory 128M grisu48/orion

Make sure that the configuration file `/srv/orion/conf/orion.conf` exists and is configured to your needs. Checkout the example [orion.conf](orion.conf) in this directory.

Also ensure that the certificate and key files are located in `/srv/orion/conf/` and configured properly in your `orion.conf`. See the section [create self-signed certificate](#create-self-signed-certificates) for more information.

`orion` can also be configured via [environmental variables](variables.md), which should be particularly useful for containerized applications.

### Build and run the binary

Compile the `orion` binary

    make               # Default build
    make static        # Build static binary

Then edit the configuration file `orion.conf` to your wishes and launch the program

    ./orion -config orion.conf

### Create self-signed certificates

> Disclaimer:
> A self-signed certificate allows for a whole class of attack scenarios e.g. man-in-the-middle attacks without additional safety guards like TOFU. Be aware that a self-signed certificate does not give you the same protection as a signed certificate by a trusted CA.

That being said, in the gemini universe self-signed certificated are kind of the common use case.

To create self-signed certificates for quick testing, you can use the following make recipe:

    make cert

Alternatively you can also run the openssl commands directly:

```bash
openssl genrsa -out orion.key 2048
openssl req -x509 -nodes -days 3650 -key orion.key -out orion.crt
```

### Build podman/docker container

`orion` is able to launch from a podman/docker container, however you need to first build the container yourself.

    make podman         # Build container for podman
    make docker         # Build container for docker

The container expects the `/conf` volume to contain your configuration file `orion.conf` and suggests to place your date into the `/data` volume. If you want to use the later depends on your configuration in `orion.conf`.

After building and configuration, this is how you can run your container:

    # Replace `podman` with `docker` for a docker container
    podman run --rm -ti --name orion -v /srv/orion/conf:/conf -v /srv/orion/data/:/data -p 1965 feldspaten.org/orion

It's recommended to place your certificates in the `/conf` direcory and use the following configuration

    Certfile = /conf/orion.crt
    Keyfile = /conf/orion.key

Note: Use the `chroot` setting in containers for additional security.

See also the supported [environmental variables](variables.md) for additional configuration possibilities.

## Credits

* This project was inspired by the [titan2](https://gitlab.com/lostleonardo/titan2) minimalistic Gemini server written by lostleonardo.