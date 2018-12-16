// html creates a html index for the test data
package main

import (
	"encoding/json"
	"html/template"
	"log"
	"os"
	"sort"

	"bramp.net/dataflow-art/palette"
)

type record struct {
	Key      string           `json:"key"`
	Files    []string         `json:"files"`
	Palettes [][]palette.RGBA `json:"palettes"`
}

const tmplHtml = `
<html>
	<head>
		<style>
			#container {}
			.palette {
				vertical-align: middle;
				display: inline-block;
			}
			.palette div {
				width: 30px;
				height: 30px;
				display: inline-block;
			}
		</style>
	</head>
	<body>
		<div id="container">
			{{range .Records -}}
				<div class=row>
					{{range .Palettes}}
						{{if .}}
							<div class=palette>
							{{range . -}}
								<div style="background-color: {{.Hex}}"></div>
							{{- end}}
							</div>
						{{end}}
					{{end}}
				{{.Key}} ({{len .Files}} paintings)
				</div>
			{{- end}}
		</div>
	</body>
</html>`

var (
	tmpl = template.Must(template.New("tmplHtml").Parse(tmplHtml))
)

func process(name string) {
	fd, err := os.Open(name + ".json")
	if err != nil {
		log.Fatal(err)
	}
	defer fd.Close()

	var index []record

	d := json.NewDecoder(fd)
	if err := d.Decode(&index); err != nil {
		log.Fatal(err)
	}

	// TODO Sort by key
	var records []record
	for _, v := range index {
		if len(v.Files) < 10 {
			// Filter out small sets of images
			continue
		}

		records = append(records, v)
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].Key < records[j].Key
	})

	out, err := os.Create(name + ".html")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	err = tmpl.Execute(out, struct {
		Index   string
		Records []record
	}{
		Index:   name,
		Records: records,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	process("artist")
	process("style")
	process("decade")
	process("year")
}
