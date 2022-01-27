package upload

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	pfs "github.com/Shivam010/upload/pfsblob"
	"gocloud.dev/blob"
	file "gocloud.dev/blob/fileblob"
	gcs "gocloud.dev/blob/gcsblob"
	mem "gocloud.dev/blob/memblob"
	s3 "gocloud.dev/blob/s3blob"
)

// Provider describes the information about the bucket service provider
type Provider int

const (
	InMemory Provider = iota
	FileSystem
	GoogleCloud
	AmazonWebServices
	ProxiedFileSystem
)

var ProviderName = map[Provider]string{
	InMemory:          "In-Memory",
	FileSystem:        "Local File System",
	GoogleCloud:       "Google Cloud Console",
	AmazonWebServices: "Amazon Web Services",
	ProxiedFileSystem: "Proxied File System",
}

func (p Provider) String() string {
	return ProviderName[p]
}

type Bucket struct {
	url      string
	name     string
	provider Provider
	bucket   *blob.Bucket
	metadata map[string]string
}

// NewBucket will return the blob bucket using the provided bucket url
func NewBucket(bucket string) *Bucket {
	b := &Bucket{url: bucket, metadata: map[string]string{}}
	b.parse()
	return b
}

// Open will return the blob bucket with an open bucket, it is recommended
// It should be used in constructor
func Open(bucket string) (*Bucket, error) {
	b := NewBucket(bucket)
	return b, b.Open()
}

// parse will parse the bucket url and extra the meaningful information
func (b *Bucket) parse() {
	if b == nil || b.bucket != nil {
		return
	}
	if b.url == "" {
		b.url = mem.Scheme + "://"
	}
	u, _ := url.Parse(b.url)
	if u == nil || u.Scheme == "" {
		u = &url.URL{Scheme: mem.Scheme}
	}
	switch u.Scheme {
	case mem.Scheme:
		b.name = u.Host
		b.provider = InMemory
	case file.Scheme:
		if b.url[len(b.url)-1] == '/' {
			b.url = b.url[:len(b.url)-1]
		}
		b.provider = FileSystem
		b.name = u.Host + "/" + u.Path
	case gcs.Scheme:
		b.provider = GoogleCloud
		b.name = u.Host
	case s3.Scheme:
		b.provider = AmazonWebServices
		b.name = u.Host
		b.metadata["region"] = u.Query().Get("region")
	case pfs.Scheme:
		b.provider = ProxiedFileSystem
		// storage directory of the file system
		b.metadata["storage"] = u.Path
		// listening route region
		route := strings.Trim(u.Query().Get("route"), "/")
		b.metadata["route"] = route

		// name of the bucket (for the link purposes)
		b.name = "http://"
		isSecure := u.Query().Get("secure")
		if isSecure == "true" {
			b.name = "https://"
		}
		b.name += u.Host
		if route != "" {
			b.name += "/" + route
		}
	default:
		b.name = u.Host
		b.url = mem.Scheme + "://" + b.name
		b.provider = InMemory
	}
}

// Provider returns the type of service provider in use
func (b *Bucket) Provider() Provider {
	return b.provider
}

// String implements stringer on Bucket
func (b *Bucket) String() string {
	return b.Provider().String() + " Bucket"
}

// Name returns name of the bucket
func (b *Bucket) Name() string {
	return b.name
}

// URL returns the url of the bucket
func (b *Bucket) URL() string {
	return b.url
}

// GetMetadata returns the value stored inside the key of the metadata of bucket
func (b *Bucket) GetMetadata(key string) string {
	return b.metadata[key]
}

// OpenContext opens a new bucket connection
func (b *Bucket) OpenContext(ctx context.Context) (err error) {
	b.parse()
	b.bucket, err = blob.OpenBucket(ctx, b.url)
	return err
}

// Open opens a new bucket connection
func (b *Bucket) Open() error {
	return b.OpenContext(context.Background())
}

// Close opens a new bucket connection
func (b *Bucket) Close() error {
	if b.bucket != nil {
		return b.bucket.Close()
	}
	return nil
}

// WriteAll will upload the content of data, under the provided path/name in name
// and returns the corresponding access url or error if any
func (b *Bucket) WriteAll(ctx context.Context, name string, data []byte) (string, error) {
	if name == "" {
		return "", errors.New("bucket: name of file-content is required")
	}
	if b.bucket == nil {
		if err := b.OpenContext(ctx); err != nil {
			return "", err
		}
	}

	w, err := b.bucket.NewWriter(ctx, name, nil)
	if err != nil {
		return "", err
	}
	if _, err := w.Write(data); err != nil {
		_ = w.Close()
		return "", err
	}
	return b.GetUrl(name), w.Close()
}

// GetUrl returns the access url path for the provided name in the corresponding provider
func (b *Bucket) GetUrl(name string) string {
	switch b.provider {
	case InMemory:
		return name
	case FileSystem:
		return b.url + "/" + name
	case GoogleCloud:
		return fmt.Sprintf("https://storage.googleapis.com/%v/%v", b.name, name)
	case AmazonWebServices:
		return fmt.Sprintf("https://%v.s3.%v.amazonaws.com/%v", b.name, b.GetMetadata("region"), name)
	case ProxiedFileSystem:
		return b.name + "/" + name
	}
	return b.name + "/" + name
}

// GetName returns the actual blob key path in the bucket for provided link in the corresponding provider
func (b *Bucket) GetName(link string) string {
	prefix := ""
	switch b.provider {
	case InMemory:
		return link
	case FileSystem:
		prefix = b.url + "/"
	case GoogleCloud:
		prefix = fmt.Sprintf("https://storage.googleapis.com/%v/", b.name)
	case AmazonWebServices:
		prefix = fmt.Sprintf("https://%v.s3.%v.amazonaws.com/", b.name, b.GetMetadata("region"))
	case ProxiedFileSystem:
		prefix = b.name + "/"
	}
	if len(link) > len(prefix) {
		return link[len(prefix):]
	}
	return link
}

// Reader will return the io.ReadCloser against the file name provided, remember to close reader
// * name should be file name, not the http-link to get name from link use GetName method
func (b *Bucket) Reader(ctx context.Context, name string) (io.ReadCloser, error) {
	if name == "" {
		return nil, errors.New("bucket: name of file-content is required")
	}
	if b.bucket == nil {
		if err := b.OpenContext(ctx); err != nil {
			return nil, err
		}
	}
	return b.bucket.NewReader(ctx, name, nil)
}

// ReadAll will read all content of file name in bucket
// * name should be file name, not the http-link to get name from link use GetName method
func (b *Bucket) ReadAll(ctx context.Context, name string) ([]byte, error) {
	if name == "" {
		return nil, errors.New("bucket: name of file-content is required")
	}
	if b.bucket == nil {
		if err := b.OpenContext(ctx); err != nil {
			return nil, err
		}
	}
	return b.bucket.ReadAll(ctx, name)
}

// Delete will delete the file name provided from corresponding provider
// * name should be file name, not the http-link to get name from link use GetName method
func (b *Bucket) Delete(ctx context.Context, name string) error {
	if name == "" {
		return errors.New("bucket: name of file-content is required")
	}
	if b.bucket == nil {
		if err := b.OpenContext(ctx); err != nil {
			return err
		}
	}
	return b.bucket.Delete(ctx, name)
}
