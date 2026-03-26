import { state, ui } from './state.js';
import { render } from './render.js';

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

function setActiveTab(tab) {
  const isDetails = tab === 'details';
  ui.detailContent.classList.toggle('hidden', !isDetails);
  ui.commandsContent.classList.toggle('hidden', isDetails);

  // Tab details button
  ui.tabDetailsBtn.classList.toggle('text-theme-blue', isDetails);
  ui.tabDetailsBtn.classList.toggle('text-theme-muted', !isDetails);
  const detailsIndicator = ui.tabDetailsBtn.querySelector('span');
  if (detailsIndicator) detailsIndicator.classList.toggle('scale-x-100', isDetails);
  if (detailsIndicator) detailsIndicator.classList.toggle('scale-x-0', !isDetails);

  // Tab commands button
  ui.tabCommandsBtn.classList.toggle('text-theme-text', !isDetails);
  ui.tabCommandsBtn.classList.toggle('text-theme-muted', isDetails);
  const commandsIndicator = ui.tabCommandsBtn.querySelector('span');
  if (commandsIndicator) commandsIndicator.classList.toggle('scale-x-100', !isDetails);
  if (commandsIndicator) commandsIndicator.classList.toggle('scale-x-0', isDetails);
}

function updateCommands(node) {
  if (!ui.commandsList) return;
  const hash = node.hash || node.id;

  const commands = [
    { label: 'Switch to Commit', cmd: `git switch --detach ${hash}` },
    { label: 'Revert Changes', cmd: `git revert ${hash}` },
    { label: 'Diff with Parent', cmd: `git diff ${hash}^ ${hash}` },
    { label: 'Diff with HEAD', cmd: `git diff HEAD ${hash}` }
  ];

  ui.commandsList.innerHTML = '';

  commands.forEach(item => {
    // 1. Container for Label + Action
    const row = document.createElement('div');
    row.className = 'space-y-2';

    // 2. Small Label
    const label = document.createElement('div');
    label.className = 'text-[10px] font-bold text-theme-dim uppercase tracking-widest px-1';
    label.textContent = item.label;

    // 3. Unified Action Box (The Container)
    const actionWrapper = document.createElement('div');
    actionWrapper.className = 'group flex items-stretch bg-theme-bg border border-theme-border rounded-lg overflow-hidden hover:border-theme-blue/40 transition-all shadow-sm';

    // 4. The Code Segment
    const code = document.createElement('code');
    code.className = 'flex-1 font-mono text-[11px] text-theme-text px-3 py-2.5 truncate select-all';
    code.textContent = item.cmd;

    // 5. The Integrated Copy Button
    const copyButton = document.createElement('button');
    copyButton.type = 'button';
    copyButton.className = 'px-3 bg-theme-border/20 border-l border-theme-border text-theme-muted hover:bg-theme-blue/10 hover:text-theme-blue transition-all flex items-center justify-center';
    copyButton.title = `Copy ${item.label}`;

    // Using Lucide-style SVG for consistency
    copyButton.innerHTML = `
      <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <rect width="14" height="14" x="8" y="8" rx="2" ry="2"/><path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2"/>
      </svg>
    `;

    copyButton.addEventListener('click', (e) => {
      e.stopPropagation(); // Prevents parent clicks if necessary
      copyText(item.cmd);

      // Optional: Quick feedback on click
      const originalIcon = copyButton.innerHTML;
      copyButton.innerHTML = '<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-green-500"><polyline points="20 6 9 17 4 12"/></svg>';
      setTimeout(() => copyButton.innerHTML = originalIcon, 1500);
    });

    // Assemble
    actionWrapper.appendChild(code);
    actionWrapper.appendChild(copyButton);

    row.appendChild(label);
    row.appendChild(actionWrapper);

    ui.commandsList.appendChild(row);
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

export function toggleMessages() {
  state.showMessages = !state.showMessages;
  ui.toggleMessagesBtn.textContent = `Messages: ${state.showMessages ? 'ON' : 'OFF'}`;
  render();
}

export function showDetail(node) {
  state.selectedNode = node;

  const fullHash = node.hash || node.id;
  document.getElementById('detail-hash').textContent = fullHash;
  if (ui.detailHashShort) {
    ui.detailHashShort.textContent = fullHash.substring(0, 12);
  }
  if (ui.copyHashBtn) {
    ui.copyHashBtn.dataset.value = fullHash;
  }
  document.getElementById('detail-message').textContent = node.message.trim();
  document.getElementById('detail-author').textContent = node.author;
  document.getElementById('detail-date').textContent = formatRelativeDate(node.date);

  const branchEl = document.getElementById('detail-branch');
  if (branchEl) {
    const refs = Array.isArray(node.refs) ? node.refs.filter(Boolean) : [];
    if (refs.length > 0) {
      branchEl.textContent = refs.join(', ');
    } else {
      const predicted = (node.predictedBranch || '').trim();
      if (predicted) {
        branchEl.textContent = predicted;
      } else {
        const laneMatch = state.graphData?.lanes?.find(lane => lane.index === node.lane);
        branchEl.textContent = laneMatch?.name?.trim() || '--';
      }
    }
  }

  const filesContainer = document.getElementById('detail-files');
  filesContainer.innerHTML = '';

  if (ui.filesCountSpan) {
    ui.filesCountSpan.textContent = node.files ? node.files.length : '0';
  }

  if (node.files && node.files.length > 0) {
    node.files.forEach(file => {
      const div = document.createElement('div');
      div.className = 'group flex items-center gap-3 text-xs py-2 px-3 hover:bg-theme-blue/5 rounded-lg border border-transparent hover:border-theme-blue/10 transition-all truncate';

      const statusIcon = document.createElement('div');
      let statusColor = 'bg-theme-blue/10 text-theme-blue ring-theme-blue/20';
      if (file.status === 'A') statusColor = 'bg-theme-green/10 text-theme-green ring-theme-green/20';
      if (file.status === 'D') statusColor = 'bg-theme-red/10 text-theme-red ring-theme-red/20';
      if (file.status === 'M') statusColor = 'bg-theme-yellow/10 text-theme-yellow ring-theme-yellow/20';

      statusIcon.className = `size-5 flex items-center justify-center rounded text-[9px] font-bold ring-1 ring-inset ${statusColor}`;
      statusIcon.textContent = file.status;

      const nameSpan = document.createElement('span');
      nameSpan.className = 'text-theme-text font-medium truncate flex-1';
      nameSpan.textContent = file.name;
      nameSpan.title = file.name;

      div.appendChild(statusIcon);
      div.appendChild(nameSpan);
      filesContainer.appendChild(div);
    });
  } else {
    const div = document.createElement('div');
    div.className = 'text-xs text-theme-dim italic p-2';
    div.textContent = 'No files changed';
    filesContainer.appendChild(div);
  }

  ui.panelBackdrop.classList.remove('hidden');
  ui.detailPanel.classList.remove('translate-x-full');
  updateCommands(node);
  render();
}

export function hideDetail() {
  state.selectedNode = null;
  ui.detailPanel.classList.add('translate-x-full');
  ui.panelBackdrop.classList.add('hidden');
  render();
}

export function initUI() {
  setActiveTab('details');

  ui.toggleMessagesBtn.addEventListener('click', () => {
    toggleMessages();
  });

  ui.tabDetailsBtn.addEventListener('click', () => setActiveTab('details'));
  ui.tabCommandsBtn.addEventListener('click', () => setActiveTab('commands'));

  ui.closeBtn.addEventListener('click', hideDetail);
  ui.panelBackdrop.addEventListener('click', hideDetail);

  if (ui.copyHashBtn) {
    ui.copyHashBtn.addEventListener('click', () => {
      copyText(ui.copyHashBtn.dataset.value);
    });
  }
}
