package version

var (
	Version, GitCommit, GitState, BuildDate string
)

func VersionInfoToSlice() []interface{} {
	s := []interface{}{
		"version", Version,
		"commit", GitCommit,
		"build", BuildDate,
	}

	return s
}
