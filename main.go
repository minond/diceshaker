package main

import (
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"runtime"

	"github.com/google/uuid"
	expl "github.com/minond/expresslane"
	nats "github.com/nats-io/go-nats"
)

const (
	ROLE_SERVER = "server"
	ROLE_CLIENT = "client"
	SUBJ_ROLL   = "roll"
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

func photo(id string) (string, error) {
	name := fmt.Sprintf("%s.jpg", id)
	args := []string{"--nopreview", "--timeout", "1", "--output", name}
	cmd := exec.Command("raspistill", args...)

	log.Print(args)
	log.Print("taking photo...")
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

	q.Register(SUBJ_ROLL, func(i expl.Item) expl.Ack {
		id, err := str(i.Data)
		if err != nil {
			return expl.Ack{Err: err}
		}

		name, err := photo(id)
		return expl.Ack{Data: name, Err: err}
	})

	log.Printf("subscribing to '%s'\n", SUBJ_ROLL)
	nc.Subscribe(SUBJ_ROLL, func(m *nats.Msg) {
		id := string(m.Data)
		ch := q.Push(SUBJ_ROLL, id)

		log.Printf("roll %s\n", id)
		acks := <-ch

		if len(acks) == 0 {
			log.Println("error, expecting 1 ack")
			nc.Publish(id, nil)
			return
		}

		ack := acks[0]
		if ack.Err != nil {
			log.Printf("error running photo task: %v\n", ack.Err)
			nc.Publish(id, nil)
			return
		}

		file, err := str(ack.Data)
		if err != nil {
			log.Printf("invalid ack data type: %v\n", ack.Data)
			nc.Publish(id, nil)
			return
		}

		log.Printf("file is %s\n", file)
		log.Printf("publishing to %s\n", id)
		nc.Publish(id, []byte(file))
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

		nc.Publish(SUBJ_ROLL, []byte(id))
		nc.Flush()

		fmt.Fprintf(w, <-ch)
		sub.Unsubscribe()
	})

	log.Fatal(http.ListenAndServe(*host, nil))
}

func str(i interface{}) (string, error) {
	switch val := i.(type) {
	case string:
		return val, nil
	default:
		return "", errors.New("bad string")
	}
}

func main() {
	switch *role {
	case ROLE_CLIENT:
		client()
	case ROLE_SERVER:
		server()
	default:
		log.Fatalf("bad role '%s', expecting %s or %s\n",
			*role, ROLE_CLIENT, ROLE_SERVER)
	}
}
