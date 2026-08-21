package web

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/go-benchmark/internal/i18n"
)

//go:embed templates/*.html static/*
var FS embed.FS

type Renderer struct {
	tmpl *template.Template
}

func NewRenderer() (*Renderer, error) {
	funcMap := template.FuncMap{
		"T": func(lang, key string, args ...any) string {
			return i18n.T(lang, key, args...)
		},
		"printf": fmt.Sprintf,
		"fmtFloat": func(v float64, prec int) string {
			return fmt.Sprintf("%.*f", prec, v)
		},
	}
	tmpl, err := template.New("").Funcs(funcMap).ParseFS(FS, "templates/*.html")
	if err != nil {
		return nil, err
	}
	return &Renderer{tmpl: tmpl}, nil
}

func (r *Renderer) HTML(c *gin.Context, name string, data any) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Status(http.StatusOK)
	if err := r.tmpl.ExecuteTemplate(c.Writer, name, data); err != nil {
		_ = c.Error(err)
	}
}

func (r *Renderer) Execute(w io.Writer, name string, data any) error {
	return r.tmpl.ExecuteTemplate(w, name, data)
}

func RegisterStatic(engine *gin.Engine) {
	sub, err := fs.Sub(FS, "static")
	if err != nil {
		panic(err)
	}
	engine.StaticFS("/static", http.FS(sub))
}
