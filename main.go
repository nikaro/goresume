package main

import (
	"bytes"
	"embed"
	"errors"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/gomarkdown/markdown"
	"github.com/kataras/i18n"
	"github.com/playwright-community/playwright-go"
	"github.com/santhosh-tekuri/jsonschema/v5"
	_ "github.com/santhosh-tekuri/jsonschema/v5/httploader"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var defaultSchema string = "https://raw.githubusercontent.com/jsonresume/resume-schema/v1.0.0/schema.json"

//go:embed themes/*.html
var defaultTemplates embed.FS

//go:embed locales/*.yml
var defaultLocales embed.FS

var exportToHtml bool
var exportToPdf bool
var lang string
var outputHtml string
var outputPdf string
var resume Resume
var resumePath string
var schemaPath string
var themeName string
var themeBasePath string
var version string

var rootCmd = &cobra.Command{
	Use:              "goresume",
	Version:          version,
	Short:            "JSON Resume Builder",
	PersistentPreRun: root,
}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate resume against schema",
	Run:   validate,
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export resume",
	Run:   export,
}

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.PersistentFlags().StringVar(
		&resumePath, "resume", "resume.json", "path to the resume",
	)
	validateCmd.PersistentFlags().StringVar(
		&schemaPath, "schema", "", "override schema, can be a path or an url",
	)
	rootCmd.AddCommand(exportCmd)
	exportCmd.PersistentFlags().StringVar(
		&lang, "lang", "", "override language for templates",
	)
	exportCmd.PersistentFlags().BoolVar(
		&exportToHtml, "html", true, "export to html",
	)
	exportCmd.PersistentFlags().StringVar(
		&outputHtml, "html-output", "resume.html", "html file output",
	)
	exportCmd.PersistentFlags().BoolVar(
		&exportToPdf, "pdf", true, "export to pdf",
	)
	exportCmd.PersistentFlags().StringVar(
		&outputPdf, "pdf-output", "resume.pdf", "pdf file output",
	)
	exportCmd.PersistentFlags().StringVar(
		&resumePath, "resume", "resume.json", "path to the resume",
	)
	exportCmd.PersistentFlags().StringVar(
		&themeName, "theme", "", "override template theme",
	)
	exportCmd.PersistentFlags().StringVar(
		&themeBasePath, "theme-path", "themes", "directory to search for themes",
	)
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	err := rootCmd.Execute()
	check(err)
}

func root(_ *cobra.Command, _ []string) {
	viper.AddConfigPath(".")
	viper.SetConfigName(resumePath)
	resumePathSplit := strings.Split(resumePath, ".")
	ext := resumePathSplit[len(resumePathSplit)-1]
	viper.SetConfigType(ext)

	viper.SetDefault("$schema", defaultSchema)
	errBindSchema := viper.BindPFlag("$schema", validateCmd.PersistentFlags().Lookup("schema"))
	check(errBindSchema)

	viper.SetDefault("meta.theme", "simple")
	errBindTheme := viper.BindPFlag("meta.theme", exportCmd.PersistentFlags().Lookup("theme"))
	check(errBindTheme)

	viper.SetDefault("meta.lang", "en")
	errBindLang := viper.BindPFlag("meta.lang", exportCmd.PersistentFlags().Lookup("lang"))
	check(errBindLang)

	errViper := viper.ReadInConfig()
	check(errViper)
	errUnmarshal := viper.Unmarshal(&resume)
	check(errUnmarshal)
}

func validate(_ *cobra.Command, _ []string) {
	jsonSchema, errLoadSchema := jsonschema.Compile(viper.GetString("$schema"))
	check(errLoadSchema)
	errValidate := jsonSchema.Validate(viper.AllSettings())
	check(errValidate)
}

func export(_ *cobra.Command, _ []string) {
	themePath := filepath.Join(themeBasePath, viper.GetString("meta.theme")+".html")
	themeTemplate := func() string {
		if _, err := os.Stat(themePath); !errors.Is(err, fs.ErrNotExist) {
			template, errTemplate := os.ReadFile(themePath)
			check(errTemplate)
			return string(template)
		} else {
			template, errTemplate := defaultTemplates.ReadFile(themePath)
			check(errTemplate)
			return string(template)
		}
	}()
	templates, errTemplate := template.New("html").Funcs(templatesFn).Parse(themeTemplate)
	check(errTemplate)
	buf := bytes.NewBuffer([]byte{})
	errOutput := templates.Execute(buf, resume)
	check(errOutput)

	if exportToPdf {
		errInstall := playwright.Install(&playwright.RunOptions{Browsers: []string{"chromium"}})
		check(errInstall)
		output, errOutput := setOutputFile(outputPdf)
		check(errOutput)
		defer output.Close()
		pw, errRun := playwright.Run()
		check(errRun)
		browser, errLaunch := pw.Chromium.Launch()
		check(errLaunch)
		context, errContext := browser.NewContext()
		check(errContext)
		page, errPage := context.NewPage()
		check(errPage)
		errContent := page.SetContent(buf.String())
		check(errContent)
		pdf, errPdf := page.PDF(playwright.PagePdfOptions{Format: playwright.String("A4")})
		check(errPdf)
		_, errWrite := output.Write(pdf)
		check(errWrite)
		check(browser.Close())
		check(pw.Stop())
	}

	if exportToHtml {
		output, errOutput := setOutputFile(outputHtml)
		check(errOutput)
		defer output.Close()
		_, errWrite := buf.WriteTo(output)
		check(errWrite)
	}
}

var templatesFn template.FuncMap = template.FuncMap{
	"tr": func(s string) string {
		// https://github.com/kataras/i18n/issues/2
		assets := i18n.Assets(
			func() (filenames []string) {
				errWalk := fs.WalkDir(defaultLocales, ".", func(path string, d fs.DirEntry, err error) error {
					if err == nil && !d.IsDir() {
						filenames = append(filenames, path)
					}
					return err
				})
				check(errWalk)
				return filenames
			},
			defaultLocales.ReadFile,
		)
		locale, errLocale := i18n.New(assets, viper.GetString("meta.lang"))
		check(errLocale)
		return locale.Tr(viper.GetString("meta.lang"), s)
	},
	"join": func(sep string, s []string) string {
		return strings.Join(s, sep)
	},
	"md": func(s string) string {
		return string(markdown.ToHTML([]byte(s), nil, nil))
	},
	"trimPrefix": func(s string, prefix string) string {
		return strings.TrimPrefix(s, prefix)
	},
}

func setOutputFile(outputString string) (outputFile *os.File, err error) {
	if outputString == "-" {
		outputFile = os.Stdout
	} else {
		outputFile, err = os.Create(outputString)
	}
	return outputFile, err
}
