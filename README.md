# GitMind (gm)

An AI-powered Git CLI manager that makes complex Git workflows simple through intelligent automation.

## ✨ Features

- **AI-Generated Commit Messages**: Analyzes your changes and generates meaningful, contextual commit messages
- **Smart Workflow Decisions**: Interactive TUI helps you decide between direct commits and feature branches
- **Free Tier Friendly**: Optimized for free API tiers with graceful degradation and token reduction
- **Production Ready**: Clean architecture, comprehensive tests, and robust error handling
- **Windows Support**: Compiled binary ready to use

## 🚀 Quick Start

**See [QUICKSTART.md](QUICKSTART.md) for detailed setup instructions!**

### TL;DR

1. **Get a free Cerebras API key**: https://cloud.cerebras.ai/
2. **Build the binary**:
   ```bash
   go build -o bin/gm.exe ./cmd/gm
   # Or use the build script: scripts\build.bat
   ```
3. **Configure GitMind**:
   ```bash
   bin\gm.exe config
   ```
4. **Use it**:
   ```bash
   # Make some changes to your code...
   bin\gm.exe commit
   ```

## 📦 Installation

### Option 1: Build from Source (Current)
```bash
git clone <this-repo>
cd gitman
go build -o gm.exe ./cmd/gm
```

### Option 2: Download Release (Coming Soon)
Pre-built binaries will be available in GitHub Releases.

### Option 3: Package Managers (Future)
```bash
# Scoop (planned)
scoop bucket add gitman https://github.com/yourusername/scoop-gitman
scoop install gitman

# Chocolatey (planned)
choco install gitman
```

## Architecture

Built with Clean Architecture and Domain-Driven Design principles:

- **Domain Layer**: Core business logic (commits, repositories, decisions)
- **Use Case Layer**: Application workflows (analyze, commit, configure)
- **Adapter Layer**: External integrations (Git, AI providers, config)
- **UI Layer**: Bubble Tea TUI components

## Development

### Prerequisites
- Go 1.21+
- Git 2.30+

### Build
```bash
go build -o gm ./cmd/gm
```

### Test
```bash
go test -v ./...
go test -coverprofile=coverage.out ./...
```

### Project Structure
```
gitman/
├── cmd/gm/              # CLI entry point
├── internal/
│   ├── domain/          # Business entities
│   ├── usecase/         # Application logic
│   ├── adapter/         # External services
│   └── ui/              # TUI components
├── test/                # Integration tests
└── configs/             # Configuration templates
```

## License

MIT License - see LICENSE file for details

## Contributing

Contributions welcome! Please read CONTRIBUTING.md first.
