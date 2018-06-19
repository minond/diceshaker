package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os/exec"
	"runtime"

	"github.com/google/uuid"
	expl "github.com/minond/expresslane"
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

	certVerify = flag.Bool("verify", false, "Controls whether a client verifies the server's certificate chain and host name")
	certFile   = flag.String("certfile", "cert.pem", "Path to certificate file")
	keyFile    = flag.String("keyfile", "key.pem", "Path to key file")
)

func init() {
	flag.Parse()
}

func connect() *nats.Conn {
	log.Printf("connecting to server\n")

	pair, err := tls.LoadX509KeyPair(*certFile, *keyFile)
	if err != nil {
		log.Fatalf("error loading certificate key pair (%s, %s): %v\n",
			*certFile, *keyFile, err)
	}

	config := &tls.Config{
		InsecureSkipVerify: !*certVerify,
		Certificates:       []tls.Certificate{pair},
	}

	nc, err := nats.Connect(*url, nats.Secure(config))
	if err != nil {
		log.Fatalf("error connecting to server on %s: %v\n", *url, err)
	}

	return nc
}

func photo() (string, error) {
	name := fmt.Sprintf("%d.jpg", rand.Int())
	cmd := exec.Command("raspistill", "--nopreview",
		"--timeout", "1", "--output", name)

	log.Print("taking photo... ")
	if err := cmd.Run(); err != nil {
		log.Printf("error running command: %v\n", err)
		return name, err
	}

	log.Println("great success!")
	return name, nil
}

func client() {
	nc := connect()
	q := expl.New().Start()

	q.Register("photo", func(i expl.Item) expl.Ack {
		name, err := photo()
		return expl.Ack{Data: name, Err: err}
	})

	log.Printf("subscribing to '%s'\n", subjRoll)
	nc.Subscribe(subjRoll, func(m *nats.Msg) {
		log.Printf("roll (%s)\n", string(m.Data))
		ch := q.Push(expl.Item{Topic: "photo", Data: m.Data})
		ack := <-ch
		if len(ack) > 0 {
			log.Printf("file: %s\n", ack[0].Data)
		}
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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		nc.Publish(subjRoll, []byte(uuid.New().String()))
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
		log.Fatalf("bad role '%s', expecting %s or %s\n",
			*role, roleClient, roleServer)
	}
}
