# Environmental variables


The following table lists all supported environmental variables for usage in `orion`. Environmental variables are in particular useful when working with `orion` as containerized application.

| Name | Description |
|------|-------------|
| `ORION_HOSTNAME` | Server hostname to use |
| `ORION_CERTFILE` | TLS certificate file |
| `ORION_KEYFILE` | TLS key file |
| `ORION_BIND` | Bind address for orion (e.g. `:1965`) |
| `ORION_CONTENTDIR` | Content/data directory |
| `ORION_CHROOT` | If set, orion will chroot into this directory |
| `ORION_UID` | If set, orion will perform a setuid to this uid after initialization |
| `ORION_GID` | If set, orion will perform a setgid to this gid after initialization |