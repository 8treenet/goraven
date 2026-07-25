package infra

import (
	"fmt"
	"net/http"
	"os"

	iris "github.com/8treenet/iris/v12"
)

func ServeFile(ctx iris.Context, absPath string, displayName string) {
	f, err := os.Open(absPath)
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.WriteString(err.Error())
		return
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		ctx.StatusCode(http.StatusInternalServerError)
		ctx.WriteString(err.Error())
		return
	}

	etag := fmt.Sprintf(`"%x%x"`, info.ModTime().UnixNano()/1e6, info.Size())
	ctx.Header("Etag", etag)
	if ctx.GetHeader("If-None-Match") == etag {
		ctx.StatusCode(http.StatusNotModified)
		return
	}

	ctx.Header("Cache-Control", "no-cache")
	http.ServeContent(ctx.ResponseWriter(), ctx.Request(), displayName, info.ModTime(), f)
}
