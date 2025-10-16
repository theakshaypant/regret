package regret

// Version information for the regret library.
const (
	// Version is the current version of the library.
	Version = "0.1.0"

	// VersionMajor is the major version number.
	VersionMajor = 0

	// VersionMinor is the minor version number.
	VersionMinor = 1

	// VersionPatch is the patch version number.
	VersionPatch = 0

	// VersionPrerelease indicates this is a pre-release version.
	VersionPrerelease = "alpha"
)

// FullVersion returns the full version string including pre-release suffix.
func FullVersion() string {
	if VersionPrerelease != "" {
		return Version + "-" + VersionPrerelease
	}
	return Version
}
