run:
	go run main.go
build:
	go build -ldflags "-w -s" -o pipe main.go
fmt:
	go fmt
