"""Mihan against Python and Go on the programs in this folder.

    python3 bench.py [path/to/mihan]

Each timing is the best of 5 runs of the whole process (start-up included).
The Go programs are built first, into a temporary folder. HTTP is requests
per second from go/load.go (8 connections, 3 seconds).
"""
import os, subprocess, sys, tempfile, time

here = os.path.dirname(os.path.abspath(__file__))
os.chdir(here)
mihan = sys.argv[1] if len(sys.argv) > 1 else "mihan"
tmp = tempfile.mkdtemp()

def gobuild(name):
    out = os.path.join(tmp, name)
    subprocess.run(["go", "build", "-o", out, os.path.join("go", name + ".go")], check=True)
    return out

def best(cmd, n=5):
    b = 1e9
    for _ in range(n):
        t = time.perf_counter()
        subprocess.run(cmd, stdout=subprocess.DEVNULL, check=True)
        b = min(b, time.perf_counter() - t)
    return b

rows = [("fib(30)", "fib"), ("5M-step loop", "loop"), ("dict+list+sort", "c"),
        ("string building", "str"), ("JSON write+read", "js"), ("file write+read", "file"),
        ("objects + lambdas", "oop"), ("start-up (hello)", "hello")]
print("| test | Mihan | Python 3 | Go |")
print("|---|---|---|---|")
for name, f in rows:
    m = best([mihan, "run", f + ".mhn"])
    p = best(["python3", f + ".py"])
    g = best([gobuild(f)])
    print(f"| {name} | {m:.3f} s | {p:.3f} s | {g:.3f} s |", flush=True)

def rps(cmd, url):
    srv = subprocess.Popen(cmd, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    try:
        for _ in range(100):
            try:
                subprocess.run(["curl", "-s", "-o", "/dev/null", url], check=True, timeout=1)
                break
            except Exception:
                time.sleep(0.1)
        out = subprocess.run([load, url], capture_output=True, text=True, check=True).stdout
        return float(out)
    finally:
        srv.kill()
        srv.wait()

load = gobuild("load")
m = rps([mihan, "run", "web.mhn"], "http://127.0.0.1:8811/x")
p = rps(["python3", "web.py"], "http://127.0.0.1:8812/x")
g = rps([gobuild("web")], "http://127.0.0.1:8813/x")
print(f"| HTTP (requests/s, higher is better) | {m:.0f} | {p:.0f} | {g:.0f} |")
