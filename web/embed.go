package web

import (
	"embed"
	"io/fs"
	"log"
)

//go:embed dist/*
var dist embed.FS

var Dist fs.FS

func init() {
	var err error
	Dist, err = fs.Sub(dist, "dist")
	if err != nil {
		log.Fatal(err)
	}
}
