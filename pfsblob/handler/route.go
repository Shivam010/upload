package handler

import (
	"errors"
	"github.com/Shivam010/upload"
	"net/http"
	"os"
)

func BucketRouteAndHandler(buck *upload.Bucket) (route string, handler func(http.ResponseWriter, *http.Request), err error) {
	if buck == nil || buck.Provider() != upload.ProxiedFileSystem {
		return "", nil, errors.New("handler: can only work on Proxied File System Bucket provider")
	}
	// Bucket region is alias for the route
	// Bucket account is alias for the storage directory
	route = "/" + buck.GetMetadata("route")
	handlerFunc := http.StripPrefix(
		route,
		http.FileServer(wrappedFileSystem{
			fs: http.Dir(buck.GetMetadata("storage")),
		}),
	).ServeHTTP

	return route + "/",
		func(w http.ResponseWriter, r *http.Request) {
			// security and caching headers
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("Referrer-Policy", "strict-origin")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("X-Frame-Options", "SAMEORIGIN")
			w.Header().Set("Strict-Transport-Security", "max-age=31536000")
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			handlerFunc(w, r)
		},
		nil
}

type wrappedFileSystem struct {
	fs http.FileSystem
}

func (w wrappedFileSystem) Open(path string) (http.File, error) {
	file, err := w.fs.Open(path)
	if err != nil {
		return nil, err
	}

	st, err := file.Stat()
	if st.IsDir() {
		return nil, os.ErrNotExist
	}
	return file, nil
}
