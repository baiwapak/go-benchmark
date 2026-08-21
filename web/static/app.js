(function () {
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

  function applyProgress(data) {
    if (!data) return;
    status = data.status;
    if (fill) fill.style.width = (data.progress_pct || 0) + "%";
    if (totalEl) totalEl.textContent = data.total;
    if (rpsEl) rpsEl.textContent = Number(data.rps || 0).toFixed(2);
    if (rateEl) rateEl.textContent = Number(data.success_rate || 0).toFixed(2) + "%";
    if (latEl) latEl.textContent = Number(data.avg_latency_ms || 0).toFixed(2) + " ms";
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
})();
