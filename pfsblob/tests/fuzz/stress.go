package main

import (
	"fmt"
	"golang.org/x/sync/errgroup"
	"net/http"
	"time"
)

func main() {
	const (
		limit  = 100
		domain = "http://localhost:8080"
	)
	var routes = map[string]int{
		"/srv":                        http.StatusNotFound,
		"/srv/main.go":                http.StatusNotFound,
		"/srv/root.txt":               http.StatusOK,
		"/srv/../main.go":             http.StatusNotFound,
		"/srv/stress.go":              http.StatusOK,
		"/srv/dir/":                   http.StatusNotFound,
		"/srv/dir/root.txt":           http.StatusNotFound,
		"/srv/dir/buzz":               http.StatusOK,
		"/srv/dir/fuzz":               http.StatusOK,
		"/../srv/dir/fuzz":            http.StatusOK,
		"/srv/dir/buzz.txt":           http.StatusNotFound,
		"/srv/new/dir":                http.StatusNotFound,
		"/srv/new/dir/fuzz.txt":       http.StatusOK,
		"/srv/new/dir/fuzz.txt.attrs": http.StatusOK,
	}

	now := time.Now()
	gr := errgroup.Group{}
	for i := 0; i < limit; i++ {
		gr.Go(func() error {
			for r, c := range routes {
				res, err := http.Get(domain + r)
				if err != nil {
					fmt.Println("failed", r, err)
					continue
				}
				if res.StatusCode != c {
					fmt.Println("status mismatch", r, "got", res.StatusCode, "want", c)
				} else {
					//fmt.Println(res.Request.URL, res.Request.Referer(), res.StatusCode)
				}
			}
			return nil
		})
	}
	fmt.Println(time.Since(now))
	_ = gr.Wait()
}
