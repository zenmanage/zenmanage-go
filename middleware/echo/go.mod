module github.com/zenmanage/zenmanage-go/middleware/echo

go 1.26.0

require (
	github.com/labstack/echo/v4 v4.16.3
	github.com/zenmanage/zenmanage-go v1.1.1
)

require (
	github.com/labstack/gommon v0.5.0 // indirect
	github.com/mattn/go-colorable v0.1.15 // indirect
	github.com/mattn/go-isatty v0.0.22 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	golang.org/x/crypto v0.57.0 // indirect
	golang.org/x/net v0.60.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
	golang.org/x/text v0.42.0 // indirect
)

// Build against the sibling module source until zenmanage-go has a tagged
// release on the Go module proxy. Ignored by anyone importing this module as
// a dependency — replace directives only apply when this go.mod is the main
// module being built (our own CI/dev), never transitively.
replace github.com/zenmanage/zenmanage-go => ../..
