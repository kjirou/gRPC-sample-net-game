get:
	go get -v -t -d ./...

run-client:
	go run client-main.go

run-client-with-debug-mode:
	go run client-main.go -debug

test:
	go test -v ./...
