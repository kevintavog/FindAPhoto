package files

import (
	"bytes"
	"errors"
	"net/url"
	"path"
	"strings"

	"github.com/kevintavog/findaphoto/common"
	"github.com/labstack/echo"
)

func ConfigureRouting(e *echo.Echo) {
	files := e.Group("/files")
	files.GET("/thumbs/*", thumbFiles)
	files.GET("/slides/*", slideFiles)
	files.GET("/media/*", mediaFiles)
}

func toRepositoryId(itemUrl string) (string, error) {
	return url.QueryUnescape(strings.Replace(itemUrl, "/", "\\", -1))
}

func aliasedToFullPath(aliasedPath string) (string, error) {
	aliasIndex := strings.Index(aliasedPath, "/")
	if aliasIndex < 1 {
		return "", errors.New("Can't find alias token: '" + aliasedPath + "'")
	}

	alias := aliasedPath[0:aliasIndex]
	basePath, err := common.PathForAlias(alias)
	if err != nil {
		return "", errors.New("Can't find alias '" + alias + "'; " + err.Error())
	}

	unescapedPath, err := url.QueryUnescape(aliasedPath[aliasIndex+1:])
	if err != nil {
		return "", errors.New("Badly escaped alias '" + aliasedPath + "'; " + err.Error())
	}
	return path.Join(basePath, unescapedPath), nil
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
