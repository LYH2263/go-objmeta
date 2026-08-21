const out = document.getElementById('out');
document.getElementById('btnList').onclick = async () => {
  const prefix = document.getElementById('prefix').value;
  const r = await fetch('/api/list?prefix=' + encodeURIComponent(prefix));
  out.textContent = await r.text();
};
document.getElementById('btnPut').onclick = async () => {
  const key = document.getElementById('key').value;
  const body = document.getElementById('body').value;
  const r = await fetch('/api/put?key=' + encodeURIComponent(key), {
    method: 'POST', body, headers: {'Content-Type': 'text/plain'}
  });
  out.textContent = await r.text();
};
