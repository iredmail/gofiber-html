package gofiber_html

import (
	"bytes"
	"embed"
	"io/fs"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"testing"
)

var (
	//go:embed views
	embedViews embed.FS
)

func trim(str string) string {
	trimmed := strings.TrimSpace(regexp.MustCompile(`\s+`).ReplaceAllString(str, " "))
	trimmed = strings.Replace(trimmed, " <", "<", -1)
	trimmed = strings.Replace(trimmed, "> ", ">", -1)
	return trimmed
}

func Test_Render(t *testing.T) {
	engine := New("./views", ".html")
	engine.AddLayouts("layouts/main")

	// Partials
	var buf bytes.Buffer
	err := engine.Render(&buf, "index", nil)
	if err != nil {
		t.Fatalf("render: %v\n", err)
	}

	expect := `<!DOCTYPE html><html><head><title>Index</title></head><body><h2>Index</h2></body></html>`
	result := trim(buf.String())
	if expect != result {
		t.Fatalf("Expected:\n%s\nResult:\n%s\n", expect, result)
	}
	// Single
	buf.Reset()
	engine.Render(&buf, "errors/404", map[string]interface{}{
		"Error": "404 Not Found!",
	})
	expect = `<h1>404 Not Found!</h1>`
	result = trim(buf.String())
	if expect != result {
		t.Fatalf("Expected:\n%s\nResult:\n%s\n", expect, result)
	}
}

func Test_AddFunc(t *testing.T) {
	engine := New("./views", ".html")
	engine.AddFunc("isAdmin", func(user string) bool {
		return user == "admin"
	})
	if err := engine.Load(); err != nil {
		t.Fatalf("load: %v\n", err)
	}

	// Func is admin
	var buf bytes.Buffer
	err := engine.Render(&buf, "admin", map[string]interface{}{
		"User": "admin",
	})
	if err != nil {
		t.Fatalf("render: %v\n", err)
	}

	expect := `<h1>Hello, Admin!</h1>`
	result := trim(buf.String())
	if expect != result {
		t.Fatalf("Expected:\n%s\nResult:\n%s\n", expect, result)
	}

	// Func is not admin
	buf.Reset()
	err = engine.Render(&buf, "admin", map[string]interface{}{
		"User": "john",
	})
	if err != nil {
		t.Fatalf("render: %v\n", err)
	}

	expect = `<h1>Access denied!</h1>`
	result = trim(buf.String())
	if expect != result {
		t.Fatalf("Expected:\n%s\nResult:\n%s\n", expect, result)
	}
}

func Test_AddFuncMap(t *testing.T) {
	// Create a temporary directory
	dir, _ := ioutil.TempDir(".", "")
	defer os.RemoveAll(dir)

	// Create a temporary template file.
	_ = ioutil.WriteFile(dir+"/func_map.html", []byte(`<h2>{{lower .Var1}}</h2><p>{{upper .Var2}}</p>`), 0700)

	engine := New(dir, ".html")

	fm := map[string]interface{}{
		"lower": func(s string) string {
			return strings.ToLower(s)
		},
		"upper": func(s string) string {
			return strings.ToUpper(s)
		},
	}

	engine.AddFuncMap(fm)

	if err := engine.Load(); err != nil {
		t.Fatalf("load: %v\n", err)
	}

	var buf bytes.Buffer
	engine.Render(&buf, "func_map", map[string]interface{}{
		"Var1": "LOwEr",
		"Var2": "upPEr",
	})
	expect := `<h2>lower</h2><p>UPPER</p>`
	result := trim(buf.String())
	if expect != result {
		t.Fatalf("Expected:\n%s\nResult:\n%s\n", expect, result)
	}
}

func Test_Layout(t *testing.T) {
	engine := New("./views", ".html")
	engine.AddLayouts("layouts/main")

	engine.AddFunc("isAdmin", func(user string) bool {
		return user == "admin"
	})

	var buf bytes.Buffer
	engine.Render(&buf, "index", nil)
	expect := `<!DOCTYPE html><html><head><title>Index</title></head><body><h2>Index</h2></body></html>`
	result := trim(buf.String())
	if expect != result {
		t.Fatalf("Expected:\n%s\nResult:\n%s\n", expect, result)
	}
}

func Test_Empty_Layout(t *testing.T) {
	engine := New("./views", ".html")

	engine.AddFunc("isAdmin", func(user string) bool {
		return user == "admin"
	})

	var buf bytes.Buffer
	engine.Render(&buf, "empty_layout", nil)
	expect := `<p>Empty</p>`
	result := trim(buf.String())
	if expect != result {
		t.Fatalf("Expected:\n%s\nResult:\n%s\n", expect, result)
	}
}

// Test_Layout_Multi checks if the layout can be rendered multiple times
func Test_Layout_Multi(t *testing.T) {
	engine := New("./views", ".html")
	engine.AddLayouts("layouts/main")

	engine.AddFunc("isAdmin", func(user string) bool {
		return user == "admin"
	})

	for i := 0; i < 2; i++ {
		var buf bytes.Buffer
		err := engine.Render(&buf, "index", nil)
		if err != nil {
			t.Fatalf("render: %v\n", err)
		}

		expect := `<!DOCTYPE html><html><head><title>Index</title></head><body><h2>Index</h2></body></html>`
		result := trim(buf.String())
		if expect != result {
			t.Fatalf("Expected:\n%s\nResult:\n%s\n", expect, result)
		}
	}

}

func Test_FileSystem(t *testing.T) {
	fsViews, err := fs.Sub(embedViews, "views")
	if err != nil {
		t.Fatalf("embed: %v\n", err)
	}
	engine := NewFileSystem(fsViews, ".html")
	engine.AddLayouts("layouts/main")

	engine.AddFunc("isAdmin", func(user string) bool {
		return user == "admin"
	})

	var buf bytes.Buffer
	err = engine.Render(&buf, "index", nil)
	if err != nil {
		t.Fatalf("render: %v\n", err)
	}

	expect := `<!DOCTYPE html><html><head><title>Index</title></head><body><h2>Index</h2></body></html>`
	result := trim(buf.String())
	if expect != result {
		t.Fatalf("Expected:\n%s\nResult:\n%s\n", expect, result)
	}
}
