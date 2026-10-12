package org.mihan.runner;

import android.app.Activity;
import android.content.Context;
import android.content.Intent;
import android.os.Build;
import android.os.VibrationEffect;
import android.os.Vibrator;
import android.webkit.JavascriptInterface;
import android.webkit.ValueCallback;
import android.widget.Toast;
import android.os.Bundle;
import android.os.Handler;
import android.os.Looper;
import android.text.Html;
import android.webkit.WebSettings;
import android.webkit.WebView;
import android.webkit.WebViewClient;

import java.io.File;
import java.io.FileInputStream;
import java.io.InputStream;
import java.nio.charset.StandardCharsets;
import java.util.Map;

public class MainActivity extends Activity {
    private WebView web;
    private Process proc;
    private final StringBuilder out = new StringBuilder();
    private final Handler ui = new Handler(Looper.getMainLooper());
    private File urlFile;
    private String shown = "";
    private boolean finished;

    @Override
    protected void onCreate(Bundle saved) {
        super.onCreate(saved);
        web = new WebView(this);
        WebSettings s = web.getSettings();
        s.setJavaScriptEnabled(true);
        s.setDomStorageEnabled(true);
        web.setWebViewClient(new WebViewClient());
        web.addJavascriptInterface(new Bridge(), "MihanAndroid");
        setContentView(web);
        showText("…");
        start();
    }

    private void start() {
        urlFile = new File(getFilesDir(), "mihan-window.txt");
        urlFile.delete();
        String exe = getApplicationInfo().nativeLibraryDir + "/libmihanapp.so";
        try {
            ProcessBuilder pb = new ProcessBuilder(exe);
            Map<String, String> env = pb.environment();
            env.put("HOME", getFilesDir().getAbsolutePath());
            env.put("XDG_CACHE_HOME", getCacheDir().getAbsolutePath());
            env.put("TMPDIR", getCacheDir().getAbsolutePath());
            env.put("MIHAN_WINDOW_FILE", urlFile.getAbsolutePath());
            env.put("MIHAN_ANDROID", "1");
            pb.directory(getFilesDir());
            pb.redirectErrorStream(true);
            proc = pb.start();
        } catch (Exception e) {
            showText("could not start the program: " + e);
            return;
        }
        Thread reader = new Thread(new Runnable() { public void run() {
            byte[] buf = new byte[4096];
            try (InputStream in = proc.getInputStream()) {
                int n;
                while ((n = in.read(buf)) > 0) {
                    synchronized (out) {
                        out.append(new String(buf, 0, n, StandardCharsets.UTF_8));
                    }
                }
            } catch (Exception ignored) {
            }
            try {
                proc.waitFor();
            } catch (InterruptedException ignored) {
            }
            finished = true;
        }});
        reader.setDaemon(true);
        reader.start();
        ui.postDelayed(poller, 100);
    }

    private final Runnable poller = new Runnable() {
        public void run() {
            poll();
        }
    };

    private void poll() {
        String url = readUrl();
        if (url != null && !url.equals(shown)) {
            shown = url;
            web.loadUrl(url);
        } else if (shown.isEmpty() || !shown.contains("://")) {
            String text;
            synchronized (out) {
                text = out.toString();
            }
            if (!text.equals(shown)) {
                showText(text);
                shown = text;
            }
        }
        if (!finished || url == null) {
            ui.postDelayed(poller, finished ? 1000 : 150);
        }
    }

    private String readUrl() {
        if (!urlFile.exists()) {
            return null;
        }
        try (FileInputStream f = new FileInputStream(urlFile)) {
            byte[] b = new byte[(int) urlFile.length()];
            int n = f.read(b);
            String s = new String(b, 0, Math.max(n, 0), StandardCharsets.UTF_8).trim();
            return s.isEmpty() ? null : s;
        } catch (Exception e) {
            return null;
        }
    }

    private void showText(String text) {
        String html = "<html><head><meta name='viewport' content='width=device-width'>"
                + "<style>body{font:15px monospace;padding:12px;white-space:pre-wrap;word-wrap:break-word}</style>"
                + "</head><body>" + Html.escapeHtml(text) + "</body></html>";
        web.loadDataWithBaseURL(null, html, "text/html", "utf-8", null);
    }

    public class Bridge {
        @JavascriptInterface
        public void toast(final String text) {
            ui.post(new Runnable() {
                public void run() {
                    Toast.makeText(MainActivity.this, text, Toast.LENGTH_SHORT).show();
                }
            });
        }

        @JavascriptInterface
        public void vibrate(int ms) {
            Vibrator v = (Vibrator) getSystemService(Context.VIBRATOR_SERVICE);
            if (v == null) {
                return;
            }
            if (Build.VERSION.SDK_INT >= 26) {
                v.vibrate(VibrationEffect.createOneShot(Math.max(1, ms), VibrationEffect.DEFAULT_AMPLITUDE));
            } else {
                v.vibrate(ms);
            }
        }

        @JavascriptInterface
        public void share(final String text, final String title) {
            ui.post(new Runnable() {
                public void run() {
                    Intent i = new Intent(Intent.ACTION_SEND);
                    i.setType("text/plain");
                    i.putExtra(Intent.EXTRA_TEXT, text);
                    if (title != null && !title.isEmpty()) {
                        i.putExtra(Intent.EXTRA_SUBJECT, title);
                    }
                    startActivity(Intent.createChooser(i, title == null || title.isEmpty() ? null : title));
                }
            });
        }
    }

    @Override
    public void onBackPressed() {

        web.evaluateJavascript("(window.__mihanBack && window.__mihanBack()) ? 'yes' : 'no'", new ValueCallback<String>() {
            public void onReceiveValue(String used) {
                if ("\"yes\"".equals(used)) {
                    return;
                }
                if (web.canGoBack()) {
                    web.goBack();
                } else {
                    finish();
                }
            }
        });
    }

    @Override
    protected void onPause() {
        super.onPause();
        if (web != null) {
            web.evaluateJavascript("window.dispatchEvent(new Event('mihan-pause'))", null);
        }
    }

    @Override
    protected void onResume() {
        super.onResume();
        if (web != null) {
            web.evaluateJavascript("window.dispatchEvent(new Event('mihan-resume'))", null);
        }
    }

    @Override
    protected void onDestroy() {
        if (proc != null) {
            proc.destroy();
        }
        super.onDestroy();
    }
}
