# gitView
Single-binary Git repository visualizer with an interactive DAG canvas.

## Features

- Horizontal branch layout with consistent lane coloring
- Clickable commits with detail panel (message, author, date, files)
- Zoom, pan, and keyboard shortcuts
- Embedded frontend assets (no external build step)

## Project Structure

- `src/` Go backend (server + graph logic)
- `src/web/` Frontend assets (HTML, JS, CSS)

## Installation

### From Release

Download the appropriate binary from the releases page:
https://github.com/bouajilaProg/gitView/releases

### Build from Source

```bash
go build -o gitView ./src
```

## Usage

Run inside any Git repository:

```bash
./gitView
```

Open http://localhost:6060 in your browser.

## Controls

- Scroll: zoom in/out
- Drag: pan the canvas
- Click: select commit
- R: reset view
- M: toggle messages
- Esc: close detail panel
