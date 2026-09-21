// Handles the pairing session and the WebSocket used to receive the file from
// the phone (relayed by the Go backend).
(() => {
  "use strict";

  const DirectAttachment = globalThis.DirectAttachment;
  const config = DirectAttachment.config;

  DirectAttachment.transfer = {
    ws: null,
    currentSession: null,

    origin() {
      return String(config.backendOrigin || "").replace(/\/+$/, "");
    },

    async createSession() {
      const res = await fetch(this.origin() + "/api/sessions", { method: "POST" });
      if (!res.ok) {
        throw new Error("Não foi possível criar a sessão.");
      }
      return res.json();
    },

    openWS(sessionId, token, onReceive, onError) {
      const wsUrl =
        this.origin().replace(/^http/, "ws") +
        "/ws/transfer?session=" +
        encodeURIComponent(sessionId) +
        "&token=" +
        encodeURIComponent(token);

      const ws = new WebSocket(wsUrl);
      ws.binaryType = "arraybuffer";

      this.ws = ws;
      this.currentSession = { sessionId, token };

      let meta = null;

      ws.onmessage = (event) => {
        if (typeof event.data === "string") {
          try {
            const msg = JSON.parse(event.data);
            if (msg.type === "meta") meta = msg;
          } catch (_) {
            // ignore malformed frames
          }
          return;
        }

        // Binary frame: the file bytes, following a metadata frame.
        if (meta && onReceive) {
          onReceive(meta, new Uint8Array(event.data));
          meta = null;
        }
      };

      ws.onerror = () => {
        if (onError) onError("Falha de conexão com o servidor.");
      };

      ws.onclose = () => {
        if (this.ws === ws) this.ws = null;
      };
    },

    async cancel() {
      if (this.ws) {
        try {
          this.ws.close();
        } catch (_) {
          // already closed
        }
        this.ws = null;
      }

      const session = this.currentSession;
      this.currentSession = null;
      if (session) {
        try {
          await fetch(
            this.origin() +
              "/api/sessions/" +
              encodeURIComponent(session.sessionId) +
              "?t=" +
              encodeURIComponent(session.token),
            { method: "DELETE" },
          );
        } catch (_) {
          // best-effort cleanup; the server will expire the session anyway
        }
      }
    },
  };
})();
