all:
	go test -coverprofile=c.out
	go run main.go > c.modified.out
	go tool cover -html=c.modified.out
