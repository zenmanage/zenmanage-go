module github.com/zenmanage/zenmanage-go/middleware/echo

go 1.25.12

require (
	github.com/labstack/echo/v4 v4.12.0
	github.com/zenmanage/zenmanage-go v1.0.0
)

require (
	github.com/labstack/gommon v0.4.2 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	golang.org/x/crypto v0.22.0 // indirect
	golang.org/x/net v0.24.0 // indirect
	golang.org/x/sys v0.19.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)

// Build against the sibling module source until zenmanage-go has a tagged
// release on the Go module proxy. Ignored by anyone importing this module as
// a dependency — replace directives only apply when this go.mod is the main
// module being built (our own CI/dev), never transitively.
replace github.com/zenmanage/zenmanage-go => ../..
