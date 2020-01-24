package terraform

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// TODO remove this file when the gardener provider is on the official terraform registry

const (
	providerURL     = "https://github.com/kyma-incubator/terraform-provider-gardener/releases/download/%s/terraform-provider-gardener-%s-%s"
	providerName    = "terraform-provider-gardener"
	providerVersion = "v0.0.5"
)

// initGardenerProvider will check if the gardener provider is available and download it if not.

func initGardenerProvider() error {
	pluginDirs, err := globalPluginDirs()
	if err != nil {
		return err
	}
	providerPath := filepath.Join(pluginDirs[1], fmt.Sprintf("%s_%s", providerName, providerVersion))

	// check if plugin is in the plugins dir
	if _, err := os.Stat(providerPath); !os.IsNotExist(err) {
		if runtime.GOOS == "windows" {
			err = generateWindowsBinary(providerPath)
			if err != nil {
				return err
			}
		}
		return nil
	}

	// Download the plugin for the OS and arch
	r, err := downloadBinary(fmt.Sprintf(providerURL, providerVersion, runtime.GOOS, runtime.GOARCH))
	if err != nil {
		return err
	}
	defer r.Close()

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	// save the file
	if _, err := os.Stat(pluginDirs[1]); os.IsNotExist(err) {
		err = os.MkdirAll(pluginDirs[1], 0700)
		if err != nil {
			return err
		}
	}
	if err := ioutil.WriteFile(providerPath, data, 0700); err != nil {
		return err
	}

	if runtime.GOOS == "windows" {
		// Create exe for windows if it doesn't exist
		err = generateWindowsBinary(providerPath)
		if err != nil {
			return err
		}
	}

	// if just downloaded a new version successfully, delete any old ones
	err = filepath.Walk(pluginDirs[1], func(path string, info os.FileInfo, err error) error {

		if err != nil {
			return err
		}

		if strings.HasPrefix(info.Name(), providerName) && !strings.HasSuffix(info.Name(), providerVersion) {
			return os.Remove(path)
		}
		return nil
	})
	return err
}

func generateWindowsBinary(providerPath string) error {
	windowsProviderPath := providerPath + ".exe"
	if _, err := os.Stat(windowsProviderPath); os.IsNotExist(err) {
		providerFile, err := ioutil.ReadFile(providerPath)
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(windowsProviderPath, providerFile, 0700)
		if err != nil {
			return err
		}
	}
	return nil
}

func downloadBinary(url string) (io.ReadCloser, error) {
	c := &http.Client{
		Timeout: 1 * time.Minute,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/octet-stream")

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
