PIUSER ?= pi
PIADDR ?= 192.168.1.4
SERVUSER ?= marcos
SERVADDR ?= minond.xyz

CERTFILE ?= cert.pem
KEYFILE ?= key.pem
BINARYFILE ?= diceshaker

FILES = $(BINARYFILE) $(CERTFILE) $(KEYFILE)

build: arm

deploy: cert server pi

amd64: deps
	GOOS=linux GOARCH=amd64 go build

arm: deps
	GOOS=linux GOARCH=arm GOARM=5 go build

pi: arm
	scp $(FILES) $(PIUSER)@$(PIADDR):~/diceshaker/

server:
	scp $(FILES) $(SERVUSER)@$(SERVADDR):~/diceshaker/

deps:
	dep ensure

cert:
	@if [ ! -f $(CERTFILE) ]; then \
	openssl req -newkey rsa:2048 -nodes -subj "/CN=$(SERVADDR)" -x509 -days 3650 \
		-keyout $(KEYFILE) \
		-out $(CERTFILE); \
	else \
		echo "detected cert/key file, skipping"; \
	fi
