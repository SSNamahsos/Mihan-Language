

(function () {
  var A = window.MihanAndroid;
  var framed = window.parent && window.parent !== window;
  function toFrame(kind, data) {
    try { window.parent.postMessage({ mihan: kind, data: data }, "*"); } catch (e) {}
  }
  function smallToast(text) {
    var d = document.createElement("div");
    d.textContent = text;
    d.style.cssText = "position:fixed;left:50%;bottom:48px;transform:translateX(-50%);" +
      "background:rgba(30,30,30,.92);color:#fff;padding:10px 18px;border-radius:20px;" +
      "font:14px sans-serif;z-index:2147483647;transition:opacity .3s";
    document.body.appendChild(d);
    setTimeout(function () { d.style.opacity = "0"; }, 1700);
    setTimeout(function () { d.remove(); }, 2100);
  }
  var backHandlers = [];
  var mihan = {
    android: !!A || /[?&]mihan_android=1/.test(location.search) || framed && document.referrer.indexOf("/_preview") >= 0,
    toast: function (text) {
      text = String(text);
      if (A) return A.toast(text);
      if (framed) return toFrame("toast", text);
      smallToast(text);
    },
    vibrate: function (ms) {
      ms = ms || 200;
      if (A) return A.vibrate(ms);
      if (framed) return toFrame("vibrate", ms);
      if (navigator.vibrate) navigator.vibrate(ms);
    },
    share: function (text, title) {
      text = String(text);
      if (A) return A.share(text, title || "");
      if (framed) return toFrame("share", { text: text, title: title || "" });
      if (navigator.share) return navigator.share({ text: text, title: title || "" });
      smallToast(text);
    },

    onBack: function (f) { backHandlers.push(f); },
    onPause: function (f) { window.addEventListener("mihan-pause", f); },
    onResume: function (f) { window.addEventListener("mihan-resume", f); }
  };

  window.__mihanBack = function () {
    for (var i = backHandlers.length - 1; i >= 0; i--) {
      try { if (backHandlers[i]() === true) return true; } catch (e) { console.error(e); }
    }
    return false;
  };
  window.addEventListener("message", function (ev) {
    var m = ev.data || {};
    if (m.mihan === "back") {
      if (!window.__mihanBack()) {
        if (history.length > 1) history.back(); else toFrame("exit", null);
      }
    } else if (m.mihan === "pause" || m.mihan === "resume") {
      window.dispatchEvent(new Event("mihan-" + m.mihan));
    }
  });
  if (framed) {
    window.addEventListener("error", function (e) { toFrame("jserror", String(e.message) + " (" + e.filename + ":" + e.lineno + ")"); });
    var log = console.log;
    console.log = function () { toFrame("console", Array.prototype.join.call(arguments, " ")); log.apply(console, arguments); };
    toFrame("page", location.pathname);
  }
  window.mihan = mihan;
})();
