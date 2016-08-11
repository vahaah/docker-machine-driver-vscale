# Docker Machine Vscale Driver

This is a plugin for [Docker Machine](https://docs.docker.com/machine/) allowing
to create Docker hosts on [Vscale]( http://vscale.io ) cloud services.

## Installation

To install this plugin manually, download the binary `docker-machine-driver-vscale`
and  make it available by `$PATH`, for example by putting it to `/usr/local/bin/`:

```console
$ curl -L https://github.com/vahaah/docker-machine-driver-vscale/releases/download/1.1.0/docker-machine-driver-vscale > /usr/local/bin/docker-machine-driver-vscale

$ chmod +x /usr/local/bin/docker-machine-driver-vscale
```

The latest version of `docker-machine-driver-vscale` binary is available on
the ["Releases"](https://github.com/vahaah/docker-machine-driver-vscale/releases) page.

For Ubuntu 16.04:

```console
$ sudo curl -L https://github.com/vahaah/docker-machine-driver-vscale/releases/download/1.1.0/docker-machine-driver-vscale.ubuntu.16.04.x64 > /usr/local/bin/docker-machine-driver-vscale

$ sudo chmod +x /usr/local/bin/docker-machine-driver-vscale
```

## Usage

After compile you can use driver for creating docker hosts.
Get Vscale access token from [your profile](https://vscale.io/panel/settings/tokens/) then run:

```console
$ docker-machine create -d vscale --vscale-access-token YOUR_VSCALE_ACCESS_TOKEN machine_name
```

Available options:

 - `--vscale-access-token`: Access token.
 - `--vscale-location`: Server location.
 - `--vscale-rplan`: Server size.
 - `--vscale-made-from`: Server type
 - `--vscale-swap-file`: Swap size in MB

Environment variables and default values:

| CLI option                    | Environment variable        | Default                     |
|-------------------------------|-----------------------------|-----------------------------|
| `--vscale-access-token`       | `VSCALE_ACCESS_TOKEN`       | -                           |
| `--vscale-location`           | `VSCALE_LOCATION`           | `spb0`                      |
| `--vscale-rplan`              | `VSCALE_RPLAN`              | `small`                     |
| `--vscale-made-from`          | `VSCALE_MADE_FROM`          | `ubuntu_14.04_64_002_master`|
| `--vscale-swap-file`          | `VSCALE_SWAP_FILE`          | `0`                         |

## Development

### Build from Source
If you wish to work on Vscale Driver for Docker machine, you'll first need
[Go](http://www.golang.org) installed (version 1.5+ is required).
Make sure Go is properly installed, including setting up a [GOPATH](http://golang.org/doc/code.html#GOPATH).

Run these commands to build the plugin binary:

```bash
$ go get -d github.com/vahaah/docker-machine-driver-vscale
$ cd $GOPATH/github.com/vahaah/docker-machine-driver-vscale
$ make build
```

After the build is complete, `bin/docker-machine-driver-vscale` binary will
be created. If you want to copy it to the `${GOPATH}/bin/`, run `make install`.

