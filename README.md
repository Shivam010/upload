# upload 
Package upload provides an easy and portable way to interact with CDN, buckets or blobs within any cloud/local storage location. And provides methods to read or write or upload or delete files to blob storage on GCP, AWS, Azure, in-memory, local and more.

It wraps `gocloud.dev/blob` (https://github.com/google/go-cloud/tree/master/blob) for further simplicity.

### Example
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

### License
This project is licensed under the [MIT License](./LICENSE)
