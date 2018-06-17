IRL Random Number Generator using a [Raspberry Pi](raspberry-pi) with a [Camera
Module](camera-module). I'm using a Model B but any model with a camera module
and wireless should work.

## Build

You'll need `openssl`, Go + dep for building the project, and NATS running on a
server. NATS is used to communicate between the server (running on a server)
and the client (running on a Raspberry Pi.) Once you have `openssl`, you can
generate a self-signed cert with `make cert`. This generates `cert.pem` and
`key.pem` which will be pushed to the server and client.

You can build the project for the Raspberry Pi with `make arm`. And assuming
your server is running Linux on an amd64 architecture, you can build for that
with `make amd64`. Once built, deploy with `make deploy`. This last command
uses `scp` to copy the files over to the two devices. This can be configured
using the _PI*_ and _SERV*_ variables in the Makefile. In short, this is how
you can build and deploy to both the (amd64) server and client:

```bash
make amd64 deploy
```


## Setting up the server

Install `gnatsd` with the command below:

```bash
go get github.com/nats-io/gnatsd
```

After deploying your code, you can start up the NATS server and application
server with the following commands:

```bash
cd ~/diceshaker
gnatsd --tls --tlscert cert.pem --tlskey key.pem
./diceshaker -role server
```

The last command will bind to port 8080 on localhost by default and listen to
HTTP requests. Going to `/` will trigger a roll event.


## Setting up the client (Raspberry Pi)

Update the system if you haven't already and open the Raspberry Pi's
configuration program:

```bash
sudo apt-get update && sudo apt-get upgrade
sudo raspi-config
```

Once opened, arrow down to "Interface Options", then find the "Camera" setting,
and when asked to enable it, select "<Yes>". You'll now be prompted to reboot,
which you should do. After the reboot, you can test things out by running
`raspistill -o img.jpg`. If things are working you should get a popup ui window
with a preview of the camera's image. It'll take a photo in five seconds.

After deploying your code, you can start up the client with the following
commands:

```bash
cd ~/diceshaker
./diceshaker -role client -connect nats://<NATSSERVERHOST>:4222
```

[raspberry-pi]: https://www.raspberrypi.org/products/raspberry-pi-3-model-b/
[camera-module]: https://www.raspberrypi.org/products/camera-module-v2/
