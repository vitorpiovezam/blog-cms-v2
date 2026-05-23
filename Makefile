.PHONY: build package serve offline deploy clean setup

BIN := .bin

build:
	@mkdir -p $(BIN)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(BIN)/bootstrap .
	@chmod +x $(BIN)/bootstrap
	@echo "✓ binary built"

package: build
	@python3 scripts/mkzip.py
	@echo "✓ deployment zip ready at $(BIN)/bootstrap.zip"

serve:
	SERVE_LOCAL=true go run .

offline: serve

deploy: package
	serverless deploy --aws-profile thali

setup:
	go mod tidy

clean:
	rm -rf $(BIN)
