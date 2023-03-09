package gofiber_html

import (
	"html/template"
	"io"
	"io/fs"
	"os"
	"path"
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
	// fs.FS supports embedded files
	fileSystem fs.FS
	// views extension
	extension string
	// layouts specifies the layout template files you need frequently, so that you don't need to specify it
	// in each `ctx.Render()`.
	layouts []string
	// debug prints the parsed templates
	debug bool
	// lock for funcmap and templates
	mutex sync.RWMutex
	// template funcmap
	funcmap map[string]interface{}
	// templates are loaded and parsed first time it is called (when client is requesting the web pag).
	templates map[string]*template.Template
}

// New returns a HTML render engine for Fiber
func New(directory, extension string) *Engine {
	engine := &Engine{
		left:      "{{",
		right:     "}}",
		directory: directory,
		extension: extension,
		funcmap:   make(map[string]interface{}),
		templates: make(map[string]*template.Template),
	}

	return engine
}

// NewFileSystem ...
func NewFileSystem(fs fs.FS, extension string) *Engine {
	engine := &Engine{
		left:       "{{",
		right:      "}}",
		fileSystem: fs,
		extension:  extension,
		funcmap:    make(map[string]interface{}),
		templates:  make(map[string]*template.Template),
	}

	return engine
}

func (e *Engine) AddLayouts(layouts ...string) *Engine {
	e.layouts = layouts
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

// Debug will print the parsed templates when Load is triggered.
func (e *Engine) Debug(enabled bool) *Engine {
	e.debug = enabled
	return e
}

// Load satisfies Fiber's `Template` interface.
func (e *Engine) Load() error {
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

	for _, pth := range paths {
		var name string
		var b []byte
		if e.fileSystem != nil {
			name, b, err = readFileFS(e.fileSystem, pth)
		} else {
			name, b, err = readFileOS(pth)
		}

		if err != nil {
			return nil, err
		}
		var t *template.Template
		if tmpl == nil {
			tmpl = template.New(name)
			tmpl.Delims(e.left, e.right)
			tmpl.Funcs(e.funcmap)
		}

		if name == tmpl.Name() {
			t = tmpl
		} else {
			t = tmpl.New(name)
		}

		s := string(b)
		_, err = t.Parse(s)
		if err != nil {
			return nil, err
		}
	}

	e.templates[key] = tmpl

	return tmpl, nil
}

// Render will execute the template name along with the given values.
func (e *Engine) Render(out io.Writer, template string, binding interface{}, layout ...string) error {
	filenames := []string{template}
	if len(layout) > 0 {
		filenames = append(filenames, layout...)
	} else {
		filenames = append(filenames, e.layouts...)
	}

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

func (e *Engine) FuncMap() map[string]any {
	return e.funcmap
}

func readFileOS(file string) (name string, b []byte, err error) {
	name = filepath.Base(file)
	b, err = os.ReadFile(file)

	return
}

func readFileFS(fsys fs.FS, file string) (name string, b []byte, err error) {
	name = path.Base(file)
	b, err = fs.ReadFile(fsys, file)

	return
}
