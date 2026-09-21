// Single source of truth for backend configuration.
//
// Set backendOrigin to the public URL of your deployed backend. During local
// development the default http://localhost:8080 works because Chrome treats
// localhost as a secure context (and exempts it from mixed-content blocking).
(() => {
  "use strict";

  const DirectAttachment = (globalThis.DirectAttachment = globalThis.DirectAttachment || {});

  DirectAttachment.config = {
    backendOrigin: "https://direct-attachment-production.up.railway.app",
  };
})();
