(() => {
  "use strict";
  const $ = id => document.getElementById(id);
  const esc = value => String(value ?? "").replace(/[&<>"']/g, c => ({"&":"&amp;","<":"&lt;",">":"&gt;",'"':"&quot;","'":"&#39;"}[c]));
  const date = value => value ? new Date(String(value).replace(" ","T")).toLocaleString("es-MX",{dateStyle:"medium",timeStyle:"short"}) : "—";
  const api = async (url, options={}) => { const response=await fetch(url,{headers:{Accept:"application/json",...(options.body?{"Content-Type":"application/json"}: {})},...options}); if(response.status===401){location.assign("/login");throw new Error("Sesión expirada");} const data=await response.json().catch(()=>({})); if(!response.ok)throw new Error(data.error||`HTTP ${response.status}`); return data; };
  const postForm = (url, form) => api(url,{method:"POST",body:JSON.stringify(Object.fromEntries(new FormData(form)))});
  const render = data => {
    const m=data.metrics||{}; $("m-organizations").textContent=m.organizations||0; $("m-locations").textContent=m.locations||0; $("m-devices").textContent=m.devices||0; $("m-codes").textContent=m.available_codes||0; $("m-alerts").textContent=m.alerts||0;
    const organizations=data.organizations||[];
    $("organizations-list").innerHTML=organizations.length?organizations.map(org=>`<div class="data-row"><div><strong>${esc(org.name)}</strong><small>${esc(org.contact_email)} · ${esc(org.plan)}</small></div><span class="tag">${esc(org.status)}</span></div>`).join(""):'<p class="empty-state">No hay empresas registradas.</p>';
    $("activation-organization").innerHTML=organizations.map(org=>`<option value="${Number(org.id)}">${esc(org.name)}</option>`).join("");
    const analyses=data.analyses||[]; $("admin-analysis").innerHTML=analyses.length?analyses.map(a=>`<div class="analysis-item ${esc(a.severity)}"><span class="severity"></span><div><strong>${esc(a.verdict)} · ${Number(a.indoor_avg_c).toFixed(1)}°C</strong><p>${esc(a.recommendation)}</p></div><time>${date(a.created_at)}</time></div>`).join(""):'<p class="empty-state">Sin análisis todavía.</p>';
    $("admin-updated").textContent=`Actualizado ${date(data.generated_at)}`;
  };
  const load=()=>api("/api/v1/admin/summary").then(render).catch(console.error);
  document.querySelectorAll("[data-open]").forEach(button=>button.addEventListener("click",()=>$(button.dataset.open).showModal()));
  document.querySelectorAll("[data-close]").forEach(button=>button.addEventListener("click",()=>$(button.dataset.close).close()));
  $("company-form").addEventListener("submit",event=>{event.preventDefault();postForm("/api/v1/admin/organizations",event.currentTarget).then(()=>{event.currentTarget.reset();$("company-dialog").close();load();}).catch(error=>$("company-message").textContent=error.message);});
  $("activation-form").addEventListener("submit",event=>{event.preventDefault();postForm("/api/v1/admin/activation-codes",event.currentTarget).then(data=>{$("activation-output").textContent=data.activation_code;load();}).catch(error=>$("activation-output").textContent=error.message);});
  document.querySelector("[data-logout]").addEventListener("click",()=>api("/api/v1/auth/logout",{method:"POST",body:"{}"}).then(()=>location.assign("/login")));
  $("refresh-admin").addEventListener("click",load); load();
})();
