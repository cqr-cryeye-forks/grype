package template

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"text/template"

	"github.com/mitchellh/go-homedir"

	"github.com/anchore/grype/grype/presenter/models"

	"github.com/anchore/grype/grype/match"
	"github.com/anchore/grype/grype/pkg"
	"github.com/anchore/grype/grype/vulnerability"
)

// Presenter is an implementation of presenter.Presenter that formats output according to a user-provided Go text template.
type Presenter struct {
	matches            match.Matches
	packages           []pkg.Package
	context            pkg.Context
	metadataProvider   vulnerability.MetadataProvider
	appConfig          interface{}
	dbStatus           interface{}
	pathToTemplateFile string
}

// NewPresenter returns a new template.Presenter.
func NewPresenter(matches match.Matches, packages []pkg.Package, context pkg.Context, metadataProvider vulnerability.MetadataProvider, appConfig interface{}, dbStatus interface{}, pathToTemplateFile string) *Presenter {
	return &Presenter{
		matches:            matches,
		packages:           packages,
		metadataProvider:   metadataProvider,
		context:            context,
		appConfig:          appConfig,
		dbStatus:           dbStatus,
		pathToTemplateFile: pathToTemplateFile,
	}
}

// Present creates output using a user-supplied Go template.
func (pres *Presenter) Present(output io.Writer) error {
	expandedPathToTemplateFile, err := homedir.Expand(pres.pathToTemplateFile)
	if err != nil {
		return fmt.Errorf("unable to expand path %q", pres.pathToTemplateFile)
	}

	templateContents, err := ioutil.ReadFile(expandedPathToTemplateFile)
	if err != nil {
		return fmt.Errorf("unable to get output template: %w", err)
	}

	templateName := expandedPathToTemplateFile
	tmpl, err := template.New(templateName).Funcs(funcMap).Parse(string(templateContents))
	if err != nil {
		return fmt.Errorf("unable to parse template: %w", err)
	}

	document, err := models.NewDocument(pres.packages, pres.context, pres.matches, pres.metadataProvider, pres.appConfig, pres.dbStatus)
	if err != nil {
		return err
	}

	// write to file output.json
	// create file
	path := "/tmp/output.json"
	f, err := os.Create(path)
	if err != nil {
		log.Fatalf("An error ocured while creating new file: %s. Error: %v", path, err)
	}

	err = tmpl.Execute(f, document)
	if err != nil {
		return fmt.Errorf("unable to execute supplied template: %w", err)
	}

	err = tmpl.Execute(output, document)
	if err != nil {
		return fmt.Errorf("unable to execute supplied template: %w", err)
	}

	return nil
}

// These are custom functions available to template authors.
var funcMap = template.FuncMap{
	"getLastIndex": func(collection interface{}) int {
		if v := reflect.ValueOf(collection); v.Kind() == reflect.Slice {
			return v.Len() - 1
		}

		return 0
	},
}
