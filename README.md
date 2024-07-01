Build HTML/PDF resume from JSON/YAML/TOML

## Installation

```
# with Go
go install github.com/nikaro/goresume@latest

# with Homebrew
brew install nikaro/tap/goresume

# on ArchLinux from AUR
yay -S goresume-bin
```

You can also download one of the binaries or packages from the
[Releases](https://github.com/nikaro/goresume/releases) page.

## Usage

```
> goresume --help
Resume Builder

Usage:
  goresume [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  export      Export resume to HTML/PDF
  help        Help about any command
  validate    Validate resume against schema

Flags:
  -h, --help               help for goresume
  -l, --log-level string   logging level (default "warn")
  -r, --resume string      path to the resume (default "resume.json")
  -v, --version            version for goresume

Use "goresume [command] --help" for more information about a command.
```

### Examples

Export `resume.json` to `resume.html` and `resume.pdf`:

```
goresume export
```

Use a custom theme for PDF output:

```
goresume export --pdf-theme actual
```

Export HTML to stdout:

```
goresume export --html-output -
```

## Themes

### Embeded

goresume comes with a few embeded themes:

* simple: [HTML](https://nikaro.github.io/goresume/simple.html)
* simple-compact: [HTML](https://nikaro.github.io/goresume/simple-compact.html)
* actual: [HTML](https://nikaro.github.io/goresume/actual.html) •
  [PDF](https://nikaro.github.io/goresume/actual.pdf)

### Custom

You can also use your own themes, by creating a `themes/my-theme.html` file
next to your `resume.json`. goresume use [Go template engine](https://pkg.go.dev/text/template)
augmented with [sprout functions](https://docs.atom.codes/sprout).
