const form = document.querySelector("#form");
const out = document.querySelector("#out");

form.addEventListener("submit", async (event) => {
  event.preventDefault();
  const data = new FormData(form);
  const res = await fetch("/api/hesab", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      a: Number(data.get("a")),
      b: Number(data.get("b"))
    })
  });
  const j = await res.json();
  out.textContent = "jam " + j.jam + "   zarb " + j.zarb + "   kam " + j.kam;
});
