function copyCommand() {
  const cmd = document.getElementById('install-cmd');
  const toast = document.getElementById('toast');

  navigator.clipboard.writeText(cmd.textContent.trim()).then(() => {
    toast.classList.remove('hidden');
    setTimeout(() => toast.classList.add('hidden'), 2000);
  }).catch(() => {
    const range = document.createRange();
    range.selectNodeContents(cmd);
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
    if (sel.includes('datakraften.no/bootstrap')) {
      const toast = document.getElementById('toast');
      toast.classList.remove('hidden');
      setTimeout(() => toast.classList.add('hidden'), 2000);
    }
  }
});
