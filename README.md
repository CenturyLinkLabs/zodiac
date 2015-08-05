![Zodiac](http://panamax.ca.tier3.io/logos/logo_zodiac.png)


##### A lightweight tool, built on top of [Docker Compose](https://docs.docker.com/compose/), for easy deployment and rollback of dockerized applications.

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

### Readying the remote environment
While Zodiac can be used just locally, most will be pointing it at a remote environment. In that case zodiac needs to be able to communicate with the docker daemon on that remote host. __We strongly encourage the use of Docker Machine__ for provisioning and installing Docker on the remote host. Zodiac is designed to work out of the box with a Machine-provisioned endpoint. However, one can also set `DOCKER_OPTS` manually on the remote host. Instructions for setting `DOCKER_OPTS` at runtime will vary by OS.



#### TLS
If you are going to use Zodiac to deploy to remote hosts you will want to ensure that your remote Docker daemon is protected with TLS security.
Zodiac ships with TLS support out of the box. Both host verification and client authentication are done via TLS.
TLS is enabled by default, though it may be disabled for debugging purposes, when using private networks, etc.

##### Configuring TLS

__NOTE:__ We assume you're using TLS by default, if not using TLS (strongly discouraged), you'll need to set `--tls=false`, or set the `DOCKER_TLS` envirnoment variable to false.

###### Let Docker Machine do all the heavy lifting
If you used Docker machine to provision Docker on the remote Host, it will have generated TLS certificates and keys for you on both the client and host machine.

You can use the `docker-machine env` command to automatically set-up the environment variables necessary for Zodiac to communicate with the remote host via TLS:

    eval "$(docker-machine env your_remote_name)"

If you do this, you should not need to use the `--endpoint` flag or any of the `--tls*` flags when running the Zodiac client.

###### OR Set up the certificates manually
If you choose to generate the TLS certificates manually, you'll want to genrate a certificate (be it self-signed or CA signed) for the host.
You'll also want to generate a client certificate for authentication purposes. Consult the Docker docs for [securing the Docker daemon](https://docs.docker.com/articles/https/).

Use the `zodiac help` command to see the options for passing in the necassary certificate files.






### Install the local Zodiac client
* Install [Docker Compose](http://docs.docker.com/compose/) **v1.3**. See: http://docs.docker.com/compose/
* Install the Zodiac Binary for your platform. See: https://github.com/CenturyLinkLabs/zodiac/releases/

or if you're a risk taker:

```
curl -sSL https://raw.githubusercontent.com/CenturyLinkLabs/zodiac/master/install.sh | bash
```

## Usage

The zodiac client supports the following commands:

* `verify` - verify that the target Docker endpoint is reachable and running a compatible version of the API.
* `deploy` - deploy the Docker Compose-defined application to the target Docker endpoint.
* `rollback` - roll to a previous Zodiac deployment.
* `list` - list all previous application deployments.
* `teardown` - remove running services and deployment history for the application.

**NOTE:** Zodiac stores all deployment history on the containers, so manually removing containers can destroy all Zodiac history.

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

## Remote Target Configuration

The remote Docker host must be exposed over a TCP port to enable remote communication from the local Zodiac CLI. This is typically done by setting DOCKER_OPTS to something like: `DOCKER_OPTS="-H tcp://0.0.0.0:2375 -H unix:///var/run/docker.sock"`. It's worth noting that Docker Machine will do this for you.

## Desired features / fixes
- [x] Support for compose's `build` option
- [ ] Support for compose's `volumes_from` option
- [ ] Support for private repos on the Docker hub
- [ ] Compose scale support
