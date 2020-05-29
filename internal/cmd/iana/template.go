package main

import (
	"fmt"
	"log"
	"os"
	"text/template"

	"github.com/Xuanwo/templateutils"
)

func newT(name string) *template.Template {
	return template.Must(
		template.New(name).
			Funcs(templateutils.FuncMap()).
			Parse(string(MustAsset(name + ".tmpl"))))
}

func generateT(tmpl *template.Template, filePath string, data interface{}) {
	errorMsg := fmt.Sprintf("generate template %s to %s", tmpl.Name(), filePath) + ": %v"

	file, err := os.Create(filePath)
	if err != nil {
		log.Fatalf(errorMsg, err)
	}
	err = tmpl.Execute(file, data)
	if err != nil {
		log.Fatalf(errorMsg, err)
	}
}
