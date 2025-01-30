package templates

import (
	"html/template"
	"sync"
)

const MetricsHTML = `
<!DOCTYPE html>
<html>
<head>
    <title>Metrics</title>
</head>
<body>
    <h1>Metrics</h1>
        <h2>Gauges</h2>
        <ul>
            {{range $name, $value := .Gauges}}
                <li>{{$name}}: {{$value}}</li>
            {{end}}
        </ul>
        <h2>Counters</h2>
        <ul>
            {{range $name, $value := .Counters}}
                <li>{{$name}}: {{$value}}</li>
            {{end}}
        </ul>
</body>
</html>
`

var once sync.Once
var metricsTemplate *template.Template

func GetMetricsTemplate() *template.Template {
	once.Do(func() {
		metricsTemplate = template.Must(template.New("metrics").Parse(MetricsHTML))
	})
	return metricsTemplate
}
