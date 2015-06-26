# zodiac

A lightweight tool, built on top of [Docker Compose](https://docs.docker.com/compose/), for tactical deployment of a dockerized application.

## Usage

### Installation
* Install [Docker Compose](http://docs.docker.com/compose/) **v1.2**. See: http://docs.docker.com/compose/
* Install the Zodiac Binary for your platform. See: https://github.com/CenturyLinkLabs/zodiac/releases/

### TLS
If you are going to use Zodiac to deploy to remote hosts you will want to ensure that your remote Docker daemon is protected with TLS security.
Zodiac ships with TLS support out of the box. Both host verification and client authentication are done via TLS.
TLS is enabled by default, though it may be disabled for debugging purposes, when using private networks, etc.

#### Configuring TLS

##### Let Docker Machine do all the heavy lifting

If you used Docker machine to provision Docker on the remote Host, it will have generated TLS certificates and keys for you on both the client and host machine.

You can use the `docker-machine env` command to automatically set-up the environment variables necessary for Zodiac to communicate with the remote host via TLS:

    eval "$(docker-machine env your_remote_name)"

##### OR Set up the certificates manually

If you choose to generate the TLS certificates manually, you'll want to genrate a certificate (be it self-signed or CA signed) for the host.
You'll also want to generate a client certificate for authentication purposes. Consult the Docker docs for [securing the Docker daemon](https://docs.docker.com/articles/https/).

Use the `zodiac help` command to see the options for passing in the necassary certificate files.


## Desired features / fixes
- [ ] Compose v1.3 support
- [ ] Support for compose's `build` option
- [ ] Support for private repos on the Docker hub
