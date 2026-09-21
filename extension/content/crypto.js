// End-to-end encryption helpers. The PC generates an ephemeral key, embeds it
// in the QR code fragment (#k=...) so it never reaches the server, and uses it
// to decrypt the bytes relayed back from the phone.
(() => {
  "use strict";

  const DirectAttachment = globalThis.DirectAttachment;

  const NONCE_LEN = 24;
  const KEY_LEN = 32;

  function bytesToB64(bytes) {
    let bin = "";
    for (let i = 0; i < bytes.length; i++) bin += String.fromCharCode(bytes[i]);
    return btoa(bin).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "");
  }

  DirectAttachment.crypto = {
    NONCE_LEN,
    KEY_LEN,

    generateKey() {
      return nacl.randomBytes(KEY_LEN);
    },

    keyToB64(key) {
      return bytesToB64(key);
    },

    // Decrypts a payload of the form nonce(24) || secretbox(ciphertext).
    // Returns the plaintext bytes, or null on failure.
    decrypt(payload, key) {
      if (!payload || payload.length <= NONCE_LEN) return null;
      const nonce = payload.subarray(0, NONCE_LEN);
      const box = payload.subarray(NONCE_LEN);
      const out = nacl.secretbox.open(box, nonce, key);
      return out || null;
    },
  };
})();
