// Renders the overlay (QR code + actions) inside a Shadow DOM so the page's
// CSS cannot interfere with it, and vice-versa.
(() => {
  "use strict";

  const DirectAttachment = globalThis.DirectAttachment;

  function formatSize(bytes) {
    if (bytes < 1024) return bytes + " B";
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + " KB";
    return (bytes / (1024 * 1024)).toFixed(1) + " MB";
  }

  // Cleans a user-entered file name: trims, strips path separators, falls back
  // to the default when empty, and re-attaches the original extension if the
  // user removed it.
  function sanitizeName(input, fallback) {
    let name = String(input || "").trim().replace(/[/\\]/g, "");
    if (!name) return fallback;

    const extMatch = /(\.[A-Za-z0-9]{1,10})$/.exec(fallback);
    if (extMatch && !/\.[A-Za-z0-9]{1,10}$/.test(name)) {
      name += extMatch[1];
    }
    return name;
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

      const css = DirectAttachment.overlayCSS || "";
      const style = document.createElement("style");
      style.textContent = css;

      const backdrop = document.createElement("div");
      backdrop.className = "da-backdrop";
      const card = document.createElement("div");
      card.className = "da-card";
      card.setAttribute("role", "dialog");
      card.setAttribute("aria-label", "Anexar pelo celular");

      const title = document.createElement("h2");
      title.textContent = "Anexar pelo celular";
      const subtitle = document.createElement("p");
      subtitle.className = "da-sub";
      subtitle.textContent = "Escaneie o QR Code para enviar uma foto do celular.";
      this.qrEl = document.createElement("div");
      this.qrEl.className = "da-qr";
      this.statusEl = document.createElement("p");
      this.statusEl.className = "da-status";

      const actions = document.createElement("div");
      actions.className = "da-actions";
      const nativeButton = document.createElement("button");
      nativeButton.className = "da-btn";
      nativeButton.type = "button";
      nativeButton.textContent = "Usar arquivo do PC";
      const cancelButton = document.createElement("button");
      cancelButton.className = "da-btn";
      cancelButton.type = "button";
      cancelButton.textContent = "Cancelar";

      actions.append(nativeButton, cancelButton);
      card.append(title, subtitle, this.qrEl, this.statusEl, actions);
      backdrop.appendChild(card);
      this.shadow.replaceChildren(style, backdrop);

      nativeButton.addEventListener("click", () => {
        if (this.onNative) this.onNative();
      });
      cancelButton.addEventListener("click", () => {
        if (this.onCancel) this.onCancel();
      });
      backdrop.addEventListener("click", (e) => {
        if (e.target.classList.contains("da-backdrop") && this.onCancel) this.onCancel();
      });

      this.setState("Gerando QR Code...", "loading");
    }

    setQR(url) {
      if (!this.qrEl) return;
      this.qrEl.replaceChildren();
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

    // Received-file state: lets the user rename the file (pre-filled with the
    // default name) before attaching or downloading it.
    renderReceived(file, handlers) {
      if (!this.shadow) return;

      const card = this.shadow.querySelector(".da-card");
      if (!card) return;

      const isFirefox = DirectAttachment.browser.isFirefox;

      const title = document.createElement("h2");
      title.textContent = "Foto recebida";
      const size = document.createElement("p");
      size.className = "da-sub";
      size.textContent = formatSize(file.size);
      const label = document.createElement("label");
      label.className = "da-label";
      label.htmlFor = "da-name";
      label.textContent = "Nome do arquivo";
      const nameInput = document.createElement("input");
      nameInput.className = "da-input";
      nameInput.id = "da-name";
      nameInput.type = "text";
      nameInput.value = file.name;
      const actions = document.createElement("div");
      actions.className = "da-actions";
      const attachButton = document.createElement("button");
      attachButton.className = "da-btn" + (isFirefox ? "" : " da-btn-primary");
      attachButton.type = "button";
      attachButton.textContent = "Anexar";
      const downloadButton = document.createElement("button");
      downloadButton.className = "da-btn" + (isFirefox ? " da-btn-primary" : "");
      downloadButton.type = "button";
      downloadButton.textContent = "Baixar";
      const closeButton = document.createElement("button");
      closeButton.className = "da-btn da-btn-ghost da-close-btn";
      closeButton.type = "button";
      closeButton.textContent = "Fechar";

      actions.append(attachButton, downloadButton);
      card.replaceChildren(title, size, label, nameInput, actions, closeButton);
      const defaultName = file.name;

      const finalName = () => sanitizeName(nameInput.value, defaultName);

      attachButton.addEventListener("click", () => {
        if (handlers.onAttach) handlers.onAttach(finalName());
      });
      downloadButton.addEventListener("click", () => {
        if (handlers.onDownload) handlers.onDownload(finalName());
      });
      closeButton.addEventListener("click", () => {
        if (handlers.onClose) handlers.onClose();
      });

      nameInput.select();
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
