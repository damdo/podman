## podman for gokrazy

This package contains https://github.com/mgoltzsche/podman-static, a static
build of podman.

### Usage

Please see [gokrazy → Available packages → Docker
containers](https://gokrazy.org/packages/docker-containers/) for instructions on
how to use this repository.

The sections below assume you are logged into to your gokrazy device using
[breakglass](https://github.com/gokrazy/breakglass).


#### Run a container

```
podman run --rm -ti docker.io/library/debian:sid
```

#### Docker API socket

To run the podman API server in the foreground (e.g. via breakglass):

```
podman --socket /run/podman/podman.sock
```

The process creates a symlink from `/var/run/docker.sock` to the given path
(unless the path is `/var/run/docker.sock` itself), sets `DOCKER_HOST` for the
service process, and runs `podman system service` in the foreground.

On gokrazy, configure this as a service in your `config.json`:

```json
{
    "PackageConfig": {
        "github.com/gokrazy/podman": {
            "CommandLineFlags": [
                "--socket", "/run/podman/podman.sock"
            ]
        }
    }
}
```

#### Optional: tmpfs

By default, containers are stored on disk (`/var` is a symlink to `/perm/var` on
the permanent data partition). If you only want to try something out without
keeping the containers around across reboots, it is faster to work in RAM:

```
mount -t tmpfs tmpfs /var
```
