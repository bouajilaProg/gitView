// Constants - HORIZONTAL layout
const NODE_RADIUS = 26;
const LANE_HEIGHT = 80; // Vertical spacing between branches (lanes)
const COLUMN_WIDTH = 120; // Horizontal spacing between commits
const HIT_RADIUS = 25;

// Branch colors - each lane gets consistent color
const LANE_COLORS = [
  '#22c55e', // green-500
  '#3b82f6', // blue-500
  '#f59e0b', // amber-500
  '#ec4899', // pink-500
  '#8b5cf6', // violet-500
  '#06b6d4', // cyan-500
  '#f97316', // orange-500
  '#84cc16' // lime-500
];

// State
let graphData = { nodes: [], edges: [], lanes: [] };
let selectedNode = null;
let showMessages = true;

// Camera object for zoom/pan
const camera = {
  x: 50,
  y: 100,
  scale: 1,
  minScale: 0.2,
  maxScale: 3
};

// Canvas setup
const canvas = document.getElementById('graph-canvas');
const ctx = canvas.getContext('2d');

// UI Elements
const detailPanel = document.getElementById('detail-panel');
const panelBackdrop = document.getElementById('panel-backdrop');
const loadingState = document.getElementById('loading-state');
const emptyState = document.getElementById('empty-state');
const closeBtn = document.getElementById('close-panel');
const toggleMessagesBtn = document.getElementById('toggle-messages');
const tabDetailsBtn = document.getElementById('tab-details');
const tabCommandsBtn = document.getElementById('tab-commands');
const detailContent = document.getElementById('detail-content');
const commandsContent = document.getElementById('commands-content');
const commandsList = document.getElementById('detail-commands');
const copyHashBtn = document.getElementById('copy-hash');

// Resize handler
function resizeCanvas() {
  canvas.width = window.innerWidth;
  canvas.height = window.innerHeight;
  render();
}

// Initialize
window.addEventListener('resize', resizeCanvas);
resizeCanvas();
setActiveTab('details');

// Toggle messages
toggleMessagesBtn.addEventListener('click', () => {
  showMessages = !showMessages;
  toggleMessagesBtn.textContent = `Messages: ${showMessages ? 'ON' : 'OFF'}`;
  render();
});

// Fetch graph data
async function fetchGraph() {
  try {
    const response = await fetch('/api/graph');
    const data = await response.json();

    loadingState.classList.add('hidden');

    if (data.error) {
      emptyState.querySelector('p').textContent = data.error;
      emptyState.classList.remove('hidden');
      return;
    }

    if (!data.nodes || data.nodes.length === 0) {
      emptyState.classList.remove('hidden');
      return;
    }

    graphData = data;
    render();
  } catch (err) {
    loadingState.classList.add('hidden');
    emptyState.querySelector('p').textContent = 'Failed to load repository';
    emptyState.classList.remove('hidden');
    console.error('Failed to fetch graph:', err);
  }
}

// Get node position - HORIZONTAL layout
// Oldest commits on RIGHT, newest on LEFT
// X = (total - index) * spacing (reversed)
// Y = lane (branch)
function getNodePosition(node, index) {
  const totalNodes = graphData.nodes.length;
  return {
    x: (totalNodes - 1 - index) * COLUMN_WIDTH,
    y: node.lane * LANE_HEIGHT
  };
}

// Transform world coordinates to screen coordinates
function worldToScreen(x, y) {
  return {
    x: (x + camera.x) * camera.scale,
    y: (y + camera.y) * camera.scale
  };
}

// Transform screen coordinates to world coordinates
function screenToWorld(x, y) {
  return {
    x: x / camera.scale - camera.x,
    y: y / camera.scale - camera.y
  };
}

// Get lane color - consistent per branch
function getLaneColor(lane) {
  return LANE_COLORS[lane % LANE_COLORS.length];
}

function getEdgeColor(edge) {
  if (typeof edge.colorLane === 'number') {
    return getLaneColor(edge.colorLane);
  }
  return getLaneColor(0);
}

// Draw horizontal line between two points with curves for lane changes
function drawEdge(edge, source, target, sourceIndex, targetIndex) {
  const start = getNodePosition(source, sourceIndex);
  const end = getNodePosition(target, targetIndex);

  const startScreen = worldToScreen(start.x, start.y);
  const endScreen = worldToScreen(end.x, end.y);

  ctx.beginPath();
  ctx.moveTo(startScreen.x, startScreen.y);

  if (start.y === end.y) {
    // Same lane - straight horizontal line
    ctx.lineTo(endScreen.x, endScreen.y);
  } else {
    // Different lanes - use bezier curve
    const midX = (startScreen.x + endScreen.x) / 2;
    ctx.bezierCurveTo(
      midX, startScreen.y,
      midX, endScreen.y,
      endScreen.x, endScreen.y
    );
  }

  ctx.strokeStyle = getEdgeColor(edge);
  ctx.lineWidth = 3 * camera.scale;
  ctx.stroke();
}

// Draw commit node with hash inside
function drawNode(node, index) {
  const pos = getNodePosition(node, index);
  const screen = worldToScreen(pos.x, pos.y);
  const radius = NODE_RADIUS * camera.scale;
  const color = getLaneColor(node.lane);

  // Draw circle
  ctx.beginPath();
  ctx.arc(screen.x, screen.y, radius, 0, Math.PI * 2);

  if (selectedNode && selectedNode.id === node.id) {
    ctx.fillStyle = '#ffffff';
    ctx.strokeStyle = color;
    ctx.lineWidth = 4 * camera.scale;
    ctx.fill();
    ctx.stroke();
  } else {
    ctx.fillStyle = color;
    ctx.fill();
    ctx.strokeStyle = '#18181b';
    ctx.lineWidth = 2 * camera.scale;
    ctx.stroke();
  }

  // Draw hash inside circle
  if (camera.scale > 0.4) {
    const hashColor = selectedNode && selectedNode.id === node.id ? '#18181b' : '#ffffff';
    ctx.fillStyle = hashColor;
    ctx.font = `bold ${Math.max(8, 10 * camera.scale)}px monospace`;
    ctx.textAlign = 'center';
    ctx.textBaseline = 'middle';
    ctx.fillText(node.id.substring(0, 7), screen.x, screen.y);
  }

  // Draw commit message below
  if (showMessages && camera.scale > 0.5) {
    const message = node.message.split('\n')[0];
    const truncated = message.length > 20 ? message.substring(0, 20) + '...' : message;
    ctx.fillStyle = '#a1a1aa';
    ctx.font = `${Math.max(8, 9 * camera.scale)}px monospace`;
    ctx.textAlign = 'center';
    ctx.textBaseline = 'top';
    ctx.fillText(truncated, screen.x, screen.y + radius + 8 * camera.scale);
  }

  // Draw refs (branch names) above node
  if (node.refs && node.refs.length > 0) {
    const refsString = node.refs.join(', ');
    ctx.font = `600 ${Math.max(9, 11 * camera.scale)}px monospace`;
    const textWidth = ctx.measureText(refsString).width;
    const paddingX = 8 * camera.scale;
    const paddingY = 4 * camera.scale;
    
    const tagX = screen.x;
    const tagY = screen.y - radius - 12 * camera.scale;

    ctx.fillStyle = '#16161e'; // bg-tokyo-panel
    ctx.strokeStyle = color;
    ctx.lineWidth = 1 * camera.scale;
    ctx.beginPath();
    ctx.roundRect(tagX - textWidth/2 - paddingX, tagY - paddingY, textWidth + paddingX*2, 14 * camera.scale + paddingY*2, 4 * camera.scale);
    ctx.fill();
    ctx.stroke();

    ctx.fillStyle = color;
    ctx.textAlign = 'center';
    ctx.textBaseline = 'top';
    ctx.fillText(refsString, tagX, tagY);
  }
}

// Resolve lanes for labeling
function getLaneList() {
  if (graphData.lanes && graphData.lanes.length > 0) {
    return [...graphData.lanes].sort((a, b) => a.index - b.index);
  }
  const lanes = [...new Set(graphData.nodes.map(n => n.lane))].sort((a, b) => a - b);
  return lanes.map(index => ({ index, name: '' }));
}

// Draw branch lane lines (horizontal guidelines)
function drawLaneLines() {
  if (graphData.nodes.length === 0) return;

  const lanes = getLaneList();
  const maxIndex = graphData.nodes.length - 1;
  const startX = -50;
  const endX = maxIndex * COLUMN_WIDTH + 50;
  const labelPadding = 12;

  lanes.forEach(lane => {
    const y = lane.index * LANE_HEIGHT;
    const startScreen = worldToScreen(startX, y);
    const endScreen = worldToScreen(endX, y);

    ctx.beginPath();
    ctx.moveTo(startScreen.x, startScreen.y);
    ctx.lineTo(endScreen.x, endScreen.y);
    ctx.strokeStyle = getLaneColor(lane.index) + '33'; // 20% opacity
    ctx.lineWidth = 2 * camera.scale;
    ctx.setLineDash([5 * camera.scale, 5 * camera.scale]);
    ctx.stroke();
    ctx.setLineDash([]);

    if (lane.name) {
      const fontSize = Math.max(10, 12 * camera.scale);
      ctx.font = `600 ${fontSize}px monospace`;
      ctx.fillStyle = getLaneColor(lane.index);
      ctx.textBaseline = 'middle';

      const leftLabel = worldToScreen(startX - labelPadding, y);
      ctx.textAlign = 'right';
      ctx.fillText(lane.name, leftLabel.x, leftLabel.y);

      const rightLabel = worldToScreen(endX + labelPadding, y);
      ctx.textAlign = 'left';
      ctx.fillText(lane.name, rightLabel.x, rightLabel.y);
    }
  });
}

// Main render function
function render() {
  ctx.clearRect(0, 0, canvas.width, canvas.height);

  if (graphData.nodes.length === 0) return;

  // Build node index map
  const nodeIndexMap = {};
  graphData.nodes.forEach((node, i) => {
    nodeIndexMap[node.id] = i;
  });

  // Draw lane guidelines first
  drawLaneLines();

  // Draw edges
  graphData.edges.forEach(edge => {
    const sourceIndex = nodeIndexMap[edge.source];
    const targetIndex = nodeIndexMap[edge.target];

    if (sourceIndex !== undefined && targetIndex !== undefined) {
      const source = graphData.nodes[sourceIndex];
      const target = graphData.nodes[targetIndex];
      drawEdge(edge, source, target, sourceIndex, targetIndex);
    }
  });

  // Draw nodes on top
  graphData.nodes.forEach((node, index) => {
    drawNode(node, index);
  });
}

// Hit detection - find node at screen coordinates
function findNodeAt(screenX, screenY) {
  const world = screenToWorld(screenX, screenY);

  for (let i = 0; i < graphData.nodes.length; i++) {
    const node = graphData.nodes[i];
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

// Format relative date
function formatRelativeDate(dateStr) {
  const date = new Date(dateStr);
  const now = new Date();
  const diffMs = now - date;
  const diffSec = Math.floor(diffMs / 1000);
  const diffMin = Math.floor(diffSec / 60);
  const diffHour = Math.floor(diffMin / 60);
  const diffDay = Math.floor(diffHour / 24);
  const diffMonth = Math.floor(diffDay / 30);
  const diffYear = Math.floor(diffDay / 365);

  if (diffYear > 0) return `${diffYear} year${diffYear > 1 ? 's' : ''} ago`;
  if (diffMonth > 0) return `${diffMonth} month${diffMonth > 1 ? 's' : ''} ago`;
  if (diffDay > 0) return `${diffDay} day${diffDay > 1 ? 's' : ''} ago`;
  if (diffHour > 0) return `${diffHour} hour${diffHour > 1 ? 's' : ''} ago`;
  if (diffMin > 0) return `${diffMin} minute${diffMin > 1 ? 's' : ''} ago`;
  return 'just now';
}

// Show detail panel (overlay)
function showDetail(node) {
  selectedNode = node;

  const fullHash = node.hash || node.id;
  document.getElementById('detail-hash').textContent = fullHash;
  if (copyHashBtn) {
    copyHashBtn.dataset.value = fullHash;
  }
  document.getElementById('detail-message').textContent = node.message.trim();
  document.getElementById('detail-author').textContent = node.author;
  document.getElementById('detail-date').textContent = formatRelativeDate(node.date);

  const filesContainer = document.getElementById('detail-files');
  filesContainer.innerHTML = '';

  if (node.files && node.files.length > 0) {
    node.files.forEach(file => {
      const div = document.createElement('div');
      div.className = 'flex items-center gap-2 text-xs py-1 px-2 bg-tokyo-bg rounded truncate';
      
      const statusSpan = document.createElement('span');
      let statusColor = 'text-tokyo-blue';
      if (file.status === 'A') statusColor = 'text-tokyo-green';
      if (file.status === 'D') statusColor = 'text-tokyo-red';
      if (file.status === 'M') statusColor = 'text-tokyo-yellow';
      
      statusSpan.className = `font-bold ${statusColor}`;
      statusSpan.textContent = file.status;
      
      const nameSpan = document.createElement('span');
      nameSpan.className = 'text-tokyo-text truncate';
      nameSpan.textContent = file.name;
      nameSpan.title = file.name;
      
      div.appendChild(statusSpan);
      div.appendChild(nameSpan);
      filesContainer.appendChild(div);
    });
  } else {
    const div = document.createElement('div');
    div.className = 'text-xs text-tokyo-dim italic';
    div.textContent = 'No files changed';
    filesContainer.appendChild(div);
  }

  panelBackdrop.classList.remove('hidden');
  detailPanel.classList.remove('translate-x-full');
  updateCommands(node);
  render();
}

// Hide detail panel
function hideDetail() {
  selectedNode = null;
  detailPanel.classList.add('translate-x-full');
  panelBackdrop.classList.add('hidden');
  render();
}

// Event: Close panel button
closeBtn.addEventListener('click', hideDetail);

// Event: Click backdrop to close
panelBackdrop.addEventListener('click', hideDetail);

function setActiveTab(tab) {
  const isDetails = tab === 'details';
  detailContent.classList.toggle('hidden', !isDetails);
  commandsContent.classList.toggle('hidden', isDetails);
  tabDetailsBtn.classList.toggle('bg-tokyo-blue', isDetails);
  tabDetailsBtn.classList.toggle('text-tokyo-bg', isDetails);
  tabDetailsBtn.classList.toggle('text-tokyo-muted', !isDetails);
  tabDetailsBtn.classList.toggle('bg-transparent', !isDetails);
  
  tabCommandsBtn.classList.toggle('bg-tokyo-blue', !isDetails);
  tabCommandsBtn.classList.toggle('text-tokyo-bg', !isDetails);
  tabCommandsBtn.classList.toggle('text-tokyo-muted', isDetails);
  tabCommandsBtn.classList.toggle('bg-transparent', isDetails);
}

tabDetailsBtn.addEventListener('click', () => setActiveTab('details'));
tabCommandsBtn.addEventListener('click', () => setActiveTab('commands'));

function updateCommands(node) {
  if (!commandsList) return;
  const hash = node.hash || node.id;
  const commands = [
    { label: 'moveTo', cmd: `git switch --detach ${hash}` },
    { label: 'delete', cmd: `git revert ${hash}` },
    { label: 'diff with one before', cmd: `git diff ${hash}^ ${hash}` },
    { label: 'diff with my position now', cmd: `git diff HEAD ${hash}` }
  ];

  commandsList.innerHTML = '';
  commands.forEach(item => {
    const row = document.createElement('div');
    row.className = 'space-y-1';

    const label = document.createElement('div');
    label.className = 'text-xs text-tokyo-dim uppercase tracking-wide';
    label.textContent = item.label;

    const commandRow = document.createElement('div');
    commandRow.className = 'flex items-start gap-2';

    const code = document.createElement('code');
    code.className = 'block text-xs text-tokyo-text bg-tokyo-bg px-2 py-1 rounded break-all flex-1 border border-tokyo-border';
    code.textContent = item.cmd;

    const copyButton = document.createElement('button');
    copyButton.type = 'button';
    copyButton.className = 'size-7 inline-flex items-center justify-center rounded hover:bg-tokyo-border/80 text-tokyo-muted flex-shrink-0';
    copyButton.setAttribute('aria-label', `Copy ${item.label} command`);
    copyButton.innerHTML = `
      <svg class="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
        stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
        <rect x="9" y="9" width="13" height="13" rx="2" />
        <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
      </svg>
    `;
    copyButton.addEventListener('click', () => copyText(item.cmd));

    commandRow.appendChild(code);
    commandRow.appendChild(copyButton);

    row.appendChild(label);
    row.appendChild(commandRow);
    commandsList.appendChild(row);
  });
}

function copyText(value) {
  if (!value) return;
  if (navigator.clipboard && navigator.clipboard.writeText) {
    navigator.clipboard.writeText(value).catch(() => {
      window.prompt('Copy to clipboard:', value);
    });
    return;
  }
  window.prompt('Copy to clipboard:', value);
}

if (copyHashBtn) {
  copyHashBtn.addEventListener('click', () => {
    copyText(copyHashBtn.dataset.value);
  });
}

function findLatestNodeIndexForBranch(startNode) {
  if (!startNode) return -1;
  const targetId = startNode.hash || startNode.id;
  
  // Find all children maps
  const childrenMap = {};
  graphData.edges.forEach(e => {
      if (!childrenMap[e.target]) childrenMap[e.target] = [];
      childrenMap[e.target].push(e.source);
  });

  // Simple BFS up to find the nearest branch tip
  const queue = [targetId];
  const visited = new Set([targetId]);
  let bestNode = null;
  let bestIndex = Infinity;

  while(queue.length > 0) {
      const currId = queue.shift();
      const currIdx = graphData.nodes.findIndex(n => (n.hash || n.id) === currId);
      if (currIdx === -1) continue;
      
      const currNode = graphData.nodes[currIdx];
      
      if (currNode.refs && currNode.refs.length > 0) {
          if (currIdx < bestIndex) {
              bestIndex = currIdx;
              bestNode = currNode;
          }
      }

      const children = childrenMap[currId] || [];
      for (const childId of children) {
          if (!visited.has(childId)) {
              visited.add(childId);
              queue.push(childId);
          }
      }
  }

  // Fallback to latest in lane if no branch tip found
  if (bestIndex === Infinity) {
    for (let i = 0; i < graphData.nodes.length; i++) {
      if (graphData.nodes[i].lane === startNode.lane) {
        return i;
      }
    }
  }

  return bestIndex !== Infinity ? bestIndex : -1;
}

function focusOnNode(node, index) {
  const pos = getNodePosition(node, index);
  camera.x = canvas.width / (2 * camera.scale) - pos.x;
  camera.y = canvas.height / (2 * camera.scale) - pos.y;
}

// Event: Canvas click
canvas.addEventListener('click', (e) => {
  const rect = canvas.getBoundingClientRect();
  const x = e.clientX - rect.left;
  const y = e.clientY - rect.top;

  const node = findNodeAt(x, y);
  if (node) {
    showDetail(node);
  }
});

// Pan state
let isPanning = false;
let panStartX = 0;
let panStartY = 0;

// Event: Mouse down (start pan)
canvas.addEventListener('mousedown', (e) => {
  isPanning = true;
  panStartX = e.clientX;
  panStartY = e.clientY;
  canvas.style.cursor = 'grabbing';
});

// Event: Mouse move (pan)
canvas.addEventListener('mousemove', (e) => {
  if (!isPanning) return;

  const dx = e.clientX - panStartX;
  const dy = e.clientY - panStartY;

  camera.x += dx / camera.scale;
  camera.y += dy / camera.scale;

  panStartX = e.clientX;
  panStartY = e.clientY;

  render();
});

// Event: Mouse up (end pan)
canvas.addEventListener('mouseup', () => {
  isPanning = false;
  canvas.style.cursor = 'default';
});

canvas.addEventListener('mouseleave', () => {
  isPanning = false;
  canvas.style.cursor = 'default';
});

// Event: Wheel (zoom)
canvas.addEventListener('wheel', (e) => {
  e.preventDefault();

  const rect = canvas.getBoundingClientRect();
  const mouseX = e.clientX - rect.left;
  const mouseY = e.clientY - rect.top;

  // Get world position before zoom
  const worldBefore = screenToWorld(mouseX, mouseY);

  // Apply zoom
  const zoomFactor = e.deltaY > 0 ? 0.9 : 1.1;
  camera.scale = Math.min(camera.maxScale, Math.max(camera.minScale, camera.scale * zoomFactor));

  // Get world position after zoom
  const worldAfter = screenToWorld(mouseX, mouseY);

  // Adjust camera to keep mouse position stable
  camera.x += worldAfter.x - worldBefore.x;
  camera.y += worldAfter.y - worldBefore.y;

  render();
}, { passive: false });

// Keyboard shortcuts
document.addEventListener('keydown', (e) => {
  // Escape to close panel
  if (e.key === 'Escape') {
    hideDetail();
    return;
  }

  // M to toggle messages
  if (e.key === 'm' || e.key === 'M') {
    showMessages = !showMessages;
    toggleMessagesBtn.textContent = `Messages: ${showMessages ? 'ON' : 'OFF'}`;
    render();
    return;
  }

  // Navigation with arrow keys (now horizontal, reversed direction)
  if (selectedNode && (e.key === 'ArrowLeft' || e.key === 'ArrowRight')) {
    e.preventDefault();
    const currentIndex = graphData.nodes.findIndex(n => n.id === selectedNode.id);
    let newIndex;

    // Left = newer (lower index), Right = older (higher index)
    if (e.key === 'ArrowLeft') {
      newIndex = Math.max(0, currentIndex - 1);
    } else {
      newIndex = Math.min(graphData.nodes.length - 1, currentIndex + 1);
    }

    if (newIndex !== currentIndex) {
      showDetail(graphData.nodes[newIndex]);
    }
  }

  // Reset view with 'r' key
  if (e.key === 'r' && !e.metaKey && !e.ctrlKey) {
    let targetIndex = -1;
    if (selectedNode) {
      targetIndex = findLatestNodeIndexForBranch(selectedNode);
    }
    if (targetIndex === -1 && graphData.nodes.length > 0) {
      targetIndex = 0;
    }
    if (targetIndex !== -1) {
      const targetNode = graphData.nodes[targetIndex];
      focusOnNode(targetNode, targetIndex);
      render();
    }
    return;
  }

  // Zoom with +/- keys
  if (e.key === '=' || e.key === '+') {
    camera.scale = Math.min(camera.maxScale, camera.scale * 1.2);
    render();
  }
  if (e.key === '-') {
    camera.scale = Math.max(camera.minScale, camera.scale * 0.8);
    render();
  }
});

// Start
fetchGraph();
