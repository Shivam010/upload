package upload_test

import (
	"context"
	"fmt"

	"github.com/Shivam010/upload"
)

const content = `Lorem ipsum is placeholder text commonly used in the graphic, print, and publishing industries for previewing layouts and visual mockups`

func Example() {
	ctx := context.TODO()

	// Different types of blob/file systems
	list := []string{
		"",                           // or "mem://" (In-memory)
		"gs://name",                  // GCP cloud bucket url
		"file:///tmp/bin/",           // Local file system url
		"s3://name?region=us-east-2", // S3 bucket url
	}
	for _, name := range list {
		// Open or create bucket using name
		buck, err := upload.Open(name) // or buck := upload.NewBucket(name)
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
}
