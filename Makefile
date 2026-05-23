.PHONY: build package serve deploy clean

BIN := .bin

# ─── build single binary (Linux amd64 — same arch as Lambda) ─────────────────

build:
	@mkdir -p $(BIN)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(BIN)/bootstrap .
	@chmod +x $(BIN)/bootstrap
	@echo "✓ binary built"

# ─── deployment zip (bootstrap + bundled src/posts markdown) ─────────────────

package: build
	@python3 scripts/mkzip.py
	@echo "✓ deployment zip ready at $(BIN)/bootstrap.zip"

# ─── local development ────────────────────────────────────────────────────────

serve:
	SERVE_LOCAL=true go run .

# ─── deploy to AWS ────────────────────────────────────────────────────────────

deploy: package
	npx serverless deploy --aws-profile thali

# ─── housekeeping ─────────────────────────────────────────────────────────────

setup:
	go mod tidy
	npm install

clean:
	rm -rf $(BIN)
