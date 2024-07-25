# Check to see if we can use ash, in Alpine images, or default to BASH.
SHELL_PATH = /bin/ash
SHELL = $(if $(wildcard $(SHELL_PATH)),/bin/ash,/bin/bash)

# ==============================================================================
# Games

connect:
	go run cmd/connect/main.go

connect-train:
	go run cmd/connect/main.go --train true

# ==============================================================================
# Examples

example1:
	go run examples/example1/main.go

example2:
	go run examples/example2/main.go

example3:
	go run -exec "env DYLD_LIBRARY_PATH=$$GOPATH/src/github.com/ardanlabs/ai-training/foundation/word2vec/libw2v/lib" examples/example3/main.go

example4:
	go run examples/example4/main.go

example5:
	go run examples/example5/main.go

example6:
	go run examples/example6/main.go

example7:
	go run examples/example7/main.go

# ==============================================================================
# Install dependencies

install:
	brew install mongosh

docker:
	docker pull mongodb/mongodb-atlas-local
	docker pull ollama/ollama
	docker pull dyrnq/open-webui:main

# ==============================================================================
# Manage project

dev-up:
	docker compose -f zarf/docker/compose.yaml up

dev-down:
	docker compose -f zarf/docker/compose.yaml down

dev-logs:
	docker compose logs -n 100

download-data:
	curl -o zarf/data/example3.gz -X GET http://snap.stanford.edu/data/amazon/productGraph/categoryFiles/reviews_Cell_Phones_and_Accessories_5.json.gz \
	&& gunzip -k -d zarf/data/example3.gz \
	&& mv zarf/data/example3 zarf/data/example3.json

clean-data:
	go run cmd/cleaner/main.go

mongo:
	mongosh -u ardan -p ardan mongodb://localhost:27017

ollama-pull:
	ollama pull mxbai-embed-large
	ollama pull llama3.1

openwebui:
	open -a "Google Chrome" http://localhost:3000/

# ==============================================================================
# Modules support

tidy:
	go mod tidy
	go mod vendor

deps-upgrade:
	go get -u -v ./...
	go mod tidy
	go mod vendor
