.PHONY: build
build:
	go build -o bin/main main.go

.PHONY: dev
dev: build
	mv bin/main ~/go/bin/
