package handlers

import (
	"log"
	"net/http"

	"github.com/CloudyKit/jet/v6"
)

// get the views from the folder containing the *.jet files
var views = jet.NewSet(
	jet.NewOSFileSystemLoader("./html"),
	jet.InDevelopmentMode(),
)

// renders the Home page
func Home(w http.ResponseWriter, r *http.Request) {
	err := renderPage(w, "home.jet", nil)
	if err != nil {
		log.Println(err)
	}
}

// function to render any page using `ResponseWriter` to write the response, `tmpl` to get the template name inside `views`
// and `data` to show any data during render
func renderPage(w http.ResponseWriter, tmpl string, data jet.VarMap) error {
	view, err := views.GetTemplate(tmpl)
	if err != nil {
		log.Println(err)
		return err
	}

	err = view.Execute(w, data, nil)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}
