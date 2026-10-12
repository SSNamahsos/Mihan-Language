const list = document.querySelector("#list");
const form = document.querySelector("#form");

function show(items) {
  list.replaceChildren();
  for (const item of items) {
    const li = document.createElement("li");
    li.textContent = item.matn;
    list.append(li);
  }
}

async function load() {
  const res = await fetch("/api/yaddasht");
  show(await res.json());
}

form.addEventListener("submit", async (event) => {
  event.preventDefault();
  const data = new FormData(form);
  const res = await fetch("/api/yaddasht", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ matn: data.get("matn") })
  });
  show(await res.json());
  form.reset();
});

load();
