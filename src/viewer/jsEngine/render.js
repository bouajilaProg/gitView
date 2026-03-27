import {
  state,
  ui,
  NODE_RADIUS,
  LANE_HEIGHT,
  COLUMN_WIDTH,
  LANE_COLORS,
  getThemeVar
} from './state.js';

export function getNodePosition(node, index) {
  const totalNodes = state.graphData.nodes.length;
  return {
    x: (totalNodes - 1 - index) * COLUMN_WIDTH,
    y: node.lane * LANE_HEIGHT
  };
}

export function worldToScreen(x, y) {
  return {
    x: (x + state.camera.x) * state.camera.scale,
    y: (y + state.camera.y) * state.camera.scale
  };
}

export function screenToWorld(x, y) {
  return {
    x: x / state.camera.scale - state.camera.x,
    y: y / state.camera.scale - state.camera.y
  };
}

function getLaneColor(lane) {
  return LANE_COLORS[lane % LANE_COLORS.length];
}

function getEdgeColor(edge) {
  if (typeof edge.colorLane === 'number') {
    return getLaneColor(edge.colorLane);
  }
  return getLaneColor(0);
}

function drawEdge(edge, source, target, sourceIndex, targetIndex) {
  const start = getNodePosition(source, sourceIndex);
  const end = getNodePosition(target, targetIndex);

  const startScreen = worldToScreen(start.x, start.y);
  const endScreen = worldToScreen(end.x, end.y);

  ui.ctx.beginPath();
  ui.ctx.moveTo(startScreen.x, startScreen.y);

  if (start.y === end.y) {
    // Same lane - draw straight horizontal line
    ui.ctx.lineTo(endScreen.x, endScreen.y);
  } else {
    // Different lanes - draw U-shape
    // Always run horizontal on the lane that is further from lane 0 (main)
    // This prevents overlap with main line
    const sourceLane = source.lane;
    const targetLane = target.lane;
    
    // Determine which lane is "lower" (further from main = higher lane number)
    const lowerLaneY = Math.max(startScreen.y, endScreen.y);
    
    // U-shape: vertical to lower lane, horizontal on lower lane, vertical to target
    ui.ctx.lineTo(startScreen.x, lowerLaneY);   // Vertical from source to lower lane
    ui.ctx.lineTo(endScreen.x, lowerLaneY);     // Horizontal on lower lane
    ui.ctx.lineTo(endScreen.x, endScreen.y);    // Vertical to target
  }

  ui.ctx.strokeStyle = getEdgeColor(edge);
  ui.ctx.lineWidth = 3 * state.camera.scale;
  ui.ctx.stroke();
}

function drawNode(node, index) {
  const pos = getNodePosition(node, index);
  const screen = worldToScreen(pos.x, pos.y);
  const radius = NODE_RADIUS * state.camera.scale;
  const color = getLaneColor(node.lane);

  const bgTheme = getThemeVar('--bg');
  const textTheme = getThemeVar('--text');
  const panelTheme = getThemeVar('--panel');
  const dimTheme = getThemeVar('--dim');
  const hashTextTheme = getThemeVar('--hash-text');

  ui.ctx.beginPath();
  ui.ctx.arc(screen.x, screen.y, radius, 0, Math.PI * 2);

  if (state.selectedNode && state.selectedNode.id === node.id) {
    ui.ctx.fillStyle = textTheme;
    ui.ctx.strokeStyle = color;
    ui.ctx.lineWidth = 4 * state.camera.scale;
    ui.ctx.fill();
    ui.ctx.stroke();
  } else {
    ui.ctx.fillStyle = color;
    ui.ctx.fill();
    ui.ctx.strokeStyle = bgTheme;
    ui.ctx.lineWidth = 2 * state.camera.scale;
    ui.ctx.stroke();
  }

  if (state.camera.scale > 0.4) {
    const hashColor = state.selectedNode && state.selectedNode.id === node.id ? bgTheme : hashTextTheme;
    ui.ctx.fillStyle = hashColor;
    ui.ctx.font = `bold ${Math.max(8, 10 * state.camera.scale)}px monospace`;
    ui.ctx.textAlign = 'center';
    ui.ctx.textBaseline = 'middle';
    ui.ctx.fillText(node.id.substring(0, 7), screen.x, screen.y);
  }

  if (state.showMessages && state.camera.scale > 0.5) {
    const message = node.message.split('\n')[0];
    const truncated = message.length > 20 ? message.substring(0, 20) + '...' : message;
    ui.ctx.fillStyle = dimTheme;
    ui.ctx.font = `${Math.max(8, 9 * state.camera.scale)}px monospace`;
    ui.ctx.textAlign = 'center';
    ui.ctx.textBaseline = 'top';
    ui.ctx.fillText(truncated, screen.x, screen.y + radius + 8 * state.camera.scale);
  }

  if (node.refs && node.refs.length > 0) {
    const refsString = node.refs.join(', ');
    ui.ctx.font = `600 ${Math.max(9, 11 * state.camera.scale)}px monospace`;
    const textWidth = ui.ctx.measureText(refsString).width;
    const paddingX = 8 * state.camera.scale;
    const paddingY = 4 * state.camera.scale;

    const tagX = screen.x;
    const tagY = screen.y - radius - 12 * state.camera.scale;

    ui.ctx.fillStyle = panelTheme;
    ui.ctx.strokeStyle = color;
    ui.ctx.lineWidth = 1 * state.camera.scale;
    ui.ctx.beginPath();
    ui.ctx.roundRect(
      tagX - textWidth / 2 - paddingX,
      tagY - paddingY,
      textWidth + paddingX * 2,
      14 * state.camera.scale + paddingY * 2,
      4 * state.camera.scale
    );
    ui.ctx.fill();
    ui.ctx.stroke();

    ui.ctx.fillStyle = color;
    ui.ctx.textAlign = 'center';
    ui.ctx.textBaseline = 'top';
    ui.ctx.fillText(refsString, tagX, tagY);
  }
}

function getLaneList() {
  if (state.graphData.lanes && state.graphData.lanes.length > 0) {
    return [...state.graphData.lanes].sort((a, b) => a.index - b.index);
  }
  const lanes = [...new Set(state.graphData.nodes.map(n => n.lane))].sort((a, b) => a - b);
  return lanes.map(index => ({ index, name: '' }));
}

function drawLaneLines() {
  if (state.graphData.nodes.length === 0) return;

  const lanes = getLaneList();
  const labelPadding = 12;

  const laneBounds = {};
  state.graphData.nodes.forEach((node, index) => {
    const pos = getNodePosition(node, index);
    if (!laneBounds[node.lane]) {
      laneBounds[node.lane] = { minX: pos.x, maxX: pos.x };
    } else {
      laneBounds[node.lane].minX = Math.min(laneBounds[node.lane].minX, pos.x);
      laneBounds[node.lane].maxX = Math.max(laneBounds[node.lane].maxX, pos.x);
    }
  });

  lanes.forEach(lane => {
    const bounds = laneBounds[lane.index];
    if (!bounds) return;

    const startX = bounds.minX;
    const endX = bounds.maxX;
    const y = lane.index * LANE_HEIGHT;

    const startScreen = worldToScreen(startX, y);
    const endScreen = worldToScreen(endX, y);

    ui.ctx.beginPath();
    ui.ctx.moveTo(startScreen.x, startScreen.y);
    ui.ctx.lineTo(endScreen.x, endScreen.y);
    ui.ctx.strokeStyle = getLaneColor(lane.index) + '33';
    ui.ctx.lineWidth = 2 * state.camera.scale;
    ui.ctx.setLineDash([5 * state.camera.scale, 5 * state.camera.scale]);
    ui.ctx.stroke();
    ui.ctx.setLineDash([]);

    if (lane.name) {
      const fontSize = Math.max(10, 12 * state.camera.scale);
      ui.ctx.font = `600 ${fontSize}px monospace`;
      ui.ctx.fillStyle = getLaneColor(lane.index);
      ui.ctx.textBaseline = 'middle';

      const leftLabel = worldToScreen(startX - labelPadding, y);
      ui.ctx.textAlign = 'right';
      ui.ctx.fillText(lane.name, leftLabel.x, leftLabel.y);

      const rightLabel = worldToScreen(endX + labelPadding, y);
      ui.ctx.textAlign = 'left';
      ui.ctx.fillText(lane.name, rightLabel.x, rightLabel.y);
    }
  });
}

export function render() {
  ui.ctx.clearRect(0, 0, ui.canvas.width, ui.canvas.height);

  if (state.graphData.nodes.length === 0) return;

  const nodeIndexMap = {};
  state.graphData.nodes.forEach((node, i) => {
    nodeIndexMap[node.id] = i;
  });


  state.graphData.edges.forEach(edge => {
    const sourceIndex = nodeIndexMap[edge.source];
    const targetIndex = nodeIndexMap[edge.target];

    if (sourceIndex !== undefined && targetIndex !== undefined) {
      const source = state.graphData.nodes[sourceIndex];
      const target = state.graphData.nodes[targetIndex];
      drawEdge(edge, source, target, sourceIndex, targetIndex);
    }
  });

  state.graphData.nodes.forEach((node, index) => {
    drawNode(node, index);
  });
}

export function findNodeAt(screenX, screenY) {
  const world = screenToWorld(screenX, screenY);

  for (let i = 0; i < state.graphData.nodes.length; i++) {
    const node = state.graphData.nodes[i];
    const pos = getNodePosition(node, i);
    const dx = world.x - pos.x;
    const dy = world.y - pos.y;
    const distance = Math.sqrt(dx * dx + dy * dy);
    if (distance <= NODE_RADIUS) {
      return node;
    }
  }
  return null;
}
