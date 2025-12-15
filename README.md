# Glimpse: AI-Powered Micro-Reviewer

**Glimpse** is a lightweight, local daemon that watches your active repository. When files change, it captures the git diff and recent backend logs, feeds them to an LLM, and prints a concise engineering review to stdout.

It is designed for "Vibecoding"—maintaining flow state while getting instant feedback on correctness, security, and idiomatic Go patterns.

## 🚀 Installation

### Go Install (Recommended)

```bash
go install github.com/revrost/glimpse@latest
```

This installs `glimpse` to your `GOPATH/bin` (usually `~/go/bin/`). Make sure `~/go/bin` is in your `PATH`.

### Download Binaries

Grab pre-built binaries from the [Releases page](https://github.com/revrost/glimpse/releases):
- Linux: `glimpse-linux-amd64`
- macOS: `glimpse-darwin-amd64` (Intel) or `glimpse-darwin-arm64` (Apple Silicon)
- Windows: `glimpse-windows-amd64.exe`

```bash
# Make executable (Linux/macOS)
chmod +x glimpse-*
mv glimpse-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m) glimpse
```

### From Source

```bash
git clone https://github.com/revrost/glimpse
cd glimpse
make build  # or: go build -o glimpse .
```

## Features

- 🔄 **Real-time file watching** with intelligent debouncing
- 📝 **Git diff capture** for staged and unstaged changes
- 📋 **Log tailing** integration with structured slog output
- 🤖 **Multi-provider LLM support** (OpenAI, Z.AI, Gemini, Ollama)
- ⚙️ **Minimal YAML configuration** with sensible defaults
- 🎯 **Focused reviews** for bugs, performance, and security issues
- 🖥️ **Simple stdout interface** with optional chat mode

## Quick Start

1. **Configure API Key**
   ```bash
   export OPENAI_API_KEY="your-api-key-here"
   # or for Gemini
   export GEMINI_API_KEY="your-gemini-key-here"
   # or for Z.AI
   export ZAI_API_KEY="your-zai-key-here"
   ```

2. **Create Configuration** (`.glimpse.yaml` in repo root)
   ```yaml
   watch:
     - "./internal/**/*.go"
     - "./pkg/**/*.go"
   ignore:
     - "*_test.go"
   logs:
     file: "./tmp/server.log"
     lines: 50
   llm:
     provider: "openai"
     model: "gpt-4o"
     system_prompt: "You are a Principal Go Engineer. Review strictly for bugs, perf, and slog context."
   ```

3. **Start Your Application with Logging**
   ```bash
   # Pipe app output to a file
   go run . | tee tmp/server.log
   ```

4. **Start Glimpse**
   ```bash
   glimpse
   # or if using local build:
   ./glimpse
   ```


## Architecture
```
fsnotify → bounded batcher → event‑scoped diffs → async LLM
                ↓
         safe timers + shutdown
```


## Managing Your Installation

### Update to Latest Version
```bash
go install github.com/revrost/glimpse@latest
```

### Check Version
```bash
glimpse -version
```

### Uninstall
```bash
rm $(which glimpse)
```

5. **Make Code Changes**

   Glimpse will automatically detect changes and provide reviews:

   ```
   --- Reviewing: internal/handlers/payment.go ---
   Analyzing with openai (gpt-4o)...

   🔍 **Code Review**:

   **Bug Risk**: Line 45 - Potential nil pointer dereference if payment.Customer is nil
   **Performance**: Consider using sync.Pool for频繁创建的 structs
   **Security**: Validate payment.Amount before processing to avoid overflow

   **Logs Context**: Recent errors show payment validation failures at 2023-12-07T14:23:11Z
   ```

## Configuration

Glimpse uses a `.glimpse.yaml` file in the repository root. If not found, it uses sensible defaults.

### Watch Configuration

```yaml
watch:
  - "./internal/**/*.go"    # Watch Go files in internal directory
  - "./pkg/**/*.go"         # Watch Go files in pkg directory
  - "./src/**/*.rs"         # Watch Rust files
ignore:
  - "*_test.go"             # Ignore test files
  - "*.generated.go"         # Ignore generated files
  - "vendor/**"             # Ignore vendor directory
```

### Log Configuration

```yaml
logs:
  file: "./tmp/server.log"   # Log file to tail
  lines: 50                 # Number of recent lines to include
```

### LLM Configuration

```yaml
llm:
  provider: "openai"         # openai, gemini, ollama, zai
  model: "gpt-4o"          # Model to use
  api_key: "optional-key"    # Can use env vars instead
  system_prompt: "You are a Principal Go Engineer. Review for bugs, performance, and security."
```

### API Keys

API keys can be provided via:
1. Environment variables: `OPENAI_API_KEY`, `GEMINI_API_KEY`, `ZAI_API_KEY`
2. Configuration file: `api_key` field in llm section

#### Z.AI Setup

Z.AI provides access to GLM models including GLM-4.6:

```yaml
llm:
  provider: "zai"
  model: "glm-4.6"          # GLM-4.6, glm-4-air, etc.
```

Set your API key:
```bash
export ZAI_API_KEY="your-zai-key-here"
```

Available Z.AI models:
- `glm-4.6` - High-performance model
- `glm-4-air` - Lightweight, efficient model
- `glm-4-32b` - Large context model

## Integration Workflow

### The Unix Way

Glimpse follows Unix philosophy - do one thing well and compose with other tools:

```bash
# Terminal 1: Run your application
go run . | tee debug.log

# Terminal 2: Run Glimpse
./glimpse

# Terminal 3: Work on your code
vim internal/handlers/payment.go
```

### Log Integration

Glimpse works excellently with structured slog output:

```go
// In your application
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

logger.Error("Payment processing failed",
    "customer_id", customerID,
    "amount", amount,
    "error", err,
)
```

Glimpse captures these structured logs and the LLM can parse them to correlate errors with your code changes.

## LLM Providers

### OpenAI

```yaml
llm:
  provider: "openai"
  model: "gpt-4o"
  # Supports: gpt-4, gpt-4-turbo, gpt-3.5-turbo
```

### Gemini (Coming Soon)

```yaml
llm:
  provider: "gemini"
  model: "gemini-pro"
  # TODO: Implementation in progress
```

### Ollama (Coming Soon)

```yaml
llm:
  provider: "ollama"
  model: "llama2"
  # TODO: Local model support
```

## Context Window Strategy

Glimpse sends a structured prompt to the LLM:

```
SYSTEM: [System Prompt]

CONTEXT:
1. File: internal/handlers/payment.go
2. Git Diff:
   [Output of git diff --unified=0 internal/handlers/payment.go]

3. Recent Runtime Logs (tail -n 50):
   [Raw content of server.log]

TASK:
Review the diff. If the logs show errors related to this logic, highlight them immediately.
Be concise.
```

## Development

### Building

```bash
go build -o glimpse
```

### Testing

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./config
go test ./watcher

# Run with coverage
go test -cover ./...
```

### Project Structure

```
glimpse/
├── main.go              # Entry point and orchestration
├── config/              # Configuration parsing
├── watcher/             # File system watching with debouncing
├── git/                 # Git operations (diff capture)
├── logs/                # Log tailing functionality
├── llm/                 # LLM client interface
└── test/                # Integration tests
```

## Troubleshooting

### No File Events Detected

- Ensure watch patterns match your file paths
- Check that directories exist when Glimpse starts
- Use `**` for recursive patterns

### LLM API Errors

- Verify API key is set correctly
- Check API quota and billing
- Try different model if rate limited

### No Logs Available

- Ensure your application is piping output to the configured log file
- Check file permissions
- Verify log file path in configuration

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes with tests
4. Run `go test ./...` to ensure all tests pass
5. Submit a pull request

## License

MIT License - see LICENSE file for details.

## Roadmap

- [ ] Gemini API integration
- [ ] Ollama local model support
- [ ] Stdin chat mode for follow-up questions
- [ ] Configuration hot-reloading
- [ ] Webhook integration for CI/CD
- [ ] VS Code extension
- [ ] IntelliJ plugin

---

**Built with ❤️ for developers who value flow state and code quality.**
