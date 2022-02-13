# orion

![Github workflow badge](https://github.com/grisu48/orion/workflows/orion/badge.svg)

**orion is still alpha** *It mostly works, but expect some bugs and possible security issues.*

orion is a minimalistic gemini server written in go with the aim of being easy-to-use and minimal requirements.

Requirements: `go >= 1.14`

## Usage

Compile the `orion` binary

    make

Then edit the configuration file `orion.conf` to your wishes and launch the program

    ./orion -config orion.conf

`orion` can also be configured via [environmental variables](variables.md), which should be particularly useful for containerized applications.

### Create self-signed certificates

To create self-signed certificates for quick testing, you can use the following recipe

    make cert

### podman/docker container

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
