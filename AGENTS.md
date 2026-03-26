# AGENTS.md

Guidelines for AI agents working on this codebase.

## Project Overview

gitView is a single-binary Git repository visualizer with an interactive DAG canvas.
- **Backend**: Go 1.26+ with go-git/go-git/v5
- **Frontend**: Vanilla JavaScript (ES modules) + Tailwind CSS
- **Build tool**: Wave (see Wavefile)

## Build Commands

```bash
# Development (build + run with file watching)
wave dev

# Production build
wave build

# Run tests
wave test

# CI pipeline (CSS + test + build)
wave ci
```

### Direct Commands (without Wave)

```bash
# Build Go binary
go build -o gitView ./src

# Run all tests
go test ./src

# Run a single test by name
go test -run TestTopologicalSortLinearOrder ./src

# Run tests matching a pattern
go test -run "TestTopological.*" ./src

# Run tests with verbose output
go test -v ./src

# Compile Tailwind CSS
./tools/tailwindcss -c tailwind.config.js -i src/viewer/style/tailwind.css -o src/viewer/style/output.css
```

## Project Structure

```
src/
├── main.go          # Entry point, HTTP server
├── http.go          # Routes & handlers
├── graph.go         # Git graph building logic
├── git.go           # File change detection
├── lanes.go         # Lane assignment algorithm
├── topo.go          # Topological sort
├── topo_test.go     # Tests
├── types.go         # Data structures
└── viewer/          # Embedded frontend
    ├── index.html
    ├── jsEngine/    # JavaScript modules
    └── style/       # CSS files
```

## Go Code Style

### Imports
Group imports: standard library first, then third-party, separated by blank line.
```go
import (
    "bufio"
    "os"
    "time"

    "github.com/go-git/go-git/v5"
    "github.com/go-git/go-git/v5/plumbing"
)
```

### Naming
- Exported types: `PascalCase` (`CommitData`, `FileStat`, `Node`)
- Unexported functions: `camelCase` (`buildGraph`, `assignLanes`)
- Short loop variables: `c` for commit, `i`, `j` for indices
- Maps: descriptive names (`childrenOf`, `commitRefs`, `branchLaneNames`)

### Error Handling
- Early return pattern with `if err != nil`
- Return errors as second value
- Use `_` to ignore errors only when safe to do so
```go
if err != nil {
    return nil, err
}
```

### Comments
- Doc comments for exported types and complex functions
- Single-line comments for inline explanations
```go
// FileStat represents a changed file and its status
type FileStat struct { ... }
```

### Testing
- Test file naming: `*_test.go`
- Test function naming: `Test<FunctionName><Scenario>`
- Use `reflect.DeepEqual` for slice/map comparisons
- Use `t.Fatalf` for assertion failures
```go
func TestTopologicalSortLinearOrder(t *testing.T) {
    got := topologicalSortOptimized(commits)
    want := []string{"C", "B", "A"}
    if !reflect.DeepEqual(got, want) {
        t.Fatalf("unexpected order: got %v want %v", got, want)
    }
}
```

## JavaScript Code Style

### Imports (ES Modules)
```javascript
import { state, ui } from './state.js';
import { render } from './render.js';
```

### Naming
- Functions: `camelCase` (`fetchGraph`, `initControls`)
- Constants: `SCREAMING_SNAKE_CASE` (`NODE_RADIUS`, `LANE_COLORS`)
- DOM elements: `camelCase` (`panelBackdrop`, `detailPanel`)

### State Management
- Centralized state object in `state.js`
- Export mutable objects for UI elements

## Git Workflow (from .skills/git-skills.md)

### Commit Format
- Format: `type(scope): subject`
- Use present tense, imperative mood
- Keep subject under 50 characters

### Commit Types
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Formatting, no code change
- `refactor`: Code restructuring
- `test`: Adding tests
- `chore`: Maintenance tasks

### Branching
- Use kebab-case: `feature/add-zoom`, `fix/lane-colors`
- Keep branches short-lived
- Rebase before merging

## UI Guidelines (from .skills/ui-skills.md)

### Required
- Use Tailwind CSS defaults unless custom values exist
- Add `aria-label` to icon-only buttons
- Use `h-dvh` not `h-screen`
- Respect `safe-area-inset` for fixed elements
- Show errors next to where action happens

### Forbidden
- No animation unless explicitly requested
- No gradients unless explicitly requested
- No purple/multicolor gradients
- No glow effects as primary affordances
- Never animate layout properties (`width`, `height`, `top`, `left`)
- Never exceed 200ms for interaction feedback

### Preferred
- Use `size-*` for square elements instead of `w-*` + `h-*`
- Use `truncate` or `line-clamp` for dense UI
- Use fixed z-index scale (no arbitrary `z-*`)
- Animate only compositor props (`transform`, `opacity`)
