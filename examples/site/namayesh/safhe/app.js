
const list = document.getElementById("list");
const fill = document.getElementById("fill");
function add(p) {
  const li = document.createElement("li");
  li.textContent = p.matn;
  list.appendChild(li);
  fill.style.width = Math.min(100, list.children.length * 12) + "%";
}
async function load() {
  const r = await fetch("/api/peygham");
  const j = await r.json();
  list.innerHTML = "";
  j.peygham_ha.forEach(add);
  document.getElementById("where").textContent = j.android
    ? "running as an Android app: the buttons below use the phone"
    : "running in a browser or window: the buttons fall back to the page";
}
document.getElementById("form").onsubmit = async (e) => {
  e.preventDefault();
  const box = document.getElementById("matn");
  if (!box.value) return;
  await fetch("/api/peygham", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ matn: box.value }) });
  box.value = "";
};
document.getElementById("clear").onclick = () => fetch("/api/peygham", { method: "DELETE" });
document.getElementById("toast").onclick = () => mihan.toast("salam az Mihan");
document.getElementById("buzz").onclick = () => mihan.vibrate(200);
document.getElementById("share").onclick = () => mihan.share("Mihan showcase: " + list.children.length + " messages");
mihan.onBack(() => { if (list.lastChild) { list.lastChild.remove(); return true; } return false; });
const live = new EventSource("/_mihan/rooydad/peygham");
live.onmessage = (m) => { const p = JSON.parse(m.data); if (p.pak) { list.innerHTML = ""; fill.style.width = "0"; } else add(p); };
load();
