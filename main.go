package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
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
	host = flag.String("http", ":8080", "Host and port for HTTP requests")
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
		log.Fatalf("client subscription error: %v\n", err)
	}

	log.Printf("listening for messages\n")
	runtime.Goexit()
}

func server() {
	nc := connect()
	defer nc.Close()

	http.HandleFunc("/"+subjRoll, func(w http.ResponseWriter, r *http.Request) {
		nc.Publish(subjRoll, nil)
		nc.Flush()

		if err := nc.LastError(); err != nil {
			log.Printf("server publishing error: %v\n", err)
			fmt.Fprintf(w, "err")
		} else {
			log.Printf("published message\n")
			fmt.Fprintf(w, "ok")
		}
	})

	log.Fatal(http.ListenAndServe(*host, nil))
}

func main() {
	switch *role {
	case roleClient:
		client()
	case roleServer:
		server()
	default:
		log.Fatalf("bad role '%s', expecting %s or %s\n", *role, roleClient, roleServer)
	}
}
