// Injects the received file into the original <input type="file"> so the page
// behaves exactly as if the user had picked it from disk.
(() => {
  "use strict";

  const DirectAttachment = globalThis.DirectAttachment;

  DirectAttachment.injector = {
    inject(input, bytes, meta) {
      const type = meta.contentType || "application/octet-stream";
      const name = meta.name || "photo";

      const blob = new Blob([bytes], { type });
      const file = new File([blob], name, { type, lastModified: Date.now() });

      const dt = new DataTransfer();
      dt.items.add(file);

      input.files = dt.files;

      // Dispatch both events: 'change' for standard inputs and 'input' for
      // frameworks that listen to it (React, Vue, etc.).
      input.dispatchEvent(new Event("change", { bubbles: true }));
      input.dispatchEvent(new Event("input", { bubbles: true }));
    },
  };
})();
