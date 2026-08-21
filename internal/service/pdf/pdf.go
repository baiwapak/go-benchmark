package pdf

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/yourorg/go-benchmark/internal/dto"
	"github.com/yourorg/go-benchmark/internal/i18n"
)

// Render builds a PDF report in the report language.
func Render(rep *dto.BenchReport) ([]byte, error) {
	lang := i18n.Normalize(rep.Lang)
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetTitle(i18n.T(lang, "pdf.title"), true)
	pdf.SetAuthor("Go Benchmark", true)

	fontFamily, ok := registerCJKFont(pdf)
	if !ok {
		fontFamily = "Helvetica"
	}

	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont(fontFamily, "", 8)
		pdf.CellFormat(0, 10, fmt.Sprintf("%s | %d", i18n.T(lang, "pdf.footer"), pdf.PageNo()), "", 0, "C", false, 0, "")
	})

	pdf.AddPage()
	pdf.SetFont(fontFamily, "B", 18)
	pdf.MultiCell(0, 10, i18n.T(lang, "pdf.title"), "", "", false)
	pdf.Ln(4)

	pdf.SetFont(fontFamily, "", 11)
	writeKV(pdf, fontFamily, "ID", rep.ID)
	writeKV(pdf, fontFamily, i18n.T(lang, "report.target"), rep.URL)
	writeKV(pdf, fontFamily, i18n.T(lang, "home.method"), rep.Method)
	writeKV(pdf, fontFamily, "Status", rep.Status)
	writeKV(pdf, fontFamily, i18n.T(lang, "home.concurrency"), strconv.Itoa(rep.Concurrency))
	writeKV(pdf, fontFamily, i18n.T(lang, "home.duration"), strconv.Itoa(rep.DurationSec))
	writeKV(pdf, fontFamily, i18n.T(lang, "home.timeout"), strconv.Itoa(rep.TimeoutSec))
	writeKV(pdf, fontFamily, i18n.T(lang, "report.generated_at"), rep.FinishedAt.Format(time.RFC3339))
	writeKV(pdf, fontFamily, i18n.T(lang, "report.duration"), fmt.Sprintf("%.2fs", rep.ElapsedSec))
	pdf.Ln(4)

	section(pdf, fontFamily, i18n.T(lang, "report.metrics"))
	writeKV(pdf, fontFamily, i18n.T(lang, "report.total"), strconv.FormatInt(rep.Total, 10))
	writeKV(pdf, fontFamily, i18n.T(lang, "report.success"), strconv.FormatInt(rep.Success, 10))
	writeKV(pdf, fontFamily, i18n.T(lang, "report.failed"), strconv.FormatInt(rep.Failed, 10))
	writeKV(pdf, fontFamily, i18n.T(lang, "report.success")+" %", fmt.Sprintf("%.2f", rep.SuccessRate))
	writeKV(pdf, fontFamily, i18n.T(lang, "report.rps"), fmt.Sprintf("%.2f", rep.RPS))
	writeKV(pdf, fontFamily, i18n.T(lang, "report.bytes"), strconv.FormatInt(rep.BytesTotal, 10))
	pdf.Ln(2)

	section(pdf, fontFamily, i18n.T(lang, "report.latency"))
	writeKV(pdf, fontFamily, i18n.T(lang, "report.min"), fmt.Sprintf("%.2f", rep.Latency.MinMs))
	writeKV(pdf, fontFamily, i18n.T(lang, "report.avg"), fmt.Sprintf("%.2f", rep.Latency.AvgMs))
	writeKV(pdf, fontFamily, i18n.T(lang, "report.max"), fmt.Sprintf("%.2f", rep.Latency.MaxMs))
	writeKV(pdf, fontFamily, i18n.T(lang, "report.p50"), fmt.Sprintf("%.2f", rep.Latency.P50Ms))
	writeKV(pdf, fontFamily, i18n.T(lang, "report.p90"), fmt.Sprintf("%.2f", rep.Latency.P90Ms))
	writeKV(pdf, fontFamily, i18n.T(lang, "report.p95"), fmt.Sprintf("%.2f", rep.Latency.P95Ms))
	writeKV(pdf, fontFamily, i18n.T(lang, "report.p99"), fmt.Sprintf("%.2f", rep.Latency.P99Ms))
	pdf.Ln(2)

	section(pdf, fontFamily, i18n.T(lang, "report.status"))
	codes := sortedKeys(rep.StatusHistogram)
	if len(codes) == 0 {
		pdf.Cell(0, 6, "-")
		pdf.Ln(6)
	}
	for _, k := range codes {
		writeKV(pdf, fontFamily, k, strconv.FormatInt(rep.StatusHistogram[k], 10))
	}
	pdf.Ln(2)

	section(pdf, fontFamily, i18n.T(lang, "report.errors"))
	errs := sortedKeys(rep.ErrorHistogram)
	if len(errs) == 0 {
		pdf.MultiCell(0, 6, i18n.T(lang, "report.none_errors"), "", "", false)
	}
	for _, k := range errs {
		writeKV(pdf, fontFamily, k, strconv.FormatInt(rep.ErrorHistogram[k], 10))
	}
	pdf.Ln(4)

	section(pdf, fontFamily, i18n.T(lang, "report.suggestions"))
	for i, s := range rep.Suggestions {
		sev := i18n.T(lang, "report.severity."+s.Severity)
		pdf.SetFont(fontFamily, "B", 11)
		pdf.MultiCell(0, 6, fmt.Sprintf("%d. [%s] %s", i+1, sev, s.Title), "", "", false)
		pdf.SetFont(fontFamily, "", 11)
		pdf.MultiCell(0, 6, s.Detail, "", "", false)
		pdf.Ln(2)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func section(pdf *gofpdf.Fpdf, font, title string) {
	pdf.SetFont(font, "B", 14)
	pdf.MultiCell(0, 8, title, "", "", false)
	pdf.Ln(2)
	pdf.SetFont(font, "", 11)
}

func writeKV(pdf *gofpdf.Fpdf, font, k, v string) {
	pdf.SetFont(font, "B", 11)
	pdf.Cell(55, 6, k)
	pdf.SetFont(font, "", 11)
	pdf.MultiCell(0, 6, v, "", "", false)
}

func sortedKeys(m map[string]int64) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func registerCJKFont(pdf *gofpdf.Fpdf) (string, bool) {
	for _, p := range candidateFonts() {
		if st, err := os.Stat(p); err != nil || st.IsDir() {
			continue
		}
		// gofpdf UTF-8 helper expects TTF (not all TTC work).
		ext := filepath.Ext(p)
		if ext != ".ttf" && ext != ".otf" {
			continue
		}
		pdf.AddUTF8Font("BenchCJK", "", p)
		pdf.AddUTF8Font("BenchCJK", "B", p)
		pdf.AddUTF8Font("BenchCJK", "I", p)
		return "BenchCJK", true
	}
	return "", false
}

func candidateFonts() []string {
	var paths []string
	// Project-local optional font (user may place one here).
	paths = append(paths,
		"fonts/NotoSansSC-Regular.otf",
		"fonts/NotoSansSC-Regular.ttf",
		"fonts/simhei.ttf",
		filepath.Join("internal", "service", "pdf", "fonts", "NotoSansSC-Regular.ttf"),
	)
	if runtime.GOOS == "windows" {
		windir := os.Getenv("WINDIR")
		if windir == "" {
			windir = `C:\Windows`
		}
		fonts := filepath.Join(windir, "Fonts")
		paths = append(paths,
			filepath.Join(fonts, "simhei.ttf"),
			filepath.Join(fonts, "simkai.ttf"),
			filepath.Join(fonts, "simfang.ttf"),
			filepath.Join(fonts, "NotoSansSC-Regular.otf"),
			filepath.Join(fonts, "Noto Sans SC (TrueType).otf"),
		)
	} else {
		paths = append(paths,
			"/usr/share/fonts/truetype/noto/NotoSansCJK-Regular.ttc",
			"/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc",
			"/usr/share/fonts/truetype/wqy/wqy-microhei.ttc",
			"/usr/share/fonts/truetype/arphic/uming.ttc",
			"/System/Library/Fonts/PingFang.ttc",
			"/Library/Fonts/Arial Unicode.ttf",
		)
	}
	return paths
}
