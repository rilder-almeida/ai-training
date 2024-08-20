# Check to see if we can use ash, in Alpine images, or default to BASH.
SHELL_PATH = /bin/ash
SHELL = $(if $(wildcard $(SHELL_PATH)),/bin/ash,/bin/bash)

# ==============================================================================
# Mongo support
#
# use connect4
#
# db.boards.deleteMany({})
#
# db.boards.deleteOne({ board_id: '1f76202b-79a2-4f8e-940b-bc90d29895dc' })
#
# db.boards.find({board_id: "fdb01ef2-42ec-4f75-918e-63197434927d"})

# ==============================================================================
# Production support
# https://chatty-api-3erkmojm3q-uc.a.run.app/v1/chat/completions
#
# {
#   "model": "gemini-1.5-flash-001",
#   "messages": [
#     {
#       "role": "user",
#       "content": "What is the capital of spain?"
#     }
#   ],
#   "temperature": 0.5,
#   "top_p": 0.65,
#   "stream": false
# }

# ==============================================================================
# Games

connect:
	go run cmd/connect/main.go

connect-train:
	go run cmd/connect/main.go --train true

connect-save:
	git add -A
	git commit -am "saving training data"
	git push
	rm log.txt

# ==============================================================================
# Examples

example1:
	go run cmd/examples/example1/main.go

example2:
	go run cmd/examples/example2/main.go

example3:
	go run -exec "env DYLD_LIBRARY_PATH=$$GOPATH/src/github.com/ardanlabs/ai-training/foundation/word2vec/libw2v/lib" cmd/examples/example3/main.go

example4:
	go run cmd/examples/example4/main.go

example5:
	go run cmd/examples/example5/main.go

example6:
	go run cmd/examples/example6/main.go

example7:
	go run cmd/examples/example7/main.go

# ==============================================================================
# Install dependencies

install:
	brew install mongosh
	brew install ollama

docker:
	docker pull mongodb/mongodb-atlas-local
	docker pull dyrnq/open-webui:main

# ==============================================================================
# Manage project

dev-up:
	docker compose -f zarf/docker/compose.yaml up

dev-down:
	docker compose -f zarf/docker/compose.yaml down

dev-logs:
	docker compose logs -n 100

dev-ollama-up:
	export OLLAMA_MODELS="zarf/docker/ollama/models" && \
	ollama serve

dev-ollama-logs:
	tail -f -n 100 ~/.ollama/logs/server.log

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
	ollama pull gemma2:27b

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
