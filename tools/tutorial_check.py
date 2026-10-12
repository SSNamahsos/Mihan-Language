#!/usr/bin/env python3
"""Run every code sample in docs/tutorial/*.md and check its printed output.

    python3 tools/tutorial_check.py [--mihan PATH] [--fill] [files...]

A sample is a fenced block whose info string starts with `mihan`:

    ```mihan                     run it; the next ```output block is what it prints
    ```mihan stdin=Ali args=a,b  give it typed input and command-line words
    ```mihan serve curl=/,/api   start it as a server, ask those paths, then stop it
    ```mihan norun               shown only (still checked with `mihan check`)
    ```mihan fails               it is meant to stop with a mistake
    ```html file=safhe/index.html  write this file before the samples after it run

Every lesson runs in a fresh folder of its own, and the files and samples of one
lesson share it in order. --fill writes the real output into each ```output
block (for authors); without it, any difference is a failure.
"""
import os, re, shutil, subprocess, sys, tempfile, time, urllib.request

FENCE = re.compile(r"^```(\S*)(.*)$")

def blocks(text):
    lines = text.split("\n")
    i = 0
    out = []
    while i < len(lines):
        m = FENCE.match(lines[i])
        if m and m.group(1):
            lang, info = m.group(1), m.group(2).strip()
            j = i + 1
            while j < len(lines) and lines[j] != "```":
                j += 1
            out.append({"lang": lang, "info": info, "start": i, "end": j,
                        "body": "\n".join(lines[i + 1:j])})
            i = j + 1
            continue
        i += 1
    return lines, out

def attrs(info):
    a = {}
    for part in info.split():
        if "=" in part:
            k, v = part.split("=", 1)
            a[k] = v
        else:
            a[part] = True
    return a

def run_sample(mihan, cwd, body, a, n):
    path = os.path.join(cwd, f"sample{n}.mhn")
    with open(path, "w", encoding="utf-8") as f:
        f.write(body + "\n")
    if "norun" in a:
        r = subprocess.run([mihan, "check", path], capture_output=True, encoding="utf-8", errors="replace", cwd=cwd)
        return None, (r.returncode == 0, r.stdout + r.stderr)
    args = a.get("args", "")
    argv = [mihan, "run", path] + ([x for x in args.split(",")] if args else [])
    if "serve" in a:
        p = subprocess.Popen(argv, stdout=subprocess.PIPE, stderr=subprocess.STDOUT,
                             cwd=cwd, encoding="utf-8", errors="replace")
        time.sleep(1.5)
        answers = []
        port = a.get("port")
        for path_ in [x for x in a.get("curl", "").split(",") if x]:
            method = "GET"
            data = None
            if path_.startswith("POST:"):
                method = "POST"
                path_, data = path_[5:].split("|", 1)
                data = data.encode()
            try:
                req = urllib.request.Request(f"http://127.0.0.1:{port}{path_}", data=data, method=method)
                with urllib.request.urlopen(req, timeout=5) as resp:
                    body_ = resp.read().decode()
            except Exception as e:
                body_ = getattr(e, "read", lambda: str(e).encode())().decode()
            answers.append(f"{method} {path_} -> {body_.strip()[:300]}")
        p.terminate()
        try:
            out, _ = p.communicate(timeout=5)
        except subprocess.TimeoutExpired:
            p.kill(); out, _ = p.communicate()
        text = out + "".join(a_ + "\n" for a_ in answers)
        return text, (True, "")
    r = subprocess.run(argv, input=a.get("stdin", "").replace("|", "\n") + "\n" if "stdin" in a else "",
                       capture_output=True, encoding="utf-8", errors="replace", cwd=cwd, timeout=60)
    text = r.stdout
    if r.returncode != 0:
        text += r.stderr
    ok = (r.returncode == 0) != ("fails" in a)
    return text, (ok, "it stopped with a mistake" if r.returncode else "it was meant to fail")

def check_file(mihan, md, fill):
    with open(md, encoding="utf-8") as f:
        text = f.read()
    lines, bs = blocks(text)
    cwd = tempfile.mkdtemp(prefix="mihan-tut-")
    failures = 0
    samples = 0
    replace = []
    for k, b in enumerate(bs):
        a = attrs(b["info"])
        if "file" in a:
            dest = os.path.join(cwd, a["file"])
            os.makedirs(os.path.dirname(dest) or cwd, exist_ok=True)
            with open(dest, "w", encoding="utf-8") as f:
                f.write(b["body"] + "\n")
            continue
        if b["lang"] != "mihan":
            continue
        samples += 1
        got, (ok, msg) = run_sample(mihan, cwd, b["body"], a, k)
        where = f"{md}:{b['start'] + 1}"
        if got is None:
            if not ok:
                failures += 1
                print(f"FAIL {where}: does not pass `mihan check`\n{msg}")
            continue
        if not ok:
            failures += 1
            print(f"FAIL {where}: {msg}\n{got}")
        nxt = bs[k + 1] if k + 1 < len(bs) else None
        if not nxt or nxt["lang"] != "output":
            failures += 1
            print(f"FAIL {where}: no ```output block after the sample")
            continue
        want = nxt["body"].rstrip("\n")
        got_ = got.rstrip("\n")
        if fill:
            replace.append((nxt["start"], nxt["end"], got_.split("\n")))
        elif want != got_:
            failures += 1
            print(f"FAIL {where}\n--- wanted\n{want}\n--- got\n{got_}\n")
    shutil.rmtree(cwd, ignore_errors=True)
    if fill and replace:
        for s, e, new in reversed(replace):
            lines[s + 1:e] = new
        with open(md, "w", encoding="utf-8") as f:
            f.write("\n".join(lines))
    return samples, failures

def main():
    argv = sys.argv[1:]
    mihan = "mihan"
    fill = False
    files = []
    while argv:
        x = argv.pop(0)
        if x == "--mihan":
            mihan = os.path.abspath(argv.pop(0))
        elif x == "--fill":
            fill = True
        else:
            files.append(x)
    if not files:
        root = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "docs", "tutorial")
        files = sorted(os.path.join(root, f) for f in os.listdir(root) if f.endswith(".md"))
    total = bad = 0
    for md in files:
        s, f = check_file(mihan, md, fill)
        total += s; bad += f
        print(f"{'ok  ' if f == 0 else 'FAIL'} {os.path.basename(md)}: {s} samples, {f} failed")
    print(f"\n{total - bad}/{total} samples print what the tutorial says.")
    sys.exit(1 if bad else 0)

if __name__ == "__main__":
    main()
