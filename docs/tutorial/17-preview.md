# Lesson 17 — Trying an Android app on your computer (mihan preview)

> **خلاصه:** `mihan preview app.mhn --android` برنامه را درست همان‌طور که در APK اجرا
> می‌شود، روی کامپیوتر خودتان در قابِ یک گوشی نشان می‌دهد، با دکمه‌های برگشت و خانه،
> چرخاندن، اندازه‌های مختلف و یک پنجرهٔ گزارش. این یک شبیه‌ساز واقعی اندروید نیست؛
> یک پیش‌نمایشِ وفادار است.

Building an APK and carrying it to a phone (lesson 15) takes a minute. While
you work, `mihan preview` shows the app on your own computer in a phone frame:

```
mihan preview app.mhn --android
mihan preview app.mhn --android --device small     # 360×640
mihan preview app.mhn --android --device tablet    # 800×1280
```

It opens a window with the phone on the left and a log on the right.

## 17.1 What is the same as on the phone

The program runs the way the APK runs it: the same `mihan` runs it, with
`MIHAN_ANDROID` set, and the phone part does what the Android app does with it:

- A program that **serves** (`serva_bgo`) or opens a window (`baza_bgo`) is
  shown in the phone's screen, as the app's WebView shows it.
- A program that **only prints** shows its output as text, as the app does,
  and back closes it.
- `android_hast()` is `dorost`, as on the phone.

## 17.2 The buttons

- **◁ back** asks the page first (see `mihan.onBack` below), then goes back
  a page; with nowhere to go back to, the app closes, as on Android.
- **○ home** sends the app to the background (`onPause`); tapping its icon on
  the home screen brings it back (`onResume`), or starts it again if it closed.
- **▢ / rotate** turns the phone. The app keeps running, as the APK asks Android.
- **restart app** stops the program and starts it again.

## 17.3 The log

The log shows, as they happen: the app's life (`onCreate`, `onPause`, …),
everything the program prints, every request with its status and how long it
took, mistakes from the program and from the page's JavaScript, the page's
`console.log`, and every phone feature the page used. The boxes at the top
hide one kind.

## 17.4 Phone features in a page: `mihan.js`

On Android, every HTML page the program sends has `/_mihan/mihan.js` in it,
which gives the page `window.mihan`. You can also put it in yourself, so the
same page works in a browser, where each call falls back to something sensible:

```html
<script src="/_mihan/mihan.js"></script>
<button onclick="mihan.toast('saved')">toast</button>
<button onclick="mihan.vibrate(200)">vibrate</button>
<button onclick="mihan.share('look at this')">share</button>
<script>
  mihan.onBack(() => { if (menuIsOpen) { closeMenu(); return true; } return false; });
  addEventListener("mihan-pause", () => save());
</script>
```

| call | on the phone | in the preview | in a browser |
|---|---|---|---|
| `mihan.toast(t)` | Android toast | toast drawn on the screen | a small note on the page |
| `mihan.vibrate(ms)` | vibrates | the phone shakes | `navigator.vibrate` |
| `mihan.share(t)` | Android share sheet | a share sheet drawn | `navigator.share` or copy |
| `mihan.onBack(f)` | back button asks `f` first | the same | — |
| `mihan-pause` / `mihan-resume` | app leaves / returns | home button / icon | — |

A program can ask whether it is on a phone:

```mihan
agar android_hast()
    chap_khat("on a phone")
warna
    chap_khat("on a computer")
```

```output
on a computer
```

## 17.5 What it is not

This is not an Android emulator. The program runs on your computer, not on an
Android system, and the page is drawn by your browser rather than Android's
WebView, so a page that needs a very new browser feature may still look
different on an old phone. Everything Mihan itself does — the program, the
server, the bridge, the back and lifecycle rules — is the same code that the
APK runs. Before you share an APK, try it once on a real phone.
