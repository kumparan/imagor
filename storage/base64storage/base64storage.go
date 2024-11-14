package base64storage

import (
	"flag"
	"fmt"
	"github.com/cshum/imagor"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strings"
)

const formFieldName = "file"
const maxMemory int64 = 1024 * 1024 * 64

type Base64Storage struct{}

func New() *Base64Storage {
	fmt.Println("BIKIN BARU BASE")
	return &Base64Storage{}
}

func WithBase64(fs *flag.FlagSet, cb func() (*zap.Logger, bool)) imagor.Option {
	return func(app *imagor.Imagor) {
		app.Loaders = append(app.Loaders, New())
	}
}

// Get implements imagor.Storage interface
func (s *Base64Storage) Get(r *http.Request, _ string) (*imagor.Blob, error) {
	//ctx := r.Context()
	//attrs, err := object.Attrs(ctx)
	fmt.Println("AAAAH BABIK")
	if isFormBody(r) {
		fmt.Println("DARI BODY")
		b, err := readFormBody(r)
		if err != nil {
			return nil, err
		}
		return imagor.NewBlobFromBytes(b), nil
	}

	b, err := readRawBody(r)
	if err != nil {
		return nil, err
	}

	return imagor.NewBlobFromBytes(b), nil
}

func isFormBody(r *http.Request) bool {
	return strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/")
}

func readRawBody(r *http.Request) ([]byte, error) {
	fmt.Println("DARI RAW")
	return io.ReadAll(r.Body)
}

func readFormBody(r *http.Request) ([]byte, error) {
	err := r.ParseMultipartForm(maxMemory)
	if err != nil {
		return nil, err
	}

	file, _, err := r.FormFile(formFieldName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buf, err := io.ReadAll(file)
	if len(buf) == 0 {
		err = imagor.ErrEmptyBody
	}

	return buf, err
}
