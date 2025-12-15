# Glimpse: AI-Powered Micro-Reviewer

**Glimpse** is a lightweight, local daemon that watches your active repository. When you stage your changes the git diff is captured and fed to an LLM, and which it will provide a concise engineering review to stdout.

It is designed for "Vibecoding"—maintaining flow state while getting instant feedback on correctness.

## 🚀 Installation

### Go Install (Recommended)

```bash
go install github.com/revrost/glimpse@latest
```

This installs `glimpse` to your `GOPATH/bin` (usually `~/go/bin/`). Make sure `~/go/bin` is in your `PATH`.


## Features

- **Git diff capture** for staged and unstaged changes
- **Focused reviews** for bugs, performance, and security issues
- **Nice stdout interface**

## Quick Start

1. **Configure API Key**
   ```bash
   export OPENAI_API_KEY="your-api-key-here"
   # or for Gemini
   export GEMINI_API_KEY="your-gemini-key-here"
   # or for Z.AI
   export ZAI_API_KEY="your-zai-key-here"
   # or for CLAUDE
   export CLAUDE_API_KEY="your-claude-key-here"
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

3. **Start Glimpse**
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
1. Environment variables: `OPENAI_API_KEY`, `GEMINI_API_KEY`, `ZAI_API_KEY`, `CLAUDE_API_KEY`
2. Configuration file: `api_key` field in llm section

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


## License

MIT License - see LICENSE file for details.

