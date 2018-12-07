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
	Key     string         `json:"key"`
	Files   []string       `json:"files"`
	Palette []palette.RGBA `json:"palette"`
}

const tmplHtml = `
<html>
	<head>
		<style>
			#container {}
			img {
				vertical-align: middle;
			}
		</style>
	</head>
	<body>
		<div id="container">
			{{range .Records}}
				<img src="output-final/{{$.Index}}/{{.Key}}.png" width=300> {{.Key}}<br />
			{{- end}}
		</div>
	</body>
</html>`

var (
	tmpl = template.Must(template.New("tmplHtml").Parse(tmplHtml))
)

func process(name string) {
	fd, err := os.Open("output-final/" + name + "/index.json")
	if err != nil {
		log.Fatal(err)
	}
	defer fd.Close()

	var index map[string]record

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
	process("year")
}
