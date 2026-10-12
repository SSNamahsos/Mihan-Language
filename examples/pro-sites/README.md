# Professional sites built with Mihan

Three complete websites. In each, Mihan is the backend (routing, JSON API, saving
data to a JSON file "database") and the frontend is plain HTML, CSS and JavaScript
in the `safhe/` folder, served by the same program.

| Site | What it is | Run |
|---|---|---|
| `negar/` | a studio / portfolio landing page with a working contact form | `mihan run examples/pro-sites/negar/negar.mhn` |
| `ketabkhane/` | a book shop: catalogue, search, cart, checkout, orders stored on disk | `mihan run examples/pro-sites/ketabkhane/ketabkhane.mhn` |
| `dashboard/` | an admin dashboard: stats, chart, orders table, add/update/delete products | `mihan run examples/pro-sites/dashboard/dashboard.mhn` |

Each prints its address (it picks the port in `PORT`, or 8080/8081/8082). Open it
in a browser. Data is kept in a `dade/` folder beside the program.
