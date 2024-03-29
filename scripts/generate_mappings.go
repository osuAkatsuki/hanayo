//go:build ignore
// +build ignore

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"golang.org/x/exp/slog"
)

func main() {
	var toGoMap bool
	flag.BoolVar(&toGoMap, "g", false, "Set if you want to export data to mappings.go")
	flag.Parse()

	semantic, err := getMappings("https://raw.githubusercontent.com/Semantic-Org/Semantic-UI/master/src/themes/default/elements/icon.overrides", semanticRegex)
	if err != nil {
		panic(err)
	}
	fontawesome, err := getMappings("https://maxcdn.bootstrapcdn.com/font-awesome/latest/css/font-awesome.css", fontAwesomeRegex)
	if err != nil {
		panic(err)
	}
	classMappings := make(map[string]string, len(semantic))
	for k, v := range semantic {
		if equivalent, ok := fontawesome[k]; ok {
			classMappings[equivalent] = v
		}
	}
	b, err := json.MarshalIndent(classMappings, "", "\t")
	if err != nil {
		panic(err)
	}

	if toGoMap {
		f, err := os.Create("modules/fa-semantic-mappings/mappings.go")
		defer f.Close()
		if err != nil {
			slog.Error("Error creating mappings.go", "error", err.Error())
			panic(err)
		}
		f.Write([]byte(fileHeader))
		fmt.Fprintf(f, "var Mappings = %#v\n", classMappings)
		slog.Info("generate: mappings.go")
	} else {
		slog.Info(string(b))
	}
}

const fileHeader = `// THIS FILE WAS AUTOMATICALLY GENERATED BY A TOOL
// Use ` + "`go generate`" + ` to generate this.

package fasuimappings

// Mappings is a map containing the Semantic UI icon equivalent of FontAwesome
// icons.
`

var semanticRegex = regexp.MustCompile(`i\.([\.a-zA-Z0-9-]+):before { content: "(.{5})"; }`)
var fontAwesomeRegex = regexp.MustCompile(`.([a-zA-Z0-9-]+):before {
  content: "(.{5})";
}`)

func getMappings(url string, regex *regexp.Regexp) (map[string]string, error) {
	ov, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	b, err := io.ReadAll(ov.Body)
	if err != nil {
		return nil, err
	}
	strs := regex.FindAllStringSubmatch(string(b), -1)
	m := make(map[string]string, len(strs))
	for _, strs := range strs {
		m[strs[2]] = strings.Replace(strs[1], ".", " ", -1)
	}
	return m, nil
}
