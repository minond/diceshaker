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
		id := string(m.Data)
		task := expl.Item{Topic: "photo", Data: m.Data}
		ch := q.Push(task)

		log.Printf("roll %s\n", id)
		acks := <-ch

		if len(acks) == 0 {
			log.Println("error, expecting 1 ack")
			return
		}

		ack := acks[0]
		if ack.Err != nil {
			log.Printf("error running photo task: %v\n", ack.Err)
			return
		}

		switch file := ack.Data.(type) {
		case string:
			log.Printf("file is %s\n", file)
			log.Printf("publishing to %s\n", id)
			nc.Publish(id, []byte(file))
		default:
			log.Printf("invalid ack data type: %v\n", ack.Data)
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
		id := uuid.New().String()
		ch := make(chan string, 1)

		log.Printf("subscribing to %s\n", id)
		sub, _ := nc.Subscribe(id, func(m *nats.Msg) {
			log.Printf("got response for %s: %s\n", id, m.Data)
			ch <- string(m.Data)
		})

		nc.Publish(subjRoll, []byte(id))
		nc.Flush()

		fmt.Fprintf(w, <-ch)
		sub.Unsubscribe()
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
