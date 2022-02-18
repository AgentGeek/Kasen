package server

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	. "kasen/cache"
	"kasen/config"

	"github.com/gin-gonic/gin"
)

type RenderOptions struct {
	Status int
	Name   string
	Key    string
	Data   map[string]interface{}
}

const htmlContentType = "text/html; charset=utf-8"

var (
	templates           *template.Template
	mu                  sync.Mutex
	ErrTemplateNotFound = errors.New("Template not found")
)

func LoadTemplates() {
	mu.Lock()
	defer mu.Unlock()

	var files []string
	err := filepath.Walk(filepath.Join(config.GetDirectories().Root, "templates"),
		func(path string, info fs.FileInfo, err error) error {
			if info == nil || err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(path, ".html") {
				files = append(files, path)
			}
			return err
		})
	if err != nil {
		log.Fatalln(err)
	}

	templates, err = template.New("").Funcs(helper).ParseFiles(files...)
	if err != nil {
		log.Fatalln(err)
	}
}

func parseTemplate(name string, data interface{}) ([]byte, error) {
	if gin.Mode() == gin.DebugMode {
		LoadTemplates()
	}

	t := templates.Lookup(name)
	if t == nil {
		return nil, ErrTemplateNotFound
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func getTemplate(name, key string) ([]byte, bool) {
	var v interface{}
	var err error

	if len(key) > 0 {
		v, err = TemplatesCache.Get(fmt.Sprintf("%s:%s", name, key))
	} else {
		v, err = TemplatesCache.Get(name)
	}

	if err != nil {
		return nil, false
	}
	return v.([]byte), true
}

func setTemplate(name, key string, data interface{}) ([]byte, error) {
	buf, err := parseTemplate(name, data)
	if err != nil {
		return nil, err
	}

	if len(key) > 0 {
		TemplatesCache.Set(fmt.Sprintf("%s:%s", name, key), buf, 0)
	} else {
		TemplatesCache.Set(name, buf, 0)
	}
	return buf, nil
}

func renderTemplate(c *Context, cache bool, o *RenderOptions) {
	var buf []byte
	if cache {
		var ok bool
		if buf, ok = getTemplate(o.Name, o.Key); !ok {
			var err error
			buf, err = setTemplate(o.Name, o.Key, o.Data)
			if err != nil {
				c.Status(http.StatusInternalServerError)
				return
			}
		}
	} else {
		buf, _ = parseTemplate(o.Name, o.Data)
	}
	c.Data(o.Status, htmlContentType, buf)
}
