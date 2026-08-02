(() => {
  "use strict";

  const tokenEndpoint = "/api/v1/auth/ws-token";
  const socketPath = "/ws/dashboard";

  function start(options = {}) {
    const refresh = typeof options.onRefresh === "function" ? options.onRefresh : () => {};
    const status = typeof options.onStatus === "function" ? options.onStatus : () => {};
    const fallbackMs = Number(options.fallbackMs) > 0 ? Number(options.fallbackMs) : 30000;
    let socket = null;
    let stopped = false;
    let ready = false;
    let attempts = 0;
    let reconnectTimer = null;
    let refreshTimer = null;
    let handshakeTimer = null;
    let generation = 0;

    const scheduleRefresh = () => {
      clearTimeout(refreshTimer);
      refreshTimer = setTimeout(() => refresh(), 200);
    };

    const scheduleReconnect = () => {
      if (stopped || reconnectTimer || navigator.onLine === false) return;
      const delay = Math.min(30000, 1000 * (2 ** Math.min(attempts, 5))) + Math.floor(Math.random() * 500);
      attempts += 1;
      status("reconnecting");
      reconnectTimer = setTimeout(() => {
        reconnectTimer = null;
        connect();
      }, delay);
    };

    const connect = async () => {
      if (stopped || navigator.onLine === false) return;
      const ownGeneration = ++generation;
      ready = false;
      status("connecting");
      try {
        const response = await fetch(tokenEndpoint, {
          headers: { Accept: "application/json" },
          credentials: "same-origin",
          cache: "no-store"
        });
        if (response.status === 401) {
          location.assign("/login");
          return;
        }
        if (!response.ok) throw new Error(`HTTP ${response.status}`);
        const data = await response.json();
        if (!data.token) throw new Error("token WebSocket ausente");
        if (stopped || ownGeneration !== generation) return;

        const protocol = location.protocol === "https:" ? "wss:" : "ws:";
        socket = new WebSocket(`${protocol}//${location.host}${socketPath}`);
        socket.addEventListener("open", () => {
          socket.send(JSON.stringify({ type: "authenticate", token: data.token }));
          clearTimeout(handshakeTimer);
          handshakeTimer = setTimeout(() => {
            if (!ready && socket) socket.close();
          }, 10000);
        });
        socket.addEventListener("message", (event) => {
          let message;
          try { message = JSON.parse(event.data); } catch (_) { return; }
          if (message.type === "ready") {
            clearTimeout(handshakeTimer);
            ready = true;
            attempts = 0;
            status("live");
            return;
          }
          if (message.type === "refresh") {
            scheduleRefresh();
            if (message.resource === "scope_changed") {
              setTimeout(() => reconnect(), 300);
            }
          }
        });
        socket.addEventListener("error", () => status("fallback"));
        socket.addEventListener("close", () => {
          clearTimeout(handshakeTimer);
          ready = false;
          socket = null;
          status("fallback");
          scheduleReconnect();
        });
      } catch (_) {
        ready = false;
        status("fallback");
        scheduleReconnect();
      }
    };

    const reconnect = () => {
      generation += 1;
      clearTimeout(reconnectTimer);
      reconnectTimer = null;
      if (socket) {
        const previous = socket;
        socket = null;
        previous.close();
      } else {
        connect();
      }
    };

    const fallbackTimer = setInterval(() => {
      if (!stopped) refresh();
    }, fallbackMs);

    const stop = () => {
      stopped = true;
      generation += 1;
      clearInterval(fallbackTimer);
      clearTimeout(reconnectTimer);
      clearTimeout(refreshTimer);
      clearTimeout(handshakeTimer);
      if (socket) socket.close();
      socket = null;
    };

    window.addEventListener("online", reconnect);
    window.addEventListener("offline", () => status("offline"));
    document.addEventListener("visibilitychange", () => {
      if (document.visibilityState === "visible" && !ready) reconnect();
    });
    window.addEventListener("beforeunload", stop, { once: true });
    connect();
    return { stop, reconnect, isReady: () => ready };
  }

  window.ClimaSenseLive = { start };
})();
