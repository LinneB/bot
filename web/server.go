package web

import (
	"bot/internal/commands"
	"embed"
	"html/template"
	FS "io/fs"
	"log"
	"net/http"
	"time"
)

// Writer wrapper with exposed statusCode, for logging purposes
type wrappedWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *wrappedWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.statusCode = statusCode
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &wrappedWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}
		next.ServeHTTP(wrapped, r)
		log.Println(wrapped.statusCode, r.RemoteAddr, r.Method, r.URL.Path, time.Since(start))
	})
}

//go:embed public
var fs embed.FS

func New(addr string) (*http.ServeMux, error) {
	tmplCommand, err := template.ParseFS(fs, "public/command.tmpl")
	if err != nil {
		return nil, err
	}
	tmplIndex, err := template.ParseFS(fs, "public/index.tmpl")
	if err != nil {
		return nil, err
	}

	router := http.NewServeMux()
	router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		commands := commands.Handler.GetAllCommands()
		err = tmplIndex.Execute(w, commands)
		if err != nil {
			log.Printf("Could not execute template: %s", err)
		}
	})
	router.HandleFunc("GET /command/{name}", func(w http.ResponseWriter, r *http.Request) {
		command, found := commands.Handler.GetCommandByName(r.PathValue("name"))
		var err error
		if found {
			err = tmplCommand.Execute(w, command)
		} else {
			err = tmplCommand.Execute(w, nil)
		}
		if err != nil {
			log.Printf("Could not execute template: %s", err)
			return
		}
	})

	staticFS, _ := FS.Sub(fs, "public/static")
	router.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(staticFS)))

	return router, nil
}
