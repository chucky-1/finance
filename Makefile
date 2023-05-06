build:
	go build -o ./.bin/finance ./main.go

run: build
	./.bin/finance