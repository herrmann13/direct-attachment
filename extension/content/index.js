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
          const key = DirectAttachment.crypto.generateKey();
          const session = await DirectAttachment.transfer.createSession();
          if (currentInput !== input) return;

          DirectAttachment.transfer.openWS(
            session.sessionId,
            session.token,
            (meta, bytes) => {
              const plain = DirectAttachment.crypto.decrypt(bytes, key);
              if (!plain) {
                DirectAttachment.transfer.cancel();
                DirectAttachment.overlay.setState("Falha ao decifrar a foto.", "error");
                return;
              }

              const detected = DirectAttachment.injector.detectType(plain);
              if (!detected) {
                DirectAttachment.transfer.cancel();
                DirectAttachment.overlay.setState(
                  "O arquivo recebido não é uma imagem válida.",
                  "error",
                );
                return;
              }

              const file = DirectAttachment.injector.makeFile(plain, {
                name: meta.name,
                contentType: detected,
              });
              DirectAttachment.transfer.cancel();

              DirectAttachment.overlay.renderReceived(file, {
                onAttach(name) {
                  const target = currentInput;
                  if (target) {
                    const toAttach = DirectAttachment.injector.makeFile(plain, {
                      name,
                      contentType: detected,
                    });
                    DirectAttachment.injector.attach(target, toAttach);
                  }
                  DirectAttachment.overlay.close();
                  reset();
                },
                onDownload(name) {
                  DirectAttachment.browser.downloadBlob(file, name);
                  DirectAttachment.overlay.close();
                  reset();
                },
                onClose() {
                  DirectAttachment.overlay.close();
                  reset();
                },
              });
            },
            (message) => {
              DirectAttachment.overlay.setState(message, "error");
            },
          );

          // The key travels only in the URL fragment, which browsers never
          // send to the server.
          const qrUrl = session.url + "#k=" + DirectAttachment.crypto.keyToB64(key);
          DirectAttachment.overlay.setQR(qrUrl);
        } catch (_) {
          DirectAttachment.overlay.setState("Não foi possível criar a sessão.", "error");
        }
      });
    },
  };
})();
