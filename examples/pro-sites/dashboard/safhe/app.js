const $ = (id) => document.getElementById(id);
const toman = (n) => n >= 1e6 ? (n / 1e6).toFixed(1) + 'M' : n.toLocaleString('en-US');
const statuses = ['tahvil shod', 'ersal shod', 'dar hal-e amade sazi', 'jadid'];
const labels = { 'tahvil shod': 'Delivered', 'ersal shod': 'Shipped', 'dar hal-e amade sazi': 'Preparing', 'jadid': 'New' };
const api = async (url, method = 'GET', body) => {
  const r = await fetch(url, { method, body: body && JSON.stringify(body) });
  return { ok: r.ok, data: await r.json() };
};

async function stats() {
  const { data: a } = await api('/api/amar');
  const days = a.ruz_ha.map(d => d.mablagh), half = days.length / 2;
  const sum = (xs) => xs.reduce((s, x) => s + x, 0);
  const change = (sum(days.slice(half)) / sum(days.slice(0, half)) - 1) * 100;
  const k = [
    ['Revenue', toman(a.dar_amad) + ' T', `${change >= 0 ? '+' : ''}${change.toFixed(1)}% last 7 days vs first 7`, '◎'],
    ['Orders', a.sefaresh, '14 days', '☰'],
    ['Average order', toman(a.miangin) + ' T', 'per order', '≈'],
    ['Customers', a.moshtari, 'unique buyers', '☺'],
  ];
  $('kpis').innerHTML = k.map(([t, v, s, i]) => `<div class="kpi"><span class="ic">${i}</span><span>${t}</span><b>${v}</b><small>${s}</small></div>`).join('');

  const pts = a.ruz_ha.map(d => d.mablagh), max = Math.max(...pts) * 1.15, W = 700, H = 260;
  const xy = pts.map((v, i) => [i * W / (pts.length - 1), H - v / max * H]);
  const line = xy.map((p, i) => (i ? 'L' : 'M') + p[0].toFixed(1) + ' ' + p[1].toFixed(1)).join(' ');
  const grid = [0.25, 0.5, 0.75].map(f => `<line x1="0" x2="${W}" y1="${H * f}" y2="${H * f}" stroke="#e2e8f0" stroke-dasharray="4 4"/>`).join('');
  $('chart').innerHTML = `<defs><linearGradient id="g" x1="0" x2="0" y1="0" y2="1"><stop offset="0" stop-color="#4f46e5" stop-opacity=".35"/><stop offset="1" stop-color="#4f46e5" stop-opacity="0"/></linearGradient></defs>
    ${grid}<path d="${line} L${W} ${H} L0 ${H} Z" fill="url(#g)"/><path d="${line}" fill="none" stroke="#4f46e5" stroke-width="3" stroke-linejoin="round"/>
    ${xy.map(p => `<circle cx="${p[0]}" cy="${p[1]}" r="4" fill="#fff" stroke="#4f46e5" stroke-width="2"/>`).join('')}`;
  $('xlabels').innerHTML = a.ruz_ha.map(d => `<span>${d.ruz.split(' ')[0]}</span>`).join('');

  $('best').innerHTML = a.behtarin.map((b, i) => `<li><span>${i + 1}. ${b.naam}</span><b>${b.tedad} sold</b></li>`).join('');
  const cmax = Math.max(...a.daste_ha.map(c => c.mablagh));
  $('cats').innerHTML = a.daste_ha.map(c => `<div class="cat"><div><span>${c.naam}</span><span>${toman(c.mablagh)} T</span></div><div class="bar"><i style="width:${c.mablagh / cmax * 100}%"></i></div></div>`).join('');
  $('low').textContent = a.mojudi_kam.length ? `${a.mojudi_kam.length} low on stock` : 'stock ok';
}

async function orders() {
  const { data } = await api('/api/sefaresh-ha');
  $('orderRows').innerHTML = data.map(o => `<tr><td>#${o.id}</td><td>${o.moshtari}</td><td>${o.kala} × ${o.tedad}</td><td>${o.ruz}</td><td>${o.mablagh.toLocaleString('en-US')} T</td>
    <td><button class="st st-${statuses.indexOf(o.vaziat)}" data-id="${o.id}" data-v="${o.vaziat}" title="Click to advance">${labels[o.vaziat]}</button></td></tr>`).join('');
  document.querySelectorAll('.st').forEach(b => b.onclick = async () => {
    const i = statuses.indexOf(b.dataset.v), next = statuses[(i + 3) % 4];
    await api('/api/sefaresh/' + b.dataset.id, 'POST', { vaziat: next }); orders();
  });
}

async function products() {
  const { data } = await api('/api/kala');
  $('productRows').innerHTML = data.map(p => `<tr><td><b>${p.naam}</b></td><td>${p.daste}</td><td>${p.gheymat.toLocaleString('en-US')} T</td>
    <td class="${p.mojudi < 6 ? 'low' : ''}"><input type="number" value="${p.mojudi}" data-id="${p.id}"></td>
    <td><button class="btn-s btn-d" data-del="${p.id}">Delete</button></td></tr>`).join('');
  document.querySelectorAll('#productRows input').forEach(i => i.onchange = async () => {
    await api('/api/kala/' + i.dataset.id, 'PUT', { mojudi: +i.value }); products(); stats();
  });
  document.querySelectorAll('[data-del]').forEach(b => b.onclick = async () => {
    if (confirm('Delete this product?')) { await api('/api/kala/' + b.dataset.del, 'DELETE'); products(); stats(); }
  });
}

$('add').onsubmit = async (e) => {
  e.preventDefault();
  const f = e.target;
  const r = await api('/api/kala', 'POST', { naam: f.naam.value, daste: f.daste.value || 'Sayer', gheymat: +f.gheymat.value, mojudi: +f.mojudi.value });
  $('addMsg').textContent = r.ok ? '' : r.data.khata;
  if (r.ok) { f.reset(); products(); stats(); }
};

stats(); orders(); products();
