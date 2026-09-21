// Entrypoint: orchestrates the flow from an intercepted file input through
// session creation, QR display, file reception and injection.
(() => {
  "use strict";

  const DirectAttachment = globalThis.DirectAttachment;

  let currentInput = null;

  function reset() {
    currentInput = null;
  }

  DirectAttachment.index = {
    handleInput(input) {
      // If another flow is already open, release it first.
      if (currentInput) {
        DirectAttachment.transfer.cancel();
        DirectAttachment.overlay.close();
      }
      currentInput = input;

      DirectAttachment.overlay.onNative = () => {
        DirectAttachment.transfer.cancel();
        const target = currentInput;
        DirectAttachment.overlay.close();
        reset();
        if (target) DirectAttachment.interceptor.openNative(target);
      };

      DirectAttachment.overlay.onCancel = () => {
        DirectAttachment.transfer.cancel();
        DirectAttachment.overlay.close();
        reset();
      };

      DirectAttachment.overlay.open().then(async () => {
        // The user may have cancelled while the overlay was mounting.
        if (currentInput !== input) return;

        try {
          const session = await DirectAttachment.transfer.createSession();
          if (currentInput !== input) return;

          DirectAttachment.transfer.openWS(
            session.sessionId,
            session.token,
            (meta, bytes) => {
              const target = currentInput;
              if (target) {
                try {
                  DirectAttachment.injector.inject(target, bytes, meta);
                } catch (err) {
                  // Ignore injection errors; the page remains untouched.
                }
              }
              DirectAttachment.transfer.cancel();
              DirectAttachment.overlay.close();
              reset();
            },
            (message) => {
              DirectAttachment.overlay.setState(message, "error");
            },
          );

          DirectAttachment.overlay.setQR(session.url);
        } catch (_) {
          DirectAttachment.overlay.setState("Não foi possível criar a sessão.", "error");
        }
      });
    },
  };
})();
