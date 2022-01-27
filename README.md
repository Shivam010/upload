# upload

[![Build](https://github.com/Shivam010/upload/actions/workflows/build.yml/badge.svg)](https://github.com/Shivam010/upload/actions/workflows/build.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/Shivam010/upload?dropcache)](https://goreportcard.com/report/github.com/Shivam010/upload)
[![Go Reference](https://pkg.go.dev/badge/github.com/Shivam010/upload)](https://pkg.go.dev/github.com/Shivam010/upload)
[![License](https://img.shields.io/badge/license-MIT-mildgreen.svg)](https://github.com/Shivam010/upload/blob/master/LICENSE)
[![GitHub release](https://img.shields.io/github/release/Shivam010/upload.svg)](https://github.com/Shivam010/upload/releases)

Package upload provides an easy and portable way to interact with CDN, buckets or blobs within any cloud/local storage
location. And provides methods to read or write or upload or delete files to blob storage on GCP, AWS, Azure, in-memory,
local and more.

It wraps `gocloud.dev/blob` (https://github.com/google/go-cloud/tree/master/blob) for further simplicity.

## Proxied File system with HTTP Router

You can store the files in using file system blob and can serve them using http file server. Upload library provides
everything you need to do that in-built. Just use `pfsblob` as your file system handler. See the example
in [pfsblob/tests](./pfsblob/tests).

### URL format for pfsblob

1. URL Scheme: `pfs`
2. URL Host: Your domain name (without protocol) e.g. `localhost:8080` or `example.com`
3. URL Path: Your storage directory in which you want to manage your files e.g. `/tmp/files`
4. Domain Protocol: If your domain uses secure `https` protocol, set a query parameter: `?secure=true`
5. Route: By default, files will be served at the root of the domain, if you want to change it, use query
   parameter: `?route=static`

Complete Query will look something like these:

1. **`pfs://localhost:8080/tmp/files?secure=false`** <br/>
   This will serve files at `http://localhost:8080/...` <br/>
   e.g. the link for the file `/tmp/files/image.png` will be `http://localhost:8080/image.png`

2. **`pfs://example.com/tmp/files?secure=true&route=public`** <br/>
   This will serve files at `https://example.com/public/...` <br/>
   e.g. the link for the file `/tmp/files/image.png` will be `https://example.com/public/image.png`

## Example

```go
package main

import (
	"context"
	"fmt"

	"github.com/Shivam010/upload"
)

const content = `Lorem ipsum is placeholder text commonly used in the graphic, print, and publishing industries.`

func main() {
	ctx := context.TODO()
	s3Url := "s3://name?region=us-east-2"
	// Open or create bucket using name
	buck, err := upload.Open(s3Url) // or buck := upload.NewBucket(s3Url)
	if err != nil {
		panic(err) // handle error
	}

	fmt.Println("Provider: ", buck.Provider())

	// Upload a new file and get its link
	link, err := buck.WriteAll(ctx, "loren.ipsum", []byte(content))
	if err != nil {
		panic(err) // handle error
	}
	fmt.Println("Link to the file (loren.ipsum)", link)

	// Name of the uploaded file from its link
	name := buck.GetName(link)

	// read content of the uploaded file
	cont, err := buck.ReadAll(ctx, name)
	if err != nil {
		panic(err) // handle error
	}
	fmt.Println("Content of the file (loren.ipsum)", cont)

	// delete the uploaded file
	if err = buck.Delete(ctx, name); err != nil {
		panic(err) // handle error
	}

	// Close the bucket
	if err = buck.Close(); err != nil {
		panic(err) // handle error
	}
}
```

## License

This project is licensed under the [MIT License](./LICENSE)
