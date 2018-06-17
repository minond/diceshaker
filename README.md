WIP!!

IRL Random Number Generator using a [Raspberry Pi](raspberry-pi) with a [Camera
Module](camera-module). I'm using a Model B but any model with a camera module
and wireless should work.

You'll need `openssl`, Go + dep for building the project, and NATS running on a
server. NATS is used to communicate between the server (running on a server)
and the client (running on a Raspberry Pi.)

```bash
go get github.com/golang/dep/cmd/dep
go get github.com/nats-io/gnatsd
```

Once installed, you can generate a self-signed cert with `make cert`. This
generates `cert.pem` and `key.pem` which will be pushed to the server and
client.

You can build the project for the Raspberry Pi with `make arm`. And assuming
your server is running Linux on an amd64 architecture, you can build for that
with `make amd64`. Once built, deploy with `make deploy`. This last command
uses `scp` to copy the files over to the two devices. This can be configured
using the PI* and SERV* variables in the Makefile. Once on the client/server,
start things up with the following commands:

```bash
# on server
gnatsd --tls --tlscert cert.pem --tlskey key.pem
./diceshaker/diceshaker -role server
```

```bash
# on client
./diceshaker/diceshaker -role client -connect nats://<NATSSERVERHOST>:4222
```

[raspberry-pi]: https://www.raspberrypi.org/products/raspberry-pi-3-model-b/
[camera-module]: https://www.raspberrypi.org/products/camera-module-v2/
