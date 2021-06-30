package debug

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const EnvVarManifestDumper = "INSTALLER_EXPORT_MANIFEST"

type ManifestDumper interface {
	DumpHelmRelease(rel *release.Release) error
	DumpUnstructuredResource(component string, resource *unstructured.Unstructured) error
}

type DefaultManifestDumper struct {
	exportDir string
}

func NewManifestDumper() ManifestDumper {
	exportDir := os.Getenv(EnvVarManifestDumper)
	if stats, err := os.Stat(exportDir); !os.IsNotExist(err) && stats.IsDir() {
		//export dir exists, return an operative dumper
		return &DefaultManifestDumper{
			exportDir: exportDir,
		}
	}
	return &NoopManifestDumper{}
}

func (md *DefaultManifestDumper) DumpHelmRelease(rel *release.Release) error {
	return md.writeFile(fmt.Sprintf("%s.yaml", rel.Name), []byte(rel.Manifest))
}

func (md *DefaultManifestDumper) DumpUnstructuredResource(filename string, resource *unstructured.Unstructured) error {
	manifest, err := yaml.Marshal(resource.Object)
	if err != nil {
		return err
	}
	return md.writeFile(filename, manifest)
}

func (md *DefaultManifestDumper) writeFile(filename string, data []byte) error {
	return ioutil.WriteFile(
		path.Join(md.exportDir, filename),
		data, 0600)
}

type NoopManifestDumper struct {
}

func (md *NoopManifestDumper) DumpHelmRelease(rel *release.Release) error {
	return nil
}

func (md *NoopManifestDumper) DumpUnstructuredResource(component string, resource *unstructured.Unstructured) error {
	return nil
}
