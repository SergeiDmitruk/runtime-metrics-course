package templates

import (
	_ "embed"
	"html/template"
	"sync"
)

//go:embed metrics.html
var metricsHTML string

var once sync.Once
var metricsTemplate *template.Template

func GetMetricsTemplate() *template.Template {
	once.Do(func() {
		metricsTemplate = template.Must(template.New("metrics").Parse(metricsHTML))
	})
	return metricsTemplate
}
