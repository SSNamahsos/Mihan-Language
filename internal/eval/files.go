package eval

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"mihan/internal/ast"
	rt "mihan/internal/runtime"
)

const maxFileBytes = 32 << 20

func fnReadFile(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	path, err := in.needPath(node, args, 0, "bekhan")
	if err != nil {
		return nil, err
	}
	full := in.workPath(path)
	body, err := readTextFile(full)
	if err != nil {
		return nil, in.fileFault(node, full, err)
	}
	return rt.Str(body), nil
}

func fnWriteFile(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	path, err := in.needPath(node, args, 0, "benevis")
	if err != nil {
		return nil, err
	}
	text, ok := args[1].(rt.Str)
	if !ok {
		return nil, in.faultAt(node, "E4308",
			"آنچه نوشته می‌شود باید متن باشد",
			"what is written has to be text, not %s", rt.TypeName(args[1]))
	}
	full := in.workPath(path)
	parent := filepath.Dir(full)
	st, statErr := os.Stat(parent)
	if statErr != nil || !st.IsDir() {
		return nil, in.faultAt(node, "E4307",
			"پوشه‌ای که پرونده باید در آن بنشیند نیست؛ pooshe آن را می‌سازد",
			"the folder this file would go in is not there; pooshe makes one")
	}
	if st, err := os.Stat(full); err == nil && st.IsDir() {
		return nil, in.faultAt(node, "E4305",
			"این یک پوشه است؛ benevis در پوشه نمی‌نویسد",
			"this is a folder; benevis writes a file")
	}
	if err := os.WriteFile(full, []byte(string(text)), 0o644); err != nil {
		return nil, in.faultAt(node, "E4306",
			"نتوانستم در پرونده بنویسم: "+err.Error(),
			"could not write the file: %v", err)
	}
	return rt.Nil, nil
}

func fnListDir(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	path, err := in.needPath(node, args, 0, "fehrest")
	if err != nil {
		return nil, err
	}
	full := in.workPath(path)
	entries, err := os.ReadDir(full)
	if err != nil {
		return nil, in.faultAt(node, "E4310",
			"fehrest یک پوشه می‌خواهد و این یکی نیست",
			"fehrest needs a folder, and %s is not one", path)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Strings(names)
	out := make([]rt.Value, len(names))
	for i, n := range names {
		out[i] = rt.Str(n)
	}
	return rt.NewList(out...), nil
}

func fnRemove(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	path, err := in.needPath(node, args, 0, "hazf")
	if err != nil {
		return nil, err
	}
	full := in.workPath(path)
	st, statErr := os.Stat(full)
	if statErr != nil {
		return nil, in.faultAt(node, "E4311",
			"چیزی به نام «"+path+"» نیست که حذف شود",
			"there is nothing named %s to remove", path)
	}
	if st.IsDir() {
		entries, err := os.ReadDir(full)
		if err != nil || len(entries) > 0 {
			return nil, in.faultAt(node, "E4312",
				"این پوشه هنوز چیزی در خود دارد و hazf آن را پاک نمی‌کند",
				"this folder still has things in it, so hazf will not remove it")
		}
	}
	if err := os.Remove(full); err != nil {
		return nil, in.faultAt(node, "E4306",
			"نتوانستم حذف کنم: "+err.Error(),
			"could not remove it: %v", err)
	}
	return rt.Nil, nil
}

func fnMkdir(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	path, err := in.needPath(node, args, 0, "pooshe")
	if err != nil {
		return nil, err
	}
	full := in.workPath(path)
	if st, err := os.Stat(full); err == nil && !st.IsDir() {
		return nil, in.faultAt(node, "E4313",
			"اینجا یک پرونده است، نه یک پوشه",
			"there is already a file here, not a folder")
	}
	if err := os.MkdirAll(full, 0o755); err != nil {
		return nil, in.faultAt(node, "E4314",
			"نتوانستم پوشه را بسازم: "+err.Error(),
			"could not make the folder: %v", err)
	}
	return rt.Nil, nil
}

func fnJoinPath(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	parts := make([]string, len(args))
	for i, a := range args {
		s, ok := a.(rt.Str)
		if !ok {
			return nil, in.faultAt(node, "E4300",
				"هر بخش از مسیر باید متن باشد",
				"every part of a path has to be text, not %s", rt.TypeName(a))
		}
		parts[i] = string(s)
	}
	return rt.Str(slashPath(parts...)), nil
}

func fnPathExists(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	path, err := in.needPath(node, args, 0, "masir_hast")
	if err != nil {
		return nil, err
	}
	_, statErr := os.Stat(in.workPath(path))
	return rt.Bool(statErr == nil), nil
}

func fnEnv(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	name, ok := args[0].(rt.Str)
	if !ok {
		return nil, in.faultAt(node, "E4300",
			"نام متغیر باید متن باشد",
			"the name has to be text, not %s", rt.TypeName(args[0]))
	}
	v, found := os.LookupEnv(string(name))
	if !found {
		return rt.Nil, nil
	}
	return rt.Str(v), nil
}

func fnSendFile(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	root, err := in.needPath(node, args, 0, "parvande_pasokh")
	if err != nil {
		return nil, err
	}
	urlPath, ok := args[1].(rt.Str)
	if !ok {
		return nil, in.faultAt(node, "E4340",
			"مسیر صفحه باید متن باشد",
			"the page path has to be text, not %s", rt.TypeName(args[1]))
	}
	dir, found := in.findPageDir(root)
	if !found {
		return nil, in.faultAt(node, "E4341",
			"پوشهٔ صفحه‌ها پیدا نشد: "+root,
			"the folder of pages was not found: %s", root)
	}
	rel, ok := urlRel(string(urlPath))
	if !ok {
		return pageMissing(), nil
	}
	target := dir
	if rel != "" {
		target = filepath.Join(dir, filepath.FromSlash(rel))
	}
	if !within(dir, target) {
		return pageMissing(), nil
	}
	st, statErr := os.Stat(target)
	if statErr != nil {
		return pageMissing(), nil
	}
	if st.IsDir() {
		target = filepath.Join(target, "index.html")
		st, statErr = os.Stat(target)
		if statErr != nil || st.IsDir() {
			return pageMissing(), nil
		}
	}
	if !within(dir, target) {
		return pageMissing(), nil
	}
	if real, err := filepath.EvalSymlinks(target); err == nil {
		if rootReal, err := filepath.EvalSymlinks(dir); err == nil && !within(rootReal, real) {
			return pageMissing(), nil
		}
	}
	if st.Size() > maxFileBytes {
		return nil, in.faultAt(node, "E4304",
			"این پرونده بزرگ‌تر از آن است که یک‌جا فرستاده شود",
			"this file is too large to send at once")
	}
	body, err := os.ReadFile(target)
	if err != nil {
		return pageMissing(), nil
	}
	ans := pageAnswer(200, string(body), contentType(target))

	if h, ok := ans.(*rt.Dict).Get(rt.Str("sarlawah")); ok {
		hd := h.(*rt.Dict)
		hd.Set(rt.Str("ETag"), rt.Str(fileTag(body)))
		hd.Set(rt.Str("Cache-Control"), rt.Str("no-cache"))
		hd.Set(rt.Str("Last-Modified"), rt.Str(st.ModTime().UTC().Format(http.TimeFormat)))
	}
	return ans, nil
}

func (in *Interp) workPath(path string) string {
	cleaned := filepath.Clean(filepath.FromSlash(path))
	if filepath.IsAbs(cleaned) {
		return cleaned
	}
	if wd, err := os.Getwd(); err == nil {
		return filepath.Join(wd, cleaned)
	}
	return cleaned
}

func slashPath(parts ...string) string {
	native := make([]string, len(parts))
	for i, p := range parts {
		native[i] = filepath.FromSlash(p)
	}
	return filepath.ToSlash(filepath.Join(native...))
}

func (in *Interp) findPageDir(path string) (string, bool) {
	cleaned := filepath.Clean(filepath.FromSlash(path))
	var candidates []string
	if filepath.IsAbs(cleaned) {
		candidates = []string{cleaned}
	} else {
		if in.File != "" {
			candidates = append(candidates, filepath.Join(filepath.Dir(in.File), cleaned))
		}
		candidates = append(candidates, cleaned)
		if wd, err := os.Getwd(); err == nil {
			candidates = append(candidates, filepath.Join(wd, cleaned))
		}
	}
	for _, c := range candidates {
		st, err := os.Stat(c)
		if err == nil && st.IsDir() {
			return c, true
		}
	}
	return "", false
}

func urlRel(urlPath string) (string, bool) {
	urlPath = strings.ReplaceAll(urlPath, "\\", "/")
	if urlPath == "" {
		urlPath = "/"
	}
	if !strings.HasPrefix(urlPath, "/") {
		urlPath = "/" + urlPath
	}

	cleaned := path.Clean(urlPath)
	if !strings.HasPrefix(cleaned, "/") {
		return "", false
	}
	rel := strings.TrimPrefix(cleaned, "/")
	if rel == "" || rel == "." {
		return "", true
	}
	return rel, true
}

func within(root, full string) bool {
	root = filepath.Clean(root)
	full = filepath.Clean(full)
	rel, err := filepath.Rel(root, full)
	if err != nil {
		return false
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return false
	}
	return true
}

func pageMissing() rt.Value {
	return pageAnswer(404, "nist", "text/plain; charset=utf-8")
}

func pageAnswer(status int, body, ctype string) rt.Value {
	h := rt.NewDict()
	h.Set(rt.Str("Content-Type"), rt.Str(ctype))
	d := rt.NewDict()
	d.Set(rt.Str("vaziat"), rt.Int(int64(status)))
	d.Set(rt.Str("matn"), rt.Str(body))
	d.Set(rt.Str("sarlawah"), h)
	return d
}

func contentType(name string) string {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".html", ".htm":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".js", ".mjs":
		return "text/javascript; charset=utf-8"
	case ".json":
		return "application/json; charset=utf-8"
	case ".svg":
		return "image/svg+xml"
	case ".txt", ".md", ".mhn":
		return "text/plain; charset=utf-8"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".ico":
		return "image/x-icon"
	case ".woff":
		return "font/woff"
	case ".woff2":
		return "font/woff2"
	default:
		return "application/octet-stream"
	}
}

type fileProblem struct {
	kind string
	err  error
}

func (e fileProblem) Error() string {
	if e.err == nil {
		return e.kind
	}
	return e.err.Error()
}

func readTextFile(full string) (string, error) {
	st, err := os.Stat(full)
	if err != nil {
		return "", fileProblem{kind: "missing", err: err}
	}
	if st.IsDir() {
		return "", fileProblem{kind: "dir"}
	}
	if st.Size() > maxFileBytes {
		return "", fileProblem{kind: "large"}
	}
	body, err := os.ReadFile(full)
	if err != nil {
		return "", fileProblem{kind: "read", err: err}
	}
	if !utf8.Valid(body) {
		return "", fileProblem{kind: "text"}
	}
	return strings.TrimPrefix(string(body), "\uFEFF"), nil
}

func (in *Interp) fileFault(node ast.Node, full string, err error) error {
	prob, ok := err.(fileProblem)
	if !ok {
		return in.faultAt(node, "E4302",
			"نتوانستم پرونده را بخوانم: "+err.Error(),
			"could not read the file: %v", err)
	}
	switch prob.kind {
	case "missing":
		return in.faultAt(node, "E4301",
			"پرونده‌ای به این نام پیدا نشد",
			"there is no file named %s", full)
	case "dir":
		return in.faultAt(node, "E4305",
			"این یک پوشه است؛ fehrest نام‌های داخل آن را می‌دهد",
			"this is a folder; fehrest lists what is inside it")
	case "large":
		return in.faultAt(node, "E4304",
			"این پرونده بزرگ‌تر از آن است که یک‌جا خوانده شود",
			"this file is too large to read at once")
	case "text":
		return in.faultAt(node, "E4303",
			"این پرونده متن نیست",
			"this file is not text")
	default:
		msg := ""
		if prob.err != nil {
			msg = prob.err.Error()
		}
		return in.faultAt(node, "E4302",
			"نتوانستم پرونده را بخوانم: "+msg,
			"could not read the file: %v", prob.err)
	}
}

func (in *Interp) needPath(node ast.Node, args []rt.Value, i int, word string) (string, error) {
	if i >= len(args) {
		return "", in.faultAt(node, "E4300",
			word+" یک مسیر می‌خواهد",
			"%s needs a path", word)
	}
	s, ok := args[i].(rt.Str)
	if !ok {
		return "", in.faultAt(node, "E4300",
			"مسیر باید متن باشد",
			"the path has to be text, not %s", rt.TypeName(args[i]))
	}
	if strings.TrimSpace(string(s)) == "" {
		return "", in.faultAt(node, "E4300",
			"مسیر خالی است",
			"the path is empty")
	}
	return string(s), nil
}
