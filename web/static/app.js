(function () {
  initBenchForm();
  initBenchPage();
  initConfirmForms();
})();

function initConfirmForms() {
  document.querySelectorAll("form.js-confirm").forEach(function (form) {
    form.addEventListener("submit", function (ev) {
      const msg = form.getAttribute("data-confirm");
      if (msg && !window.confirm(msg)) {
        ev.preventDefault();
      }
    });
  });
}

function initBenchForm() {
  const form = document.getElementById("bench-form");
  if (!form) return;

  const msgs = {
    urlRequired: form.dataset.msgUrlRequired,
    urlInvalid: form.dataset.msgUrlInvalid,
    confirmRequired: form.dataset.msgConfirmRequired,
    concurrency: form.dataset.msgConcurrency,
    duration: form.dataset.msgDuration,
    timeout: form.dataset.msgTimeout,
    ramp: form.dataset.msgRamp,
    rampDuration: form.dataset.msgRampDuration,
    headers: form.dataset.msgHeaders,
    body: form.dataset.msgBody,
  };
  const maxConcurrency = Number(form.dataset.maxConcurrency);
  const maxDuration = Number(form.dataset.maxDuration);
  const maxTimeout = Number(form.dataset.maxTimeout);
  const maxRamp = Number(form.dataset.maxRamp);
  const maxBodyBytes = 65536;
  const methodSelect = document.getElementById("method-select");
  const bodyField = document.getElementById("body-field");
  const formError = document.getElementById("form-error");
  const fieldErrors = {};
  form.querySelectorAll(".field-error").forEach(function (el) {
    fieldErrors[el.dataset.field] = el;
  });

  function clearErrors() {
    if (formError) {
      formError.hidden = true;
      formError.textContent = "";
    }
    form.querySelectorAll(".invalid").forEach(function (el) {
      el.classList.remove("invalid");
    });
    Object.values(fieldErrors).forEach(function (el) {
      el.textContent = "";
    });
  }

  function setFieldError(name, message) {
    const input = form.elements.namedItem(name);
    const field = fieldErrors[name];
    if (input && input.classList) input.classList.add("invalid");
    if (name === "confirm") {
      const check = form.querySelector(".check");
      if (check) check.classList.add("invalid");
    }
    if (field) field.textContent = message;
    if (formError) {
      formError.textContent = message;
      formError.hidden = false;
    }
  }

  function isValidURL(value) {
    try {
      const u = new URL(value.trim());
      return (u.protocol === "http:" || u.protocol === "https:") && u.host !== "";
    } catch (_) {
      return false;
    }
  }

  function parsePositiveInt(value) {
    const n = Number.parseInt(String(value).trim(), 10);
    return Number.isFinite(n) ? n : NaN;
  }

  function validateHeaders(text) {
    const lines = text.split("\n");
    for (let i = 0; i < lines.length; i++) {
      const line = lines[i].trim();
      if (!line) continue;
      const idx = line.indexOf(":");
      if (idx <= 0) return false;
      if (!line.slice(0, idx).trim()) return false;
    }
    return true;
  }

  function validate() {
    clearErrors();

    const url = form.elements.url.value;
    if (!url.trim()) {
      setFieldError("url", msgs.urlRequired);
      return false;
    }
    if (!isValidURL(url)) {
      setFieldError("url", msgs.urlInvalid);
      return false;
    }

    const concurrency = parsePositiveInt(form.elements.concurrency.value);
    if (!Number.isFinite(concurrency) || concurrency < 1 || concurrency > maxConcurrency) {
      setFieldError("concurrency", msgs.concurrency);
      return false;
    }

    const duration = parsePositiveInt(form.elements.duration_sec.value);
    if (!Number.isFinite(duration) || duration < 1 || duration > maxDuration) {
      setFieldError("duration_sec", msgs.duration);
      return false;
    }

    const timeout = parsePositiveInt(form.elements.timeout_sec.value);
    if (!Number.isFinite(timeout) || timeout < 1 || timeout > maxTimeout) {
      setFieldError("timeout_sec", msgs.timeout);
      return false;
    }

    const rampRaw = String(form.elements.ramp_sec.value).trim();
    const ramp = rampRaw === "" ? 0 : parsePositiveInt(rampRaw);
    if (!Number.isFinite(ramp) || ramp < 0 || ramp > maxRamp) {
      setFieldError("ramp_sec", msgs.ramp);
      return false;
    }
    if (ramp > duration) {
      setFieldError("ramp_sec", msgs.rampDuration);
      return false;
    }

    const headersText = form.elements.headers_text.value;
    if (!validateHeaders(headersText)) {
      setFieldError("headers_text", msgs.headers);
      return false;
    }

    const method = methodSelect ? methodSelect.value : "GET";
    const bodyText = form.elements.body_text ? form.elements.body_text.value : "";
    if (method !== "GET" && method !== "HEAD" && bodyText.length > maxBodyBytes) {
      setFieldError("body_text", msgs.body);
      return false;
    }

    const confirm = form.elements.confirm;
    if (!confirm || !confirm.checked) {
      setFieldError("confirm", msgs.confirmRequired);
      return false;
    }

    return true;
  }

  form.addEventListener("submit", function (ev) {
    if (!validate()) {
      ev.preventDefault();
      const firstInvalid = form.querySelector(".invalid");
      if (firstInvalid && typeof firstInvalid.focus === "function") {
        firstInvalid.focus();
      }
    }
  });

  form.addEventListener("input", function (ev) {
    const target = ev.target;
    if (!target || !target.name) return;
    target.classList.remove("invalid");
    if (fieldErrors[target.name]) fieldErrors[target.name].textContent = "";
    if (target.name === "confirm") {
      const check = form.querySelector(".check");
      if (check) check.classList.remove("invalid");
    }
    if (formError && !form.querySelector(".field-error:not(:empty)")) {
      formError.hidden = true;
      formError.textContent = "";
    }
    if (target.name === "concurrency" || target.name === "duration_sec" || target.name === "timeout_sec" || target.name === "ramp_sec") {
      clearActivePreset();
    }
  });

  function syncBodyField() {
    if (!methodSelect || !bodyField) return;
    const method = methodSelect.value;
    const hide = method === "GET" || method === "HEAD";
    bodyField.hidden = hide;
  }
  if (methodSelect) {
    methodSelect.addEventListener("change", syncBodyField);
    syncBodyField();
  }

  initBenchPresets(form, maxConcurrency, maxDuration, maxTimeout, clearErrors);
}

function initBenchPresets(form, maxConcurrency, maxDuration, maxTimeout, clearErrors) {
  const panel = document.getElementById("preset-panel");
  if (!panel) return;

  const cards = panel.querySelectorAll(".preset-card");
  const rationaleBox = document.getElementById("preset-rationale");
  const rationaleText = document.getElementById("preset-rationale-text");
  let activeCard = null;

  function clamp(value, min, max) {
    return Math.min(Math.max(value, min), max);
  }

  function clearActivePreset() {
    if (activeCard) activeCard.classList.remove("active");
    activeCard = null;
    if (rationaleBox) rationaleBox.hidden = true;
    if (rationaleText) rationaleText.textContent = "";
  }

  cards.forEach(function (card) {
    card.addEventListener("click", function () {
      const concurrency = clamp(Number(card.dataset.concurrency), 1, maxConcurrency);
      const duration = clamp(Number(card.dataset.duration), 1, maxDuration);
      const timeout = clamp(Number(card.dataset.timeout), 1, maxTimeout);

      form.elements.concurrency.value = String(concurrency);
      form.elements.duration_sec.value = String(duration);
      form.elements.timeout_sec.value = String(timeout);

      if (activeCard) activeCard.classList.remove("active");
      card.classList.add("active");
      activeCard = card;

      if (rationaleText) rationaleText.textContent = card.dataset.rationale || "";
      if (rationaleBox) rationaleBox.hidden = !card.dataset.rationale;
      if (clearErrors) clearErrors();
    });
  });
}

function initBenchPage() {
  const root = document.querySelector(".bench-page");
  if (!root) return;

  const jobId = root.getAttribute("data-job-id");
  let status = root.getAttribute("data-status");
  const lang = root.getAttribute("data-lang") || "zh";

  const fill = document.getElementById("progress-fill");
  const totalEl = document.getElementById("stat-total");
  const rpsEl = document.getElementById("stat-rps");
  const rateEl = document.getElementById("stat-rate");
  const latEl = document.getElementById("stat-lat");
  const stopBtn = document.getElementById("btn-stop");
  const chartRps = document.getElementById("chart-rps");
  const chartLatency = document.getElementById("chart-latency");
  const chartSuccess = document.getElementById("chart-success");
  let seriesData = [];

  const seriesEl = document.getElementById("bench-series");
  if (seriesEl) {
    try {
      seriesData = JSON.parse(seriesEl.textContent || "[]");
    } catch (_) {
      seriesData = [];
    }
  }

  function renderCharts(series) {
    if (!series || series.length === 0) return;
    drawLineChart(chartRps, series, "rps", "#1f6b45");
    drawLineChart(chartLatency, series, "avg_latency_ms", "#b54708");
    drawLineChart(chartSuccess, series, "success_rate", "#145233", 100);
  }
  renderCharts(seriesData);

  function applyProgress(data) {
    if (!data) return;
    status = data.status;
    if (fill) fill.style.width = (data.progress_pct || 0) + "%";
    if (totalEl) totalEl.textContent = data.total;
    if (rpsEl) rpsEl.textContent = Number(data.rps || 0).toFixed(2);
    if (rateEl) rateEl.textContent = Number(data.success_rate || 0).toFixed(2) + "%";
    if (latEl) latEl.textContent = Number(data.avg_latency_ms || 0).toFixed(2) + " ms";
    if (data.series && data.series.length) {
      seriesData = data.series;
      renderCharts(seriesData);
    }
    if (data.message) {
      const title = document.getElementById("bench-status-title");
      if (title) title.textContent = data.message;
    }
  }

  if (status === "running" || status === "pending") {
    const es = new EventSource("/bench/" + jobId + "/events");
    es.addEventListener("progress", function (ev) {
      try {
        applyProgress(JSON.parse(ev.data));
      } catch (_) {}
    });
    es.addEventListener("done", function () {
      es.close();
      window.location.reload();
    });
    es.onerror = function () {
      // browser will retry; if job finished, reload once
      if (status === "done" || status === "stopped" || status === "failed") {
        es.close();
      }
    };
  }

  if (stopBtn) {
    stopBtn.addEventListener("click", async function () {
      stopBtn.disabled = true;
      try {
        await fetch("/bench/" + jobId + "/stop?lang=" + encodeURIComponent(lang), { method: "POST" });
      } catch (_) {}
    });
  }
}

function drawLineChart(canvas, series, key, color, fixedMax) {
  if (!canvas) return;
  const ctx = canvas.getContext("2d");
  const w = canvas.width;
  const h = canvas.height;
  const pad = { top: 14, right: 14, bottom: 28, left: 42 };
  const plotW = w - pad.left - pad.right;
  const plotH = h - pad.top - pad.bottom;
  ctx.clearRect(0, 0, w, h);
  if (!series || series.length === 0) return;

  const values = series.map(function (p) { return Number(p[key] || 0); });
  const maxV = fixedMax || Math.max.apply(null, values.concat([1]));
  const lastSec = series[series.length - 1].second || series.length - 1;
  const xMax = Math.max(lastSec, 1);

  ctx.strokeStyle = "#e0ebe4";
  ctx.lineWidth = 1;
  for (let i = 0; i <= 4; i++) {
    const y = pad.top + (plotH * i) / 4;
    ctx.beginPath();
    ctx.moveTo(pad.left, y);
    ctx.lineTo(w - pad.right, y);
    ctx.stroke();
  }

  ctx.beginPath();
  series.forEach(function (pt, idx) {
    const x = pad.left + (pt.second / xMax) * plotW;
    const y = pad.top + plotH - (values[idx] / maxV) * plotH;
    if (idx === 0) ctx.moveTo(x, y);
    else ctx.lineTo(x, y);
  });
  ctx.strokeStyle = color;
  ctx.lineWidth = 2;
  ctx.stroke();

  ctx.fillStyle = color;
  series.forEach(function (pt, idx) {
    const x = pad.left + (pt.second / xMax) * plotW;
    const y = pad.top + plotH - (values[idx] / maxV) * plotH;
    ctx.beginPath();
    ctx.arc(x, y, 2.5, 0, Math.PI * 2);
    ctx.fill();
  });

  ctx.fillStyle = "#5a6b60";
  ctx.font = "11px Segoe UI, PingFang SC, Microsoft YaHei, sans-serif";
  ctx.fillText("0s", pad.left, h - 8);
  ctx.fillText(xMax + "s", w - pad.right - 20, h - 8);
  ctx.fillText(String(Math.round(maxV)), 4, pad.top + 8);
}
