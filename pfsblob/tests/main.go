package main

import (
	"context"
	"fmt"
	"github.com/Shivam010/upload"
	pfs "github.com/Shivam010/upload/pfsblob/handler"
	"io"
	"net/http"
	"os"
)

const (
	route  = "/srv"
	domain = "localhost"
	port   = ":8080"
)

func main() {
	dir := fuzzDirPath()
	fmt.Println(dir)
	// pfs://localhost:8080/.../pfsblob/tests/fuzz?secure=false&route=srv
	buckUrl := "pfs://" + domain + port + dir + "?secure=false&route=" + route[1:]

	fmt.Println("Connecting", buckUrl)
	buck := upload.NewBucket(buckUrl)

	// Starting server
	go server(buck, port)

	FileReadExample(buck)
	FileWriteExample(buck)
	HttpFileReadExample(buck)

	// Wait infinitely
	//select {}
}

func server(buck *upload.Bucket, port string) {
	route, handler, _ := pfs.BucketRouteAndHandler(buck)
	http.HandleFunc(route, handler)
	_ = http.ListenAndServe(port, nil)
}

func fuzzDirPath() string {
	p, _ := os.Getwd()
	p += "/pfsblob/tests/fuzz"
	return p
}

func FileReadExample(buck *upload.Bucket) {
	ctx := context.Background()
	fileUrl := buck.GetUrl("root.txt")
	fmt.Println("Reading File:", fileUrl)
	data, err := buck.ReadAll(ctx, buck.GetName(fileUrl))
	if err != nil {
		panic("ops " + err.Error())
	}
	fmt.Println("Data in file:", fileUrl)
	fmt.Println(string(data))
	fmt.Println()
}

func FileWriteExample(buck *upload.Bucket) {
	ctx := context.Background()
	name := "new/dir/fuzz.txt"
	fmt.Println("Writing File:", name)
	link, err := buck.WriteAll(ctx, name, []byte("Writing in file: "+name+"\n link: "+buck.GetUrl(name)))
	if err != nil {
		panic("ops " + err.Error())
	}
	fmt.Println("Link of file:", name)
	fmt.Println(link)
	fmt.Println()
}

func HttpFileReadExample(buck *upload.Bucket) {
	link := buck.GetUrl("root.txt")
	res, err := http.Get(link)
	if err != nil {
		panic("ops: " + err.Error())
	}
	fmt.Println("Reading File", link, "through HTTP")
	fmt.Println("Status Code:", res.Status)
	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic("ops read: " + err.Error())
	}
	fmt.Println("File Data:", string(body))
	fmt.Println()
}
