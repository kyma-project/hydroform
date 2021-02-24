package preinstaller

// File consists of a path to the file that was a part of PreInstaller installation
// and a component name that it belongs to.
type File struct {
	component string
	path      string
}

// Output contains lists of installed and not installed files during PreInstaller installation.
type Output struct {
	installed    []File
	notInstalled []File
}
