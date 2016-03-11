package common

import (
	"path"
	"strings"
)

func ToThumbPath(aliasedPath string) string {
	thumbPath := path.Join(ThumbnailDirectory, strings.Replace(aliasedPath, "\\", "/", -1))
	if strings.ToUpper(path.Ext(thumbPath)) != ".JPG" {
		thumbPath += ".JPG"
	}
	return thumbPath
}
