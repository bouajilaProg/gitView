export const NODE_RADIUS = 26;
export const LANE_HEIGHT = 80;
export const COLUMN_WIDTH = 120;
export const HIT_RADIUS = 25;

export const LANE_COLORS = [
  '#22c55e',
  '#3b82f6',
  '#f59e0b',
  '#ec4899',
  '#8b5cf6',
  '#06b6d4',
  '#f97316',
  '#84cc16'
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
