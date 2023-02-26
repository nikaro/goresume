package main

import (
	"bytes"
	"embed"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/charmbracelet/log"
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

var logLevel string
var resume string
var version string

var rootCmd = &cobra.Command{
	Use:              "goresume",
	Version:          version,
	Short:            "Resume Builder",
	PersistentPreRun: root,
}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate resume against schema",
	Run:   validate,
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export resume to HTML/PDF",
	Run:   export,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(
		&logLevel, "log-level", "l", "warn", "logging level",
	)
	rootCmd.PersistentFlags().StringVar(
		&resume, "resume", "resume.json", "path to the resume",
	)
	rootCmd.AddCommand(validateCmd)
	validateCmd.PersistentFlags().String(
		"schema", defaultSchema, "override schema, can be a path or an url",
	)
	rootCmd.AddCommand(exportCmd)
	exportCmd.PersistentFlags().String(
		"lang", "en", "language for templates",
	)
	exportCmd.PersistentFlags().Bool(
		"html", true, "export to html",
	)
	exportCmd.PersistentFlags().String(
		"html-output", "resume.html", "html output file",
	)
	exportCmd.PersistentFlags().String(
		"html-theme", "simple", "html template theme",
	)
	exportCmd.PersistentFlags().Bool(
		"pdf", true, "export to pdf",
	)
	exportCmd.PersistentFlags().String(
		"pdf-output", "resume.pdf", "pdf output file",
	)
	exportCmd.PersistentFlags().String(
		"pdf-theme", "simple", "pdf template theme",
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
	log.SetLevel(func() (l log.Level) {
		switch logLevel {
		case "fatal":
			l = log.FatalLevel
		case "error":
			l = log.ErrorLevel
		case "warn":
			l = log.WarnLevel
		case "info":
			l = log.InfoLevel
		case "debug":
			l = log.DebugLevel
		default:
			log.Fatal("unknown loglevel", "level", logLevel)
		}
		return l
	}())

	viper.AddConfigPath(".")
	viper.SetConfigName(resume)
	resumePathSplit := strings.Split(resume, ".")
	ext := resumePathSplit[len(resumePathSplit)-1]
	viper.SetConfigType(ext)

	check(viper.BindPFlag("$schema", validateCmd.PersistentFlags().Lookup("schema")))
	check(viper.BindPFlag("meta.lang", exportCmd.PersistentFlags().Lookup("lang")))
	check(viper.BindPFlag("meta.html", exportCmd.PersistentFlags().Lookup("html")))
	check(viper.BindPFlag("meta.html-output", exportCmd.PersistentFlags().Lookup("html-output")))
	check(viper.BindPFlag("meta.html-theme", exportCmd.PersistentFlags().Lookup("html-theme")))
	check(viper.BindPFlag("meta.pdf", exportCmd.PersistentFlags().Lookup("pdf")))
	check(viper.BindPFlag("meta.pdf-output", exportCmd.PersistentFlags().Lookup("pdf-output")))
	check(viper.BindPFlag("meta.pdf-theme", exportCmd.PersistentFlags().Lookup("pdf-theme")))
	check(viper.ReadInConfig())
	log.Debug("dump", "resume", viper.AllSettings())
}

func validate(_ *cobra.Command, _ []string) {
	jsonSchema, errLoadSchema := jsonschema.Compile(viper.GetString("$schema"))
	check(errLoadSchema)
	check(jsonSchema.Validate(viper.AllSettings()))
}

func getTemplate(theme string) *bytes.Buffer {
	themePath := filepath.Join("themes", theme+".html")
	themeTemplate := func() string {
		if _, err := os.Stat(themePath); !errors.Is(err, fs.ErrNotExist) {
			log.Info("export", "from", "local", "theme", themePath)
			template, errTemplate := os.ReadFile(themePath)
			check(errTemplate)
			return string(template)
		} else {
			log.Info("export", "from", "embed", "theme", themePath)
			template, errTemplate := defaultTemplates.ReadFile(themePath)
			check(errTemplate)
			return string(template)
		}
	}()
	templates, errTemplate := template.New("html").
		Funcs(sprig.FuncMap()).
		Funcs(templatesFn).
		Parse(themeTemplate)
	check(errTemplate)
	buf := bytes.NewBuffer([]byte{})
	errOutput := templates.Execute(buf, viper.AllSettings())
	check(errOutput)
	return buf
}

func export(_ *cobra.Command, _ []string) {
	if viper.GetBool("meta.pdf") {
		buf := getTemplate(viper.GetString("meta.pdf-theme"))
		errInstall := playwright.Install(&playwright.RunOptions{Browsers: []string{"chromium"}})
		check(errInstall)
		output, errOutput := setOutputFile(viper.GetString("meta.pdf-output"))
		check(errOutput)
		log.Info("export", "to", output.Name())
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

	if viper.GetBool("meta.html") {
		buf := getTemplate(viper.GetString("meta.html-theme"))
		output, errOutput := setOutputFile(viper.GetString("meta.html-output"))
		check(errOutput)
		log.Info("export", "to", output.Name())
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
