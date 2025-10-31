install:
	@go install -v .
test:
	@go test -v ./... -cover

golden:
	@UPDATE_GOLDEN=1 go test ./asciidoc -run TestProducerGenerateGolden