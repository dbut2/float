package static

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"dbut.dev/float/web"
)

func Register(r *gin.Engine) {
	httpFS := http.FS(web.Dist)

	r.StaticFS("/assets", httpFS)
	r.GET("/", func(c *gin.Context) {
		c.FileFromFS("index.html", httpFS)
	})
	r.NoRoute(func(c *gin.Context) {
		c.FileFromFS("index.html", httpFS)
	})
}
