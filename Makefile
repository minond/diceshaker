PIUSER ?= pi
PIADDR ?= 192.168.1.4
SERVUSER ?= marcos
SERVADDR ?= minond.xyz

arm:
	GOOS=linux GOARCH=arm GOARM=5 go build

pi: arm
	scp diceshaker cert.pem key.pem $(PIUSER)@$(PIADDR):~/diceshaker/

server:
	scp cert.pem key.pem $(SERVUSER)@$(SERVADDR):~/diceshaker/

cert:
	@if [ ! -f cert.pem ]; then \
	openssl req -newkey rsa:2048 -nodes -subj "/CN=$(SERVADDR)" -x509 -days 3650 \
		-keyout key.pem \
		-out cert.pem; \
	else \
		echo "detected cert/key file, skipping"; \
	fi
