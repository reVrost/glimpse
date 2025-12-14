# Changelog

All notable changes to Glimpse will be documented in this file.

## [Unreleased]

### Added
- Z.AI integration with GLM-4.6 support
- Go installable distribution via `go install github.com/revrost/glimpse@latest`
- Version flag support (`-version`)
- Automated GitHub releases with binary artifacts
- Makefile for building and development

### Changed
- Improved documentation with Z.AI setup instructions
- Enhanced configuration examples

## [0.1.0] - 2025-12-14

### Added
- Initial release of Glimpse
- Real-time file watching with debouncing
- Git diff capture for code changes
- Log tailing integration
- OpenAI API integration
- YAML configuration system
- Streaming LLM responses
- Stdin chat mode (TODO)

### Core Features
- File system monitoring with pattern matching
- Structured log parsing and context extraction
- Multi-provider LLM architecture
- Configuration-driven operation
- Cross-platform support

---

## How to Upgrade

### From source to go install
1. Remove old binary: `rm $(which glimpse)` (if you built manually)
2. Install new version: `go install github.com/revrost/glimpse@latest`

### Between versions
```bash
go install github.com/revrost/glimpse@latest
```