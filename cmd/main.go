package main

import (
	"context"
	"flag"
	"image-download-tool/internal/app"
	"image-download-tool/internal/config"
	"runtime"
)

var linksSource = flag.String("src", "links.json", "a file with links in json format (a simple array of strings)")
var workers = flag.Int("wrk", runtime.NumCPU()*8, "amount of workers, default=runtime.NumCPU()*8")
var dst = flag.String("dst", "./data", "destination")

func main() {
	flag.Parse()

	cfg := config.Config{
		TargetDir:  *dst,
		SourceFile: *linksSource,
		Workers:    *workers,
	}

	app.New(cfg).Start(context.Background())
}
