package pfsblob

import (
	"context"
	"gocloud.dev/blob"
	file "gocloud.dev/blob/fileblob"
	"net/url"
)

// Scheme is the URL scheme pfsblob registers its URLOpener under on
// blob.DefaultMux.
const Scheme = "pfs"

func init() {
	blob.DefaultURLMux().RegisterBucket(Scheme, &URLOpener{fileOpener: &file.URLOpener{}})
}

// URLOpener opens file bucket URLs like "pfs:///foo/bar/baz?secure=false&route=", using file system blob
type URLOpener struct {
	// File System blob opener
	fileOpener *file.URLOpener
}

func (o *URLOpener) OpenBucketURL(ctx context.Context, u *url.URL) (*blob.Bucket, error) {
	fu := changeURLToFileScheme(u)
	return o.fileOpener.OpenBucketURL(ctx, fu)
}

// changeURLToFileScheme will change the scheme of the url to "file" and return a new url.URL
func changeURLToFileScheme(u *url.URL) *url.URL {
	fu := *u
	fu.Host = ""
	fu.RawQuery = ""
	fu.Scheme = file.Scheme
	return &fu
}
