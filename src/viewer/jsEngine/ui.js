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
  ui.tabDetailsBtn.classList.toggle('bg-theme-blue', isDetails);
  ui.tabDetailsBtn.classList.toggle('text-white', isDetails);
  ui.tabDetailsBtn.classList.toggle('text-theme-muted', !isDetails);
  ui.tabDetailsBtn.classList.toggle('bg-transparent', !isDetails);

  ui.tabCommandsBtn.classList.toggle('bg-theme-blue', !isDetails);
  ui.tabCommandsBtn.classList.toggle('text-white', !isDetails);
  ui.tabCommandsBtn.classList.toggle('text-theme-muted', isDetails);
  ui.tabCommandsBtn.classList.toggle('bg-transparent', isDetails);
}

function updateCommands(node) {
  if (!ui.commandsList) return;
  const hash = node.hash || node.id;
  const commands = [
    { label: 'moveTo', cmd: `git switch --detach ${hash}` },
    { label: 'delete', cmd: `git revert ${hash}` },
    { label: 'diff with one before', cmd: `git diff ${hash}^ ${hash}` },
    { label: 'diff with my position now', cmd: `git diff HEAD ${hash}` }
  ];

  ui.commandsList.innerHTML = '';
  commands.forEach(item => {
    const row = document.createElement('div');
    row.className = 'space-y-1.5';

    const label = document.createElement('div');
    label.className = 'text-[10px] font-bold text-theme-dim uppercase tracking-wider';
    label.textContent = item.label;

    const commandRow = document.createElement('div');
    commandRow.className = 'flex items-stretch gap-2';

    const code = document.createElement('code');
    code.className = 'block text-xs text-theme-text bg-theme-panel/30 px-2.5 py-1.5 rounded-md break-all flex-1 border border-theme-border/30 shadow-sm';
    code.textContent = item.cmd;

    const copyButton = document.createElement('button');
    copyButton.type = 'button';
    copyButton.className = 'w-8 inline-flex items-center justify-center rounded-md bg-theme-panel/30 border border-theme-border/30 hover:bg-theme-border text-theme-muted flex-shrink-0 transition-colors shadow-sm';
    copyButton.setAttribute('aria-label', `Copy ${item.label} command`);
    copyButton.innerHTML = `
      <svg class="size-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
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
  if (ui.copyHashBtn) {
    ui.copyHashBtn.dataset.value = fullHash;
  }
  document.getElementById('detail-message').textContent = node.message.trim();
  document.getElementById('detail-author').textContent = node.author;
  document.getElementById('detail-date').textContent = formatRelativeDate(node.date);

  const filesContainer = document.getElementById('detail-files');
  filesContainer.innerHTML = '';

  if (ui.filesCountSpan) {
    ui.filesCountSpan.textContent = node.files ? node.files.length : '0';
  }

  if (node.files && node.files.length > 0) {
    node.files.forEach(file => {
      const div = document.createElement('div');
      div.className = 'flex items-center gap-2 text-xs py-1.5 px-2.5 hover:bg-theme-border/30 rounded-md transition-colors truncate bg-theme-panel/30';

      const statusSpan = document.createElement('span');
      let statusColor = 'text-theme-blue';
      if (file.status === 'A') statusColor = 'text-theme-green';
      if (file.status === 'D') statusColor = 'text-theme-red';
      if (file.status === 'M') statusColor = 'text-theme-yellow';

      statusSpan.className = `font-bold ${statusColor} w-4 text-center`;
      statusSpan.textContent = file.status;

      const nameSpan = document.createElement('span');
      nameSpan.className = 'text-theme-text truncate flex-1';
      nameSpan.textContent = file.name;
      nameSpan.title = file.name;

      div.appendChild(statusSpan);
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
