package static

import (
	"encoding/json"
	"fmt"
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
	assetsFS, err := fs.Sub(web.Dist, "assets")
	if err != nil {
		panic(err)
	}
	r.StaticFS("/assets", http.FS(assetsFS))

	configMap := make(map[string]string)
	for _, key := range configKeys {
		configMap[key] = os.Getenv("VITE_" + key)
	}

	indexHTML, err := fs.ReadFile(web.Dist, "index.html")
	if err != nil {
		panic(err)
	}

	configJSON, err := json.Marshal(configMap)
	if err != nil {
		panic(err)
	}
	configScript := fmt.Sprintf(`<script>window.__CONFIG__=%s</script>`, configJSON)
	indexStr := strings.Replace(string(indexHTML), "</head>", configScript+"</head>", 1)
	indexBytes := []byte(indexStr)

	serveIndex := func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexBytes)
	}

	// Replace placeholders in service worker JS
	swJS, err := fs.ReadFile(web.Dist, "firebase-messaging-sw.js")
	if err == nil {
		swStr := string(swJS)
		for _, key := range configKeys {
			swStr = strings.ReplaceAll(swStr, "__"+key+"__", configMap[key])
		}
		swBytes := []byte(swStr)

		r.GET("/firebase-messaging-sw.js", func(c *gin.Context) {
			c.Header("Service-Worker-Allowed", "/")
			c.Header("Cache-Control", "no-store")
			c.Data(http.StatusOK, "application/javascript; charset=utf-8", swBytes)
		})
	}

	r.GET("/", serveIndex)
	r.NoRoute(serveIndex)
}
