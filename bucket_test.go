package upload

import (
	"context"
	"io/ioutil"
	"os"
	"testing"
)

type unitData struct {
	name   string
	bucket string
	ioList []struct{ input, output string }
}

func GetTestData(t *testing.T) []unitData {
	t.Helper()
	_ = os.Mkdir("bin", 0777)
	return []unitData{
		{
			name:   "empty",
			bucket: "",
			ioList: []struct{ input, output string }{
				{input: "number/one.in", output: "number/one.in"},
				{input: "two", output: "two"},
			},
		},
		{
			name:   "fake",
			bucket: "fake://",
			ioList: []struct{ input, output string }{
				{input: "number/one.in", output: "number/one.in"},
				{input: "two", output: "two"},
			},
		},
		{
			name:   FileSystem.String(),
			bucket: "file://" + pwd() + "/bin/",
			ioList: []struct{ input, output string }{
				{input: "number/one.in", output: "file://" + pwd() + "/bin/number/one.in"},
				{input: "two", output: "file://" + pwd() + "/bin/two"},
			},
		},
		{
			name:   AmazonWebServices.String(),
			bucket: getenv("S3_URL", "s3://name?region=us-east-2"),
			ioList: []struct{ input, output string }{
				{
					input:  "number/one.in",
					output: getenv("S3_BASE_URL", "https://name.s3.us-east-2.amazonaws.com/") + "number/one.in",
				},
				{
					input:  "two",
					output: getenv("S3_BASE_URL", "https://name.s3.us-east-2.amazonaws.com/") + "two",
				},
			},
		},
		{
			name:   GoogleCloud.String(),
			bucket: getenv("GCP_URL", "gs://name"),
			ioList: []struct{ input, output string }{
				{
					input:  "number/one.in",
					output: getenv("GCP_BASE_URL", "https://storage.googleapis.com/name/") + "number/one.in",
				},
				{
					input:  "two",
					output: getenv("GCP_BASE_URL", "https://storage.googleapis.com/name/") + "two",
				},
			},
		},
		{
			name:   ProxiedFileSystem.String(),
			bucket: "pfs://localhost:8080" + pwd() + "/bin?route=web/file/srv",
			ioList: []struct{ input, output string }{
				{
					input:  "number/three.in",
					output: "http://localhost:8080/web/file/srv/number/three.in",
				},
				{
					input:  "four",
					output: "http://localhost:8080/web/file/srv/four",
				},
			},
		},
		{
			name:   ProxiedFileSystem.String() + " (Secure)",
			bucket: "pfs://example.com" + pwd() + "/bin/?route=web/file/srv&secure=true",
			ioList: []struct{ input, output string }{
				{
					input:  "number/three.in",
					output: "https://example.com/web/file/srv/number/three.in",
				},
				{
					input:  "four",
					output: "https://example.com/web/file/srv/four",
				},
			},
		},
		{
			name:   ProxiedFileSystem.String() + " (without route)",
			bucket: "pfs://example.com" + pwd() + "/bin",
			ioList: []struct{ input, output string }{
				{
					input:  "number/three.in",
					output: "http://example.com/number/three.in",
				},
				{
					input:  "four",
					output: "http://example.com/four",
				},
			},
		},
	}
}

func TestUpload(t *testing.T) {
	ctx := context.Background()
	_ = os.Mkdir("bin", 0777)
	tests := GetTestData(t)
	for _, tt := range tests {
		t.Run("Open/"+tt.name, func(t *testing.T) {
			for _, io := range tt.ioList {
				bucket, err := Open(tt.bucket)
				if err != nil {
					t.Errorf("Open(%v), got: %v \n", tt.name, err)
					continue
				}
				t.Log(bucket.Provider())
				if bucket.Provider() == AmazonWebServices {
					t.Skip("Credentials are removed")
				}
				// upload content
				got, err := bucket.WriteAll(ctx, io.input, []byte(io.input))
				if err != nil {
					t.Errorf("WriteAll(%v), got: %v \n", io.input, err)
					continue
				}
				if got != io.output {
					t.Errorf("GetUrl(%v), got: %v want: %v \n", io.input, got, io.output)
				}
				// read content
				con, err := bucket.ReadAll(ctx, bucket.GetName(io.output))
				if err != nil {
					t.Errorf("ReadAll(%v), got: %v \n", io.output, err)
					continue
				}
				if string(con) != io.input {
					t.Errorf("Content(%v), got: %s want: %v \n", io.output, con, io.input)
				}
				// delete file
				err = bucket.Delete(ctx, bucket.GetName(io.output))
				if err != nil {
					t.Errorf("Delete(%v), got: %v \n", io.output, err)
				}
				if err = bucket.Close(); err != nil {
					t.Errorf("Close(), got: %v \n", err)
				}
			}
		})
		t.Run("New/"+tt.name, func(t *testing.T) {
			for _, io := range tt.ioList {
				bucket := NewBucket(tt.bucket)
				t.Log(bucket.Provider())
				if bucket.Provider() == AmazonWebServices {
					t.Skip("Credentials are removed")
				}
				// upload content
				got, err := bucket.WriteAll(ctx, io.input, []byte(io.input))
				if err != nil {
					t.Errorf("WriteAll(%v), got: %v \n", io.input, err)
					continue
				}
				if got != io.output {
					t.Errorf("GetUrl(%v), got: %v want: %v \n", io.input, got, io.output)
				}
				// read content
				r, err := bucket.Reader(ctx, bucket.GetName(io.output))
				if err != nil {
					t.Errorf("Reader(%v), got: %v \n", io.output, err)
					continue
				}
				con, err := ioutil.ReadAll(r)
				if err != nil {
					t.Errorf("ReadAll(%v), got: %v \n", io.output, err)
				}
				if err := r.Close(); err != nil {
					t.Errorf("Close(%v), got: %v \n", io.output, err)
				}
				if string(con) != io.input {
					t.Errorf("Content(%v), got: %s want: %v \n", io.output, con, io.input)
				}
				// delete file
				err = bucket.Delete(ctx, bucket.GetName(io.output))
				if err != nil {
					t.Errorf("Delete(%v), got: %v \n", io.output, err)
				}
				if err = bucket.Close(); err != nil {
					t.Errorf("Close(), got: %v \n", err)
				}
			}
		})
	}
}

func TestNegative(t *testing.T) {
	t.Run("failedContext", func(t *testing.T) {
		b := &Bucket{}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := b.WriteAll(ctx, "one.in", []byte("one.in"))
		if err == nil {
			t.Error("WriteAll should failed due to cancelled context")
		}
	})
	t.Run("empty-name", func(t *testing.T) {
		b := &Bucket{}
		ctx := context.Background()
		_, err := b.WriteAll(ctx, "", []byte("one.in"))
		if err == nil {
			t.Error("WriteAll should failed due to cancelled context")
		}
	})
	t.Run("invalid-bucket", func(t *testing.T) {
		b := &Bucket{name: "wun"}
		ctx := context.Background()
		_, err := b.WriteAll(ctx, "one-in", []byte("one.in"))
		if err != nil {
			t.Error("WriteAll should not fail, it should work with mem://")
		}
	})
}

func pwd() string {
	d, _ := os.Getwd()
	return d
}

func getenv(key, _default string) string {
	val := os.Getenv(key)
	if val != "" {
		return val
	}
	return _default
}
