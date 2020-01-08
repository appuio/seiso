package git

// SortTagBy type for sort order
type SortTagBy string

const(
	// Version sorts by version
	Version SortTagBy = "version"
	// Alphabetic sorts in alphabetical order
	Alphabetic SortTagBy = "alphabetic"
)

func (c SortTagBy) String() string {
	return string(c)
}

// IsValidSortValue function tries to cast the string to SortedTagBy type
func IsValidSortValue(sortValue string) bool {
	switch sortValue {
	case Version.String(), Alphabetic.String():
		return true
	}
	return false
}