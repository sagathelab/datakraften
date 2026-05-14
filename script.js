let activeTab = 'dk';

function switchTab(tab, event) {
  activeTab = tab;

  document.querySelectorAll('.tab-btn').forEach(btn => btn.classList.remove('active'));
  event.currentTarget.classList.add('active');

  document.querySelectorAll('.terminal-command').forEach(el => el.classList.add('hidden'));
  document.getElementById('tab-' + tab).classList.remove('hidden');

  const titles = { dk: 'bash — datakraften.no', bash: 'bash — datakraften.no (legacy)' };
  document.getElementById('terminal-title').textContent = titles[tab];
}

function getActiveCommand() {
  const el = document.getElementById('install-cmd-' + activeTab);
  return el ? el.textContent.trim() : '';
}

function copyCommand() {
  const cmd = getActiveCommand();
  const toast = document.getElementById('toast');

  navigator.clipboard.writeText(cmd).then(() => {
    toast.classList.remove('hidden');
    setTimeout(() => toast.classList.add('hidden'), 2000);
  }).catch(() => {
    const el = document.getElementById('install-cmd-' + activeTab);
    const range = document.createRange();
    range.selectNodeContents(el);
    const selection = window.getSelection();
    selection.removeAllRanges();
    selection.addRange(range);
    document.execCommand('copy');
    selection.removeAllRanges();
    toast.classList.remove('hidden');
    setTimeout(() => toast.classList.add('hidden'), 2000);
  });
}

document.addEventListener('keydown', (e) => {
  if ((e.metaKey || e.ctrlKey) && e.key === 'c') {
    const sel = window.getSelection().toString();
    if (sel.includes('datakraften.no')) {
      const toast = document.getElementById('toast');
      toast.classList.remove('hidden');
      setTimeout(() => toast.classList.add('hidden'), 2000);
    }
  }
});
