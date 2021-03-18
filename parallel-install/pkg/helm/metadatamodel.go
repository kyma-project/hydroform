package helm

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	kymaMetadataCount int64 = time.Now().Unix()
	mu                sync.Mutex
)

type KymaMetadata struct {
	Profile      string
	Version      string
	Component    bool //indicator flag to which is always set to 'true' (used in lookups)
	OperationID  string
	CreationTime int64
	Counter      int64 //count metaData usage
}

func (km *KymaMetadata) isValid() bool {
	//check whether all mandatory fields are defined
	return km.Version != "" && km.Component && km.OperationID != "" && km.CreationTime > 0
}

func (km *KymaMetadata) increment() *KymaMetadata {
	mu.Lock()
	kymaMetadataCount++
	km.Counter = kymaMetadataCount
	mu.Unlock()
	return km
}

type KymaVersionSet struct {
	Versions []*KymaVersion
}

func (kvs *KymaVersionSet) Count() int {
	return len(kvs.Versions)
}

func (kvs *KymaVersionSet) Names() []string {
	names := []string{}
	for _, version := range kvs.Versions {
		names = append(names, version.Version)
	}
	return names
}

func (kvs KymaVersionSet) String() string {
	return strings.Join(kvs.Names(), ", ")
}

func (kvs *KymaVersionSet) Empty() bool {
	return kvs.Count() == 0
}

type KymaVersion struct {
	Version      string
	Profile      string
	OperationID  string
	CreationTime int64
	components   []*KymaComponent
}

//Components returns all components of this version in increasing priority order (first installed to the latest installed components)
func (v *KymaVersion) Components() []*KymaComponent {
	sort.Slice(v.components, func(i, j int) bool { return v.components[i].Priority < v.components[j].Priority })
	return v.components
}

func (v *KymaVersion) ComponentNames() []string {
	result := []string{}
	for _, comp := range v.Components() {
		result = append(result, comp.Name)
	}
	return result
}

func (v *KymaVersion) String() string {
	return fmt.Sprintf("%s:%s(%d)", v.Version, v.OperationID, v.CreationTime)
}

type KymaComponent struct {
	Name      string
	Namespace string
	Priority  int64
}

func NewKymaMetadata(version, profile string) *KymaMetadata {
	return &KymaMetadata{
		Profile:      profile,
		Version:      version,
		Component:    true, //flag will always be set for any Kyma component
		OperationID:  uuid.New().String(),
		CreationTime: time.Now().Unix(),
	}
}
