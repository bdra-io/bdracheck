# bdracheck 🚀

> "If an architectural boundary can be bypassed by a junior engineer on a Friday afternoon, it isn't a boundary, it's a suggestion."

`bdracheck` is the automated static analysis engine and architecture guardian for the **BDRA (Business-Domain Ring Architecture)** ecosystem. By analyzing Go source streams directly at the Abstract Syntax Tree (AST) level, it mathematically intercepts code rot, illegal cross-package imports, and architectural leaks before they reach your compilation binary.

Unlike generic code style linters or coarse text-matching grep routines, `bdracheck` maps your filesystem topology directly against structural domain rules to keep your core business logic pristine, decoupled, and 100% testable.

---

## 🏛️ How It Works: AST vs. Text Parsing

Traditional import checkers rely on regular expressions or file string searching. This approach falls apart when handling complex formatting, nested blocks, multi-line aliased imports, or conditional build tags. 

`bdracheck` parses the target Go files using Go's compiler toolchain (`go/parser`, `go/ast`). It breaks the code down into an unambiguous, syntactic token tree. It extracts the raw, compiled import specification arrays directly from the file node, evaluating boundaries with mathematical accuracy.

```text
[ Raw Go Code Stream ] ──► [ go/parser Engine ] ──► [ AST Structural Node Graph ] ──► [ Boundary Rules Evaluation ]
```

---

## ✨ Key Features

- **Pure Layer Zero-Dependency Isolation**: Enforces that files living in any `*/pure/` directory contain absolute zero third-party framework or I/O imports, safeguarding your high-velocity unit testing matrix.
- **Protected Layer Data Lock**: Verifies that your contract boundaries (`*/protected/`) are strictly free from raw infrastructural dependencies (`database/sql`, `net/http`, etc.).
- **Concentric Inward-Flow Enforcement**: Evaluates cross-ring references to guarantee that outer analytical rings can access inner rings, but inner transactional loops can never depend on components outside their domain.
- **Zero Overhead**: Operates entirely out-of-band by analyzing compilation tokens. It processes complex, large-scale modular codebases in milliseconds.

---

## 🛠️ Installation

To install the executable engine globally on your system or build agent workspace, execute the standard remote compile command:

```bash
go install [github.com/bdra-io/bdracheck@latest](https://github.com/bdra-io/bdracheck@latest)
```

---

## ⚙️ Configuration File Schema (`bdracheck.json`)

The engine requires a `bdracheck.json` metadata file located directly in the root directory of the target project folder. This maps your modular namespace layout and outlines authorized dependency workflows.

```json
{
  "$schema": "[https://raw.githubusercontent.com/bdra-io/bdra-spec/main/schemas/v1/linter.json](https://raw.githubusercontent.com/bdra-io/bdra-spec/main/schemas/v1/linter.json)",
  "projectName": "modular-monolith",
  "architectureStrategy": "BDRA-Lite",
  "rules": {
    "disallowExternalImportsInPure": {
      "targetDirs": ["internal/ring*/pure/**/*.go"],
      "allowedPrefixes": ["modular-monolith/internal/ring*/pure"]
    },
    "disallowIOInProtected": {
      "targetDirs": ["internal/ring*/protected/**/*.go"],
      "forbiddenPackages": ["database/sql", "net/http", "os", "io"]
    },
    "enforceInwardDependencyFlow": {
      "rings": [
        { "id": "ring0", "path": "internal/ring0" },
        { "id": "ring1", "path": "internal/ring1", "allowedDependencies": ["ring0"] },
        { "id": "ring2", "path": "internal/ring2", "allowedDependencies": ["ring0", "ring1"] }
      ]
    }
  }
}
```

---

## 🚀 Usage

Execute verification passes by targeting the directory containing your configuration rules manifest:

### 1. Run Analysis Locally

Navigate to the root directory of your project (where `bdracheck.json` resides) and run:

```bash
bdracheck verify
```

### 2. Specify a Custom Directory Workspace Path

```bash
bdracheck verify --config=/path/to/your/bdracheck.json
```

---

## 🤖 Continuous Integration Gate Setup

To turn abstract architecture goals into an unbreakable automation pipeline gate, integrate `bdracheck` directly inside your automated GitHub Actions CI/CD workflows:

Create a file named `.github/workflows/bdra-lint.yml` inside your application repository:

```yaml
name: BDRA Architecture Governance Gate

on:
  push:
    branches: [ main, master ]
  pull_request:
    branches: [ main, master ]

jobs:
  governance:
    name: Enforce Ring and Layer Boundaries
    runs-on: ubuntu-latest
    
    steps:
    - name: Checkout Code Repository
      uses: actions/checkout@v4

    - name: Set Up Go Runtime Environment
      uses: actions/setup-go@v5
      with:
        go-version: '1.22'

    - name: Install BDRA Static Analysis Linter Engine
      run: go install [github.com/bdra-io/bdracheck@latest](https://github.com/bdra-io/bdracheck@latest)

    - name: Execute AST Invariant Verification
      run: bdracheck verify
```

---

## 📋 Exit Code Rules & Verification Output Matrix

The binary sets standard shell exit values for clean pipeline integration:


| Exit Code | Evaluation Meaning                                                 | Pipeline Status          |
| --------- | ------------------------------------------------------------------ | ------------------------ |
| `**0**`   | System clean. Zero layer or ring dependency violations identified. | ✅ Pass (Build Continues) |
| `**1**`   | Structural boundary bypass intercepted. Breakdown printed to logs. | 🛑 Fail (Build Stopped)  |


---

## 🎨 Credits & Licensing

- **Authorship**: Core parsing framework managed by the [BDRA Open Source Initiative](https://github.com/bdra-io).
- **License**: Open-source software distributed under the terms of the **Apache 2.0 License**.

