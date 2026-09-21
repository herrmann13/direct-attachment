// Intercepts clicks on <input type="file"> so we can offer the phone flow
// instead of the native file picker.
//
// A capture-phase listener on the document catches every click on the input,
// including clicks synthesized by <label> wrappers and programmatic
// input.click(). Using composedPath() also catches inputs living inside a
// page's Shadow DOM (where event.target is retargeted to the shadow host).
(() => {
  "use strict";

  const DirectAttachment = globalThis.DirectAttachment;

  let allowNative = false;

  function findFileInput(target) {
    if (!target || target.nodeType !== Node.ELEMENT_NODE) return null;

    const path = typeof target.composedPath === "function" ? target.composedPath() : null;
    const candidates = path && path.length ? path : [target];

    for (const el of candidates) {
      if (el instanceof HTMLInputElement && el.type === "file") return el;
    }
    return null;
  }

  document.addEventListener(
    "click",
    (event) => {
      const input = findFileInput(event.target);
      if (!input) return;

      // The fallback "Usar arquivo do PC" button re-opens the native picker;
      // let that click through once.
      if (allowNative) {
        allowNative = false;
        return;
      }

      event.preventDefault();
      event.stopPropagation();

      if (DirectAttachment.index && DirectAttachment.index.handleInput) {
        DirectAttachment.index.handleInput(input);
      }
    },
    true,
  );

  DirectAttachment.interceptor = {
    // Re-opens the native file picker for the given input.
    openNative(input) {
      allowNative = true;
      input.click();
    },
  };
})();
