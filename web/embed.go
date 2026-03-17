package web

import (
	"embed"
	"io/fs"
	"log"
)

//go:embed dist/*
var dist embed.FS

var Dist = func() fs.FS {
	d, err := fs.Sub(dist, "dist")
	if err != nil {
		log.Fatal(err)
	}
	return d
}()
