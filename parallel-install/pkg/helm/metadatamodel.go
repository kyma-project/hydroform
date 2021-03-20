package helm

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var (
	kymaComponentPriority int64 = 0 //used to sequentially order the created Kyma component metadata
	mu                    sync.Mutex
)

//KymaComponentMetadataTemplate is used as template (factory) to create KymaComponentMetadata instances
type KymaComponentMetadataTemplate struct {
	Profile      string
	Version      string
	Prerequisite bool   //flag to mark prerequisite components
	Component    bool   //indicator flag to which is always set to 'true' (used in lookups)
	OperationID  string //unique ID used to distinguish versions with the same name
	CreationTime int64  //timestamp when the version was installed
	ready        bool   //indicates whether the the ForPrerequisites() or ForComponents() function was called
}

//ForPrerequisites creates a copy of the template instance which has to be used for prerequisite components
func (kmt *KymaComponentMetadataTemplate) ForPrerequisites() *KymaComponentMetadataTemplate {
	return kmt.clone(true)
}

//ForComponents creates a copy of the template instance which has to be used for components (non-prerequisites)
func (kmt *KymaComponentMetadataTemplate) ForComponents() *KymaComponentMetadataTemplate {
	return kmt.clone(false)
}

//clone creates a copy of the template instance
func (kmt *KymaComponentMetadataTemplate) clone(isPrerequisiteTemplate bool) *KymaComponentMetadataTemplate {
	return &KymaComponentMetadataTemplate{
		Profile:      kmt.Profile,
		Version:      kmt.Version,
		Component:    kmt.Component,
		OperationID:  kmt.OperationID,
		CreationTime: kmt.CreationTime,
		Prerequisite: isPrerequisiteTemplate,
		ready:        true,
	}
}

//Build creates a KymaComponentMetadata
func (kmt *KymaComponentMetadataTemplate) Build(namespace, name string) (*KymaComponentMetadata, error) {
	if !kmt.ready {
		return nil, fmt.Errorf("KymaComponentMetadataTemplate is not ready: call ForPrerequisite() or ForComponent()")
	}

	mu.Lock()
	kymaComponentPriority++
	mu.Unlock()
	compMeta := &KymaComponentMetadata{
		Profile:      kmt.Profile,
		Version:      kmt.Version,
		Component:    kmt.Component,
		OperationID:  kmt.OperationID,
		CreationTime: kmt.CreationTime,
		Name:         name,
		Namespace:    namespace,
		Priority:     kymaComponentPriority,
		Prerequisite: kmt.Prerequisite,
	}
	if err := compMeta.isValid(); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Kyma component '%s' is invalid", compMeta))
	}
	return compMeta, nil
}

//KymaComponentMetadata stores metadata of a component. These metadata fields are persisted as labels of a Kyma component Helm secret.
type KymaComponentMetadata struct {
	Profile      string
	Version      string
	Component    bool //indicator flag to which is always set to 'true' (used in lookups)
	OperationID  string
	CreationTime int64
	Name         string
	Namespace    string
	Priority     int64
	Prerequisite bool
}

//isValid verifies the completeness of a metadata instance
func (km *KymaComponentMetadata) isValid() error {
	//check whether all mandatory fields are defined
	if km.Version == "" {
		return fmt.Errorf("Version is missing")
	}
	if !km.Component {
		return fmt.Errorf("Component flag not set to true")
	}
	if km.OperationID == "" {
		return fmt.Errorf("Operation ID is empty")
	}
	if km.CreationTime == 0 {
		return fmt.Errorf("Creation time is 0")
	}
	if km.Priority == 0 {
		return fmt.Errorf("Priority is 0")
	}
	if km.Name == "" {
		return fmt.Errorf("Name is missing")
	}
	if km.Namespace == "" {
		return fmt.Errorf("Namespace is missing")
	}
	return nil
}

func (km *KymaComponentMetadata) String() string {
	return fmt.Sprintf("%s:%s:%d", km.Namespace, km.Name, km.Priority)
}

//KymaVersionSet bundles multiple Kyma versions
type KymaVersionSet struct {
	Versions []*KymaVersion
}

//Count counts the different Kyma versions
func (kvs *KymaVersionSet) Count() int {
	return len(kvs.Versions)
}

//Names returns the name of all Kyma versions
func (kvs *KymaVersionSet) Names() []string {
	names := []string{}
	for _, version := range kvs.Versions {
		names = append(names, version.Version)
	}
	return names
}

//InstalledComponents returns a list of all components sorted by their installation sequence
func (kvs *KymaVersionSet) InstalledComponents() []*KymaComponentMetadata {
	var comps []*KymaComponentMetadata
	for _, version := range kvs.Versions {
		comps = append(comps, version.Components...)
	}
	return sortComponents(comps)
}

func (kvs KymaVersionSet) String() string {
	return strings.Join(kvs.Names(), ", ")
}

//Empty verifies whether a Kyma version was found
func (kvs *KymaVersionSet) Empty() bool {
	return kvs.Count() == 0
}

//KymaVersion stores metadata of an installed Kyma version
type KymaVersion struct {
	Version      string
	Profile      string
	OperationID  string
	CreationTime int64
	Components   []*KymaComponentMetadata
}

//InstalledComponents returns a list of all components in the version sorted by their installation order
func (v *KymaVersion) InstalledComponents() []*KymaComponentMetadata {
	return sortComponents(v.Components)
}

//ComponentNames returns the names of the installed components
func (v *KymaVersion) ComponentNames() []string {
	result := []string{}
	for _, comp := range v.InstalledComponents() {
		result = append(result, comp.Name)
	}
	return result
}

func (v *KymaVersion) String() string {
	return fmt.Sprintf("%s:%s(%d)", v.Version, v.OperationID, v.CreationTime)
}

//NewKymaComponentMetadataTemplate creates a new KymaComponentMetadataTemplate
func NewKymaComponentMetadataTemplate(version, profile string) *KymaComponentMetadataTemplate {
	return &KymaComponentMetadataTemplate{
		Profile:      profile,
		Version:      version,
		Component:    true, //flag will always be set for any Kyma component
		OperationID:  uuid.New().String(),
		CreationTime: time.Now().Unix(),
	}
}

//sortComponents is a function used for sorting installed components (first prerequisites followed by their installation sequence)
func sortComponents(comps []*KymaComponentMetadata) []*KymaComponentMetadata {
	sort.Slice(comps, func(i, j int) bool {
		prio1 := comps[i].Priority
		if comps[i].Prerequisite { //boost if pre-requisite
			prio1 -= 100000000
		}
		prio2 := comps[j].Priority
		if comps[j].Prerequisite { //boost if pre-requisite
			prio2 -= 100000000
		}
		return prio1 < prio2
	})
	return comps
}
