package main

import (
	"flag"
	"log"
	"runtime"

	nats "github.com/nats-io/go-nats"
)

const (
	roleServer = "server"
	roleClient = "client"
	subjRoll   = "roll"
)

var (
	url  = flag.String("connect", nats.DefaultURL, "NATS server URL")
	role = flag.String("role", "", "Is this a server or a client?")
)

func init() {
	flag.Parse()
}

func connect() *nats.Conn {
	log.Printf("connecting to server\n")
	nc, err := nats.Connect(*url)

	if err != nil {
		log.Printf("error connecting to server on %s\n", *url)
		log.Fatal(err)
	}

	return nc
}

func client() {
	nc := connect()

	log.Printf("subscribing to '%s'\n", subjRoll)
	nc.Subscribe(subjRoll, func(m *nats.Msg) {
		log.Printf("got roll message\n")
	})
	nc.Flush()

	if err := nc.LastError(); err != nil {
		log.Printf("client subscription error:")
		log.Fatal(err)
	}

	log.Printf("listening for messages\n")
	runtime.Goexit()
}

func server() {
	nc := connect()
	defer nc.Close()

	nc.Publish(subjRoll, nil)
	nc.Flush()

	if err := nc.LastError(); err != nil {
		log.Printf("server publishing error:")
		log.Fatal(err)
	} else {
		log.Printf("published message\n")
	}
}

func main() {
	switch *role {
	case roleClient:
		client()
	case roleServer:
		server()
	default:
		log.Fatal("bad role '%s', expecting %s or %s", *role, roleClient, roleServer)
	}
}
