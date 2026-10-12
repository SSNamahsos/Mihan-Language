const $ = (id) => document.getElementById(id);
const toman = (n) => n.toLocaleString('en-US') + ' T';
let books = [], cats = [], cat = '', cart = JSON.parse(localStorage.getItem('cart') || '{}');

async function load() {
  const p = new URLSearchParams();
  if ($('q').value.trim()) p.set('q', $('q').value.trim());
  if (cat) p.set('daste', cat);
  books = await (await fetch('/api/ketab?' + p)).json();
  render();
}

function render() {
  $('chips').innerHTML = ['', ...cats].map(c =>
    `<button class="${c === cat ? 'on' : ''}" data-c="${c}">${c || 'All books'}</button>`).join('');
  document.querySelectorAll('#chips button').forEach(b => b.onclick = () => { cat = b.dataset.c; load(); });
  $('grid').innerHTML = books.map(b => `
    <article class="book">
      <div class="cover" style="--c:${b.rang}"><small>${b.daste}</small><b>${b.onvan}</b><small>${b.nevisande}</small></div>
      <h3>${b.onvan}</h3><p class="by">${b.nevisande}</p>
      <div class="row"><span class="price">${toman(b.gheymat)}</span><span class="rate">★ ${b.emtiaz}</span></div>
      <button class="add" data-id="${b.id}" ${b.mojudi < 1 ? 'disabled' : ''}>${b.mojudi < 1 ? 'Out of stock' : 'Add to cart'}</button>
    </article>`).join('');
  $('empty').hidden = books.length > 0;
  document.querySelectorAll('.add').forEach(b => b.onclick = () => { add(+b.dataset.id, 1); openCart(); });
  renderCart();
}

const all = {};
function add(id, n) {
  cart[id] = Math.max(0, (cart[id] || 0) + n);
  if (!cart[id]) delete cart[id];
  localStorage.setItem('cart', JSON.stringify(cart));
  renderCart();
}

async function renderCart() {
  for (const id of Object.keys(cart)) if (!all[id]) all[id] = await (await fetch('/api/ketab/' + id)).json();
  books.forEach(b => all[b.id] = b);
  const ids = Object.keys(cart);
  $('count').textContent = ids.reduce((s, id) => s + cart[id], 0);
  $('lines').innerHTML = ids.length ? ids.map(id => { const b = all[id]; return `
    <div class="line"><span class="mini" style="--c:${b.rang}"></span>
      <div><b>${b.onvan}</b><br><small>${toman(b.gheymat)}</small></div>
      <span class="qty"><button data-id="${id}" data-n="-1">−</button>${cart[id]}<button data-id="${id}" data-n="1">+</button></span>
    </div>`; }).join('') : '<p style="color:#78716c">Your cart is empty.</p>';
  document.querySelectorAll('.qty button').forEach(b => b.onclick = () => add(+b.dataset.id, +b.dataset.n));
  $('subtotal').textContent = toman(ids.reduce((s, id) => s + cart[id] * all[id].gheymat, 0));
}

function openCart() { $('drawer').classList.add('open'); $('shade').classList.add('open'); }
function closeCart() { $('drawer').classList.remove('open'); $('shade').classList.remove('open'); }
$('openCart').onclick = openCart; $('closeCart').onclick = closeCart; $('shade').onclick = closeCart;

let t; $('q').oninput = () => { clearTimeout(t); t = setTimeout(load, 200); };

$('checkout').onsubmit = async (e) => {
  e.preventDefault();
  const f = e.target, msg = $('msg');
  const body = { naam: f.naam.value, address: f.address.value,
                 aghlam: Object.keys(cart).map(id => ({ id: +id, tedad: cart[id] })) };
  const r = await fetch('/api/sefaresh', { method: 'POST', body: JSON.stringify(body) });
  const d = await r.json();
  if (r.ok) {
    msg.className = 'msg ok'; msg.textContent = `Order #${d.id} placed — total ${toman(d.kol)}. Thank you!`;
    cart = {}; localStorage.removeItem('cart'); f.reset(); load();
  } else { msg.className = 'msg err'; msg.textContent = d.khata; }
};

fetch('/api/daste-ha').then(r => r.json()).then(c => { cats = c; load(); });

if (location.hash === '#demo') { cart = { 4: 2, 7: 1 }; $('drawer').style.transition = 'none'; setTimeout(() => { renderCart(); openCart(); }, 300); }
