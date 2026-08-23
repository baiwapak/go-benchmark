package web

import (
	"encoding/json"
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
		"dict": func(values ...any) (map[string]any, error) {
			if len(values)%2 != 0 {
				return nil, fmt.Errorf("dict: odd number of arguments")
			}
			m := make(map[string]any, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, fmt.Errorf("dict: key must be string")
				}
				m[key] = values[i+1]
			}
			return m, nil
		},
		"printf": fmt.Sprintf,
		"fmtFloat": func(v float64, prec int) string {
			return fmt.Sprintf("%.*f", prec, v)
		},
		"toJSON": func(v any) (template.JS, error) {
			b, err := json.Marshal(v)
			if err != nil {
				return "[]", err
			}
			return template.JS(b), nil
		},
		"deltaClass": func(a, b float64, higherIsBetter bool) string {
			if a == b {
				return "delta-neutral"
			}
			improved := b > a
			if !higherIsBetter {
				improved = b < a
			}
			if improved {
				return "delta-good"
			}
			return "delta-bad"
		},
		"deltaText": func(a, b float64) string {
			d := b - a
			if d > 0 {
				return fmt.Sprintf("+%.2f", d)
			}
			return fmt.Sprintf("%.2f", d)
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
