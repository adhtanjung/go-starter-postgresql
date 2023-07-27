package renderer

import (
	"html/template"
	"io"

	"github.com/labstack/echo/v4"
)

type Renderer struct {
	templates *template.Template
	debug     bool
	location  string
}

func NewRenderer(location string, debug bool) *Renderer {
	tpl := &Renderer{
		location: location,
		debug:    debug,
	}

	tpl.ReloadTemplates()

	return tpl
}

func (t *Renderer) ReloadTemplates() {
	t.templates = template.Must(template.ParseGlob(t.location))
}

func (t *Renderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	if t.debug {
		t.ReloadTemplates()
	}

	return t.templates.ExecuteTemplate(w, name, data)
}
