package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func runFmt(args []string) {
	fs := flag.NewFlagSet("fmt", flag.ExitOnError)
	write := fs.Bool("write", false, "write result back to file")
	fs.Parse(args)
	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "picf fmt: missing <file.pic>")
		os.Exit(2)
	}
	path := fs.Arg(0)
	src, err := os.ReadFile(path)
	if err != nil {
		fatal("read", err)
	}
	out := format(string(src))
	if *write {
		if err := os.WriteFile(path, []byte(out), 0644); err != nil {
			fatal("write", err)
		}
		return
	}
	fmt.Print(out)
}

// format es un formateador conservador:
//   - indentación con tabs según llaves { }
//   - trim trailing whitespace
//   - colapsa espacios múltiples fuera de strings y comentarios
//   - asegura un único \n final
func format(src string) string {
	lines := strings.Split(strings.ReplaceAll(src, "\r\n", "\n"), "\n")
	depth := 0
	var out []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			out = append(out, "")
			continue
		}

		// Ajuste de profundidad si la línea empieza cerrando bloque.
		open := strings.Count(trimmed, "{")
		close := strings.Count(trimmed, "}")
		if strings.HasPrefix(trimmed, "}") && depth > 0 {
			depth--
		}

		body := collapseSpaces(trimmed)
		out = append(out, strings.Repeat("\t", depth)+body)

		depth += open - close
		if strings.HasPrefix(trimmed, "}") {
			depth++
		}
		if depth < 0 {
			depth = 0
		}
	}

	// Quita líneas vacías al inicio/fin y asegura newline final.
	for len(out) > 0 && out[0] == "" {
		out = out[1:]
	}
	for len(out) > 0 && out[len(out)-1] == "" {
		out = out[:len(out)-1]
	}
	return strings.Join(out, "\n") + "\n"
}

// collapseSpaces reduce runs de espacios/tabs a uno, excepto dentro de
// strings ("...") y tras // (comentario).
func collapseSpaces(s string) string {
	var b strings.Builder
	inStr := false
	inComment := false
	prevSpace := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if inComment {
			b.WriteByte(c)
			continue
		}
		if inStr {
			b.WriteByte(c)
			if c == '\\' && i+1 < len(s) {
				b.WriteByte(s[i+1])
				i++
			} else if c == '"' {
				inStr = false
			}
			continue
		}
		if c == '/' && i+1 < len(s) && s[i+1] == '/' {
			inComment = true
			b.WriteByte(c)
			continue
		}
		if c == '"' {
			inStr = true
			prevSpace = false
			b.WriteByte(c)
			continue
		}
		if c == ' ' || c == '\t' {
			if !prevSpace {
				b.WriteByte(' ')
				prevSpace = true
			}
			continue
		}
		prevSpace = false
		b.WriteByte(c)
	}
	return b.String()
}
