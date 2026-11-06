package uri_template_expander

import (
	"errors"
	"testing"
)

func TestExpandSimple(t *testing.T) {
	got, err := Expand("{var}", map[string]interface{}{"var": "value"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "value" {
		t.Fatalf("got %q, want %q", got, "value")
	}
}

func TestExpandSimpleEncodesReserved(t *testing.T) {
	got, err := Expand("{var}", map[string]interface{}{"var": "a/b?c"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "a%2Fb%3Fc"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestExpandSimpleEncodesSpace(t *testing.T) {
	got, err := Expand("{var}", map[string]interface{}{"var": "hello world"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "hello%20world"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestExpandReserved(t *testing.T) {
	got, err := Expand("{+path}", map[string]interface{}{"path": "/foo/bar"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "/foo/bar" {
		t.Fatalf("got %q, want %q", got, "/foo/bar")
	}
}

func TestExpandReservedStillEncodesSpace(t *testing.T) {
	got, err := Expand("{+path}", map[string]interface{}{"path": "/foo bar"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "/foo%20bar"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestExpandFragment(t *testing.T) {
	got, err := Expand("{#frag}", map[string]interface{}{"frag": "section"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "#section" {
		t.Fatalf("got %q, want %q", got, "#section")
	}
}

func TestExpandFragmentEmptyStillEmitsHash(t *testing.T) {
	got, err := Expand("{#frag}", map[string]interface{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "#" {
		t.Fatalf("got %q, want %q", got, "#")
	}
}

func TestExpandMissingVariableIsEmpty(t *testing.T) {
	got, err := Expand("a{missing}b", map[string]interface{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "ab" {
		t.Fatalf("got %q, want %q", got, "ab")
	}
}

func TestExpandMultipleVariables(t *testing.T) {
	got, err := Expand("{a},{b}", map[string]interface{}{"a": "1", "b": "2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "1,2" {
		t.Fatalf("got %q, want %q", got, "1,2")
	}
}

func TestExpandList(t *testing.T) {
	got, err := Expand("{list}", map[string]interface{}{"list": []string{"red", "green", "blue"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "red,green,blue" {
		t.Fatalf("got %q, want %q", got, "red,green,blue")
	}
}

func TestExpandListReserved(t *testing.T) {
	got, err := Expand("{+list}", map[string]interface{}{"list": []string{"a/b", "c d"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "a/b,c%20d"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestExpandUnsupportedOperator(t *testing.T) {
	_, err := Expand("{?q}", map[string]interface{}{"q": "x"})
	if err == nil {
		t.Fatal("expected error for unsupported operator")
	}
	if !errors.Is(err, ErrUnsupportedOperator) {
		t.Fatalf("got %v, want ErrUnsupportedOperator", err)
	}
}

func TestExpandMalformedUnclosed(t *testing.T) {
	_, err := Expand("{var", map[string]interface{}{"var": "x"})
	if err == nil {
		t.Fatal("expected error for unclosed brace")
	}
	if !errors.Is(err, ErrMalformedTemplate) {
		t.Fatalf("got %v, want ErrMalformedTemplate", err)
	}
}

func TestExpandMalformedEmpty(t *testing.T) {
	_, err := Expand("{}", map[string]interface{}{})
	if err == nil {
		t.Fatal("expected error for empty expression")
	}
	if !errors.Is(err, ErrMalformedTemplate) {
		t.Fatalf("got %v, want ErrMalformedTemplate", err)
	}
}

func TestExpandStrayClosingBrace(t *testing.T) {
	_, err := Expand("a}b", map[string]interface{}{})
	if err == nil {
		t.Fatal("expected error for stray closing brace")
	}
	if !errors.Is(err, ErrMalformedTemplate) {
		t.Fatalf("got %v, want ErrMalformedTemplate", err)
	}
}

func TestExpandNonStringValue(t *testing.T) {
	got, err := Expand("{n}", map[string]interface{}{"n": 42})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "42" {
		t.Fatalf("got %q, want %q", got, "42")
	}
}

func TestExpandLiteralOnly(t *testing.T) {
	got, err := Expand("https://example.org/path", map[string]interface{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "https://example.org/path" {
		t.Fatalf("got %q, want %q", got, "https://example.org/path")
	}
}

func TestExpandWhitespaceInVariableList(t *testing.T) {
	got, err := Expand("{a, b}", map[string]interface{}{"a": "1", "b": "2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "1,2" {
		t.Fatalf("got %q, want %q", got, "1,2")
	}
}
