package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/piclang/piclang/internal/lexer"
	"github.com/piclang/piclang/internal/parser"
)

type diag struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Col      int    `json:"col"`
	EndLine  int    `json:"endLine,omitempty"`
	EndCol   int    `json:"endCol,omitempty"`
	Severity string `json:"severity"` // error | warning
	Message  string `json:"message"`
}

// Formato que emite el compilador: "file:line:col: msg"
var errRe = regexp.MustCompile(`^(.+?):(\d+):(\d+):\s*(.+)$`)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "check":
		runCheck(args)
	case "fmt":
		runFmt(args)
	case "-h", "--help", "help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "picf: unknown command %q\n", cmd)
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `usage:
  picf check [--json] <file.pic>
  picf fmt   [--write] <file.pic>`)
}

// ── check ────────────────────────────────────────────────────────────────────

func runCheck(args []string) {
	fs := flag.NewFlagSet("check", flag.ExitOnError)
	jsonOut := fs.Bool("json", false, "emit diagnostics as JSON")
	fs.Parse(args)
	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "picf check: missing <file.pic>")
		os.Exit(2)
	}
	path := fs.Arg(0)

	src, err := os.ReadFile(path)
	if err != nil {
		fatal("read", err)
	}

	diags:=[]diag{}

	// Errores de sintaxis vía parser.
	parseErrs := parseFile(string(src))
	for _, e := range parseErrs {
		diags = append(diags, toDiag(path, e)...)
	}

	// Si el parser pasó, chequear semántica con el loader+sema.
	if len(diags) == 0 {
		diags = append(diags, semaCheck(path)...)
	}

	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(diags)
	} else {
		for _, d := range diags {
			fmt.Printf("%s:%d:%d: %s: %s\n", d.File, d.Line, d.Col, d.Severity, d.Message)
		}
	}
	if len(diags) > 0 {
		os.Exit(1)
	}
}

// parseFile intenta parsear y devuelve los errores como strings
// con formato "file:line:col: msg" (o solo "msg" si no hay posición).
//
// ⚠️ Ajusta esta función si la API real de parser es distinta.
// La suposición es:
//   p := parser.New(lexer.New(src))
//   file := p.ParseFile()
//   errs := p.Errors()
func parseFile(src string) (errs []string) {
	defer func() {
		if r := recover(); r != nil {
			errs = append(errs, fmt.Sprintf("parser panic: %v", r))
		}
	}()

	lex := lexer.New(src)
	p := parser.New(lex)

	// Firma actual del compilador: ParseFile(src string) -> algo (probablemente *ast.File)
	_ = p.ParseFile(src)

	// Si el Parser expone Errors(), los recogemos.
	if pe, ok := interface{}(p).(interface{ Errors() []string }); ok {
		errs = append(errs, pe.Errors()...)
	}
	return errs
}
func semaCheck(path string) []diag {
	// Importa aquí loader y sema si quieres chequeo completo con imports.
	// Lo dejo fuera para que picf no dependa del loader y sea rápido.
	// Si lo quieres, descomenta:
	//
	// prog, err := loader.LoadEntryWith(path, loader.Options{})
	// if err != nil { return []diag{{File: path, Line: 1, Col: 1, Severity: "error", Message: err.Error()}} }
	// chk := sema.NewProgram(prog)
	// if err := chk.Check(); err != nil {
	//     out := []diag{}
	//     for _, e := range chk.Errors() { out = append(out, toDiag(path, e)...) }
	//     return out
	// }
	return nil
}

var (
	// "archivo.pic:5:8: msg"
	errReWithFile = regexp.MustCompile(`^(.+?):(\d+):(\d+):\s*(.+)$`)
	// "5:8: msg" (formato que emite el parser de Piclang)
	errReNoFile = regexp.MustCompile(`^(\d+):(\d+):\s*(.+)$`)
)

func toDiag(file, raw string) []diag {
	// 1) Con archivo: "file.pic:5:8: msg"
	if m := errReWithFile.FindStringSubmatch(raw); m != nil {
		line, _ := strconv.Atoi(m[2])
		col, _ := strconv.Atoi(m[3])
		f := m[1]
		if !filepath.IsAbs(f) {
			f = file
		}
		return []diag{{File: f, Line: line, Col: col, Severity: "error", Message: m[4]}}
	}

	// 2) Sin archivo: "5:8: msg" (parser de Piclang)
	if m := errReNoFile.FindStringSubmatch(raw); m != nil {
		line, _ := strconv.Atoi(m[1])
		col, _ := strconv.Atoi(m[2])
		return []diag{{File: file, Line: line, Col: col, Severity: "error", Message: m[3]}}
	}

	// 3) Sin posición: fallback a (1,1)
	return []diag{{File: file, Line: 1, Col: 1, Severity: "error", Message: raw}}
}
func fatal(stage string, err error) {
	fmt.Fprintf(os.Stderr, "picf: %s: %v\n", stage, err)
	os.Exit(1)
}
