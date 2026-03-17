package static

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"dbut.dev/float/web"
)

var configKeys = []string{
	"FIREBASE_API_KEY",
	"FIREBASE_AUTH_DOMAIN",
	"FIREBASE_PROJECT_ID",
	"FIREBASE_STORAGE_BUCKET",
	"FIREBASE_MESSAGING_SENDER_ID",
	"FIREBASE_APP_ID",
	"FIREBASE_VAPID_KEY",
}

func Register(r *gin.Engine) {
	configMap := make(map[string]string)
	for _, key := range configKeys {
		configMap[key] = os.Getenv("VITE_" + key)
	}

	configJSON, err := json.Marshal(configMap)
	if err != nil {
		panic(err)
	}

	r.GET("/config.json", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/json; charset=utf-8", configJSON)
	})

	swJS, err := fs.ReadFile(web.Dist, "firebase-messaging-sw.js")
	if err == nil {
		swStr := string(swJS)
		for _, key := range configKeys {
			swStr = strings.ReplaceAll(swStr, "__"+key+"__", configMap[key])
		}
		swBytes := []byte(swStr)
		r.GET("/firebase-messaging-sw.js", func(c *gin.Context) {
			c.Data(http.StatusOK, "application/javascript; charset=utf-8", swBytes)
		})
	}

	indexHTML, err := fs.ReadFile(web.Dist, "index.html")
	if err != nil {
		panic(err)
	}

	serveIndex := func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
	}

	distFS := http.FS(web.Dist)
	r.GET("/", serveIndex)
	r.NoRoute(func(c *gin.Context) {
		path := strings.TrimPrefix(c.Request.URL.Path, "/")
		if f, err := web.Dist.Open(path); err == nil {
			_ = f.Close()
			c.FileFromFS(c.Request.URL.Path, distFS)
			return
		}
		serveIndex(c)
	})
}
