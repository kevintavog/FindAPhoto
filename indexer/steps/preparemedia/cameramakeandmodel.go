package preparemedia

import (
	//	"fmt"
	"strings"

	"github.com/kevintavog/findaphoto/common"
)

var cameraMakeSubstitution = map[string]string{
	"CASIO COMPUTER CO.,LTD":  "Casio",
	"EASTMAN KODAK COMPANY":   "Kodak",
	"FUJIFILM":                "Fuji",
	"Minolta Co., Ltd.":       "Minolta",
	"NIKON":                   "Nikon",
	"NIKON CORPORATION":       "Nikon",
	"OLYMPUS IMAGING CORP.":   "Olympus",
	"OLYMPUS OPTICAL CO.,LTD": "Olympus",
	"SAMSUNG":                 "Samsung",
	"SONY":                    "Sony",
	"TOSHIBA":                 "Toshiba",
}

// Camera manufacturers are both inconsistent with each other and, at times, inconsistent
// with themselves. Try to make it better.
func populateCameraMakeAndModel(media *common.Media, candidate *common.CandidateFile) {

	cameraMake := candidate.Exif.EXIF.Make
	if len(cameraMake) < 1 {
		cameraMake = candidate.Exif.XMP.Make // Videos seem to have make & model in XMP, if at all
	}

	cameraModel := candidate.Exif.EXIF.Model
	if len(cameraModel) < 1 {
		cameraModel = candidate.Exif.XMP.Model
	}

	media.OriginalCameraMake = cameraMake
	media.OriginalCameraModel = cameraModel

	if override, ok := cameraMakeSubstitution[cameraMake]; ok {
		cameraMake = override
	}

	// Special handling per make: Remove the make from the model name, proper case a few names.
	switch cameraMake {
	case "Canon":
		cameraModel = strings.Replace(cameraModel, "Canon ", "", 1)
		cameraModel = strings.Replace(cameraModel, "DIGITAL ", "", 1)
		cameraModel = strings.Replace(cameraModel, "REBEL", "Rebel", 1)

	case "Kodak":
		cameraModel = strings.Replace(cameraModel, "KODAK ", "", 1)
		cameraModel = strings.Replace(cameraModel, "DIGITAL CAMERA", "", 1)
		cameraModel = strings.Replace(cameraModel, "EASYSHARE", "Easyshare", 1)
		cameraModel = strings.Replace(cameraModel, "ZOOM", "Zoom", 1)

	case "Nikon":
		cameraModel = strings.Replace(cameraModel, "NIKON ", "", 1)
	}

	media.CameraMake = strings.Trim(cameraMake, " ")
	media.CameraModel = strings.Trim(cameraModel, " ")
}
