VERSION := $(shell cat version.txt)

build-dist:
	env GOOS=linux GOARCH=amd64 go build -o dist/linux ./src
	env GOOS=windows GOARCH=amd64 go build -o dist/windows ./src
	env GOOS=darwin GOARCH=amd64 go build -o dist/macos ./src
.PHONY: build-dist

tag:
	git tag -a v$(VERSION) -m "Version $(VERSION)"
	git push --tags
.PHONY: tag

run-local:
	./run.sh $(args)
.PHONY: run-local

