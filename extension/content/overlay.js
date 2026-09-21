// Renders the overlay (QR code + actions) inside a Shadow DOM so the page's
// CSS cannot interfere with it, and vice-versa.
(() => {
  "use strict";

  const DirectAttachment = globalThis.DirectAttachment;

  let cssPromise = null;

  function loadCSS() {
    if (!cssPromise) {
      cssPromise = fetch(chrome.runtime.getURL("overlay.css"))
        .then((res) => (res.ok ? res.text() : ""))
        .catch(() => "");
    }
    return cssPromise;
  }

  class Overlay {
    constructor() {
      this.host = null;
      this.shadow = null;
      this.qrEl = null;
      this.statusEl = null;
      this.onNative = null;
      this.onCancel = null;
    }

    ensure() {
      if (this.host) return;

      const host = document.createElement("div");
      host.id = "direct-attachment-overlay-host";
      host.style.cssText = "all: initial; position: fixed; inset: 0; z-index: 2147483647;";
      this.shadow = host.attachShadow({ mode: "open" });
      document.documentElement.appendChild(host);
      this.host = host;
    }

    async open() {
      this.ensure();

      const css = await loadCSS();
      this.shadow.innerHTML = `
        <style>${css}</style>
        <div class="da-backdrop">
          <div class="da-card" role="dialog" aria-label="Anexar pelo celular">
            <h2>Anexar pelo celular</h2>
            <p class="da-sub">Escaneie o QR Code para enviar uma foto do celular.</p>
            <div class="da-qr" id="da-qr"></div>
            <p class="da-status" id="da-status"></p>
            <div class="da-actions">
              <button class="da-btn" id="da-native" type="button">Usar arquivo do PC</button>
              <button class="da-btn" id="da-cancel" type="button">Cancelar</button>
            </div>
          </div>
        </div>
      `;

      this.qrEl = this.shadow.querySelector("#da-qr");
      this.statusEl = this.shadow.querySelector("#da-status");

      this.shadow.querySelector("#da-native").addEventListener("click", () => {
        if (this.onNative) this.onNative();
      });
      this.shadow.querySelector("#da-cancel").addEventListener("click", () => {
        if (this.onCancel) this.onCancel();
      });
      this.shadow.querySelector(".da-backdrop").addEventListener("click", (e) => {
        if (e.target.classList.contains("da-backdrop") && this.onCancel) this.onCancel();
      });

      this.setState("Gerando QR Code...", "loading");
    }

    setQR(url) {
      if (!this.qrEl) return;
      this.qrEl.innerHTML = "";
      try {
        new QRCode(this.qrEl, {
          text: url,
          width: 220,
          height: 220,
          correctLevel: QRCode.CorrectLevel.M,
        });
      } catch (err) {
        this.setState("Não foi possível gerar o QR Code.", "error");
        return;
      }
      this.setState("Aguarde o envio da foto...", "waiting");
    }

    setState(message, kind) {
      if (!this.statusEl) return;
      this.statusEl.textContent = message || "";
      this.statusEl.className = "da-status" + (kind ? " da-" + kind : "");
    }

    close() {
      if (this.host) {
        this.host.remove();
      }
      this.host = null;
      this.shadow = null;
      this.qrEl = null;
      this.statusEl = null;
      this.onNative = null;
      this.onCancel = null;
    }
  }

  DirectAttachment.overlay = new Overlay();
})();
