# Lesson 15 — Your program on an Android phone (mihan build --os android)

> **خلاصه:** `mihan build app.mhn --os android` یک فایل `.apk` می‌سازد که روی گوشی
> اندروید نصب می‌شود. برنامه‌ای که پنجره باز می‌کند (`baza_bgo`) یا سرور راه می‌اندازد،
> در خودِ برنامهٔ گوشی نشان داده می‌شود؛ برنامه‌ای که فقط چاپ می‌کند، نوشته‌هایش را نشان می‌دهد.

The same program that becomes a window on Windows (lesson 12) becomes an app
on a phone. You need nothing extra: no Android Studio, no Java, no SDK.
`mihan` writes and signs the APK itself.

## 15.1 Building the APK

```mihan file=shomar/shomar.mhn norun
shomar = 0

funct safhe()
    bargard $"<h1 style='font-size:4rem'>{shomar}</h1><a href='/bishtar'>+1</a>"

funct pasokh(darakht)
    agar darakht["path"] == "/bishtar"
        shomar += 1
    bargard {"vaziat": 200, "matn": safhe()}

funct baz_kon()
    baza_bgo(serva_khod())

serva_bgo(0, pasokh, baz_kon)
```

```
mihan build shomar/shomar.mhn --os android
```

```
built shomar.apk (android/arm64, package org.mihan.app.shomar, 1 files, 3.4 MB, signed with a debug key)
```

| option | meaning |
|---|---|
| `--arch arm64` | for phones (the default for Android) |
| `--arch amd64` | for the Android emulator on a PC |
| `--arch all` | both in one APK (bigger) |
| `--name "Shomarande"` | the name under the icon (default: the file name) |
| `--package org.example.shomar` | the app's id (default: `org.mihan.app.<file name>`) |

## 15.2 Putting it on the phone

Copy `shomar.apk` to the phone and open it. Android will ask you to allow
installing apps from that source, because the app does not come from a store. Or, with the phone connected
by USB and *USB debugging* turned on:

```
adb install shomar.apk
```

The app opens straight onto your program. The back button goes back a page;
closing the app stops the program.

## 15.3 What happens inside

* The APK holds a small Android app (a full-screen WebView) and, as its native
  library, the **Linux runtime for the phone's processor with your program
  appended**, exactly like a Linux build from lesson 14.
* The app starts that program. When the program opens a window or starts a
  server, the app shows that address. If it only prints, the app shows what it
  printed, so `chap_khat` programs work too. `khanak` (typed input) does not
  work: a phone has no terminal to type into.
* Files the program writes go to the app's own private folder on the phone.

## 15.4 Signing

Android only installs signed apps. `mihan` signs with a **debug key** that it
makes the first time and keeps in your settings folder (`~/.config/mihan/android-debug.pem`
on Linux, `%AppData%\mihan\android-debug.pem` on Windows). Keep that file: an
update installs over the old app only if it is signed with the same key. A
debug-signed app is fine for your own phones and for friends. To publish in a
store you would sign it again with your own release key (for example with
`apksigner` from the Android SDK).
