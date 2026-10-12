<div dir="rtl">

# یادگیری میهن — دوره‌ی آموزشی

English: [../README.md](../README.md)

> میهن را از صفر یاد بگیرید. هر درس توضیح، مثال قابل اجرا، خروجی واقعی و تمرین با جواب دارد.

درس‌ها را به ترتیب بخوانید. هر نمونه‌کد یک برنامه‌ی واقعی است: آن را در یک فایل `.mhn` کپی
کنید، با `mihan run file.mhn` اجرایش کنید، و دقیقاً همان خروجی‌ای را می‌بینید که زیرش آمده.
(`python3 tools/tutorial_check.py` همه‌ی نمونه‌ها را دوباره اجرا می‌کند و اگر خروجی‌ای فرق
کند شکست می‌خورد، پس دوره هیچ‌وقت از زبان عقب نمی‌ماند.)

| # | درس | چه یاد می‌گیرید |
|---|---|---|
| ۱ | [شروع](01-shoru.md) | نصب، اولین برنامه، چاپ، REPL |
| ۲ | [مقدارها و متغیرها](02-arzesh-ha.md) | عدد، متن، درست/غلط، نوع‌ها، تبدیل |
| ۳ | [متن](03-matn-ha.md) | `$"..."`، برش، متدهای متن، قالب‌بندی |
| ۴ | [لیست‌ها، دیکشنری‌ها و مجموعه‌ها](04-majmooe-ha.md) | مجموعه‌ها، `filter`، `tabdil_kol`، بازه‌ها |
| ۵ | [تصمیم‌ها و حلقه‌ها](05-tasmim-va-tekrar.md) | `agar`، `warna`، `har`، `ta`، `shek`، `edame` |
| ۶ | [تابع‌ها](06-tabe-ha.md) | `funct`، مقدار پیش‌فرض، آرگومان اسم‌دار، لامبدا، closure |
| ۷ | [کلاس‌ها](07-darghan-ha.md) | `darghan`، `khod`، سازنده، ارث‌بری |
| ۸ | [خطاها](08-khata-ha.md) | `koshish`، `giroft`، `portab`، `nemigir` |
| ۹ | [فایل‌ها، JSON و اسکریپت‌ها](09-script-ha.md) | فایل، JSON، ورودی، آرگومان‌ها، `biar` |
| ۱۰ | [بک‌اند یک وب‌سایت](10-backend.md) | `serva_bgo`، درخواست‌ها، یک REST API از نوع JSON |
| ۱۱ | [فرانت‌اند یک وب‌سایت](11-frontend.md) | فایل‌های HTML/CSS/JS، یا صفحه‌هایی از تکه‌های میهن |
| ۱۲ | [یک برنامه‌ی ویندوزی](12-windows-app.md) | `baza_bgo`، تحویل برنامه |
| ۱۳ | [ابزارها](13-abzar.md) | `mihan test`، `check`، `fmt`، VS Code |
| ۱۴ | [اشتراک به‌صورت یک فایل](14-build.md) | `mihan build` برای ویندوز و لینوکس |
| ۱۵ | [روی گوشی اندروید](15-android.md) | `mihan build --os android`، یک APK |
| ۱۶ | [یک API واقعی وب](16-web-api.md) | مسیرهای `:id`، متدها، CORS، کوکی، `jalase`، آپلود، `qaleb`، رویدادهای زنده |
| ۱۷ | [امتحان کردن اپ اندروید](17-preview.md) | `mihan preview --android`، `mihan.toast/vibrate/share/onBack` |

فایل‌های کوتاه `NN-*.mhn` در پوشه‌ی [`docs/tutorial/`](../) برنامه‌های مرجع یک‌صفحه‌ای اولیه
هستند؛ `mihan test` اجرایشان می‌کند و مکمل درس‌ها هستند.

## خلاصه‌ی کلمه‌های کلیدی

</div>

| Mihan | English | | Mihan | English |
|---|---|---|---|---|
| `chap_khat` | println | | `funct` | def / func |
| `agar` | if | | `bargard` | return |
| `warna` | else | | `darghan` | class |
| `har … dar` | for … in | | `khod` | self |
| `har i az a ta b` | for i in range(a, b) | | `koshish` / `giroft` | try / catch |
| `ta` | while | | `portab` | raise / throw |
| `shek` / `edame` | break / continue | | `biar` | import |
| `va` / `ya` / `na` | and / or / not | | `dorost` / `ghalat` / `hich` | true / false / null |
