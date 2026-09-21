// Injects the received file into the original <input type="file">.
//
// Behavior differs per browser because Chrome is the only engine that lets you
// assign input.files (via DataTransfer), which makes the file appear exactly as
// if the user had picked it from disk. Firefox keeps `input.files` read-only,
// so there we fall back to a synthetic drop event plus a manual download.
(() => {
  "use strict";

  const DirectAttachment = globalThis.DirectAttachment;

  // Detects the real image type from magic bytes. Used to re-validate the
  // decrypted payload on the PC, since the server cannot inspect it.
  function detectImageType(bytes) {
    if (!bytes || bytes.length < 12) return null;

    // JPEG: FF D8 FF
    if (bytes[0] === 0xff && bytes[1] === 0xd8 && bytes[2] === 0xff) {
      return "image/jpeg";
    }
    // PNG: 89 50 4E 47 0D 0A 1A 0A
    if (
      bytes[0] === 0x89 && bytes[1] === 0x50 && bytes[2] === 0x4e && bytes[3] === 0x47 &&
      bytes[4] === 0x0d && bytes[5] === 0x0a && bytes[6] === 0x1a && bytes[7] === 0x0a
    ) {
      return "image/png";
    }
    // WebP: "RIFF" .... "WEBP"
    if (
      bytes[0] === 0x52 && bytes[1] === 0x49 && bytes[2] === 0x46 && bytes[3] === 0x46 &&
      bytes[8] === 0x57 && bytes[9] === 0x45 && bytes[10] === 0x42 && bytes[11] === 0x50
    ) {
      return "image/webp";
    }
    return null;
  }

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
    // Re-validates decrypted bytes and returns the detected MIME type, or null
    // when the payload is not a supported image.
    detectType(bytes) {
      return detectImageType(bytes);
    },

    // Builds a File from decrypted bytes and metadata (name, contentType).
    makeFile(bytes, meta) {
      return makeFile(bytes, meta);
    },

    // Delivers a File to the page: injects into the input on Chrome, or fires a
    // synthetic drop on Firefox.
    attach(input, file) {
      if (DirectAttachment.browser.isFirefox) {
        tryDrop(input, file);
      } else {
        injectNative(input, file);
      }
    },
  };
})();
