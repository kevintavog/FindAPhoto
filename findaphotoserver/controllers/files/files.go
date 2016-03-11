package files

import (
	"bytes"
	"errors"
	"net/url"
	"path"
	"strings"

	"github.com/go-playground/lars"

	"github.com/kevintavog/findaphoto/findaphotoserver/configuration"
)

func ConfigureRouting(l *lars.LARS) {
	files := l.Group("/files")
	files.Get("/thumbs/*", Thumbs)
	files.Get("/slides/*", Slides)
	files.Get("/media/*", Media)
}

func toRepositoryId(itemUrl string) (string, error) {
	return url.QueryUnescape(strings.Replace(itemUrl, "/", "\\", -1))
}

func aliasedToFullPath(aliasedPath string) (string, error) {
	// 1\2014\IMG_8489.JPG
	// Grab string up to first '\' (excluding '\')
	// Find matching alias
	// Append remaining string to path
	aliasIndex := strings.Index(aliasedPath, "/")
	if aliasIndex < 1 {
		return "", errors.New("Can't find alias token: '" + aliasedPath + "'")
	}

	alias := aliasedPath[0:aliasIndex]
	for _, pathAndAlias := range configuration.Current.PathAndAliases {
		if alias == pathAndAlias.Alias {
			return path.Join(pathAndAlias.Path, aliasedPath[aliasIndex+1:]), nil
		}
	}

	return "", errors.New("Can't find alias '" + alias + "' from '" + aliasedPath + "'")
}

type FileBuffer struct {
	Buffer bytes.Buffer
	Index  int64
}

func NewFileBuffer() FileBuffer {
	return FileBuffer{}
}

func (fbuffer *FileBuffer) Bytes() []byte {
	return fbuffer.Buffer.Bytes()
}

func (fbuffer *FileBuffer) Read(p []byte) (int, error) {
	n, err := bytes.NewBuffer(fbuffer.Buffer.Bytes()[fbuffer.Index:]).Read(p)

	if err == nil {
		if fbuffer.Index+int64(len(p)) < int64(fbuffer.Buffer.Len()) {
			fbuffer.Index += int64(len(p))
		} else {
			fbuffer.Index = int64(fbuffer.Buffer.Len())
		}
	}

	return n, err
}

func (fbuffer *FileBuffer) Write(p []byte) (int, error) {
	n, err := fbuffer.Buffer.Write(p)

	if err == nil {
		fbuffer.Index = int64(fbuffer.Buffer.Len())
	}

	return n, err
}

func (fbuffer *FileBuffer) Seek(offset int64, whence int) (int64, error) {
	var err error
	var Index int64 = 0

	switch whence {
	case 0:
		if offset >= int64(fbuffer.Buffer.Len()) || offset < 0 {
			err = errors.New("Invalid Offset.")
		} else {
			fbuffer.Index = offset
			Index = offset
		}
	default:
		err = errors.New("Unsupported Seek Method")
	}

	return Index, err
}
