export const NODE_RADIUS = 26;
export const LANE_HEIGHT = 80;
export const COLUMN_WIDTH = 120;
export const HIT_RADIUS = 25;

export const LANE_COLORS = [
  '#4ade80', // Electric Green (Cool)
  '#f87171', // Soft Red (Warm)
  '#38bdf8', // Sky Blue (Cool)
  '#fbbf24', // Amber (Warm)
  '#818cf8', // Indigo (Cool)
  '#f472b6', // Pink (Warm)
  '#2dd4bf', // Teal (Cool)
  '#fb923c', // Orange (Warm)
  '#a78bfa', // Lavender (Cool)
  '#e879f9', // Fuchsia (Warm/Cool bridge)
  '#4ade80', // Mint (Cool)
  '#facc15', // Bright Yellow (Warm)
  '#60a5fa', // Bright Blue (Cool)
  '#fb7185', // Rose (Warm)
  '#22d3ee', // Cyan (Cool)
  '#c084fc', // Purple (Warm/Cool bridge)
  '#bef264', // Lime (Warm-ish)
  '#94a3b8'  // Steel Blue/Slate (Neutral contrast)
];

export const state = {
  graphData: { nodes: [], edges: [], lanes: [] },
  selectedNode: null,
  showMessages: true,
  camera: {
    x: 50,
    y: 100,
    scale: 1,
    minScale: 0.2,
    maxScale: 3
  },
  mouse: {
    x: 0,
    y: 0
  }
};

export const ui = {
  canvas: document.getElementById('graph-canvas'),
  ctx: document.getElementById('graph-canvas').getContext('2d'),
  detailPanel: document.getElementById('detail-panel'),
  panelBackdrop: document.getElementById('panel-backdrop'),
  loadingState: document.getElementById('loading-state'),
  emptyState: document.getElementById('empty-state'),
  closeBtn: document.getElementById('close-panel'),
  toggleMessagesBtn: document.getElementById('toggle-messages'),
  tabDetailsBtn: document.getElementById('tab-details'),
  tabCommandsBtn: document.getElementById('tab-commands'),
  detailContent: document.getElementById('detail-content'),
  commandsContent: document.getElementById('commands-content'),
  commandsList: document.getElementById('detail-commands'),
  copyHashBtn: document.getElementById('copy-hash'),
  filesCountSpan: document.getElementById('files-count'),
  detailHashShort: document.getElementById('detail-hash-short')
};

export function getThemeVar(name) {
  return getComputedStyle(document.documentElement).getPropertyValue(name).trim();
}
