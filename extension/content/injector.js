// Injects the received file into the original <input type="file">.
//
// Behavior differs per browser because Chrome is the only engine that lets you
// assign input.files (via DataTransfer), which makes the file appear exactly as
// if the user had picked it from disk. Firefox keeps `input.files` read-only,
// so there we fall back to a synthetic drop event plus a manual download.
(() => {
  "use strict";

  const DirectAttachment = globalThis.DirectAttachment;

  function makeFile(bytes, meta) {
    const type = meta.contentType || "application/octet-stream";
    const name = meta.name || "photo";
    const blob = new Blob([bytes], { type });
    return new File([blob], name, { type, lastModified: Date.now() });
  }

  function injectNative(input, file) {
    const dt = new DataTransfer();
    dt.items.add(file);
    input.files = dt.files;

    // Dispatch both events: 'change' for standard inputs and 'input' for
    // frameworks that listen to it (React, Vue, etc.).
    input.dispatchEvent(new Event("change", { bubbles: true }));
    input.dispatchEvent(new Event("input", { bubbles: true }));
  }

  // Attempts to deliver the file via a synthetic drop. This only helps sites
  // that have their own drag-and-drop handlers; a native file input will simply
  // ignore it. Safe to call and ignore the outcome.
  function tryDrop(input, file) {
    try {
      const dt = new DataTransfer();
      dt.items.add(file);
      input.dispatchEvent(
        new DragEvent("drop", {
          bubbles: true,
          cancelable: true,
          dataTransfer: dt,
        }),
      );
    } catch (_) {
      // Constructor support varies; the download fallback remains available.
    }
  }

  DirectAttachment.injector = {
    // Chrome-only seamless injection.
    inject(input, bytes, meta) {
      injectNative(input, makeFile(bytes, meta));
    },

    // Cross-browser delivery. Returns the outcome so the caller can decide
    // whether to close the overlay or offer a download.
    deliver(input, bytes, meta) {
      const file = makeFile(bytes, meta);

      if (!DirectAttachment.browser.isFirefox) {
        injectNative(input, file);
        return { mode: "injected" };
      }

      tryDrop(input, file);
      return { mode: "firefox", file };
    },
  };
})();
