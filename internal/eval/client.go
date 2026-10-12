package eval

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"mihan/internal/ast"
	rt "mihan/internal/runtime"
)

const (
	requestTimeout = 30 * time.Second
	maxBodyBytes   = 8 << 20
)

func fnRequest(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	raw, ok := args[0].(rt.Str)
	if !ok {
		return nil, in.faultAt(node, "E4330",
			"نشانی باید متن باشد",
			"the address has to be text, not %s", rt.TypeName(args[0]))
	}
	addr := strings.TrimSpace(string(raw))
	u, err := url.Parse(addr)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return nil, in.faultAt(node, "E4331",
			"darkhast فقط نشانیِ http و https را می‌خواند",
			"darkhast only speaks http and https")
	}
	method := "GET"
	var body string
	headers := map[string]string{}
	if len(args) == 2 {
		d, ok := args[1].(*rt.Dict)
		if !ok {
			return nil, in.faultAt(node, "E4332",
				"گزینه‌های darkhast باید یک dict باشد",
				"the options to darkhast have to be a dict, not %s", rt.TypeName(args[1]))
		}
		if v, found := d.Get(rt.Str("method")); found && v != nil {
			if _, isNone := v.(rt.Hich); !isNone {
				s, ok := v.(rt.Str)
				if !ok {
					return nil, in.faultAt(node, "E4332",
						"method باید متن باشد",
						"method has to be text, not %s", rt.TypeName(v))
				}
				method = strings.ToUpper(string(s))
			}
		}
		if v, found := d.Get(rt.Str("matn")); found && v != nil {
			if _, isNone := v.(rt.Hich); !isNone {
				s, ok := v.(rt.Str)
				if !ok {
					return nil, in.faultAt(node, "E4332",
						"matn باید متن باشد",
						"matn has to be text, not %s", rt.TypeName(v))
				}
				body = string(s)
			}
		}
		if v, found := d.Get(rt.Str("sarlawah")); found && v != nil {
			if _, isNone := v.(rt.Hich); !isNone {
				hd, ok := v.(*rt.Dict)
				if !ok {
					return nil, in.faultAt(node, "E4332",
						"sarlawah باید یک dict باشد",
						"sarlawah has to be a dict, not %s", rt.TypeName(v))
				}
				for _, key := range hd.Keys() {
					name, ok := jsonKey(key)
					if !ok {
						continue
					}
					val, _ := hd.Get(key)
					s, ok := val.(rt.Str)
					if !ok {
						return nil, in.faultAt(node, "E4332",
							"مقدارِ سرآیند باید متن باشد",
							"a header's value has to be text, not %s", rt.TypeName(val))
					}
					headers[name] = string(s)
				}
			}
		}
	}
	switch method {
	case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD":
	default:
		return nil, in.faultAt(node, "E4331",
			"این method را نمی‌شناسم",
			"darkhast does not know the method %s", method)
	}

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, addr, reader)
	if err != nil {
		return nil, in.faultAt(node, "E4333",
			"این درخواست ساخته نشد",
			"the request could not be made: %v", err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	var resp *http.Response
	whileWaiting(func() { resp, err = http.DefaultClient.Do(req) })
	if err != nil {
		return nil, in.faultAt(node, "E4333",
			"پاسخی نیامد",
			"no answer came back: %v", err)
	}
	defer resp.Body.Close()
	got, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes+1))
	if err != nil {
		return nil, in.faultAt(node, "E4333",
			"پاسخ خوانده نشد",
			"the answer could not be read: %v", err)
	}
	if len(got) > maxBodyBytes {
		return nil, in.faultAt(node, "E4304",
			"پاسخ بزرگ‌تر از آن است که یک‌جا خوانده شود",
			"the answer is too large to read at once")
	}
	if !utf8.Valid(got) {
		return nil, in.faultAt(node, "E4303",
			"پاسخ متن نیست",
			"the answer is not text")
	}
	h := rt.NewDict()
	keys := make([]string, 0, len(resp.Header))
	for k := range resp.Header {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		vs := resp.Header.Values(k)
		if len(vs) > 0 {
			h.Set(rt.Str(k), rt.Str(vs[0]))
		}
	}
	out := rt.NewDict()
	out.Set(rt.Str("vaziat"), rt.Int(int64(resp.StatusCode)))
	out.Set(rt.Str("matn"), rt.Str(string(got)))
	out.Set(rt.Str("sarlawah"), h)
	return out, nil
}
