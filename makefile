test:
	go test -coverprofile=coverage.out $(shell go list ./... | grep -vE '/pb|/mocks')
	go tool cover -func=coverage.out
