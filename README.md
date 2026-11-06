# uri-template-expander

Expands RFC 6570 URI templates up to level 3: simple (`{var}`), reserved
(`{+var}`), and fragment (`{#var}`) expressions.

```go
package main

import (
	"fmt"

	uri "uri-template-expander"
)

func main() {
	out, err := uri.Expand(
		"https://api.example.org/{+path}{#frag}",
		map[string]interface{}{
			"path": "/users/42",
			"frag": "profile",
		},
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(out)
	// https://api.example.org/users/42#profile
}
```

The exported API is small on purpose:

- `Expand(template string, values map[string]interface{}) (string, error)`
- `ErrUnsupportedOperator`
- `ErrMalformedTemplate`

## Why this exists

Most URI template libraries in Go pull in a dependency or implement the whole
RFC, including path, query, continuation, and label operators. That is a lot of
surface for callers who only need to assemble a URL from a handful of known
parts. This package covers levels 1-3 and nothing else.

The trade-off is explicit: unsupported operators are rejected with
`ErrUnsupportedOperator` instead of being expanded with a best-effort guess. A
caller who writes `{?q}` gets an error and can fix the template, rather than
silently receiving a URI that is missing its query string.

## Values and edge cases

Values are looked up by name in a `map[string]interface{}`. A missing key
expands to the empty string, as RFC 6570 requires. A `[]string` value expands
as a comma-separated list; any other value is formatted with `fmt.Sprint`.

Simple expansion percent-encodes everything outside the unreserved set, so
`{var}` with `"a/b"` yields `a%2Fb`. Reserved and fragment expansion leave
gen-delims and sub-delims intact but still encode characters such as space.
Fragment expansion always emits the leading `#`, even when the variable is
undefined.

The awkward edge is the operator set. If your template uses `{;x}`, `{/x}`,
`{.x}`, `{?x}`, or `{&x}`, `Expand` returns an error. That is intentional: the
package only claims level 3, and pretending otherwise would produce URLs that
look right until a server rejects them.
