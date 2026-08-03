(() => {
  "use strict";
  const $=id=>document.getElementById(id); const esc=value=>String(value??"").replace(/[&<>"']/g,c=>({"&":"&amp;","<":"&lt;",">":"&gt;",'"':"&quot;","'":"&#39;"}[c]));
  const date=value=>value?new Date(String(value).replace(" ","T")).toLocaleString("es-MX",{dateStyle:"medium",timeStyle:"short"}):"—";
  const api=async(url,options={})=>{const response=await fetch(url,{headers:{Accept:"application/json",...(options.body?{"Content-Type":"application/json"}:{})},...options});if(response.status===401){location.assign("/login");throw new Error("Sesión expirada");}const data=await response.json().catch(()=>({}));if(!response.ok)throw new Error(data.error||`HTTP ${response.status}`);return data;};
  let locations=[];

  let mapInstance=null, mapMarker=null;
  const updateMap=(lat,lon)=>{
    lat=Number(lat)||19.4326;
    lon=Number(lon)||-99.1332;
    if(!window.L)return;
    const mapContainer=$("location-map");
    if(!mapContainer)return;

    if(!mapInstance){
      mapInstance=L.map('location-map').setView([lat,lon],14);
      L.tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png',{
        maxZoom:19,
        attribution:'© OpenStreetMap'
      }).addTo(mapInstance);
      mapMarker=L.marker([lat,lon],{draggable:true}).addTo(mapInstance);

      mapMarker.on('dragend',()=>{
        const pos=mapMarker.getLatLng();
        $("geo-lat-input").value=pos.lat.toFixed(6);
        $("geo-lon-input").value=pos.lng.toFixed(6);
      });

      mapInstance.on('click',e=>{
        mapMarker.setLatLng(e.latlng);
        $("geo-lat-input").value=e.latlng.lat.toFixed(6);
        $("geo-lon-input").value=e.latlng.lng.toFixed(6);
      });
    }else{
      mapInstance.setView([lat,lon],14);
      mapMarker.setLatLng([lat,lon]);
    }
    setTimeout(()=>{if(mapInstance)mapInstance.invalidateSize();},250);
  };

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
    const parseTs=value=>{if(!value)return NaN;const s=String(value).trim().replace(" ","T");return new Date(s.endsWith("Z")||s.includes("+")?s:s+"Z").getTime();};
    const generatedAt=parseTs(data.generated_at);
    const isOnline=device=>{
      if(device.status!=="active"||!device.last_seen_at)return false;
      const lastSeen=parseTs(device.last_seen_at);
      return Number.isFinite(generatedAt)&&Number.isFinite(lastSeen)&&Math.abs(generatedAt-lastSeen)<=180000;
    };
    $("organization-name").textContent=data.organization?.name||"Tu ambiente, en contexto";
    $("t-locations").textContent=locations.length;
    $("t-devices").textContent=devices.length;
    $("t-online").textContent=devices.filter(isOnline).length;
    $("t-analyses").textContent=analyses.length;

    const latest=analyses[0];
    const badge=$("overall-status");
    badge.className=`status-badge ${latest?.severity||""}`;
    badge.textContent=latest?({ok:"Óptimo",warning:"Atención",alert:"Requiere acción"}[latest.severity]||latest.verdict):"Esperando datos";
    $("status-title").textContent=latest?(latest.severity==="ok"?"El ambiente está dentro del rango":"Hay una condición que revisar"):"Conecta tu primer ClimaSense Edge";
    $("status-copy").textContent=latest?.recommendation||"Cuando el dispositivo envíe muestras, el análisis aparecerá aquí.";
    $("indoor-temp").textContent=latest?`${Number(latest.indoor_avg_c).toFixed(1)}°`:"—";
    $("outdoor-temp").textContent=latest?.outdoor_c!=null?`${Number(latest.outdoor_c).toFixed(1)}°`:"—";

    window.allAnalyses=analyses;
    const recentAnalyses=analyses.slice(0,3);
    const renderAnalysisItem=(a,idx)=>`
      <div class="analysis-item ${esc(a.severity)}">
        <span class="severity"></span>
        <div style="flex:1;">
          <strong>${Number(a.indoor_avg_c).toFixed(1)}°C · ${esc(a.verdict)}</strong>
          <p style="margin:4px 0 6px 0;">${esc(a.recommendation)}</p>
          ${(a.details||a.analysis_model)?`<button type="button" class="text-button" data-read-more="${idx}" style="font-size:12px; padding:0; text-decoration:underline;">[ Leer más ]</button>`:""}
        </div>
        <time>${date(a.created_at)}</time>
      </div>`;

    $("tenant-analysis").innerHTML=recentAnalyses.length?recentAnalyses.map(renderAnalysisItem).join(""):'<p class="empty-state">Aún no hay análisis.</p>';
    
    const histBtn=$("view-history-btn");
    if(histBtn){
      histBtn.style.display=analyses.length>3?"inline-block":"none";
    }

    document.querySelectorAll("[data-read-more]").forEach(btn=>{
      btn.addEventListener("click",()=>{
        const item=window.allAnalyses[Number(btn.dataset.readMore)];
        if(!item)return;
        $("details-model-tag").textContent=esc(item.analysis_model||"ClimaSense AI");
        $("details-title").textContent=`Análisis (${Number(item.indoor_avg_c).toFixed(1)}°C · ${esc(item.verdict)})`;
        $("details-content").textContent=item.details||item.recommendation||"Sin detalles adicionales.";
        $("details-dialog").showModal();
      });
    });

    $("locations-list").innerHTML=locations.length?locations.map(l=>`
      <div class="data-row">
        <div>
          <strong>${esc(l.name)}</strong>
          <small>${esc(l.location_type)} · ${Number(l.desired_min_c).toFixed(1)}–${Number(l.desired_max_c).toFixed(1)}°C</small>
        </div>
        <div style="display:flex; align-items:center; gap:8px;">
          <span class="tag">${esc(l.timezone)}</span>
          <button type="button" class="ghost-button" data-edit-location="${l.id}" style="padding:3px 10px; font-size:11px;">✏️ Editar</button>
        </div>
      </div>`).join(""):'<p class="empty-state">Sin ubicaciones.</p>';
    
    document.querySelectorAll("[data-edit-location]").forEach(btn=>{
      btn.addEventListener("click",()=>{
        const locId=Number(btn.dataset.editLocation);
        const loc=locations.find(l=>Number(l.id)===locId);
        if(!loc)return;
        $("geo-location-id").value=loc.id;
        $("geo-name-input").value=loc.name;
        $("geo-type-select").value=loc.location_type||"general";
        $("geo-address-input").value=loc.address||"";
        $("geo-tz-input").value=loc.timezone||"America/Mexico_City";
        $("geo-lat-input").value=Number(loc.latitude).toFixed(6);
        $("geo-lon-input").value=Number(loc.longitude).toFixed(6);
        $("geo-min-input").value=loc.desired_min_c;
        $("geo-max-input").value=loc.desired_max_c;
        
        $("location-dialog-eyebrow").textContent="Editar ubicación";
        $("location-dialog-title").textContent=`Editar ${loc.name}`;
        $("location-submit-btn").textContent="Actualizar ubicación";
        
        $("location-dialog").showModal();
        updateMap(loc.latitude, loc.longitude);
      });
    });

    $("codes-list").innerHTML=codes.length?codes.map(c=>`<div class="data-row"><div><strong>${esc(c.label||"Código de activación")}</strong><small>Creado: ${date(c.created_at)}</small></div><span class="tag ${c.status==="claimed"?"claimed":"available"}">${c.status==="claimed"?"Reclamado por "+esc(c.claimed_by_device):"Disponible"}</span></div>`).join(""):'<p class="empty-state">Sin códigos generados.</p>';

    $("devices-list").innerHTML=devices.length?devices.map(d=>`<article class="device-card"><header><strong>${esc(d.name)}</strong><span class="tag">${d.status!=="active"?esc(d.status):(isOnline(d)?"EN LÍNEA":"SIN CONEXIÓN")}</span></header><p>${esc(d.device_id)}<br>Último contacto: ${date(d.last_seen_at)}</p><select data-device="${esc(d.device_id)}"><option value="">Asignar ubicación…</option>${locations.map(l=>`<option value="${Number(l.id)}" ${Number(d.location_id)===Number(l.id)?"selected":""}>${esc(l.name)}</option>`).join("")}</select></article>`).join(""):'<p class="empty-state">Sin dispositivos activados.</p>';
    document.querySelectorAll("[data-device]").forEach(select=>select.addEventListener("change",()=>{if(!select.value)return;api(`/api/v1/tenant/devices/${encodeURIComponent(select.dataset.device)}/location`,{method:"POST",body:JSON.stringify({location_id:Number(select.value)})}).then(load).catch(error=>alert(error.message));}));
  };

  let histPage=1;
  const renderHistory=()=>{
    const analyses=window.allAnalyses||[];
    const pageSize=10;
    const totalPages=Math.max(1,Math.ceil(analyses.length/pageSize));
    if(histPage>totalPages)histPage=totalPages;
    if(histPage<1)histPage=1;
    const start=(histPage-1)*pageSize;
    const pageItems=analyses.slice(start,start+pageSize);
    
    $("history-feed").innerHTML=pageItems.length?pageItems.map((a,i)=>`
      <div class="analysis-item ${esc(a.severity)}">
        <span class="severity"></span>
        <div style="flex:1;">
          <strong>${Number(a.indoor_avg_c).toFixed(1)}°C · ${esc(a.verdict)}</strong>
          <p style="margin:4px 0 6px 0;">${esc(a.recommendation)}</p>
          ${(a.details||a.analysis_model)?`<button type="button" class="text-button" data-hist-read-more="${start+i}" style="font-size:12px; padding:0; text-decoration:underline;">[ Leer más ]</button>`:""}
        </div>
        <time>${date(a.created_at)}</time>
      </div>`).join(""):'<p class="empty-state">Sin historial.</p>';

    $("hist-page-info").textContent=`Página ${histPage} de ${totalPages} (${analyses.length} total)`;
    $("hist-prev").disabled=histPage<=1;
    $("hist-next").disabled=histPage>=totalPages;

    document.querySelectorAll("[data-hist-read-more]").forEach(btn=>{
      btn.addEventListener("click",()=>{
        const item=window.allAnalyses[Number(btn.dataset.histReadMore)];
        if(!item)return;
        $("details-model-tag").textContent=esc(item.analysis_model||"ClimaSense AI");
        $("details-title").textContent=`Análisis (${Number(item.indoor_avg_c).toFixed(1)}°C · ${esc(item.verdict)})`;
        $("details-content").textContent=item.details||item.recommendation||"Sin detalles adicionales.";
        $("details-dialog").showModal();
      });
    });
  };

  const viewHistBtn=$("view-history-btn");
  if(viewHistBtn){
    viewHistBtn.addEventListener("click",()=>{
      histPage=1;
      renderHistory();
      $("history-dialog").showModal();
    });
  }

  const histPrev=$("hist-prev");
  if(histPrev){
    histPrev.addEventListener("click",()=>{
      if(histPage>1){histPage--; renderHistory();}
    });
  }

  const histNext=$("hist-next");
  if(histNext){
    histNext.addEventListener("click",()=>{
      const totalPages=Math.ceil((window.allAnalyses||[]).length/10);
      if(histPage<totalPages){histPage++; renderHistory();}
    });
  }

  const load=()=>api("/api/v1/tenant/summary").then(render).catch(console.error);

  document.querySelectorAll("[data-open]").forEach(button=>button.addEventListener("click",()=>{
    const el=$(button.dataset.open);
    if(el){
      if(button.dataset.open==="location-dialog"){
        $("geo-location-id").value="";
        $("location-form").reset();
        $("location-dialog-eyebrow").textContent="Nueva ubicación";
        $("location-dialog-title").textContent="Define el ambiente esperado";
        $("location-submit-btn").textContent="Guardar ubicación";
        updateMap(19.4326, -99.1332);
      }
      el.showModal();
    }
  }));

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

  // --- OpenStreetMap Nominatim & Leaflet Map Integration ---
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
                const lat=Number(div.dataset.lat);
                const lon=Number(div.dataset.lon);
                geoLatInput.value=lat.toFixed(6);
                geoLonInput.value=lon.toFixed(6);
                geoResults.style.display="none";
                updateMap(lat,lon);
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

  [geoLatInput, geoLonInput].forEach(inp=>{
    if(inp){
      inp.addEventListener("change",()=>{
        const lat=Number(geoLatInput.value);
        const lon=Number(geoLonInput.value);
        if(Number.isFinite(lat)&&Number.isFinite(lon)){
          updateMap(lat,lon);
        }
      });
    }
  });

  if(geoGpsBtn){
    geoGpsBtn.addEventListener("click",()=>{
      const geoNotice=$("geo-notice");
      if(!navigator.geolocation){
        if(geoNotice){
          geoNotice.textContent="📍 Tu navegador no soporta geolocalización GPS. Puedes mover el pin en el mapa para ubicar la posición.";
          geoNotice.style.display="block";
        }
        updateMap(Number(geoLatInput.value)||19.4326, Number(geoLonInput.value)||-99.1332);
        return;
      }
      geoGpsBtn.textContent="Obteniendo…";
      if(geoNotice)geoNotice.style.display="none";

      navigator.geolocation.getCurrentPosition(
        pos=>{
          const lat=pos.coords.latitude;
          const lon=pos.coords.longitude;
          geoLatInput.value=lat.toFixed(6);
          geoLonInput.value=lon.toFixed(6);
          geoGpsBtn.textContent="📍 Mi Ubicación";
          updateMap(lat,lon);

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
          if(geoNotice){
            geoNotice.textContent="📍 Ubicación GPS no concedida por el navegador. Puedes arrastrar la marca roja directamente en el mapa para ubicar la posición exacta.";
            geoNotice.style.display="block";
          }
          updateMap(Number(geoLatInput.value)||19.4326, Number(geoLonInput.value)||-99.1332);
        },
        {enableHighAccuracy:true,timeout:8000}
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
      const locId=body.location_id;
      delete body.location_id;

      const endpoint=locId?`/api/v1/tenant/locations/${encodeURIComponent(locId)}/update`:"/api/v1/tenant/locations";
      api(endpoint,{method:"POST",body:JSON.stringify(body)}).then(()=>{
        form.reset();
        $("geo-location-id").value="";
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
