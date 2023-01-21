.PHONY: generate
generate:
	go generate ./...

.PHONY: test
test:
	go test ./... -coverprofile=coverage_temp.out
	cat coverage_temp.out | grep -v enumer > coverage.out
	go tool cover -func=coverage.out