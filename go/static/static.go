package static

import (
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"

	"dbut.dev/float/web"
)

func Register(r *gin.Engine) {
	assetsFS, err := fs.Sub(web.Dist, "assets")
	if err != nil {
		panic(err)
	}
	r.StaticFS("/assets", http.FS(assetsFS))

	indexHTML, err := fs.ReadFile(web.Dist, "index.html")
	if err != nil {
		panic(err)
	}

	serveIndex := func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
	}

	r.GET("/", serveIndex)
	r.NoRoute(serveIndex)
}
