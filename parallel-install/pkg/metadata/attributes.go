package metadata

import (
	"io/ioutil"
	"path/filepath"
)

//Attributes represents common metadata attributes
type Attributes struct {
	profile           string
	version           string
	componentListData []byte
	componentListFile string
}

//NewAttributes create a new attributes entity
func NewAttributes(version string, profile string, componentListFile string) (*Attributes, error) {
	compListData, err := ioutil.ReadFile(componentListFile)
	if err != nil {
		return nil, err
	}
	return &Attributes{
		profile:           profile,
		version:           version,
		componentListData: compListData,
		componentListFile: filepath.Base(componentListFile),
	}, nil
}
