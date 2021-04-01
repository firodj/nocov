all:
	go test -coverprofile=c.out
	go run . -ignore-files 'ignore\.go' > c.modified.out
	go tool cover -html=c.modified.out
