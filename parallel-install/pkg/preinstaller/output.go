package preinstaller

type File struct {
	component string
	path      string
}

type Output struct {
	installed    []File
	notInstalled []File
}
