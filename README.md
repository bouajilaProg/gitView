# gitView

A single-binary Git repository visualizer with an interactive DAG (Directed Acyclic Graph) canvas view.

## Features

- Interactive canvas with zoom and pan controls
- Topological sorting with lane assignment for clean branch visualization
- Bezier curves connecting commits
- Click on commits to view details (message, author, date, files changed)
- Single self-contained binary (HTML embedded via go:embed)

## Installation

### From Release

Download the appropriate binary for your platform from the [Releases](../../releases) page.

### Build from Source

```bash
go build -o gitView .
```

## Usage

Navigate to any git repository and run:

```bash
./gitView
```

Then open http://localhost:6060 in your browser.

## Controls

- **Scroll**: Zoom in/out
- **Drag**: Pan the canvas
- **Click**: Select a commit to view details

## Project Structure

- `src/index.html` is the embedded frontend source
- `assets/` is reserved for static assets during development

## Tech Stack

- **Backend**: Go with [go-git](https://github.com/go-git/go-git)
- **Frontend**: Vanilla JavaScript with Canvas API
- **Styling**: Tailwind CSS (via CDN)
