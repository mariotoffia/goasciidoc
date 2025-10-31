install:
	@go install -v .
test:
	@go test -v ./... -cover