#Basic makefile

default: build

build: clean vet
	@go build -o doctorkaa-go

doc:
	@godoc -http=:6060 -index

lint:
	@golint ./...

debug:
	@fresh

run: build
	./doctorkaa-go

test:
	@go test ./...

vet:
	@go vet ./...

clean:
	@rm -f ./doctorkaa-go
