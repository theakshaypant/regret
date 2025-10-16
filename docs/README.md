# regret - Documentation

Complete documentation for the **regret** library and CLI tool.

---

## 📚 Documentation Index

### Getting Started

- **[WHEN_TO_USE.md](WHEN_TO_USE.md)** - When and who should use regret
  - Decision tree for evaluating if regret is right for you
  - User personas and specific scenarios
  - Integration patterns and examples
  - Common questions and answers

- **[GETTING_STARTED.md](GETTING_STARTED.md)** - Installation, basic usage, and configuration
  - Library installation and quick start
  - CLI tool setup
  - Common usage patterns
  - Configuration options

### Reference Guides

- **[API.md](API.md)** - Complete API reference
  - All public functions
  - Type definitions
  - Configuration options
  - Performance characteristics

- **[CLI.md](CLI.md)** - Command-line tool reference
  - All commands and subcommands
  - Flags and options
  - Output formats
  - Usage examples

### Understanding regret

- **[HOW_IT_WORKS.md](HOW_IT_WORKS.md)** - Detection algorithms and theory
  - Catastrophic backtracking explained
  - Multi-layered detection approach
  - NFA analysis and formal methods
  - EDA/IDA detection algorithms
  - Pump pattern generation
  - Academic foundations

### Technical Documentation

- **[ARCHITECTURE.md](ARCHITECTURE.md)** - Implementation details
  - Package structure
  - Implementation phases
  - Key algorithms
  - Performance targets
  - Internal design decisions


---

## 📖 Reading Guide

### New Users

Start here to get up and running:

1. [../README.md](../README.md) - Project overview and quick start
2. [WHEN_TO_USE.md](WHEN_TO_USE.md) - Decide if regret is right for you
3. [GETTING_STARTED.md](GETTING_STARTED.md) - Detailed setup and usage
4. [API.md](API.md) or [CLI.md](CLI.md) - Depending on how you'll use regret

### Understanding the Internals

To learn how regret works:

1. [HOW_IT_WORKS.md](HOW_IT_WORKS.md) - Detection algorithms explained
2. [ARCHITECTURE.md](ARCHITECTURE.md) - Technical implementation

---

## 📋 Quick Links

| Topic | Documentation |
|-------|--------------|
| **Should I use regret?** | [WHEN_TO_USE.md](WHEN_TO_USE.md) |
| **Installation** | [GETTING_STARTED.md](GETTING_STARTED.md#installation) |
| **Quick Start** | [../README.md](../README.md#quick-start) |
| **API Functions** | [API.md](API.md#core-functions) |
| **CLI Commands** | [CLI.md](CLI.md#commands) |
| **Validation Modes** | [GETTING_STARTED.md](GETTING_STARTED.md#configuration) |
| **Complexity Scoring** | [HOW_IT_WORKS.md](HOW_IT_WORKS.md#complexity-scoring) |
| **EDA/IDA Detection** | [HOW_IT_WORKS.md](HOW_IT_WORKS.md#layer-2-nfa-analysis) |
| **Pump Patterns** | [HOW_IT_WORKS.md](HOW_IT_WORKS.md#layer-3-adversarial-testing) |
| **Performance** | [API.md](API.md#performance-characteristics) |
| **Examples** | [../examples/](../examples/) |

---

## 🔄 Documentation Maintenance

### Adding New Documentation

1. **Create file** in `docs/` directory
2. **Add entry** to this index
3. **Link from** relevant documents
4. **Update** reading guide if needed

### Documentation Standards

- Use clear, descriptive headings
- Add table of contents for long documents
- Include code examples with explanations
- Link between related documents
- Keep examples up-to-date with API changes

---

## 📝 Examples

See the [examples/](../examples/) directory for runnable code examples:

- Basic validation
- Complexity analysis
- Pump pattern generation
- Custom configuration
- Error handling

---
