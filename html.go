// html creates a html index for the test data
package main

import (
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const tmplHtml = `
<html>
	<head>
		<style>
			#container {
				display: flex;
				flex-wrap: wrap;
				align-items: center;
    			justify-content: center;
			}
			#container div {
				padding: 10px
			}
		</style>
	</head>
	<body>
		<div id="container">
			{{range .Files}}
			<div>
				<img src="{{.}}" width=300><br />
				<img src="{{.}}_pal.png"><br />
			</div>
			{{end}}
		</div>
	</body>
</html>`

func main() {
	tmpl := template.Must(template.New("tmplHtml").Parse(tmplHtml))

	var files []string
	err := filepath.Walk("art", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(info.Name(), ".jpg") {
			files = append(files, path)
		}

		return nil
	})

	out, err := os.Create("index.html")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	err = tmpl.Execute(out, struct {
		Files []string
	}{
		Files: files,
	})
	if err != nil {
		log.Fatal(err)
	}
}
