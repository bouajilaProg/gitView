import { ui } from './state.js';
import { render } from './render.js';
import { initUI } from './ui.js';
import { initControls } from './controls.js';
import { fetchGraph } from './graphApi.js';

function resizeCanvas() {
  ui.canvas.width = window.innerWidth;
  ui.canvas.height = window.innerHeight;
  render();
}

window.addEventListener('resize', resizeCanvas);
window.addEventListener('theme-changed', render);

resizeCanvas();
initUI();
initControls();
fetchGraph();
