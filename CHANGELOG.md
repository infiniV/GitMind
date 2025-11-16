# Changelog

All notable changes to GitMind will be documented in this file.

## [0.1.0] - 2024-11-17

### Added
- Initial release of GitMind
- AI-powered commit message generation using Cerebras
- Interactive TUI with Bubble Tea for commit decisions
- Smart branching recommendations
- Free tier optimization with context reduction
- Configuration management system
- Support for conventional commits format
- Comprehensive error handling and user-friendly messages

### Fixed (Since Initial Test)
- **API Integration**: Fixed Cerebras API request format
  - Changed `max_tokens` to `max_completion_tokens` per latest API docs
  - Added `additionalProperties: false` to JSON schema for strict mode
  - Improved error messages with detailed debugging information

### Project Structure
- Reorganized directory structure:
  - `bin/` - Compiled binaries
  - `docs/` - Documentation (QUICKSTART, TESTING)
  - `scripts/` - Build and utility scripts
  - `cmd/` - CLI entry points
  - `internal/` - Application code (domain, adapters, use cases, UI)
  - `test/` - Test files and fixtures

### Technical Details
- Clean Architecture with DDD principles
- Domain layer: 97% test coverage
- Git adapter: 70% test coverage
- Production-ready error handling
- Retry logic with exponential backoff
- Token-efficient diff summarization
- Multi-provider AI abstraction (extensible)

## Known Limitations (MVP)
- No push/pull integration yet
- No diff preview in TUI
- Simple config format (will upgrade to JSON)
- Windows-only tested (cross-platform coming)
