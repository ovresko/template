package jet

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/CloudyKit/jet/v6"
	"github.com/CloudyKit/jet/v6/loaders/httpfs"
	"github.com/gofiber/fiber/v2"
	t "github.com/gofiber/template"
	i "github.com/gofiber/template/internal"
	"github.com/gofiber/utils"
)

// engine struct
type engine struct {
	i.Engine
	// templates
	Templates *jet.Set
}

// New returns a Jet render engine for Fiber
func New(directory, extension string) t.Engine {
	// jet library does not export or give us any option to modify the file Extension
	if extension != ".html.jet" && extension != ".jet.html" && extension != ".jet" {
		log.Fatalf("%s Extension is not a valid jet engine ['.html.jet', .jet.html', '.jet']", extension)
	}

	engine := &engine{
		Engine: i.Engine{
			Directory:  directory,
			Extension:  extension,
			LayoutName: "embed",
			Funcmap:    make(map[string]interface{}),
		},
	}

	return engine
}

// NewFileSystem returns a Jet render engine for Fiber with file system
func NewFileSystem(fs http.FileSystem, extension string) t.Engine {
	// jet library does not export or give us any option to modify the file Extension
	if extension != ".html.jet" && extension != ".jet.html" && extension != ".jet" {
		log.Fatalf("%s Extension is not a valid jet engine ['.html.jet', .jet.html', '.jet']", extension)
	}

	engine := &engine{
		Engine: i.Engine{
			Directory:  "/",
			FileSystem: fs,
			Extension:  extension,
			LayoutName: "embed",
			Funcmap:    make(map[string]interface{}),
		},
	}

	return engine
}

// Load parses the templates to the engine.
func (e *engine) Load() (err error) {
	// race safe
	e.Mutex.Lock()
	defer e.Mutex.Unlock()

	// parse templates
	// e.Templates = jet.NewHTMLSet(e.Directory)

	var loader jet.Loader

	if e.FileSystem != nil {
		loader, err = httpfs.NewLoader(e.FileSystem)

		if err != nil {
			return
		}
	} else {
		loader = jet.NewInMemLoader()
	}
	if e.Verbose {
		e.Templates = jet.NewSet(
			loader,
			jet.WithDelims(e.Left, e.Right),
			jet.InDevelopmentMode(),
		)
	} else {
		e.Templates = jet.NewSet(
			loader,
			jet.WithDelims(e.Left, e.Right),
		)
	}

	for name, fn := range e.Funcmap {
		e.Templates.AddGlobal(name, fn)
	}

	walkFn := func(path string, info os.FileInfo, err error) error {
		l := loader.(*jet.InMemLoader)
		// Return error if exist
		if err != nil {
			return err
		}
		// Skip file if it's a Directory or has no file info
		if info == nil || info.IsDir() {
			return nil
		}
		// Skip file if it does not equal the given template Extension
		if len(e.Extension) >= len(path) || path[len(path)-len(e.Extension):] != e.Extension {
			return nil
		}
		// ./views/html/index.tmpl -> index.tmpl
		rel, err := filepath.Rel(e.Directory, path)
		if err != nil {
			return err
		}
		name := strings.TrimSuffix(rel, e.Extension)
		// Read the file
		// #gosec G304
		buf, err := utils.ReadFile(path, e.FileSystem)
		if err != nil {
			return err
		}

		l.Set(name, string(buf))
		// Debugging
		if e.Verbose {
			fmt.Printf("views: parsed template: %s\n", name)
		}

		return err
	}

	e.Loaded = true

	if _, ok := loader.(*jet.InMemLoader); ok {
		return filepath.Walk(e.Directory, walkFn)
	}

	return
}

// Render will render the template by name
func (e *engine) Render(out io.Writer, template string, binding interface{}, layout ...string) error {
	if !e.Loaded || e.ShouldReload {
		if e.ShouldReload {
			e.Loaded = false
		}
		if err := e.Load(); err != nil {
			return err
		}
	}
	tmpl, err := e.Templates.GetTemplate(template)
	if err != nil || tmpl == nil {
		return fmt.Errorf("render: template %s could not be Loaded: %v", template, err)
	}
	bind := jetVarMap(binding)
	if len(layout) > 0 && layout[0] != "" {
		lay, err := e.Templates.GetTemplate(layout[0])
		if err != nil {
			return err
		}
		bind.Set(e.LayoutName, func() {
			_ = tmpl.Execute(out, bind, nil)
		})
		return lay.Execute(out, bind, nil)
	}
	return tmpl.Execute(out, bind, nil)
}

func jetVarMap(binding interface{}) jet.VarMap {
	var bind jet.VarMap
	if binding == nil {
		return bind
	}
	if binds, ok := binding.(map[string]interface{}); ok {
		bind = make(jet.VarMap)
		for key, value := range binds {
			bind.Set(key, value)
		}
	} else if binds, ok := binding.(fiber.Map); ok {
		bind = make(jet.VarMap)
		for key, value := range binds {
			bind.Set(key, value)
		}
	} else if binds, ok := binding.(jet.VarMap); ok {
		bind = binds
	}
	return bind
}
