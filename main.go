package main

import (
	"bytes"
	"embed"
	"io/fs"
	"log"
	"os"
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

//go:embed themes/_default
var defaultTemplate embed.FS

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
var themeParentPath string
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
	rootCmd.PersistentFlags().StringVar(
		&resumePath, "resume", "resume.json", "path to the resume",
	)
	rootCmd.AddCommand(validateCmd)
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
		&themeName, "theme", "", "override template theme",
	)
	exportCmd.PersistentFlags().StringVar(
		&themeParentPath, "theme-path", "themes", "directory to search for themes",
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

	viper.SetDefault("meta.theme", "_default")
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
	themePath := themeParentPath + "/" + viper.GetString("meta.theme")
	themeFs := func() fs.FS {
		if viper.GetString("meta.theme") != "_default" {
			return os.DirFS(themePath)
		} else {
			return defaultTemplate
		}
	}
	templates, errTemplate := template.New("index.html").Funcs(templatesFn).ParseFS(themeFs(), themePath+"/*.html")
	check(errTemplate)
	buf := bytes.NewBuffer([]byte{})
	errOutput := templates.ExecuteTemplate(buf, "index.html", resume)
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

// https://github.com/kataras/i18n/issues/2
var templatesFn template.FuncMap = template.FuncMap{
	"tr": func(s string) string {
		assets := i18n.Assets(
			func() (filenames []string) {
				//nolint:errcheck
				fs.WalkDir(defaultLocales, ".", func(path string, d fs.DirEntry, err error) error {
					if err == nil && !d.IsDir() {
						filenames = append(filenames, path)
					}
					return err
				})
				return filenames
			},
			defaultLocales.ReadFile,
		)
		locale, errLocale := i18n.New(assets, viper.GetString("meta.lang"))
		check(errLocale)
		return locale.Tr(viper.GetString("meta.lang"), s)
	},
	"jn": func(sep string, s []string) string {
		return strings.Join(s, sep)
	},
	"md": func(s string) string {
		return string(markdown.ToHTML([]byte(s), nil, nil))
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
