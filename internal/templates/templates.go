package templates

import "html/template"

// HTML шаблон для метрик
const MetricsHTML = `
<!DOCTYPE html>
<html>
<head>
    <title>Metrics</title>
</head>
<body>
    <h1>Metrics</h1>
    {{range $type, $values := .}}
        <h2>{{$type}}</h2>
        <ul>
            {{range $name, $value := $values}}
                <li>{{$name}}: {{$value}}</li>
            {{end}}
        </ul>
    {{end}}
</body>
</html>
`

func GetMetricsTemplate() (*template.Template, error) {
	return template.New("metrics").Parse(MetricsHTML)
}
