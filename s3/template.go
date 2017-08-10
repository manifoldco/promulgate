package s3

import (
	"reflect"
	"text/template"
)

const index = `
<!doctype html>
<html>
    <head>
        <meta charset="utf-8">
        <title>releases.manifold.co directory listing</title>
        <style>
            body {
                position: relative;
                margin: 0;
                padding: 80px 0 0;
                font-family: sans-serif;
                font-size: 18px;
                background: #f7f7f7;
            }

            body::before {
                content: '';
                position: absolute;
                top: 0;
                left: 0;
                width: 100%;
                height: 2px;
                background-image: linear-gradient(-21deg, #FF0264 0%, #FE744A 49%, #FDBC39 82%, #FDDF31 100%);
            }

            .wrapper {
                padding: 20px 30px;
                margin: 0 auto 20px;
                max-width: 840px;
                background: #FFF;
            }

            table {
                display: table;
                width: 100%;
                line-height: 1.5;
            }

            td {
                border-bottom: 1px dotted #bababa;
                padding: 5px;
            }

            td.size {
                text-align: end;
            }

            td.modified {
                text-align: end;
            }

            .parent.directory a {
                font-size: 15px;
                color: #666;
            }

            .parent.directory a:hover {
                color: #222;
            }

            .parent.directory a::before {
                content: '\2190';
                margin-right: 4px;
            }

            .footer {
                margin: 40px 0 0 0;
                padding: 20px 0;
                font-size: 13px;
                color: #858181;
                margin-top: 80px;
            }

            .copyright {
                float: right;
            }

            h1 {
                text-align: center;
                font-size: 14px;
                text-transform: uppercase;
                margin-bottom: 40px;
                font-weight: 600;
                letter-spacing: 1px;
                color: #888;
                text-shadow: 0 2px 1px #FFF;
            }

            a {
                color: #0072C4;
                text-decoration: none;
            }

            a:visited {
                color: rgba(0, 114, 196, 0.75);
            }

            a:hover {
                color: #0D86DD;
            }

            .directory .name a:before,
            .generic .name a:before {
                width: 17px;
                height: 17px;
                display: inline-block;
                margin: 0 10px 0 0;
                position: relative;
                top: 2px;
                opacity: .4;
            }

            .directory .name a:before {
                content: url(data:image/svg+xml;base64,PD94bWwgdmVyc2lvbj0iMS4wIiBlbmNvZGluZz0idXRmLTgiPz48c3ZnIHZlcnNpb249IjEuMSIgaWQ9IkxheWVyXzEiIHhtbG5zOmNjPSJodHRwOi8vY3JlYXRpdmVjb21tb25zLm9yZy9ucyMiIHhtbG5zOmRjPSJodHRwOi8vcHVybC5vcmcvZGMvZWxlbWVudHMvMS4xLyIgeG1sbnM6aW5rc2NhcGU9Imh0dHA6Ly93d3cuaW5rc2NhcGUub3JnL25hbWVzcGFjZXMvaW5rc2NhcGUiIHhtbG5zOnJkZj0iaHR0cDovL3d3dy53My5vcmcvMTk5OS8wMi8yMi1yZGYtc3ludGF4LW5zIyIgeG1sbnM6c29kaXBvZGk9Imh0dHA6Ly9zb2RpcG9kaS5zb3VyY2Vmb3JnZS5uZXQvRFREL3NvZGlwb2RpLTAuZHRkIiB4bWxuczpzdmc9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHhtbG5zOnhsaW5rPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5L3hsaW5rIiB4PSIwcHgiIHk9IjBweCIgdmlld0JveD0iMCAwIDIwIDIwIiBzdHlsZT0iZW5hYmxlLWJhY2tncm91bmQ6bmV3IDAgMCAyMCAyMDsiIHhtbDpzcGFjZT0icHJlc2VydmUiPjxnIHRyYW5zZm9ybT0idHJhbnNsYXRlKDAsLTk1Mi4zNjIxOCkiPjxwYXRoIGQ9Ik0xLjgsOTU0LjFjLTEsMC0xLjgsMC44LTEuOCwxLjh2MTIuOWMwLDEsMC44LDEuOCwxLjgsMS44aDE2LjRjMSwwLDEuOC0wLjgsMS44LTEuOFY5NTljMC0xLTAuOC0xLjgtMS44LTEuOGgtOWMtMC42LTAuOC0xLjEtMS45LTEuOC0yLjZjLTAuMy0wLjMtMC43LTAuNi0xLjItMC42SDEuOHogTTEuOCw5NTUuNWg0LjRjMC4xLDAsMC4yLDAsMC4zLDAuMmMwLjYsMC44LDEuMiwxLjcsMS44LDIuNmMwLjEsMC4yLDAuMywwLjMsMC42LDAuM2g5LjNjMC4zLDAsMC40LDAuMiwwLjQsMC40djkuOGMwLDAuMy0wLjIsMC40LTAuNCwwLjRIMS44Yy0wLjMsMC0wLjQtMC4yLTAuNC0wLjR2LTEyLjlDMS4zLDk1NS43LDEuNSw5NTUuNSwxLjgsOTU1LjVMMS44LDk1NS41eiIvPjwvZz48L3N2Zz4=);
            }

            .generic .name a:before {
                content: url(data:image/svg+xml;base64,PD94bWwgdmVyc2lvbj0iMS4wIiBlbmNvZGluZz0idXRmLTgiPz48c3ZnIHZlcnNpb249IjEuMSIgaWQ9IkxheWVyXzEiIHhtbG5zOmNjPSJodHRwOi8vY3JlYXRpdmVjb21tb25zLm9yZy9ucyMiIHhtbG5zOmRjPSJodHRwOi8vcHVybC5vcmcvZGMvZWxlbWVudHMvMS4xLyIgeG1sbnM6aW5rc2NhcGU9Imh0dHA6Ly93d3cuaW5rc2NhcGUub3JnL25hbWVzcGFjZXMvaW5rc2NhcGUiIHhtbG5zOnJkZj0iaHR0cDovL3d3dy53My5vcmcvMTk5OS8wMi8yMi1yZGYtc3ludGF4LW5zIyIgeG1sbnM6c29kaXBvZGk9Imh0dHA6Ly9zb2RpcG9kaS5zb3VyY2Vmb3JnZS5uZXQvRFREL3NvZGlwb2RpLTAuZHRkIiB4bWxuczpzdmc9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHhtbG5zOnhsaW5rPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5L3hsaW5rIiB4PSIwcHgiIHk9IjBweCIgdmlld0JveD0iMCAwIDIwIDIwIiBzdHlsZT0iZW5hYmxlLWJhY2tncm91bmQ6bmV3IDAgMCAyMCAyMDsiIHhtbDpzcGFjZT0icHJlc2VydmUiPjxnIHRyYW5zZm9ybT0idHJhbnNsYXRlKDAsLTk1Mi4zNjIxOCkiPjxwYXRoIGQ9Ik0yLjgsOTUzdjE4LjdjMCwwLjMsMC4zLDAuNywwLjcsMC43aDEzYzAuMywwLDAuNy0wLjMsMC43LTAuN3YtMTQuM2MwLTAuMi0wLjEtMC4zLTAuMi0wLjVsLTQuMy00LjNjLTAuMS0wLjEtMC4zLTAuMi0wLjUtMC4ySDMuNUMzLjEsOTUyLjQsMi44LDk1Mi43LDIuOCw5NTNMMi44LDk1M3ogTTQuMSw5NTMuN2g3LjR2My43YzAsMC4zLDAuMywwLjcsMC43LDAuN2gzLjd2MTNINC4xVjk1My43eiBNMTIuOCw5NTQuNmwyLjEsMi4xaC0yLjFMMTIuOCw5NTQuNnoiLz48L2c+PC9zdmc+);
            }

        </style>
    </head>
    <body>
        <div class="wrapper">
            <div class="listing">
                <h4>{{ .Dir }}</h4>
                <table>
                    {{ if ne .Dir "/" }}
                    <tr class="parent directory">
                        <td><a href="../">parent directory</a></td>
                        <td></td>
                        <td></td>
                    </tr>
                    {{ end }}
                    {{ range .Files }}
                    <tr class="{{ .Type }}">
                        <td class="name"><a href="{{ .Name }}">{{ .Name }}{{ if eq .Type "directory" }}/{{ end }}</a></td>
                        <td class="size">{{ filterZero .Size }}</td>
                        <td class="modified">{{ filterZero .Modified }}</td>
                    </tr>
                    {{ end }}
                </table>
            </div>
            <div class="footer">
                <span class="copyright">© 2016 - 2017 Manifold</span>
                <span class="timestamp">Index generated at {{ .Timestamp }}</span>
            </div>
        </div>
    </body>
</html>
`

var tmpl = template.Must(template.New("index.html").Funcs(template.FuncMap{
	"filterZero": func(v interface{}) interface{} {
		if v == reflect.Zero(reflect.TypeOf(v)).Interface() {
			return ""
		}

		return v
	},
}).Parse(index))
