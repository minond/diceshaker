PIUSER ?= pi
PIADDR ?= 192.168.1.4
SERVUSER ?= marcos
SERVADDR ?= minond.xyz

CERTFILE ?= cert.pem
KEYFILE ?= key.pem
BINARYFILE ?= diceshaker

FILES = $(BINARYFILE) $(CERTFILE) $(KEYFILE)

build: arm amd64

deploy: cert server pi

amd64: deps
	GOOS=linux GOARCH=amd64 go build

arm: deps
	GOOS=linux GOARCH=arm GOARM=5 go build

deploy-pi: arm
	scp $(FILES) $(PIUSER)@$(PIADDR):~/diceshaker/

deploy-server:
	scp $(FILES) $(SERVUSER)@$(SERVADDR):~/diceshaker/

deploy-systemd: arm
	scp diceshaker-client.service $(PIUSER)@$(PIADDR):~/diceshaker/
	scp diceshaker-server.service $(SERVUSER)@$(SERVADDR):~/diceshaker/

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
