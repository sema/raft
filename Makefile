
ALL_PACKAGES := $(shell go list ./... | grep -v mocks)
PBSRCS       := $(shell find . -name "*.proto" -type f | grep -v vendor)
PBOBJS       := $(patsubst %.proto, %.pb.go, $(PBSRCS))

build: ${PBOBJS}
	@echo $(ALL_PACKAGES) | xargs -n1 go build

lint:
	go fmt $(ALL_PACKAGES)
	go vet $(ALL_PACKAGES)
	dep check

test:
	go test --coverprofile=cover.out $(ALL_PACKAGES)
	go tool cover -func=cover.out

mocks:
	go generate $(ALL_PACKAGES)

%.pb.go: %.proto
	protoc $< --go_out=plugins=grpc:.
