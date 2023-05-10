package internal

import (
	t "github.com/gofiber/template"
	"net/http"
	"sync"
)

// Engine engine struct
type Engine struct {
	// delimiters
	Left  string
	Right string
	// views folder
	Directory string
	// http.FileSystem supports embedded files
	FileSystem http.FileSystem
	// views extension
	Extension string
	// layout variable name that incapsulates the template
	LayoutName string
	// determines if the engine parsed all templates
	Loaded bool
	// reload on each render
	ShouldReload bool
	// debug prints the parsed templates
	Verbose bool
	// lock for funcmap and templates
	Mutex sync.RWMutex
	// template funcmap
	Funcmap map[string]interface{}
}

// AddFunc adds the function to the template's function map.
// It is legal to overwrite elements of the default actions
func (e *Engine) AddFunc(name string, fn interface{}) t.EngineInterface {
	e.Mutex.Lock()
	e.Funcmap[name] = fn
	e.Mutex.Unlock()
	return e
}

// AddFuncMap adds the functions from a map to the template's function map.
// It is legal to overwrite elements of the default actions
func (e *Engine) AddFuncMap(m map[string]interface{}) t.EngineInterface {
	e.Mutex.Lock()
	for name, fn := range m {
		e.Funcmap[name] = fn
	}
	e.Mutex.Unlock()
	return e
}

// Debug will print the parsed templates when Load is triggered.
func (e *Engine) Debug(enabled bool) t.EngineInterface {
	e.Verbose = enabled
	return e
}

// Delims sets the action delimiters to the specified strings, to be used in
// templates. An empty delimiter stands for the
// corresponding default: {{ or }}.
func (e *Engine) Delims(left, right string) t.EngineInterface {
	e.Left, e.Right = left, right
	return e
}

// FuncMap returns the template's function map.
func (e *Engine) FuncMap() map[string]interface{} {
	return e.Funcmap
}

// Layout defines the variable name that will incapsulate the template
func (e *Engine) Layout(key string) t.EngineInterface {
	e.LayoutName = key
	return e
}

// Reload if set to true the templates are reloading on each render,
// use it when you're in development and you don't want to restart
// the application when you edit a template file.
func (e *Engine) Reload(enabled bool) t.EngineInterface {
	e.ShouldReload = enabled
	return e
}
