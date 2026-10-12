const list = document.querySelector("#list");
const form = document.querySelector("#form");

async function load() {
  const res = await fetch("/api/kala");
  const items = await res.json();
  list.replaceChildren();
  for (const item of items) {
    const li = document.createElement("li");
    li.textContent = item.nam + " — " + item.gheymat;
    list.append(li);
  }
}

form.addEventListener("submit", async (event) => {
  event.preventDefault();
  const data = new FormData(form);
  await fetch("/api/kala", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      nam: data.get("nam"),
      gheymat: Number(data.get("gheymat"))
    })
  });
  form.reset();
  load();
});

load();
