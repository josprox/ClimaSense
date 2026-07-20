(() => {
    "use strict";

    const endpoint = "/api/v1/dashboard/summary";
    const number = new Intl.NumberFormat("es-MX");
    const $ = (id) => document.getElementById(id);
    let measurements = [];

    function text(id, value) {
        const node = $(id);
        if (node) node.textContent = value;
    }

    function dateTime(value) {
        if (!value) return "Sin contacto";
        const date = new Date(value);
        return Number.isNaN(date.getTime()) ? String(value) : date.toLocaleString("es-MX", { dateStyle: "medium", timeStyle: "short" });
    }

    function renderStatus(data) {
        const ready = data.database_ready === true;
        text("system-title", ready ? "ClimaSense está recibiendo datos" : "Servidor activo · base de datos pendiente");
        text("system-detail", data.detail || "Estado actualizado");
        const pill = $("database-pill");
        pill.classList.toggle("offline", !ready);
        pill.querySelector("b").textContent = ready ? "conectada" : "sin conexión";
        $("status-orb").classList.toggle("offline", !ready);
    }

    function renderMetrics(data) {
        const metrics = data.metrics || {};
        text("metric-devices", number.format(metrics.devices || 0));
        text("metric-active", `${number.format(metrics.active_devices || 0)} activos`);
        text("metric-measurements", number.format(metrics.measurements || 0));
        text("metric-heartbeats", number.format(metrics.heartbeats || 0));
        text("metric-events", number.format(metrics.security_events || 0));
        text("device-caption", `${number.format(metrics.devices || 0)} dispositivos registrados`);
    }

    function renderLatest(rows) {
        const latest = rows && rows.length ? rows[0] : null;
        text("latest-temperature", latest ? `${Number(latest.temperature_c).toFixed(1)}°` : "—");
        text("latest-pressure", latest ? `${Number(latest.pressure_hpa).toFixed(1)}` : "—");
        text("latest-device", latest ? latest.device_id : "Sin datos");
        text("latest-sensor", latest ? `${String(latest.sensor_type || "BMP280").toUpperCase()} · ${latest.sensor_address || "—"}` : "BMP280");
        text("latest-time", latest ? dateTime(latest.measured_at) : "—");
    }

    function makeCell(content, className) {
        const cell = document.createElement("td");
        if (className) cell.className = className;
        if (content instanceof Node) cell.appendChild(content); else cell.textContent = content;
        return cell;
    }

    function renderDevices(devices) {
        const body = $("devices-body");
        body.replaceChildren();
        if (!devices || !devices.length) {
            const row = document.createElement("tr");
            const cell = makeCell("Todavía no hay dispositivos aprovisionados", "loading-row");
            cell.colSpan = 4;
            row.appendChild(cell);
            body.appendChild(row);
            return;
        }
        devices.forEach((device) => {
            const row = document.createElement("tr");
            const wrapper = document.createElement("div");
            wrapper.className = "device-cell";
            const avatar = document.createElement("span");
            avatar.className = "device-avatar";
            avatar.textContent = "◇";
            const copy = document.createElement("div");
            const name = document.createElement("strong");
            name.textContent = device.name || "ClimaSense Edge";
            const id = document.createElement("small");
            id.textContent = device.device_id || "sin identificador";
            copy.append(name, id);
            wrapper.append(avatar, copy);
            const status = document.createElement("span");
            status.className = `status-tag ${device.status === "active" ? "" : "inactive"}`;
            status.textContent = device.status || "desconocido";
            row.append(
                makeCell(wrapper),
                makeCell(device.location || "Sin ubicación"),
                makeCell(status),
                makeCell(dateTime(device.last_seen_at))
            );
            body.appendChild(row);
        });
    }

    function drawChart() {
        const canvas = $("telemetry-chart");
        const empty = $("chart-empty");
        const rows = [...measurements].reverse();
        empty.classList.toggle("visible", rows.length < 2);
        if (rows.length < 2) return;

        const rect = canvas.getBoundingClientRect();
        const ratio = window.devicePixelRatio || 1;
        canvas.width = Math.max(1, Math.floor(rect.width * ratio));
        canvas.height = Math.max(1, Math.floor(rect.height * ratio));
        const ctx = canvas.getContext("2d");
        ctx.scale(ratio, ratio);
        const width = rect.width;
        const height = rect.height;
        const pad = { left: 36, right: 36, top: 18, bottom: 28 };
        const plotW = width - pad.left - pad.right;
        const plotH = height - pad.top - pad.bottom;

        ctx.clearRect(0, 0, width, height);
        ctx.strokeStyle = "rgba(145, 180, 202, .10)";
        ctx.lineWidth = 1;
        ctx.font = "9px system-ui";
        ctx.fillStyle = "#557083";
        for (let i = 0; i <= 4; i += 1) {
            const y = pad.top + (plotH / 4) * i;
            ctx.beginPath(); ctx.moveTo(pad.left, y); ctx.lineTo(width - pad.right, y); ctx.stroke();
        }

        const temps = rows.map((row) => Number(row.temperature_c));
        const pressure = rows.map((row) => Number(row.pressure_hpa));
        const bounds = (values, fallback) => {
            const min = Math.min(...values); const max = Math.max(...values);
            return Number.isFinite(min) && Number.isFinite(max) ? [min - Math.max(1, (max - min) * .15), max + Math.max(1, (max - min) * .15)] : fallback;
        };
        const [tempMin, tempMax] = bounds(temps, [0, 40]);
        const [pressMin, pressMax] = bounds(pressure, [950, 1050]);
        const x = (index) => pad.left + (plotW * index) / (rows.length - 1);
        const yTemp = (value) => pad.top + plotH - ((value - tempMin) / (tempMax - tempMin)) * plotH;
        const yPressure = (value) => pad.top + plotH - ((value - pressMin) / (pressMax - pressMin)) * plotH;

        function series(values, y, color, fill) {
            const gradient = ctx.createLinearGradient(0, pad.top, 0, height - pad.bottom);
            gradient.addColorStop(0, fill); gradient.addColorStop(1, "rgba(0,0,0,0)");
            ctx.beginPath(); values.forEach((value, index) => index ? ctx.lineTo(x(index), y(value)) : ctx.moveTo(x(index), y(value)));
            ctx.lineTo(x(values.length - 1), height - pad.bottom); ctx.lineTo(x(0), height - pad.bottom); ctx.closePath(); ctx.fillStyle = gradient; ctx.fill();
            ctx.beginPath(); values.forEach((value, index) => index ? ctx.lineTo(x(index), y(value)) : ctx.moveTo(x(index), y(value)));
            ctx.strokeStyle = color; ctx.lineWidth = 2; ctx.lineJoin = "round"; ctx.stroke();
        }
        series(pressure, yPressure, "#9b7cff", "rgba(155,124,255,.16)");
        series(temps, yTemp, "#32d8e4", "rgba(50,216,228,.18)");

        ctx.fillStyle = "#5d788a";
        ctx.textAlign = "left";
        ctx.fillText(`${tempMax.toFixed(0)}°`, 3, pad.top + 4);
        ctx.fillText(`${tempMin.toFixed(0)}°`, 3, height - pad.bottom);
        ctx.textAlign = "right";
        ctx.fillText(pressMax.toFixed(0), width - 3, pad.top + 4);
        ctx.fillText(pressMin.toFixed(0), width - 3, height - pad.bottom);
    }

    async function load() {
        const button = $("refresh-button");
        button.classList.add("loading");
        try {
            const response = await fetch(endpoint, { headers: { Accept: "application/json" }, cache: "no-store" });
            if (!response.ok) throw new Error(`HTTP ${response.status}`);
            const data = await response.json();
            measurements = Array.isArray(data.recent_measurements) ? data.recent_measurements : [];
            renderStatus(data);
            renderMetrics(data);
            renderLatest(measurements);
            renderDevices(data.recent_devices || []);
            drawChart();
            text("last-update", dateTime(data.generated_at || new Date().toISOString()));
        } catch (error) {
            text("system-title", "No fue posible consultar la API");
            text("system-detail", error.message || "Error de conexión");
            $("status-orb").classList.add("offline");
        } finally {
            button.classList.remove("loading");
        }
    }

    $("refresh-button").addEventListener("click", load);
    window.addEventListener("resize", drawChart);
    load();
    window.setInterval(load, 10000);
})();
