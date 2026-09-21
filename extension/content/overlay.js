// Renders the overlay (QR code + actions) inside a Shadow DOM so the page's
// CSS cannot interfere with it, and vice-versa.
(() => {
  "use strict";

  const DirectAttachment = globalThis.DirectAttachment;

  let cssPromise = null;

  function loadCSS() {
    if (!cssPromise) {
      cssPromise = fetch(DirectAttachment.browser.api.runtime.getURL("overlay.css"))
        .then((res) => (res.ok ? res.text() : ""))
        .catch(() => "");
    }
    return cssPromise;
  }

  function escapeHTML(value) {
    const div = document.createElement("div");
    div.textContent = value;
    return div.innerHTML;
  }

  function formatSize(bytes) {
    if (bytes < 1024) return bytes + " B";
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + " KB";
    return (bytes / (1024 * 1024)).toFixed(1) + " MB";
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

    // Firefox-only result state: shows a download button because the file
    // cannot be injected automatically.
    renderReceived(file) {
      if (!this.shadow) return;

      const card = this.shadow.querySelector(".da-card");
      if (!card) return;

      card.innerHTML = `
        <h2>Foto recebida</h2>
        <p class="da-sub">${escapeHTML(file.name)} (${formatSize(file.size)})</p>
        <p class="da-status">
          O Firefox não permite anexar automaticamente ao campo da página.
          Baixe o arquivo abaixo e selecione-o manualmente.
        </p>
        <div class="da-actions">
          <button class="da-btn da-btn-primary" id="da-download" type="button">Baixar arquivo</button>
          <button class="da-btn" id="da-close" type="button">Fechar</button>
        </div>
      `;

      card.querySelector("#da-download").addEventListener("click", () => {
        DirectAttachment.browser.downloadBlob(file, file.name);
      });
      card.querySelector("#da-close").addEventListener("click", () => {
        if (this.onCancel) this.onCancel();
      });
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
