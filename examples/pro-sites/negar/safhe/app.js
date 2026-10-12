let work = [];
let filter = 'All';

function render() {
  const cats = ['All', ...new Set(work.map(w => w.daste))];
  document.getElementById('filters').innerHTML = cats
    .map(c => `<button class="${c === filter ? 'on' : ''}" data-c="${c}">${c}</button>`).join('');
  document.querySelectorAll('#filters button').forEach(b => b.onclick = () => { filter = b.dataset.c; render(); });
  document.getElementById('work').innerHTML = work
    .filter(w => filter === 'All' || w.daste === filter)
    .map(w => `<article class="work">
        <div class="thumb" style="background:linear-gradient(135deg, ${w.rang}, #111)"></div>
        <div class="meta"><b>${w.onvan}</b><span>${w.daste} · ${w.sal}</span></div>
      </article>`).join('');
}

fetch('/api/kar-ha').then(r => r.json()).then(d => { work = d; render(); });

document.getElementById('form').onsubmit = async (e) => {
  e.preventDefault();
  const f = e.target, status = document.getElementById('status');
  f.querySelectorAll('.bad').forEach(x => x.classList.remove('bad'));
  const body = { naam: f.naam.value, email: f.email.value, payam: f.payam.value };
  const r = await fetch('/api/tamas', { method: 'POST', body: JSON.stringify(body) });
  const d = await r.json();
  if (r.ok) {
    status.className = 'status ok';
    status.textContent = 'Thank you! Your message was received.';
    f.reset();
  } else {
    status.className = 'status err';
    status.textContent = 'Please check the highlighted fields.';
    (d.field_ha || []).forEach(n => f[n].classList.add('bad'));
  }
};
