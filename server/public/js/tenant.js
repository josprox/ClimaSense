(() => {
  "use strict";
  const $=id=>document.getElementById(id); const esc=value=>String(value??"").replace(/[&<>"']/g,c=>({"&":"&amp;","<":"&lt;",">":"&gt;",'"':"&quot;","'":"&#39;"}[c]));
  const date=value=>value?new Date(String(value).replace(" ","T")).toLocaleString("es-MX",{dateStyle:"medium",timeStyle:"short"}):"—";
  const api=async(url,options={})=>{const response=await fetch(url,{headers:{Accept:"application/json",...(options.body?{"Content-Type":"application/json"}:{})},...options});if(response.status===401){location.assign("/login");throw new Error("Sesión expirada");}const data=await response.json().catch(()=>({}));if(!response.ok)throw new Error(data.error||`HTTP ${response.status}`);return data;};
  let locations=[];

  const render=data=>{
    const noOrgBanner=$("no-org-banner");
    if(data.has_organization===false){
      if(noOrgBanner)noOrgBanner.style.display="flex";
      $("organization-name").textContent="Sin Empresa Registrada";
      $("t-locations").textContent="0";
      $("t-devices").textContent="0";
      $("t-online").textContent="0";
      $("t-analyses").textContent="0";
      return;
    }
    if(noOrgBanner)noOrgBanner.style.display="none";

    locations=data.locations||[];
    const devices=data.devices||[],analyses=data.analyses||[],codes=data.activation_codes||[];
    $("organization-name").textContent=data.organization?.name||"Tu ambiente, en contexto";
    $("t-locations").textContent=locations.length;
    $("t-devices").textContent=devices.length;
    $("t-online").textContent=devices.filter(d=>d.status==="active").length;
    $("t-analyses").textContent=analyses.length;

    const latest=analyses[0];
    const badge=$("overall-status");
    badge.className=`status-badge ${latest?.severity||""}`;
    badge.textContent=latest?({ok:"Óptimo",warning:"Atención",alert:"Requiere acción"}[latest.severity]||latest.verdict):"Esperando datos";
    $("status-title").textContent=latest?(latest.severity==="ok"?"El ambiente está dentro del rango":"Hay una condición que revisar"):"Conecta tu primer ClimaSense Edge";
    $("status-copy").textContent=latest?.recommendation||"Cuando el dispositivo envíe muestras, el análisis aparecerá aquí.";
    $("indoor-temp").textContent=latest?`${Number(latest.indoor_avg_c).toFixed(1)}°`:"—";
    $("outdoor-temp").textContent=latest?.outdoor_c!=null?`${Number(latest.outdoor_c).toFixed(1)}°`:"—";

    $("tenant-analysis").innerHTML=analyses.length?analyses.map(a=>`<div class="analysis-item ${esc(a.severity)}"><span class="severity"></span><div><strong>${Number(a.indoor_avg_c).toFixed(1)}°C · ${esc(a.verdict)}</strong><p>${esc(a.recommendation)}</p></div><time>${date(a.created_at)}</time></div>`).join(""):'<p class="empty-state">Aún no hay análisis.</p>';
    $("locations-list").innerHTML=locations.length?locations.map(l=>`<div class="data-row"><div><strong>${esc(l.name)}</strong><small>${esc(l.location_type)} · ${Number(l.desired_min_c).toFixed(1)}–${Number(l.desired_max_c).toFixed(1)}°C</small></div><span class="tag">${esc(l.timezone)}</span></div>`).join(""):'<p class="empty-state">Sin ubicaciones.</p>';
    
    $("codes-list").innerHTML=codes.length?codes.map(c=>`<div class="data-row"><div><strong>${esc(c.label||"Código de activación")}</strong><small>Creado: ${date(c.created_at)}</small></div><span class="tag ${c.status==="claimed"?"claimed":"available"}">${c.status==="claimed"?"Reclamado por "+esc(c.claimed_by_device):"Disponible"}</span></div>`).join(""):'<p class="empty-state">Sin códigos generados.</p>';

    $("devices-list").innerHTML=devices.length?devices.map(d=>`<article class="device-card"><header><strong>${esc(d.name)}</strong><span class="tag">${esc(d.status)}</span></header><p>${esc(d.device_id)}<br>Último contacto: ${date(d.last_seen_at)}</p><select data-device="${esc(d.device_id)}"><option value="">Asignar ubicación…</option>${locations.map(l=>`<option value="${Number(l.id)}" ${Number(d.location_id)===Number(l.id)?"selected":""}>${esc(l.name)}</option>`).join("")}</select></article>`).join(""):'<p class="empty-state">Sin dispositivos activados.</p>';
    document.querySelectorAll("[data-device]").forEach(select=>select.addEventListener("change",()=>{if(!select.value)return;api(`/api/v1/tenant/devices/${encodeURIComponent(select.dataset.device)}/location`,{method:"POST",body:JSON.stringify({location_id:Number(select.value)})}).then(load).catch(error=>alert(error.message));}));
  };

  const load=()=>api("/api/v1/tenant/summary").then(render).catch(console.error);

  document.querySelectorAll("[data-open]").forEach(button=>button.addEventListener("click",()=>{const el=$(button.dataset.open);if(el)el.showModal();}));
  document.querySelectorAll("[data-close]").forEach(button=>button.addEventListener("click",()=>{const el=$(button.dataset.close);if(el)el.close();}));

  const orgForm=$("org-form");
  if(orgForm){
    orgForm.addEventListener("submit",event=>{
      event.preventDefault();
      const form=event.currentTarget;
      const body=Object.fromEntries(new FormData(form));
      api("/api/v1/tenant/organizations",{method:"POST",body:JSON.stringify(body)}).then(()=>{
        form.reset();
        $("org-dialog").close();
        load();
      }).catch(error=>$("org-message").textContent=error.message);
    });
  }

  const activationForm=$("activation-form");
  if(activationForm){
    activationForm.addEventListener("submit",event=>{
      event.preventDefault();
      const form=event.currentTarget;
      const body=Object.fromEntries(new FormData(form));
      api("/api/v1/tenant/activation-codes",{method:"POST",body:JSON.stringify(body)}).then(data=>{
        form.reset();
        $("activation-dialog").close();
        $("generated-code-display").textContent=data.code;
        $("code-result-dialog").showModal();
        load();
      }).catch(error=>$("activation-message").textContent=error.message);
    });
  }

  const copyBtn=$("copy-code-btn");
  if(copyBtn){
    copyBtn.addEventListener("click",()=>{
      const code=$("generated-code-display").textContent;
      navigator.clipboard.writeText(code).then(()=>{
        copyBtn.textContent="¡Copiado!";
        setTimeout(()=>copyBtn.textContent="Copiar Código",2000);
      });
    });
  }

  // --- Integración de Geolocalización OpenStreetMap Nominatim & GPS GPS ---
  const geoAddressInput=$("geo-address-input");
  const geoResults=$("geo-results");
  const geoLatInput=$("geo-lat-input");
  const geoLonInput=$("geo-lon-input");
  const geoGpsBtn=$("geo-my-location-btn");
  let geoTimer=null;

  if(geoAddressInput&&geoResults){
    geoAddressInput.addEventListener("input",()=>{
      clearTimeout(geoTimer);
      const query=geoAddressInput.value.trim();
      if(query.length<3){
        geoResults.style.display="none";
        geoResults.innerHTML="";
        return;
      }
      geoTimer=setTimeout(()=>{
        fetch(`https://nominatim.openstreetmap.org/search?format=json&q=${encodeURIComponent(query)}&limit=5`)
          .then(res=>res.json())
          .then(items=>{
            if(!items||!items.length){
              geoResults.style.display="none";
              return;
            }
            geoResults.innerHTML=items.map(item=>`<div data-lat="${item.lat}" data-lon="${item.lon}" data-name="${esc(item.display_name)}" style="padding:8px 12px; cursor:pointer; border-bottom:1px solid rgba(255,255,255,0.05); hover:background:#334155;">${esc(item.display_name)}</div>`).join("");
            geoResults.style.display="block";
            geoResults.querySelectorAll("[data-lat]").forEach(div=>{
              div.addEventListener("click",()=>{
                geoAddressInput.value=div.dataset.name;
                geoLatInput.value=Number(div.dataset.lat).toFixed(6);
                geoLonInput.value=Number(div.dataset.lon).toFixed(6);
                geoResults.style.display="none";
              });
            });
          })
          .catch(()=>{
            geoResults.style.display="none";
          });
      },350);
    });

    document.addEventListener("click",e=>{
      if(!geoAddressInput.contains(e.target)&&!geoResults.contains(e.target)){
        geoResults.style.display="none";
      }
    });
  }

  if(geoGpsBtn){
    geoGpsBtn.addEventListener("click",()=>{
      if(!navigator.geolocation){
        alert("Tu navegador no soporta geolocalización GPS.");
        return;
      }
      geoGpsBtn.textContent="Obteniendo…";
      navigator.geolocation.getCurrentPosition(
        pos=>{
          const lat=pos.coords.latitude;
          const lon=pos.coords.longitude;
          geoLatInput.value=lat.toFixed(6);
          geoLonInput.value=lon.toFixed(6);
          geoGpsBtn.textContent="📍 Mi Ubicación";

          fetch(`https://nominatim.openstreetmap.org/reverse?format=json&lat=${lat}&lon=${lon}`)
            .then(res=>res.json())
            .then(data=>{
              if(data&&data.display_name){
                geoAddressInput.value=data.display_name;
              }
            })
            .catch(()=>{});
        },
        err=>{
          geoGpsBtn.textContent="📍 Mi Ubicación";
          alert("No se pudo obtener la ubicación GPS ("+err.message+").");
        },
        {enableHighAccuracy:true,timeout:10000}
      );
    });
  }

  const locationForm=$("location-form");
  if(locationForm){
    locationForm.addEventListener("submit",event=>{
      event.preventDefault();
      const form=event.currentTarget;
      const body=Object.fromEntries(new FormData(form));
      ["latitude","longitude","desired_min_c","desired_max_c"].forEach(k=>body[k]=Number(body[k]));
      api("/api/v1/tenant/locations",{method:"POST",body:JSON.stringify(body)}).then(()=>{
        form.reset();
        $("location-dialog").close();
        load();
      }).catch(error=>$("location-message").textContent=error.message);
    });
  }

  document.querySelector("[data-logout]").addEventListener("click",()=>api("/api/v1/auth/logout",{method:"POST",body:"{}"}).then(()=>location.assign("/login")));
  $("refresh-tenant").addEventListener("click",load);
  load();
  window.ClimaSenseLive.start({
    onRefresh:load,
    fallbackMs:20000,
    onStatus:state=>{
      const node=document.querySelector("[data-live-status]");
      const labels={live:"Panel actualizado en vivo",connecting:"Conectando panel en vivo",reconnecting:"Reconectando panel en vivo",fallback:"Respaldo por consulta periódica",offline:"Navegador sin red"};
      if(node)node.textContent=labels[state]||labels.connecting;
    }
  });
})();
