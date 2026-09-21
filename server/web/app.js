(() => {
  "use strict";

  // ---- Session info from the URL (/s/{id}?t={token}#k={key}) --------------
  const parts = window.location.pathname.split("/").filter(Boolean);
  const sessionId = parts[1] || "";
  const token = new URLSearchParams(window.location.search).get("t") || "";
  const key = readKeyFromFragment();

  const MAX_SIZE = 15 * 1024 * 1024; // 15 MiB, keep in sync with server
  const NONCE_LEN = 24;

  // Decodes the base64url encryption key carried in the URL fragment. Fragments
  // never leave the device, so the server does not know the key.
  function readKeyFromFragment() {
    try {
      const m = /(?:^|&)k=([^&]+)/.exec(window.location.hash.slice(1));
      if (!m) return null;
      const b64 = m[1].replace(/-/g, "+").replace(/_/g, "/");
      const pad = b64.length % 4 === 0 ? "" : "=".repeat(4 - (b64.length % 4));
      const bin = atob(b64 + pad);
      const bytes = new Uint8Array(bin.length);
      for (let i = 0; i < bin.length; i++) bytes[i] = bin.charCodeAt(i);
      return bytes.length === 32 ? bytes : null;
    } catch (_) {
      // Malformed fragment: degrade to "no key" instead of killing the page.
      return null;
    }
  }

  // ---- Elements -----------------------------------------------------------
  const $ = (id) => document.getElementById(id);
  const preview = $("preview");
  const previewImg = $("preview-img");
  const actions = $("actions");
  const cameraPanel = $("camera-panel");
  const video = $("video");
  const fileInput = $("file");
  const sendBtn = $("send");
  const statusEl = $("status");
  const cropPanel = $("crop-panel");
  const cropImg = $("crop-img");

  let selectedFile = null;
  let stream = null;
  let cropper = null;
  let cropSource = null; // Blob|File shown in the cropper

  function setStatus(message, kind) {
    statusEl.textContent = message;
    statusEl.className = "status" + (kind ? " " + kind : "");
  }

  function showPreview(file) {
    selectedFile = file;
    previewImg.src = URL.createObjectURL(file);
    preview.classList.remove("hidden");
    actions.classList.add("hidden");
    sendBtn.classList.remove("hidden");
    setStatus("");
  }

  function reset() {
    selectedFile = null;
    if (previewImg.src) URL.revokeObjectURL(previewImg.src);
    previewImg.removeAttribute("src");
    preview.classList.add("hidden");
    sendBtn.classList.add("hidden");
    actions.classList.remove("hidden");
    fileInput.value = "";
    setStatus("");
  }

  function stopCamera() {
    if (stream) {
      stream.getTracks().forEach((t) => t.stop());
      stream = null;
    }
    cameraPanel.classList.add("hidden");
    actions.classList.remove("hidden");
  }

  async function openCamera() {
    setStatus("");
    if (!navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
      setStatus("Câmera indisponível neste navegador.", "error");
      return;
    }
    try {
      stream = await navigator.mediaDevices.getUserMedia({
        video: {
          facingMode: "environment",
          width: { ideal: 3840 },
          height: { ideal: 2160 },
        },
        audio: false,
      });
      video.srcObject = stream;
      actions.classList.add("hidden");
      cameraPanel.classList.remove("hidden");
    } catch (err) {
      setStatus("Não foi possível acessar a câmera. Use a galeria.", "error");
      stream = null;
    }
  }

  async function capturePhoto() {
    // Prefer ImageCapture, which grabs a full-resolution frame from the camera
    // sensor (bypassing the lower-resolution preview stream).
    const track = stream && stream.getVideoTracks()[0];
    if (window.ImageCapture && track) {
      try {
        const blob = await new ImageCapture(track).takePhoto();
        stopCamera();
        startCrop(blob);
        return;
      } catch (_) {
        // fall through to the canvas path below
      }
    }

    const canvas = document.createElement("canvas");
    canvas.width = video.videoWidth;
    canvas.height = video.videoHeight;
    canvas.getContext("2d").drawImage(video, 0, 0);
    canvas.toBlob(
      (blob) => {
        if (!blob) {
          setStatus("Falha ao capturar a foto.", "error");
          return;
        }
        stopCamera();
        startCrop(blob);
      },
      "image/jpeg",
      0.92,
    );
  }

  // Opens the crop panel for a Blob|File and initializes Cropper.js.
  function startCrop(source) {
    cropSource = source;
    cropImg.src = URL.createObjectURL(source);
    actions.classList.add("hidden");
    cropPanel.classList.remove("hidden");

    if (cropper) {
      cropper.destroy();
      cropper = null;
    }
    cropper = new Cropper(cropImg, {
      viewMode: 1,
      autoCropArea: 1,
      responsive: true,
      background: false,
      guides: true,
    });
  }

  function destroyCropper() {
    if (cropper) {
      cropper.destroy();
      cropper = null;
    }
    if (cropImg.src) URL.revokeObjectURL(cropImg.src);
    cropImg.removeAttribute("src");
    cropSource = null;
    cropPanel.classList.add("hidden");
  }

  // Renders the current crop box to a JPEG and continues to the preview.
  function applyCrop() {
    if (!cropper) return;
    const canvas = cropper.getCroppedCanvas({ maxWidth: 2560, maxHeight: 2560 });
    if (!canvas) {
      skipCrop();
      return;
    }
    canvas.toBlob(
      (blob) => {
        const file = new File([blob], `photo-${Date.now()}.jpg`, { type: "image/jpeg" });
        destroyCropper();
        showPreview(file);
      },
      "image/jpeg",
      0.92,
    );
  }

  function skipCrop() {
    const source = cropSource;
    destroyCropper();
    if (source) showPreview(source instanceof File ? source : new File([source], `photo-${Date.now()}.jpg`, { type: "image/jpeg" }));
  }

  function cancelCrop() {
    destroyCropper();
    actions.classList.remove("hidden");
  }

  function validate(file) {
    if (!file) return false;
    const okType = ["image/jpeg", "image/png", "image/webp"].includes(file.type);
    if (!okType) {
      setStatus("Formato não suportado. Use JPEG, PNG ou WebP.", "error");
      return false;
    }
    if (file.size > MAX_SIZE) {
      setStatus("Arquivo muito grande (máx. 15 MB).", "error");
      return false;
    }
    return true;
  }

  async function send() {
    if (!selectedFile || !validate(selectedFile)) return;
    if (!key) {
      setStatus("Chave de criptografia ausente. Escaneie o QR code novamente.", "error");
      return;
    }

    sendBtn.disabled = true;
    setStatus("Criptografando e enviando...");

    let payload;
    try {
      payload = await encryptFile(selectedFile, key);
    } catch (err) {
      setStatus("Falha ao criptografar a foto.", "error");
      sendBtn.disabled = false;
      return;
    }

    const name = selectedFile.name || "photo.jpg";
    const form = new FormData();
    form.append("file", new Blob([payload]), name);
    form.append("name", name);
    form.append("contentType", selectedFile.type || "application/octet-stream");

    try {
      const res = await fetch(
        `/api/upload/${encodeURIComponent(sessionId)}?t=${encodeURIComponent(token)}`,
        { method: "POST", body: form },
      );

      if (res.ok) {
        selectedFile = null;
        if (previewImg.src) URL.revokeObjectURL(previewImg.src);
        previewImg.removeAttribute("src");
        preview.classList.add("hidden");
        sendBtn.classList.add("hidden");
        actions.classList.add("hidden");
        setStatus("Enviado! Você já pode voltar ao computador.", "success");
        return;
      }

      const body = await res.json().catch(() => null);
      const code = body?.error?.code || "";
      const message = body?.error?.message || "Erro ao enviar.";
      setStatus(mapError(code, message), "error");
    } catch (err) {
      setStatus("Falha de conexão. Tente novamente.", "error");
    } finally {
      sendBtn.disabled = false;
    }
  }

  // Encrypts a File with nacl.secretbox (XSalsa20-Poly1305) and returns
  // nonce(24) || ciphertext, so the server only ever sees opaque bytes.
  async function encryptFile(file, key) {
    const buf = await file.arrayBuffer();
    const message = new Uint8Array(buf);
    const nonce = nacl.randomBytes(NONCE_LEN);
    const box = nacl.secretbox(message, nonce, key);

    const payload = new Uint8Array(NONCE_LEN + box.length);
    payload.set(nonce, 0);
    payload.set(box, NONCE_LEN);
    return payload;
  }

  function mapError(code, fallback) {
    switch (code) {
      case "session_expired":
        return "Este link expirou. Escaneie o QR code novamente.";
      case "no_receiver":
        return "A sessão no computador foi encerrada.";
      case "file_too_large":
        return "Arquivo muito grande (máx. 15 MB).";
      case "unsupported_type":
        return "Formato não suportado. Use JPEG, PNG ou WebP.";
      case "invalid_token":
      case "session_not_found":
        return "Link inválido. Escaneie o QR code novamente.";
      default:
        return fallback;
    }
  }

  // ---- Wire up ------------------------------------------------------------
  $("camera").addEventListener("click", openCamera);
  $("gallery").addEventListener("click", () => fileInput.click());
  $("capture").addEventListener("click", capturePhoto);
  $("cancel-camera").addEventListener("click", stopCamera);
  $("reset").addEventListener("click", reset);
  sendBtn.addEventListener("click", send);
  $("crop-confirm").addEventListener("click", applyCrop);
  $("crop-skip").addEventListener("click", skipCrop);
  $("crop-cancel").addEventListener("click", cancelCrop);

  fileInput.addEventListener("change", () => {
    const file = fileInput.files && fileInput.files[0];
    if (file && validate(file)) startCrop(file);
  });

  if (!sessionId || !token) {
    setStatus("Link inválido. Escaneie o QR code novamente.", "error");
    actions.classList.add("hidden");
  }
})();
