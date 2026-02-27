package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/doc"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type PackageDoc struct {
	Name       string       `json:"name"`
	ImportPath string       `json:"import_path"`
	Doc        string       `json:"doc"`
	Types      []TypeDoc    `json:"types"`
	Funcs      []FuncDoc    `json:"funcs"`
	Consts     []ValueDoc   `json:"consts,omitempty"`
	Vars       []ValueDoc   `json:"vars,omitempty"`
	Examples   []ExampleDoc `json:"examples,omitempty"`
}

type TypeDoc struct {
	Name         string       `json:"name"`
	Doc          string       `json:"doc"`
	Decl         string       `json:"decl"`
	Constructors []FuncDoc    `json:"constructors,omitempty"`
	Methods      []FuncDoc    `json:"methods,omitempty"`
	Consts       []ValueDoc   `json:"consts,omitempty"`
	Vars         []ValueDoc   `json:"vars,omitempty"`
	Examples     []ExampleDoc `json:"examples,omitempty"`
}

type FuncDoc struct {
	Name      string       `json:"name"`
	Doc       string       `json:"doc"`
	Signature string       `json:"signature"`
	Recv      string       `json:"recv,omitempty"`
	Examples  []ExampleDoc `json:"examples,omitempty"`
}

type ExampleDoc struct {
	Name   string `json:"name,omitempty"`
	Doc    string `json:"doc,omitempty"`
	Code   string `json:"code"`
	Output string `json:"output,omitempty"`
}

type ValueDoc struct {
	Doc   string   `json:"doc"`
	Names []string `json:"names"`
	Decl  string   `json:"decl"`
}

func main() {
	importPath := flag.String("import", "github.com/kungfusheep/glyph", "import path for the package")
	flag.Parse()

	dir := flag.Arg(0)
	if dir == "" {
		fmt.Fprintln(os.Stderr, "usage: docgen [-import path] <source-directory>")
		os.Exit(1)
	}

	dir, err := filepath.Abs(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error resolving path: %v\n", err)
		os.Exit(1)
	}

	fset := token.NewFileSet()

	// parse source files (no tests)
	pkgs, err := parser.ParseDir(fset, dir, goFileFilter, parser.ParseComments)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing: %v\n", err)
		os.Exit(1)
	}

	// find the main package (not _test)
	var pkg *ast.Package
	for name, p := range pkgs {
		if !strings.HasSuffix(name, "_test") {
			pkg = p
			break
		}
	}
	if pkg == nil {
		fmt.Fprintln(os.Stderr, "no package found")
		os.Exit(1)
	}

	// flatten source files
	var files []*ast.File
	for _, f := range pkg.Files {
		files = append(files, f)
	}

	// also parse _test.go files for examples
	testPkgs, err := parser.ParseDir(fset, dir, testFileFilter, parser.ParseComments)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing test files: %v\n", err)
		os.Exit(1)
	}
	for _, tp := range testPkgs {
		for _, f := range tp.Files {
			files = append(files, f)
		}
	}

	dpkg, err := doc.NewFromFiles(fset, files, *importPath, doc.PreserveAST)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error building docs: %v\n", err)
		os.Exit(1)
	}

	out := PackageDoc{
		Name:       dpkg.Name,
		ImportPath: dpkg.ImportPath,
		Doc:        dpkg.Doc,
	}

	// types
	for _, t := range dpkg.Types {
		td := TypeDoc{
			Name: t.Name,
			Doc:  t.Doc,
			Decl: renderGenDecl(fset, t.Decl),
		}
		for _, f := range t.Funcs {
			td.Constructors = append(td.Constructors, funcDoc(fset, f))
		}
		for _, m := range t.Methods {
			td.Methods = append(td.Methods, funcDoc(fset, m))
		}
		for _, c := range t.Consts {
			td.Consts = append(td.Consts, valueDoc(fset, c))
		}
		for _, v := range t.Vars {
			td.Vars = append(td.Vars, valueDoc(fset, v))
		}
		for _, e := range t.Examples {
			td.Examples = append(td.Examples, exampleDoc(fset, e))
		}
		out.Types = append(out.Types, td)
	}

	// free functions
	for _, f := range dpkg.Funcs {
		out.Funcs = append(out.Funcs, funcDoc(fset, f))
	}

	// package-level examples
	for _, e := range dpkg.Examples {
		out.Examples = append(out.Examples, exampleDoc(fset, e))
	}

	// package-level constants
	for _, c := range dpkg.Consts {
		out.Consts = append(out.Consts, valueDoc(fset, c))
	}

	// package-level variables
	for _, v := range dpkg.Vars {
		out.Vars = append(out.Vars, valueDoc(fset, v))
	}

	// sort types by name for stable output
	sort.Slice(out.Types, func(i, j int) bool {
		return out.Types[i].Name < out.Types[j].Name
	})

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		fmt.Fprintf(os.Stderr, "error encoding json: %v\n", err)
		os.Exit(1)
	}
}

func funcDoc(fset *token.FileSet, f *doc.Func) FuncDoc {
	fd := FuncDoc{
		Name:      f.Name,
		Doc:       f.Doc,
		Signature: funcSignature(fset, f.Decl),
		Recv:      f.Recv,
	}
	for _, e := range f.Examples {
		fd.Examples = append(fd.Examples, exampleDoc(fset, e))
	}
	return fd
}

func exampleDoc(fset *token.FileSet, e *doc.Example) ExampleDoc {
	var buf bytes.Buffer
	if e.Code != nil {
		if err := format.Node(&buf, fset, e.Code); err == nil {
			// format.Node on a BlockStmt gives "{\n...\n}", strip the braces
			code := buf.String()
			code = strings.TrimPrefix(code, "{\n")
			code = strings.TrimSuffix(code, "\n}")
			// dedent one tab level (go/format indents the body)
			lines := strings.Split(code, "\n")
			for i, line := range lines {
				lines[i] = strings.TrimPrefix(line, "\t")
			}
			code = strings.Join(lines, "\n")
			return ExampleDoc{
				Name:   e.Suffix,
				Doc:    e.Doc,
				Code:   code,
				Output: e.Output,
			}
		}
	}
	return ExampleDoc{Name: e.Suffix, Doc: e.Doc}
}

func funcSignature(fset *token.FileSet, decl *ast.FuncDecl) string {
	if decl == nil {
		return ""
	}
	// nil out body and doc so format.Node only prints the bare signature
	savedBody := decl.Body
	savedDoc := decl.Doc
	decl.Body = nil
	decl.Doc = nil
	defer func() {
		decl.Body = savedBody
		decl.Doc = savedDoc
	}()

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, decl); err != nil {
		return ""
	}
	return buf.String()
}

func renderGenDecl(fset *token.FileSet, decl *ast.GenDecl) string {
	if decl == nil {
		return ""
	}
	// nil out doc so we get the bare declaration
	savedDoc := decl.Doc
	decl.Doc = nil
	defer func() { decl.Doc = savedDoc }()

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, decl); err != nil {
		return ""
	}
	return buf.String()
}

func valueDoc(fset *token.FileSet, v *doc.Value) ValueDoc {
	return ValueDoc{
		Doc:   v.Doc,
		Names: v.Names,
		Decl:  renderGenDecl(fset, v.Decl),
	}
}

// goFileFilter excludes test files
func goFileFilter(info os.FileInfo) bool {
	return !strings.HasSuffix(info.Name(), "_test.go")
}

// testFileFilter includes only test files
func testFileFilter(info os.FileInfo) bool {
	return strings.HasSuffix(info.Name(), "_test.go")
}
