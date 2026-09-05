package examples

import (
	"fmt"
	"io"
	"log"
	"net/http"

	zenmanage "github.com/zenmanage/zenmanage-go"
	"github.com/zenmanage/zenmanage-go/middleware"
)

// Middleware demonstrates how to integrate Zenmanage feature flags into a
// standard net/http server using the provided middleware package.
//
// The middleware injects a request-scoped FlagManager into every request
// context. If the caller sets the "X-User-ID" request header the flag manager
// is pre-configured with that user identity so rollout and targeting rules
// are evaluated per-request.
func Middleware(token string, addr string) {
	client, err := zenmanage.NewConfigBuilder().
		WithEnvironmentToken(token).
		Build()
	if err != nil {
		log.Fatalf("zenmanage config: %v", err)
	}
	zm := zenmanage.New(client)

	mux := http.NewServeMux()

	// /feature — returns whether "new-feature" is enabled for the caller.
	mux.HandleFunc("/feature", func(w http.ResponseWriter, r *http.Request) {
		enabled, err := middleware.IsEnabled(r.Context(), "new-feature")
		if err != nil {
			http.Error(w, "flag evaluation failed", http.StatusInternalServerError)
			return
		}
		if enabled {
			_, _ = fmt.Fprintln(w, "new-feature is ON")
		} else {
			_, _ = fmt.Fprintln(w, "new-feature is OFF")
		}
	})

	// /color — returns the "button-color" string flag value.
	mux.HandleFunc("/color", func(w http.ResponseWriter, r *http.Request) {
		color, err := middleware.GetString(r.Context(), "button-color", "blue")
		if err != nil {
			http.Error(w, "flag evaluation failed", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = io.WriteString(w, fmt.Sprintf("button color: %s\n", color))
	})

	// /rate — returns the "request-rate-limit" numeric flag value.
	mux.HandleFunc("/rate", func(w http.ResponseWriter, r *http.Request) {
		rate, err := middleware.GetNumber(r.Context(), "request-rate-limit", 100)
		if err != nil {
			http.Error(w, "flag evaluation failed", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = io.WriteString(w, fmt.Sprintf("rate limit: %.0f\n", rate))
	})

	// Wrap the whole mux with the Zenmanage middleware.
	handler := middleware.InjectFlags(zm, mux)

	// Plain HTTP for demo simplicity. Production deployments should either
	// serve via http.ListenAndServeTLS with real certificates, or terminate
	// TLS at a reverse proxy/load balancer in front of this server.
	log.Printf("Listening on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("server: %v", err)
	}
}
