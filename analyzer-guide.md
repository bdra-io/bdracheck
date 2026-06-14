# The `bdracheck` Engineering & Usage Guide

Architectural design principles are only as strong as the tooling that enforces them. This guide breaks down the internal mechanics, structural features, and operational blueprints of `bdracheck`—the concurrent, AST-driven architecture linter engineered to safeguard your codebase boundaries automatically.

---

## 🛠️ Under the Hood: How the Analyzer is Built

`bdracheck` is built from the ground up to replace fragile, slow string-matching script sweeps with a true compiler-grade static analysis engine. Here is how its internal architecture functions:

### 1. True AST Parsing vs. Flaky Regex Matching
Most lightweight code checkers use regular expressions to scan for banned strings or phrases. That approach easily breaks when a developer formats imports across multiple lines, uses aliased package namespaces, or includes commented-out code blocks. 

`bdracheck` utilizes Go’s native compiler utilities (`go/parser` and `go/ast`). It reads the code tokens directly, stripping away whitespace and layout variations to convert the source into an immutable, structured mathematical tree. The engine iterates over exact `ImportSpec` structural nodes, meaning it evaluates what the compiler *actually reads*, with absolute precision.

```text
[ Raw Go Code Stream ] ──► [ go/parser Engine ] ──► [ AST Structural Node Graph ] ──► [ Boundary Rules Evaluation ]
```

### 2. High-Performance Concurrent Worker Pipeline
Iterating through thousands of domain files sequentially creates an extreme performance bottleneck during local pre-commit hooks or fast-moving CI/CD pipelines.

`bdracheck` implements a concurrent multi-threaded worker model:
* A file walker harvests target paths inside the `internal/` folder.
* The orchestrator assigns each file target to an independent goroutine spinning up across your machine's CPU cores.
* Violations are securely aggregated through a thread-safe synchronized channel (`chan Violation`).
* A coordination lock (`sync.WaitGroup`) controls the execution lifecycle, allowing the tool to evaluate massive enterprise codebases in milliseconds.

### 3. Strict Token-Line Comment Mapping
To prevent developers from easily disabling architectural rules using broad, file-wide suppression keys, `bdracheck` forces inline accountability. By reading the code file's native spatial coordinates (`token.FileSet`), it tracks the exact file line of every single import declaration statement. An ignore directive (`// bdracheck:ignore`) is honored *only* if it sits on the exact same line as the targeted violation.

---

## 💎 Core Enforcement Features

The tool maps your file paths directly against the rules defined in `bdracheck.json` to enforce three structural invariants:

* **Pure Layer Zero-Dependency Isolation:** Scans all files located inside `*/pure/` directories. It strictly blocks external framework libraries, network drivers (`net/http`), database adapters (`database/sql`), and file system tools (`os`, `io`), ensuring your core domain business math is completely deterministic and fast to test without mocks.
* **Protected Layer Data Lock:** Audits your API and communication contract files (`*/protected/`). It confirms that database drivers or routing handles do not bleed backwards into your shared contracts, keeping your ring boundaries completely decoupled from third-party framework code.
* **Concentric Inward-Flow Rule:** Evaluates dependencies across your domain rings. It maps the import strings to guarantee that outer, reactive rings (like Ring 2 Analytics) safely import inner transactional cores, but inner domains (like Ring 0 Identity) never accidentally depend on or import anything sitting outside their perimeter.

---

## 🚀 How to Configure & Use `bdracheck`

### 1. Step 1: Initialize Your Configuration Matrix
Place a `bdracheck.json` manifest file directly in the root directory of your application workspace. This tells the static engine how to parse your package layout:

```json
{
  "projectName": "my-monolith",
  "architectureStrategy": "BDRA-Lite",
  "rules": {
    "disallowExternalImportsInPure": {
      "targetDirs": ["internal/ring*/pure/**/*.go"]
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

### 2. Step 2: Install the Executable
Compile and install the tool globally on your system or remote build container using Go's official distribution channel:

```bash
go install [github.com/bdra-io/bdracheck@latest](https://github.com/bdra-io/bdracheck@latest)
```

### 3. Step 3: Execute the Verification Pass
Run the verification command from your project's root folder where your config manifest lives:

```bash
bdracheck verify
```

If you need to direct the engine to a configuration file residing in a custom directory workspace path, use the `--config` option flag:
```bash
bdracheck verify --config=deployments/linters/bdracheck.json
```

---

## 🛡️ Utilizing Inline Escape Valves

If a complex production scenario forces a brief exception—such as linking an external encoding helper inside a domain validation sequence—you can easily prevent global build failures by writing an inline comment directive.

Append the `// bdracheck:ignore` token directly onto the end of the targeted import statement:

```go
package pure

import (
	"encoding/json" // bdracheck:ignore
	"errors"
)
```

The AST compiler engine will identify the comment block token aligned to that precise file line number, log the deliberate choice, and skip past evaluation constraints without stopping your continuous integration deployment gates.