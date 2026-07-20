package ctrl

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	. "github.com/mickael-kerjean/filestash/server/common"
	"github.com/mickael-kerjean/filestash/server/pkg/permissions"
)

func FileSearch(ctx *App, res http.ResponseWriter, req *http.Request) {
	path, err := PathBuilder(ctx, req.URL.Query().Get("path"))
	if err != nil {
		path = "/"
	}
	q := req.URL.Query().Get("q")
	from, err := parseOptionalInt64(req.URL.Query().Get("from"))
	if err != nil {
		SendErrorResult(res, ErrNotValid)
		return
	}
	to, err := parseOptionalInt64(req.URL.Query().Get("to"))
	if err != nil {
		SendErrorResult(res, ErrNotValid)
		return
	}
	if q == "" && from == nil && to == nil {
		SendErrorResult(res, ErrNotValid)
		return
	}
	if permissions.CanRead(ctx) == false {
		Log.Debug("ctrl::search 'can not read \"%s\"'", path)
		SendErrorResult(res, ErrPermissionDenied)
		return
	}

	searchEngine := Hooks.Get.SearchEngine()
	if searchEngine == nil {
		SendErrorResult(res, ErrMissingDependency)
		return
	}

	rangeFilter := SearchTimeRange{From: from, To: to}
	if HasSearchTimeRange(rangeFilter) {
		if ctx.Context == nil {
			ctx.Context = context.Background()
		}
		ctx.Context = WithSearchTimeRange(ctx.Context, rangeFilter)
	}

	searchResults, err := searchEngine.Query(*ctx, path, q)
	if err != nil {
		SendErrorResult(res, err)
		return
	}

	// overwrite the path of a file according to chroot
	if ctx.Session["path"] != "" {
		for i := 0; i < len(searchResults); i++ {
			searchResults[i] = File{
				FName: searchResults[i].Name(),
				FSize: searchResults[i].Size(),
				FTime: FileTimeMs(searchResults[i]),
				FType: func() string {
					if searchResults[i].IsDir() {
						return "directory"
					}
					return "file"
				}(),
				FPath: "/" + strings.TrimPrefix(
					searchResults[i].Path(),
					ctx.Session["path"],
				),
			}
		}
	}
	SendSuccessResults(res, searchResults)
}

func parseOptionalInt64(s string) (*int64, error) {
	if s == "" {
		return nil, nil
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return nil, err
	}
	return &v, nil
}
