install:
	@go install -v .
test:
	@go test -v ./... -cover

docs-goasciidoc:
	@go run main.go --type-links external -i -t --highlighter goasciidoc\
		-c "{\"author\": \"Mario Toffia\", \"email\": \"mario.toffia@xy.net\", \"web\": \"https://github.com/mariotoffia/goasciidoc\", \"images\": \"../meta/assets\", \"title\":\"Go Asciidoc Document Generator\", \"toc\": \"Table of Contents\", \"toclevel\": 3}"
docs:
	@go run main.go --type-links external -i -t \
		-c "{\"author\": \"Mario Toffia\", \"email\": \"mario.toffia@xy.net\", \"web\": \"https://github.com/mariotoffia/goasciidoc\", \"images\": \"../meta/assets\", \"title\":\"Go Asciidoc Document Generator\", \"toc\": \"Table of Contents\", \"toclevel\": 3}"
golden:
	@UPDATE_GOLDEN=1 go test ./asciidoc -run TestProducerGenerateGolden
