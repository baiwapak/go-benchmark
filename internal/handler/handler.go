package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/go-benchmark/internal/capacity"
	"github.com/yourorg/go-benchmark/internal/config"
	"github.com/yourorg/go-benchmark/internal/dto"
	"github.com/yourorg/go-benchmark/internal/glossary"
	"github.com/yourorg/go-benchmark/internal/i18n"
	"github.com/yourorg/go-benchmark/internal/middleware"
	"github.com/yourorg/go-benchmark/internal/preset"
	"github.com/yourorg/go-benchmark/internal/service"
	"github.com/yourorg/go-benchmark/internal/service/pdf"
	"github.com/yourorg/go-benchmark/internal/service/report"
	"github.com/yourorg/go-benchmark/pkg/apperr"
	"github.com/yourorg/go-benchmark/pkg/response"
	"github.com/yourorg/go-benchmark/web"
)

type Handler struct {
	cfg *config.Config
	mgr *service.Manager
	web *web.Renderer
}

func New(cfg *config.Config, mgr *service.Manager, r *web.Renderer) *Handler {
	return &Handler{cfg: cfg, mgr: mgr, web: r}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/lang/:code", h.SetLang)
	r.GET("/", h.Home)
	r.GET("/reports", h.ReportsPage)
	r.POST("/reports/clear", h.ClearReports)
	r.GET("/reports/compare", h.ComparePage)
	r.POST("/reports/:id/delete", h.DeleteReport)
	r.GET("/glossary", h.GlossaryPage)

	create := r.Group("/")
	create.Use(middleware.RateLimit(h.cfg.CreateRateLimitPerMin))
	create.POST("/bench", h.CreateBench)
	create.POST("/api/v1/bench", h.CreateBenchAPI)

	r.GET("/bench/:id", h.BenchPage)
	r.GET("/bench/:id/events", h.BenchEvents)
	r.POST("/bench/:id/stop", h.StopBench)
	r.GET("/bench/:id/report.json", h.ReportJSON)
	r.GET("/bench/:id/pdf", h.ReportPDF)

	r.GET("/api/v1/bench/:id", h.GetBenchAPI)
	r.POST("/api/v1/bench/:id/stop", h.StopBench)
	r.GET("/api/v1/bench/:id/report", h.ReportJSON)
}

func (h *Handler) langFrom(c *gin.Context) string {
	if q := c.Query("lang"); q != "" {
		return i18n.Normalize(q)
	}
	if v, err := c.Cookie("lang"); err == nil && v != "" {
		return i18n.Normalize(v)
	}
	return i18n.Normalize(h.cfg.DefaultLang)
}

func (h *Handler) SetLang(c *gin.Context) {
	lang := i18n.Normalize(c.Param("code"))
	c.SetCookie("lang", lang, 86400*365, "/", "", false, false)
	next := c.Query("next")
	if next == "" {
		next = "/"
	}
	if u, err := url.Parse(next); err != nil || u.IsAbs() || !strings.HasPrefix(u.Path, "/") {
		next = "/"
	}
	sep := "?"
	if strings.Contains(next, "?") {
		sep = "&"
	}
	c.Redirect(http.StatusFound, next+sep+"lang="+lang)
}

func parseGitHubRepo(raw string) (user, repo string) {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimSuffix(raw, "/")
	if raw == "" {
		return "", ""
	}
	const marker = "github.com/"
	idx := strings.Index(raw, marker)
	if idx < 0 {
		return "", ""
	}
	parts := strings.Split(strings.Trim(raw[idx+len(marker):], "/"), "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", ""
	}
	return parts[0], parts[1]
}

func (h *Handler) chrome(lang, nextPath string, extra gin.H) gin.H {
	ghUser, ghRepo := parseGitHubRepo(h.cfg.GitHubRepoURL)
	data := gin.H{
		"Lang":       lang,
		"NextPath":   nextPath,
		"GitHubURL":  h.cfg.GitHubRepoURL,
		"GitHubUser": ghUser,
		"GitHubRepo": ghRepo,
	}
	for k, v := range extra {
		data[k] = v
	}
	return data
}

func (h *Handler) homePageData(lang string, extra gin.H) gin.H {
	data := h.chrome(lang, "/", gin.H{
		"MaxConcurrency":  h.cfg.MaxConcurrency,
		"MaxDurationSec":  h.cfg.MaxDurationSec,
		"MaxTimeoutSec":   h.cfg.MaxTimeoutSec,
		"MaxRampSec":      h.cfg.MaxRampSec,
		"MaxInflightJobs": h.cfg.MaxInflightJobs,
		"HostCPUs":        h.cfg.Host.CPUs,
		"HostMemGB":       capacity.MemGBForDisplay(h.cfg.Host),
		"LimitsAuto":      h.cfg.LimitsAuto,
		"Methods":         []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"},
		"Presets":         preset.All,
	})
	for k, v := range extra {
		data[k] = v
	}
	return data
}

func (h *Handler) Home(c *gin.Context) {
	lang := h.langFrom(c)
	h.web.HTML(c, "home.html", h.homePageData(lang, gin.H{
		"Form": dto.CreateBenchRequest{
			Method:      "GET",
			Concurrency: 10,
			DurationSec: 10,
			TimeoutSec:  10,
		},
	}))
}

func (h *Handler) CreateBench(c *gin.Context) {
	lang := h.langFrom(c)
	var req dto.CreateBenchRequest
	_ = c.ShouldBind(&req)
	if c.PostForm("confirm") == "true" || c.PostForm("confirm") == "on" {
		req.Confirm = true
	}
	req.Lang = lang
	if v := c.PostForm("concurrency"); v != "" {
		req.Concurrency, _ = strconv.Atoi(v)
	}
	if v := c.PostForm("duration_sec"); v != "" {
		req.DurationSec, _ = strconv.Atoi(v)
	}
	if v := c.PostForm("timeout_sec"); v != "" {
		req.TimeoutSec, _ = strconv.Atoi(v)
	}
	if v := c.PostForm("ramp_sec"); v != "" {
		req.RampSec, _ = strconv.Atoi(v)
	}
	req.BodyText = c.PostForm("body_text")

	job, err := h.mgr.Create(req)
	if err != nil {
		h.renderCreateError(c, lang, req, err)
		return
	}
	c.Redirect(http.StatusFound, "/bench/"+job.ID+"?lang="+lang)
}

func (h *Handler) CreateBenchAPI(c *gin.Context) {
	lang := h.langFrom(c)
	var req dto.CreateBenchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	if req.Lang == "" {
		req.Lang = lang
	}
	job, err := h.mgr.Create(req)
	if err != nil {
		h.writeErr(c, err)
		return
	}
	response.OK(c, gin.H{"id": job.ID, "status": job.Status})
}

func (h *Handler) renderCreateError(c *gin.Context, lang string, req dto.CreateBenchRequest, err error) {
	msg := err.Error()
	var ae *apperr.AppError
	if errors.As(err, &ae) {
		msg = ae.Message
	}
	if middleware.ClientWantsJSON(c) {
		h.writeErr(c, err)
		return
	}
	req.Lang = lang
	h.web.HTML(c, "home.html", h.homePageData(lang, gin.H{
		"Error": msg,
		"Form":  req,
	}))
}

func (h *Handler) BenchPage(c *gin.Context) {
	lang := h.langFrom(c)
	job, err := h.mgr.Get(c.Param("id"))
	if err != nil {
		h.notFound(c, lang)
		return
	}
	prog := h.mgr.Progress(job)
	rep := h.mgr.ReportOf(job)
	// Prefer request lang for UI chrome; report body keeps job lang for suggestions already stored.
	status := string(job.Status)
	series := prog.Series
	if rep != nil && len(rep.Series) > 0 {
		series = rep.Series
	}
	h.web.HTML(c, "bench.html", h.chrome(lang, "/bench/"+job.ID, gin.H{
		"JobID":    job.ID,
		"Status":   status,
		"Progress": prog,
		"Report":   localizeReport(rep, lang),
		"Series":   series,
	}))
}

func localizeReport(rep *dto.BenchReport, lang string) *dto.BenchReport {
	if rep == nil {
		return nil
	}
	cp := *rep
	cp.Lang = i18n.Normalize(lang)
	cp.Suggestions = report.BuildSuggestions(cp.Lang, &cp)
	return &cp
}

func (h *Handler) BenchEvents(c *gin.Context) {
	job, err := h.mgr.Get(c.Param("id"))
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Writer.Flush()

	ch, unsub := h.mgr.Subscribe(job)
	defer unsub()

	clientGone := c.Request.Context().Done()
	for {
		select {
		case <-clientGone:
			return
		case snap, ok := <-ch:
			if !ok {
				return
			}
			b, _ := json.Marshal(snap)
			_, _ = fmt.Fprintf(c.Writer, "event: progress\ndata: %s\n\n", b)
			c.Writer.Flush()
			if snap.Status == "done" || snap.Status == "stopped" || snap.Status == "failed" {
				_, _ = fmt.Fprintf(c.Writer, "event: done\ndata: {}\n\n")
				c.Writer.Flush()
				return
			}
		}
	}
}

func (h *Handler) StopBench(c *gin.Context) {
	if err := h.mgr.Stop(c.Param("id")); err != nil {
		h.writeErr(c, err)
		return
	}
	if strings.HasPrefix(c.Request.URL.Path, "/api/") || middleware.ClientWantsJSON(c) {
		response.OKMessage(c, "ok")
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) ReportJSON(c *gin.Context) {
	job, err := h.mgr.Get(c.Param("id"))
	if err != nil {
		h.writeErr(c, err)
		return
	}
	rep := h.mgr.ReportOf(job)
	if rep == nil {
		response.Fail(c, http.StatusConflict, 409, "report not ready")
		return
	}
	lang := h.langFrom(c)
	out := localizeReport(rep, lang)
	c.Header("Content-Disposition", "attachment; filename=report-"+job.ID+".json")
	c.JSON(http.StatusOK, out)
}

func (h *Handler) ReportPDF(c *gin.Context) {
	job, err := h.mgr.Get(c.Param("id"))
	if err != nil {
		h.writeErr(c, err)
		return
	}
	rep := h.mgr.ReportOf(job)
	if rep == nil {
		response.Fail(c, http.StatusConflict, 409, "report not ready")
		return
	}
	lang := h.langFrom(c)
	out := localizeReport(rep, lang)
	b, err := pdf.Render(out)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=report-"+job.ID+".pdf")
	c.Data(http.StatusOK, "application/pdf", b)
}

func (h *Handler) GetBenchAPI(c *gin.Context) {
	job, err := h.mgr.Get(c.Param("id"))
	if err != nil {
		h.writeErr(c, err)
		return
	}
	response.OK(c, gin.H{
		"id":       job.ID,
		"status":   job.Status,
		"progress": h.mgr.Progress(job),
		"report":   localizeReport(h.mgr.ReportOf(job), h.langFrom(c)),
	})
}

func (h *Handler) writeErr(c *gin.Context, err error) {
	var ae *apperr.AppError
	if errors.As(err, &ae) {
		if ae.Code == 422 && ae.Field != "" {
			response.ValidationFail(c, []response.FieldError{{Field: ae.Field, Message: ae.Message}})
			return
		}
		response.Fail(c, ae.HTTPStatus, ae.Code, ae.Message)
		return
	}
	response.Fail(c, http.StatusInternalServerError, 500, err.Error())
}

func (h *Handler) notFound(c *gin.Context, lang string) {
	if middleware.ClientWantsJSON(c) {
		response.NotFound(c, i18n.T(lang, "err.not_found"))
		return
	}
	c.String(http.StatusNotFound, i18n.T(lang, "err.not_found"))
}

func (h *Handler) ReportsPage(c *gin.Context) {
	lang := h.langFrom(c)
	list, err := h.mgr.ListReports(50)
	if err != nil {
		list = []dto.ReportSummary{}
	}
	h.web.HTML(c, "reports.html", h.chrome(lang, "/reports", gin.H{
		"Reports": list,
	}))
}

func (h *Handler) DeleteReport(c *gin.Context) {
	lang := h.langFrom(c)
	err := h.mgr.DeleteReport(c.Param("id"))
	if middleware.ClientWantsJSON(c) {
		if err != nil {
			h.writeErr(c, err)
			return
		}
		response.OKMessage(c, "ok")
		return
	}
	c.Redirect(http.StatusFound, "/reports?lang="+lang)
}

func (h *Handler) ClearReports(c *gin.Context) {
	lang := h.langFrom(c)
	_, err := h.mgr.ClearReports()
	if middleware.ClientWantsJSON(c) {
		if err != nil {
			h.writeErr(c, err)
			return
		}
		response.OKMessage(c, "ok")
		return
	}
	c.Redirect(http.StatusFound, "/reports?lang="+lang)
}

func (h *Handler) GlossaryPage(c *gin.Context) {
	lang := h.langFrom(c)
	h.web.HTML(c, "glossary.html", h.chrome(lang, "/glossary", gin.H{
		"Sections": glossary.Sections,
	}))
}

func (h *Handler) ComparePage(c *gin.Context) {
	lang := h.langFrom(c)
	list, _ := h.mgr.ListReports(50)
	idA := strings.TrimSpace(c.Query("a"))
	idB := strings.TrimSpace(c.Query("b"))

	var repA, repB *dto.BenchReport
	if idA != "" {
		if job, err := h.mgr.Get(idA); err == nil {
			repA = localizeReport(h.mgr.ReportOf(job), lang)
		}
	}
	if idB != "" {
		if job, err := h.mgr.Get(idB); err == nil {
			repB = localizeReport(h.mgr.ReportOf(job), lang)
		}
	}

	h.web.HTML(c, "compare.html", h.chrome(lang, "/reports/compare", gin.H{
		"Reports": list,
		"IDA":     idA,
		"IDB":     idB,
		"ReportA": repA,
		"ReportB": repB,
	}))
}
