.PHONY: build run clean install-deps setup-wg

build:
	go build -o bin/p2nova-vpn cmd/api/main.go

run:
	sudo ./bin/p2nova-vpn

dev:
	go run cmd/api/main