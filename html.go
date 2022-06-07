package gofiber_html

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"
)

// Engine struct
type Engine struct {
	// delimiters
	left  string
	right string
	// views folder
	directory string
	// http.FileSystem supports embedded files
	fileSystem fs.FS
	// views extension
	extension string
	// layout variable name that incapsulates the template
	layout string
	// determines if the engine parsed all templates
	loaded bool
	// reload on each render
	reload bool
	// debug prints the parsed templates
	debug bool
	// lock for funcmap and templates
	mutex sync.RWMutex
	// template funcmap
	funcmap map[string]interface{}
	// templates
	templates map[string]*template.Template
}

// New returns a HTML render engine for Fiber
func New(directory, extension string) *Engine {
	engine := &Engine{
		left:      "{{",
		right:     "}}",
		directory: directory,
		extension: extension,
		layout:    "embed",
		funcmap:   make(map[string]interface{}),
	}
	engine.AddFunc(engine.layout, func() error {
		return fmt.Errorf("layout called unexpectedly.")
	})
	return engine
}

//NewFileSystem ...
func NewFileSystem(fs fs.FS, extension string) *Engine {
	engine := &Engine{
		left:       "{{",
		right:      "}}",
		directory:  "/",
		fileSystem: fs,
		extension:  extension,
		layout:     "embed",
		funcmap:    make(map[string]interface{}),
	}
	engine.AddFunc(engine.layout, func() error {
		return fmt.Errorf("layout called unexpectedly.")
	})
	return engine
}

// Layout defines the variable name that will incapsulate the template
func (e *Engine) Layout(key string) *Engine {
	e.layout = key
	return e
}

// Delims sets the action delimiters to the specified strings, to be used in
// templates. An empty delimiter stands for the
// corresponding default: {{ or }}.
func (e *Engine) Delims(left, right string) *Engine {
	e.left, e.right = left, right
	return e
}

// AddFunc adds the function to the template's function map.
// It is legal to overwrite elements of the default actions
func (e *Engine) AddFunc(name string, fn interface{}) *Engine {
	e.mutex.Lock()
	e.funcmap[name] = fn
	e.mutex.Unlock()
	return e
}

// AddFuncMap adds the functions from a map to the template's function map.
// It is legal to overwrite elements of the default actions
func (e *Engine) AddFuncMap(m map[string]interface{}) *Engine {
	e.mutex.Lock()
	for name, fn := range m {
		e.funcmap[name] = fn
	}
	e.mutex.Unlock()
	return e
}

// Reload if set to true the templates are reloading on each render,
// use it when you're in development and you don't want to restart
// the application when you edit a template file.
func (e *Engine) Reload(enabled bool) *Engine {
	e.reload = enabled
	return e
}

// Debug will print the parsed templates when Load is triggered.
func (e *Engine) Debug(enabled bool) *Engine {
	e.debug = enabled
	return e
}

// Parse is deprecated, please use Load() instead
func (e *Engine) Parse() error {
	fmt.Println("Parse() is deprecated, please use Load() instead.")
	return e.Load()
}

// Load reset templates.
func (e *Engine) Load() error {
	e.templates = make(map[string]*template.Template)

	return nil
}

func (e *Engine) appendTemplate(key string, filenames ...string) (tmpl *template.Template, err error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	var paths []string
	for _, filename := range filenames {
		pth := filepath.Join(e.directory, filename+e.extension)
		pth = filepath.ToSlash(pth)
		paths = append(paths, pth)
	}

	t := template.New(key)
	t.Delims(e.left, e.right)
	t.Funcs(e.funcmap)
	if e.fileSystem != nil {
		tmpl, err = t.ParseFS(e.fileSystem, paths...)
		if err != nil {
			return
		}
	} else {
		tmpl, err = t.ParseFiles(paths...)
		if err != nil {
			return
		}
	}

	e.templates[key] = tmpl

	return tmpl, nil
}

// Render will execute the template name along with the given values.
func (e *Engine) Render(out io.Writer, template string, binding interface{}, layout ...string) error {
	if !e.loaded || e.reload {
		if e.reload {
			e.loaded = false
		}
		if err := e.Load(); err != nil {
			return err
		}
	}

	filenames := []string{template}
	filenames = append(filenames, layout...)
	key := strings.Join(filenames, ",")
	tmpl, exist := e.templates[key]
	if exist {
		return tmpl.Execute(out, binding)
	}

	tmpl, err := e.appendTemplate(key, filenames...)
	if err != nil {
		return err
	}

	return tmpl.Execute(out, binding)
}
