import { state, ui } from './state.js';
import { render, findNodeAt, screenToWorld, getNodePosition } from './render.js';
import { showDetail, hideDetail, toggleMessages } from './ui.js';

function zoomAtPoint(screenX, screenY, factor) {
  const worldBefore = screenToWorld(screenX, screenY);
  state.camera.scale = Math.min(state.camera.maxScale, Math.max(state.camera.minScale, state.camera.scale * factor));
  const worldAfter = screenToWorld(screenX, screenY);
  state.camera.x += worldAfter.x - worldBefore.x;
  state.camera.y += worldAfter.y - worldBefore.y;
}

function findLatestNodeIndexForBranch(startNode) {
  if (!startNode) return -1;
  const targetId = startNode.hash || startNode.id;

  const childrenMap = {};
  state.graphData.edges.forEach(e => {
    if (!childrenMap[e.target]) childrenMap[e.target] = [];
    childrenMap[e.target].push(e.source);
  });

  const queue = [targetId];
  const visited = new Set([targetId]);
  let bestIndex = Infinity;

  while (queue.length > 0) {
    const currId = queue.shift();
    const currIdx = state.graphData.nodes.findIndex(n => (n.hash || n.id) === currId);
    if (currIdx === -1) continue;

    const currNode = state.graphData.nodes[currIdx];

    if (currNode.refs && currNode.refs.length > 0) {
      if (currIdx < bestIndex) {
        bestIndex = currIdx;
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

  if (bestIndex === Infinity) {
    for (let i = 0; i < state.graphData.nodes.length; i++) {
      if (state.graphData.nodes[i].lane === startNode.lane) {
        return i;
      }
    }
  }

  return bestIndex !== Infinity ? bestIndex : -1;
}

function focusOnNode(node, index) {
  const pos = getNodePosition(node, index);
  state.camera.x = ui.canvas.width / (2 * state.camera.scale) - pos.x;
  state.camera.y = ui.canvas.height / (2 * state.camera.scale) - pos.y;
}

export function initControls() {
  let mouseX = ui.canvas.width / 2;
  let mouseY = ui.canvas.height / 2;

  ui.canvas.addEventListener('click', (e) => {
    const rect = ui.canvas.getBoundingClientRect();
    const x = e.clientX - rect.left;
    const y = e.clientY - rect.top;

    const node = findNodeAt(x, y);
    if (node) {
      showDetail(node);
    }
  });

  let isPanning = false;
  let panStartX = 0;
  let panStartY = 0;

  ui.canvas.addEventListener('mousedown', (e) => {
    isPanning = true;
    panStartX = e.clientX;
    panStartY = e.clientY;
    ui.canvas.style.cursor = 'grabbing';
  });

  ui.canvas.addEventListener('mousemove', (e) => {
    const rect = ui.canvas.getBoundingClientRect();
    mouseX = e.clientX - rect.left;
    mouseY = e.clientY - rect.top;

    if (!isPanning) return;

    const dx = e.clientX - panStartX;
    const dy = e.clientY - panStartY;

    state.camera.x += dx / state.camera.scale;
    state.camera.y += dy / state.camera.scale;

    panStartX = e.clientX;
    panStartY = e.clientY;

    render();
  });

  ui.canvas.addEventListener('mouseup', () => {
    isPanning = false;
    ui.canvas.style.cursor = 'default';
  });

  ui.canvas.addEventListener('mouseleave', () => {
    isPanning = false;
    ui.canvas.style.cursor = 'default';
  });

  ui.canvas.addEventListener('wheel', (e) => {
    e.preventDefault();

    const rect = ui.canvas.getBoundingClientRect();
    const wheelX = e.clientX - rect.left;
    const wheelY = e.clientY - rect.top;

    const zoomFactor = e.deltaY > 0 ? 0.9 : 1.1;
    zoomAtPoint(wheelX, wheelY, zoomFactor);
    render();
  }, { passive: false });

  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') {
      hideDetail();
      return;
    }

    if (e.key === 'm' || e.key === 'M') {
      toggleMessages();
      return;
    }

    if (state.selectedNode && (e.key === 'ArrowLeft' || e.key === 'ArrowRight')) {
      e.preventDefault();
      const currentIndex = state.graphData.nodes.findIndex(n => n.id === state.selectedNode.id);
      let newIndex;

      if (e.key === 'ArrowLeft') {
        newIndex = Math.max(0, currentIndex - 1);
      } else {
        newIndex = Math.min(state.graphData.nodes.length - 1, currentIndex + 1);
      }

      if (newIndex !== currentIndex) {
        showDetail(state.graphData.nodes[newIndex]);
      }
    }

    if (e.key === 'r' && !e.metaKey && !e.ctrlKey) {
      let targetIndex = -1;
      if (state.selectedNode) {
        targetIndex = findLatestNodeIndexForBranch(state.selectedNode);
      }
      if (targetIndex === -1 && state.graphData.nodes.length > 0) {
        targetIndex = 0;
      }
      if (targetIndex !== -1) {
        const targetNode = state.graphData.nodes[targetIndex];
        focusOnNode(targetNode, targetIndex);
        render();
      }
      return;
    }

    if (e.key === '=' || e.key === '+') {
      zoomAtPoint(mouseX, mouseY, 1.2);
      render();
    }
    if (e.key === '-') {
      zoomAtPoint(mouseX, mouseY, 0.8);
      render();
    }
  });
}
