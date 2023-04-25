VERSION := $(shell cat version.txt)

fmt:
	go fmt ./...
.PHONY: fmt

vet:
	go vet ./...
.PHONY: vet

build-dist: fmt vet
	env GOOS=linux GOARCH=amd64 go build -o dist/linux ./src
	env GOOS=windows GOARCH=amd64 go build -o dist/windows ./src
	env GOOS=darwin GOARCH=amd64 go build -o dist/macos ./src
.PHONY: build-dist

tag:
	git tag --force -a v$(VERSION) -m "Version $(VERSION)"
	git push --force --tags
.PHONY: tag

run-local:
	./run.sh $(args)
.PHONY: run-local

