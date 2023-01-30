.PHONY: install-tools
install-tools:
	go mod download
	go install ./tools/pm

.PHONY: README.md
README.md:
	pm template-file docs/README.md.tmpl README.md

.PHONY: test
test:
	go test ./...

.PHONY: analyze
analyze:
	golangci-lint run ./...
