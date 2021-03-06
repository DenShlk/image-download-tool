package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image-download-tool/internal/config"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sync/atomic"
)

type App struct {
	cfg     config.Config
	counter int64
	items   chan []string
}

func (app *App) load(ctx context.Context, item []string) {

	name := item[1] + ".jpg"
	link := item[0]

	//extension := link[len(link) - 4:]
	//name += extension

	filePath := path.Join(app.cfg.TargetDir, name)

	if !app.cfg.Rewrite {
		if _, err := os.Stat(filePath); !errors.Is(err, fs.ErrNotExist) {
			// file exists
			return
		}
	}
	ensureDir(filePath)

	req, err := http.NewRequestWithContext(ctx, "GET", link, nil)

	if err != nil {
		log.Printf("failed to create a request: %e\n", err)
		return
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("failed to fetch a link: %s %e\n", link, err)
		return
	}
	if res.StatusCode != 200 {
		log.Printf("inavlid status code: %d\n", res.StatusCode)
		return
	}
	defer res.Body.Close()

	file, err := os.Create(filePath)

	if err != nil {
		log.Println("failed to open a file:", filePath, err)
		return
	}
	defer file.Close()

	if _, err := io.Copy(file, res.Body); err != nil {
		log.Printf("failed to write to a file: %e\n", err)
		return
	}

	log.Printf("successfully loaded %d", atomic.AddInt64(&app.counter, 1))
}

func (app *App) saveMetaData(ctx context.Context, link string) {
	//todo
}

func (app *App) runWorker(ctx context.Context) {
	for {
		select {
		case item := <-app.items:
			app.load(ctx, item)
		case <-ctx.Done():
			return
		}
	}
}

func (app *App) loadLinks(ctx context.Context) {
	file, err := os.Open(app.cfg.SourceFile)
	if err != nil {
		log.Println("(loading) failed to open a file:", err)
		return
	}

	defer file.Close()

	var items [][]string

	if err := json.NewDecoder(file).Decode(&items); err != nil {
		log.Println("failed to unmarshal json:", err)
		return
	}

	for _, item := range items {
		select {
		case <-ctx.Done():
			return
		case app.items <- item:
		}
	}
}

func ensureDir(fileName string) error {
	dirName := filepath.Dir(fileName)
	if _, serr := os.Stat(dirName); serr != nil {
		merr := os.MkdirAll(dirName, os.ModePerm)
		if merr != nil {
			return merr
		}
	}
	return nil
}

func (app *App) Start(ctx context.Context) {
	if err := ensureDir(app.cfg.TargetDir); err != nil {
		fmt.Println("Target directory creation failed with error: " + err.Error())
		return
	}

	for i := 0; i < app.cfg.Workers; i++ {
		go app.runWorker(ctx)
	}

	go app.loadLinks(ctx)

	select {
	case <-ctx.Done():
		return
	}
}

func New(cfg config.Config) *App {
	return &App{
		cfg:     cfg,
		counter: 0,
		items:   make(chan []string),
	}
}
