package generatethumbnail

import (
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/kevintavog/findaphoto/common"

	"github.com/disintegration/imaging"
	"github.com/ian-kent/go-log/log"
	"github.com/twinj/uuid"
)

var GeneratedImage int64
var FailedImage int64
var GeneratedVideo int64
var FailedVideo int64

type ThumbnailInfo struct {
	FullPath    string
	AliasedPath string
	MimeType    string
}

const numConsumers = 8
const thumbnailMaxHeightDimension = 170

var queue = make(chan *ThumbnailInfo, numConsumers)
var waitGroup sync.WaitGroup

func Start() {
	waitGroup.Add(numConsumers)
	for idx := 0; idx < numConsumers; idx++ {
		go func() {
			dequeue()
			waitGroup.Done()
		}()
	}
}

func Done() {
	close(queue)
}

func Wait() {
	waitGroup.Wait()
}

func Enqueue(fullPath, aliasedPath, mimeType string) {
	thumbnailInfo := &ThumbnailInfo{
		FullPath:    fullPath,
		AliasedPath: aliasedPath,
		MimeType:    mimeType,
	}
	queue <- thumbnailInfo
}

func dequeue() {
	for thumbnailInfo := range queue {
		var mediaType = strings.Split(thumbnailInfo.MimeType, "/")
		thumbPath := common.ToThumbPath(thumbnailInfo.AliasedPath)
		if len(mediaType) < 1 {
			log.Error("Invalid media type: '%s' for %s", thumbnailInfo.MimeType, thumbnailInfo.FullPath)
			continue
		}

		err := common.CreateDirectory(path.Dir(thumbPath))
		if err != nil {
			log.Error("Unable to create directory for '%s'", thumbPath)
			continue
		}

		switch strings.ToLower(mediaType[0]) {
		case "video":
			generateVideo(thumbnailInfo.FullPath, thumbPath)
		case "image":
			generateImage(thumbnailInfo.FullPath, thumbPath)
		default:
			log.Error("Unhandled mediaType: %s (%s) for %s", thumbnailInfo.MimeType, mediaType, thumbnailInfo.FullPath)
		}
	}
}

func generateImage(fullPath, thumbPath string) {
	if resize(fullPath, thumbPath) != nil {
		atomic.AddInt64(&FailedImage, 1)
	} else {
		atomic.AddInt64(&GeneratedImage, 1)
	}
}

func generateVideo(fullPath, thumbPath string) {
	tmpFilename := path.Join(os.TempDir(), "findAPhoto", "thumbnails", uuid.NewV4().String()+".JPG")
	defer os.Remove(tmpFilename)

	err := common.CreateDirectory(path.Dir(tmpFilename))
	if err != nil {
		log.Fatal("Unable to create temporary directory for thumbnail generation (%s): %s", tmpFilename, err.Error())
	}

	out, err := exec.Command(common.FfmpegPath, "-i", fullPath, "-ss", "00:00:01.0", "-vframes", "1", tmpFilename).Output()
	if err != nil {
		atomic.AddInt64(&FailedVideo, 1)
		log.Fatal("Failed executing ffmpeg for '%s': %s (%s)", fullPath, err.Error(), out)
	}

	if exists, _ := common.PathExists(tmpFilename); !exists {
		// The video may not be long enough to grab a frame at the 1 second point...
		out, err = exec.Command(common.FfmpegPath, "-i", fullPath, "-ss", "00:00:00.0", "-vframes", "1", tmpFilename).Output()
		if err != nil {
			atomic.AddInt64(&FailedVideo, 1)
			log.Error("Failed executing ffmpeg for '%s': %s (%s)", fullPath, err.Error(), out)
		}
	}

	if resize(tmpFilename, thumbPath) != nil {
		atomic.AddInt64(&FailedVideo, 1)
	} else {
		atomic.AddInt64(&GeneratedVideo, 1)
	}
}

func resize(imageFilename, thumbFilename string) error {
	image, err := imaging.Open(imageFilename)
	if err != nil {
		log.Error("Unable to open '%s': %s", imageFilename, err.Error())
		return err
	}

	thumbImage := imaging.Resize(image, 0, thumbnailMaxHeightDimension, imaging.Lanczos)
	err = imaging.Save(thumbImage, thumbFilename)
	if err != nil {
		log.Error("Unable to save to '%s': %s", thumbFilename, err.Error())
		return err
	}

	return nil
}
