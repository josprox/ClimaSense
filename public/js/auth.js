(() => {
  "use strict";
  const send = async (form, endpoint) => {
    const message = document.getElementById("form-message");
    message.className = "form-message";
    message.textContent = "Procesando…";
    const body = Object.fromEntries(new FormData(form));
    const response = await fetch(endpoint, {method:"POST",headers:{"Content-Type":"application/json","Accept":"application/json"},body:JSON.stringify(body)});
    const data = await response.json().catch(() => ({}));
    if (!response.ok) throw new Error(data.error || `HTTP ${response.status}`);
    message.classList.add("success");
    message.textContent = data.message || "Listo";
    if (data.redirect) window.location.assign(data.redirect);
  };
  const login = document.getElementById("login-form");
  if (login) login.addEventListener("submit", (event) => { event.preventDefault(); send(login,"/api/v1/auth/login").catch(error => document.getElementById("form-message").textContent=error.message); });
  const setup = document.getElementById("setup-form");
  if (setup) setup.addEventListener("submit", (event) => { event.preventDefault(); send(setup,"/api/v1/setup/admin").catch(error => document.getElementById("form-message").textContent=error.message); });
})();
