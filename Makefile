# Define the server executable name
SERVER_EXEC = server
# Define the wasm output file
WASM_FILE = static/main.wasm
# Define the static directory
STATIC_DIR = static

# Default target when `make` is run
all: run

# Run the server
run: build
	@echo "🚀 Starting server on http://localhost:8080"
	@./$(SERVER_EXEC)

# Build all necessary components
build: build-server build-wasm copy-wasm-exec

# Build the Go web server
build-server: server.go
	@echo "🛠️  Building server..."
	@go build -o $(SERVER_EXEC) server.go

# Build the Go code into a WebAssembly binary
build-wasm: main.go
	@echo "📦 Compiling Go to WASM..."
	@mkdir -p $(STATIC_DIR)
	@GOOS=js GOARCH=wasm go build -o $(WASM_FILE) main.go

# Copy the wasm_exec.js file from the Go installation path
copy-wasm-exec:
	@echo "📋 Copying wasm_exec.js..."
	@mkdir -p $(STATIC_DIR)
	@cp "$(shell go env GOROOT)/lib/wasm/wasm_exec.js" $(STATIC_DIR)/

# Serve static files using Python's built-in HTTP server
serve-static: build-wasm copy-wasm-exec
	@echo "🌐 Starting static file server on http://localhost:8000"
	@echo "📁 Serving files from $(STATIC_DIR) directory"
	@python3 -m http.server 8000

# Clean up build artifacts
clean:
	@echo "🧹 Cleaning up..."
	@rm -f $(SERVER_EXEC)
	@rm -rf $(STATIC_DIR)

# Phony targets are not files
.PHONY: all run build build-server build-wasm copy-wasm-exec serve-static clean
