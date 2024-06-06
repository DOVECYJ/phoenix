package router

import (
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/DOVECYJ/phoenix/middleware"
	"github.com/go-chi/chi/v5"
)

// A RESTful action interface
type IResource interface {
	Index(http.ResponseWriter, *http.Request)  // index: show a list of object
	Edit(http.ResponseWriter, *http.Request)   // edit: show edit form
	New(http.ResponseWriter, *http.Request)    // new: show create object form
	Show(http.ResponseWriter, *http.Request)   // show: show one object detail by id
	Create(http.ResponseWriter, *http.Request) // create: save a new object
	Update(http.ResponseWriter, *http.Request) // update: save update object
	Delete(http.ResponseWriter, *http.Request) // delete: delete a object by id
}

type Resources struct {
	IResource
}

func (s Resources) Route(r chi.Router) {
	Resource(s)(r)
}

func (s Resources) Only(r chi.Router, actions ...string) {
	ResourceOnly(s, actions...)(r)
}

func (s Resources) Except(r chi.Router, actions ...string) {
	ResourceExcept(s, actions...)(r)
}

// Route a full RESTful actions. It will panic when 'rsc' is nil.
func Resource(rsc IResource) func(chi.Router) {
	if rsc == nil {
		panic("resource can not be nil")
	}
	return func(r chi.Router) {
		r.Get("/", rsc.Index)   // index: show a list of object
		r.Get("/new", rsc.New)  // new: show create object form
		r.Post("/", rsc.Create) // create: save a new object
		r.Route("/{id}", func(r chi.Router) {
			r.Use(middleware.FetchID)
			r.Get("/edit", rsc.Edit)  // edit: show edit form
			r.Get("/", rsc.Show)      // show: show one object detail by id
			r.Patch("/", rsc.Update)  // update: save update object
			r.Put("/", rsc.Update)    // update: save update object
			r.Delete("/", rsc.Delete) // delete: delete a object by id
		})
	}
}

// Route RESTful action only 'actions'. It will panic when 'rsc' is nil.
func ResourceOnly(rsc IResource, actions ...string) func(chi.Router) {
	if rsc == nil {
		slog.Error("resource can not be nil")
		panic("nil resource")
	}
	return func(r chi.Router) {
		for _, action := range actions {
			switch action {
			case "index":
				r.Get("/", rsc.Index)
			case "edit":
				r.With(middleware.FetchID).Get("/{id}/edit", rsc.Edit)
			case "new":
				r.Get("/new", rsc.New)
			case "show":
				r.With(middleware.FetchID).Get("/{id}", rsc.Show)
			case "create":
				r.Post("/", rsc.Create)
			case "update":
				r.With(middleware.FetchID).Patch("/{id}", rsc.Update)
				r.With(middleware.FetchID).Put("/{id}", rsc.Update)
			case "delete":
				r.With(middleware.FetchID).Delete("/{id}", rsc.Delete)
			}
		}
	}
}

// Route RESTful action expect 'actions'. It will panic when 'rsc' is nil.
func ResourceExcept(rsc IResource, actions ...string) func(chi.Router) {
	if rsc == nil {
		slog.Error("resource can not be nil")
		panic("nil resource")
	}
	return func(r chi.Router) {
		if !slices.Contains(actions, "index") {
			r.Get("/", rsc.Index)
		}
		if !slices.Contains(actions, "edit") {
			r.With(middleware.FetchID).Get("/{id}/edit", rsc.Edit)
		}
		if !slices.Contains(actions, "new") {
			r.Get("/new", rsc.New)
		}
		if !slices.Contains(actions, "show") {
			r.With(middleware.FetchID).Get("/{id}", rsc.Show)
		}
		if !slices.Contains(actions, "create") {
			r.Post("/", rsc.Create)
		}
		if !slices.Contains(actions, "update") {
			r.With(middleware.FetchID).Patch("/{id}", rsc.Update)
			r.With(middleware.FetchID).Put("/{id}", rsc.Update)
		}
		if !slices.Contains(actions, "delete") {
			r.With(middleware.FetchID).Delete("/{id}", rsc.Delete)
		}
	}
}

// Print routed path.
func PrintRouters(router chi.Router) {
	chi.Walk(router,
		func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
			route = strings.Replace(route, "/*/", "/", -1)
			slog.Info(fmt.Sprintf("router: %-6s %s", method, route))
			return nil
		},
	)
}

// Serve static file on path in dir. Than means if you visit
// localhost:8080/'path'/a.txt, file 'dir'/a.txt will be served.
// And visit direacory will be forbidden.
func ServeStatic(r chi.Router, path, dir string) {
	if filepath.IsLocal(dir) {
		workDir, _ := os.Getwd()
		dir = filepath.Join(workDir, dir)
	}
	serveStatic(r, path, http.Dir(filepath.Clean(dir)), true)
}

// Serve static file on path in FS. This time visit directory will be fine.
func ServeStaticFs(r chi.Router, path string, fs fs.FS) {
	serveStatic(r, path, http.FS(fs), false)
}

// serveStatic conveniently sets up a http.serveStatic handler to serve
// static files from a http.FileSystem.
func serveStatic(r chi.Router, path string, root http.FileSystem, disableDir bool) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		if disableDir && r.URL.Path[len(r.URL.Path)-1] == '/' {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		http.StripPrefix(
			strings.TrimSuffix(chi.RouteContext(r.Context()).RoutePattern(), "/*"),
			http.FileServer(root),
		).ServeHTTP(w, r)
	})
}
