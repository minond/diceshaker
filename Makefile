PIUSER ?= pi
PIADDR ?= 192.168.1.4

arm:
	GOOS=linux GOARCH=arm GOARM=5 go build

pi: arm
	scp diceshaker $(PIUSER)@$(PIADDR):/home/$(PIUSER)/diceshaker
