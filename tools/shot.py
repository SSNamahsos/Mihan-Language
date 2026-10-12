#!/usr/bin/env python3
"""Screenshots of a page over time, through headless Chrome (needs websocket-client).

    python3 tools/shot.py URL WIDTH HEIGHT '[[seconds, js-or-null, "out.png-or-null"], ...]'
"""
import json, subprocess, sys, time, base64, urllib.request, websocket, tempfile
def run(url, w, h, shots):
    """shots: list of (seconds, js or None, path)"""
    port = 9333
    p = subprocess.Popen(["google-chrome", "--headless=new", "--no-sandbox", "--disable-gpu", "--hide-scrollbars",
        f"--remote-debugging-port={port}", f"--window-size={w},{h}", f"--user-data-dir={tempfile.mkdtemp()}", "about:blank"],
        stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    try:
        for _ in range(50):
            try:
                tabs = json.load(urllib.request.urlopen(f"http://127.0.0.1:{port}/json")); break
            except Exception: time.sleep(.2)
        tab = [t for t in tabs if t["type"] == "page"][0]
        ws = websocket.create_connection(tab["webSocketDebuggerUrl"], suppress_origin=True)
        n = [0]
        def call(m, **pr):
            n[0] += 1; ws.send(json.dumps({"id": n[0], "method": m, "params": pr}))
            while True:
                r = json.loads(ws.recv())
                if r.get("id") == n[0]: return r.get("result", {})
        call("Page.enable"); call("Page.navigate", url=url); t0 = time.time()
        for at, js, path in shots:
            time.sleep(max(0, at - (time.time() - t0)))
            if js: call("Runtime.evaluate", expression=js)
            if path:
                d = call("Page.captureScreenshot", format="png")["data"]
                open(path, "wb").write(base64.b64decode(d))
    finally:
        p.kill()
if __name__ == "__main__":
    run(sys.argv[1], int(sys.argv[2]), int(sys.argv[3]), json.loads(sys.argv[4]))
