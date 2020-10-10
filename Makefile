GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOTIDY=$(GOMOD) tidy

PROTOC=protoc -I ./proto --go_out=. --go_opt=paths=source_relative 

logentry.pb.go: ./proto/logentry.proto 
	$(PROTOC) ./proto/logentry.proto 

error.pb.go: ./proto/error.proto 
	$(PROTOC) ./proto/error.proto 


proto-build: logentry.pb.go error.pb.go


proto-clean: 
	find . -iname "*.pb.go" -exec rm -f {} \;

clean:
	$(GOCLEAN)

clean-all: clean proto-clean

tidy:
	$(GOTIDY)


build: proto-build tidy

test: proto-build tidy
	$(GOTEST) -v ./...
