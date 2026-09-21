// Browser abstraction and feature detection. Loaded first so every other
// module can rely on DirectAttachment.browser without knowing which browser is
// hosting the extension.
(() => {
  "use strict";

  const DirectAttachment = (globalThis.DirectAttachment = globalThis.DirectAttachment || {});

  const api = globalThis.browser || globalThis.chrome;

  const isFirefox =
    typeof globalThis.InstallTrigger !== "undefined" || /Firefox\//.test(globalThis.navigator.userAgent);

  DirectAttachment.browser = {
    // The WebExtension API namespace (browser.* in Firefox, chrome.* in both).
    api,

    // Chrome is the only engine that allows assigning input.files.
    isFirefox,
    isChrome: !isFirefox && !!api,

    // Fire-and-forget download of a Blob, used as the Firefox fallback.
    downloadBlob(blob, name) {
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = name;
      a.style.display = "none";
      document.documentElement.appendChild(a);
      a.click();
      setTimeout(() => {
        document.documentElement.removeChild(a);
        URL.revokeObjectURL(url);
      }, 1000);
    },
  };
})();
