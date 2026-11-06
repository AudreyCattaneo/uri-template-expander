// Package uri_template_expander implements RFC 6570 URI Template expansion
// up to level 3: simple string expansion ({var}), reserved expansion ({+var}),
// and fragment expansion ({#var}).
//
// The package deliberately stops at level 3. Path-style parameters ({;var}),
// query ({?var}), continuation ({&var}), and label ({.var}) expansions are not
// implemented; calling Expand with a template that uses them returns an error
// rather than silently producing a wrong URI.
package uri_template_expander

import (
	"errors"
	"fmt"
	"strings"
)

// ErrUnsupportedOperator is returned when a template uses an operator that is
// not part of levels 1-3.
var ErrUnsupportedOperator = errors.New("uri_template_expander: unsupported operator")

// ErrMalformedTemplate is returned for templates that cannot be parsed, such as
// an unclosed brace or an empty expression.
var ErrMalformedTemplate = errors.New("uri_template_expander: malformed template")

// Expand expands a level 1-3 RFC 6570 URI template using the supplied values.
//
// Values are looked up by name. A missing key expands to the empty string, as
// required by RFC 6570 section 3.2.1. Values may be strings or []string; any
// other type is formatted with fmt.Sprint. This keeps the API small while still
// supporting list expansion for simple and reserved expressions.
func Expand(template string, values map[string]interface{}) (string, error) {
	var b strings.Builder
	b.Grow(len(template))

	for i := 0; i < len(template); {
		switch template[i] {
		case '{':
			end := strings.IndexByte(template[i:], '}')
			if end < 0 {
				return "", fmt.Errorf("%w: unclosed brace at offset %d", ErrMalformedTemplate, i)
			}
			end += i
			expr := template[i+1 : end]
			if expr == "" {
				return "", fmt.Errorf("%w: empty expression at offset %d", ErrMalformedTemplate, i)
			}
			out, err := expandExpr(expr, values)
			if err != nil {
				return "", err
			}
			b.WriteString(out)
			i = end + 1
		case '}':
			return "", fmt.Errorf("%w: stray closing brace at offset %d", ErrMalformedTemplate, i)
		default:
			b.WriteByte(template[i])
			i++
		}
	}
	return b.String(), nil
}

// expandExpr handles a single {..} expression. The leading character selects
// the operator; everything else is a comma-separated list of variable names.
func expandExpr(expr string, values map[string]interface{}) (string, error) {
	op := byte(0)
	vars := expr
	switch expr[0] {
	case '+', '#', '.', '/', ';', '?', '&':
		op = expr[0]
		vars = expr[1:]
	}

	// Only levels 1-3 are supported. Reject the operators we do not expand
	// so callers get a clear error instead of a subtly wrong URI.
	switch op {
	case 0, '+', '#':
	default:
		return "", fmt.Errorf("%w: %q", ErrUnsupportedOperator, string(op))
	}

	if vars == "" {
		return "", fmt.Errorf("%w: no variables in expression", ErrMalformedTemplate)
	}

	names := strings.Split(vars, ",")
	parts := make([]string, 0, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			return "", fmt.Errorf("%w: empty variable name", ErrMalformedTemplate)
		}
		// Reject a name that still contains an operator or brace; that would
		// mean the caller wrote something we do not understand.
		if strings.ContainsAny(name, "{}+#./;?&") {
			return "", fmt.Errorf("%w: invalid variable name %q", ErrMalformedTemplate, name)
		}
		parts = append(parts, expandVar(name, values[name], op))
	}

	joined := strings.Join(parts, ",")
	switch op {
	case 0:
		return joined, nil
	case '+':
		return joined, nil
	case '#':
		// Fragment expansion always emits the leading '#', even when every
		// variable is undefined (RFC 6570 section 3.2.5).
		return "#" + joined, nil
	}
	return joined, nil
}

// expandVar expands one name/value pair under the given operator. Simple
// expansion percent-encodes everything that is not unreserved. Reserved and
// fragment expansion keep reserved characters intact.
func expandVar(name string, value interface{}, op byte) string {
	list, isList := value.([]string)
	if !isList {
		list = []string{stringify(value)}
	}

	encoded := make([]string, len(list))
	for i, v := range list {
		if op == '+' || op == '#' {
			encoded[i] = encodeReserved(v)
		} else {
			encoded[i] = encodeSimple(v)
		}
	}

	// For list expansion the members are joined with ','. A single value is
	// just itself; an undefined variable is the empty string.
	return strings.Join(encoded, ",")
}

// stringify converts a value to its string form. A nil value is the empty
// string, matching "undefined variable" semantics.
func stringify(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprint(v)
}

const upperhex = "0123456789ABCDEF"

// encodeSimple percent-encodes a value for simple string expansion. Only
// unreserved characters (ALPHA / DIGIT / "-" / "." / "_" / "~") survive.
func encodeSimple(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if isUnreserved(c) {
			b.WriteByte(c)
		} else {
			b.WriteByte('%')
			b.WriteByte(upperhex[c>>4])
			b.WriteByte(upperhex[c&0x0f])
		}
	}
	return b.String()
}

// encodeReserved percent-encodes a value for reserved or fragment expansion.
// Reserved characters (gen-delims and sub-delims) pass through unchanged, as
// do unreserved characters. Everything else is percent-encoded.
func encodeReserved(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if isUnreserved(c) || isReserved(c) {
			b.WriteByte(c)
		} else {
			b.WriteByte('%')
			b.WriteByte(upperhex[c>>4])
			b.WriteByte(upperhex[c&0x0f])
		}
	}
	return b.String()
}

func isUnreserved(c byte) bool {
	return (c >= 'A' && c <= 'Z') ||
		(c >= 'a' && c <= 'z') ||
		(c >= '0' && c <= '9') ||
		c == '-' || c == '.' || c == '_' || c == '~'
}

func isReserved(c byte) bool {
	switch c {
	case ':', '/', '?', '#', '[', ']', '@', // gen-delims
		'!', '$', '&', '\'', '(', ')', '*', '+', ',', ';', '=': // sub-delims
		return true
	}
	return false
}
