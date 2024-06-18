package pprofapp

import (
	"fmt"
	"net/http"
	"net/http/pprof"
)

type App struct {
	Host   string
	Port   int
	server *http.Server
}

func New(host string, port int) *App {
	return &App{
		Host: host,
		Port: port,
	}
}

func (a *App) Start() {
	address := fmt.Sprintf("%s:%d", a.Host, a.Port)
	mux := http.NewServeMux()

	// Маршруты для профилирования
	mux.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	mux.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	mux.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	mux.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	mux.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))

	a.server = &http.Server{
		Addr:    address,
		Handler: mux,
	}

	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()
}

func (a *App) Stop() {
	if a.server != nil {
		if err := a.server.Close(); err != nil {
			panic(err)
		}
	}
}
