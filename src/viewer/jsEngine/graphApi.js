import { state, ui } from './state.js';
import { render } from './render.js';

export async function fetchGraph() {
  try {
    const response = await fetch('/api/graph');
    const data = await response.json();

    ui.loadingState.classList.add('hidden');

    if (data.error) {
      ui.emptyState.querySelector('p').textContent = data.error;
      ui.emptyState.classList.remove('hidden');
      return;
    }

    if (!data.nodes || data.nodes.length === 0) {
      ui.emptyState.classList.remove('hidden');
      return;
    }

    state.graphData = data;
    render();
  } catch (err) {
    ui.loadingState.classList.add('hidden');
    ui.emptyState.querySelector('p').textContent = 'Failed to load repository';
    ui.emptyState.classList.remove('hidden');
    console.error('Failed to fetch graph:', err);
  }
}
