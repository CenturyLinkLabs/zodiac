# zodiac

A lightweight tool, built on top of [Docker Compose](https://docs.docker.com/compose/), for tactical deployment of a dockerized application (or, Docker Compose with rollback support).

Zodiac allows you to deploy Docker Compose applications while maintaining a history of all deployments. 
Among other things, this allows you to rollback to a previous, known-good deployment in the event that there are issues with your current build.

Imagine you have the following `docker-compose.yml` file

```
web:
  image: centurylink/simple-server
  ports:
    - "8888:8888"
```

Using Zodiac to deploy this application would look like this:

```
$ zodiac deploy
Deploying your application...
Creating zodiac_web_1
Successfully deployed 1 container(s)
```

You can view your previous deployments using the `list` command:

```
$ zodiac list
ID      DEPLOY DATE             SERVICES        MESSAGE
1       2015-06-30 00:40:27     zodiac_web_1
```

## Installation
* Install [Docker Compose](http://docs.docker.com/compose/) **v1.2**. See: http://docs.docker.com/compose/
* Install the Zodiac Binary for your platform. See: https://github.com/CenturyLinkLabs/zodiac/releases/

## Usage

The zodiac client supports the following commands:

* `verify` - verify that the target Docker endpoint is reachable and running a compatible version of the API.
* `deploy` - deploy the Docker Compose-defined application to the target Docker endpoint.
* `rollback` - roll to a previous Zodiac deployment.
* `list` - list all previous application deployments.
* `teardown` - remove running services and deployment history for the application.

### Global Options

The following flags apply to all of the Zodiac commands:

* `--endpoint` - Host and port for the target Docker endpoint. Should be in the form "tcp://hostname:port". Can optionally be provided by setting the `DOCKER_HOST` environment variable.
* `--tls` - Flag indicating whether or not to use TLS/SSL to communicate with the Docker daemon endpoint (defaults to *true*).
* `--tlsverify` - Flag indicating whether or not to perform TLS certificate authentication on the remote server's certificate (defaults to *true*). 
* `--tlscacert` - Path to the CA certificate which should be used to authenticate the remote server's certificate (defaults to *~/.docker/ca.pem*).
* `--tlscert` - Path to the certificate which should be used for client certificate authentication (defaults to *~/.docker/cert.pem*).
* `--tlskey` - Path to the private key which should be used for client certificate authentication (defaults to *~/.docker/key.pem*).
* `--debug` - Run the client in debug mode with verbose output.
* `--version` - Display version information for the Zodiac client.
* `--help` - Display the Zodiac client help text.

For more information about the various TLS flags, see the TLS section below.

### TLS
If you are going to use Zodiac to deploy to remote hosts you will want to ensure that your remote Docker daemon is protected with TLS security.
Zodiac ships with TLS support out of the box. Both host verification and client authentication are done via TLS.
TLS is enabled by default, though it may be disabled for debugging purposes, when using private networks, etc.

#### Configuring TLS

##### Let Docker Machine do all the heavy lifting
If you used Docker machine to provision Docker on the remote Host, it will have generated TLS certificates and keys for you on both the client and host machine.

You can use the `docker-machine env` command to automatically set-up the environment variables necessary for Zodiac to communicate with the remote host via TLS:

    eval "$(docker-machine env your_remote_name)"

If you do this, you should not need to use the `--endpoint` flag or any of the `--tls*` flags when running the Zodiac client.

##### OR Set up the certificates manually
If you choose to generate the TLS certificates manually, you'll want to genrate a certificate (be it self-signed or CA signed) for the host.
You'll also want to generate a client certificate for authentication purposes. Consult the Docker docs for [securing the Docker daemon](https://docs.docker.com/articles/https/).

Use the `zodiac help` command to see the options for passing in the necassary certificate files.


## Desired features / fixes
- [ ] Compose v1.3 support
- [ ] Support for compose's `build` option
- [ ] Support for compose's `volumes_from` option
- [ ] Support for private repos on the Docker hub
