package cmd

import (
	"fmt"
	"go/doc"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

var templateFileCmd = &cobra.Command{
	Use:   "template-file <infile> <outfile>",
	Short: "output a template file using project details",
	Args:  cobra.ExactArgs(2),
	Run:   TemplateFile,
}

var unindent = regexp.MustCompile(`(?ms)^    `)
var snipMain = regexp.MustCompile(`(?ms)^func main\(\) \{$`)
var snipStart = regexp.MustCompile(`(?ms)\s*// snip start$`)
var snipEnd = regexp.MustCompile(`(?ms)^(?:}|\s*// snip end)$`)
var nolint = regexp.MustCompile(`(?ms)//nolint.*?$`)

func TemplateFile(_ *cobra.Command, args []string) {
	infile, outfile := args[0], args[1]

	out, err := os.Create(outfile)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: unable to create %q: %v\n", outfile, err)
		os.Exit(1)
	}

	tmpl := template.New(filepath.Base(infile)).Funcs(template.FuncMap{
		"Example": func(src, name string) (string, error) {
			fset := token.NewFileSet()
			ex, err := findExample(fset, src, name)
			if err != nil {
				return "", err
			}
			b := &strings.Builder{}
			pc := &printer.Config{
				Mode:     printer.UseSpaces,
				Tabwidth: 4,
				Indent:   0,
			}
			err = pc.Fprint(b, fset, ex.Play)
			if err != nil {
				return "", err
			}
			return b.String(), nil
		},
		"ExampleCode": func(src, name string) (string, error) {
			fset := token.NewFileSet()
			ex, err := findExample(fset, src, name)
			if err != nil {
				return "", err
			}
			b := &strings.Builder{}
			pc := &printer.Config{
				Mode:     printer.UseSpaces,
				Tabwidth: 4,
				Indent:   0,
			}
			err = pc.Fprint(b, fset, ex.Play)
			if err != nil {
				return "", err
			}

			s := b.String()
			if ixs := snipMain.FindStringIndex(s); ixs != nil {
				s = s[ixs[1]:]
			}

			if ixs := snipStart.FindStringIndex(s); ixs != nil {
				s = s[ixs[1]:]
			}

			if ixs := snipEnd.FindStringIndex(s); ixs != nil {
				s = s[:ixs[0]]
			}

			s = nolint.ReplaceAllString(s, "")
			s = unindent.ReplaceAllString(s, "")
			s = strings.TrimSpace(s)
			return s, nil
		},
	})

	tmpl, err = tmpl.ParseFiles(infile)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: unable to decode template file %q: %v\n", infile, err)
		os.Exit(1)
	}

	err = tmpl.Execute(out, nil)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: cannot render template file: %v\n", err)
		os.Exit(1)
	}

	err = out.Close()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: cannot close new file %q: %v", outfile, err)
		os.Exit(1)
	}
}

func findExample(fset *token.FileSet, src, name string) (*doc.Example, error) {
	f, err := parser.ParseFile(fset, src, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	exs := doc.Examples(f)
	for _, ex := range exs {
		if ex.Name == name {
			return ex, nil
		}
	}

	return nil, fmt.Errorf("example %q not found in file %q", name, src)
}
