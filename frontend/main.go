package main

import (
	"fmt"
	"log"
	"net/http"
	"text/template"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		render(w, "test.page.gohtml")
	})

	fmt.Println("[FRONTEND service] listening on port 9090")
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal(err)
	}
}

//TODO hardcoded here, move to a config file maybe
var partialPages = []string{
	"./web/templates/base.layout.gohtml",
	"./web/templates/header.partial.gohtml",
	"./web/templates/footer.partial.gohtml",
}

func render(w http.ResponseWriter, chosenTemplateName string) {

	var templateSlice []string
	selectedTemplate := "./web/templates/" + chosenTemplateName
	templateSlice = append(templateSlice, selectedTemplate)

	//adding partials to allow correct rendering
	for _, partial := range partialPages {
		templateSlice = append(templateSlice, partial)
	}

	tmpl, err := template.ParseFiles(templateSlice...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}
