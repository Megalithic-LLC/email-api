.PHONY: compile
compile:
	go install ./...

dependencies:
	go get -u github.com/golang/dep/cmd/dep
	go get -u github.com/rakyll/statik
	dep ensure -v
	go get -u github.com/golang/protobuf/protoc-gen-go

generate:
	statik -src=privatefs
	go generate ./agentstreamendpoint/...
