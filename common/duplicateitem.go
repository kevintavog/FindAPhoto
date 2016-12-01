package common

const DuplicateTypeName = "duplicate"

type DuplicateItem struct {
	IgnoredPath  string `json:"ignoredpath"`
	ExistingPath string `json:"existingpath"`
}
