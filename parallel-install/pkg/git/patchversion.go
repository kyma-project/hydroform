package git

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type tagStruct struct {
	Tag        string `json:"tag_name"`
	IsPrelease bool   `json:"prerelease"`
}

func getDataBytes() ([]byte, error) {
	const url = "https://api.github.com/repos/kyma-project/kyma/releases"
	resp, err := http.Get(url)
	if err != nil {
		return []byte{}, fmt.Errorf("GET error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []byte{}, fmt.Errorf("status error: %v", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, fmt.Errorf("read body: %v", err)
	}

	return data, nil
}

func updatePatchVersion(version string, patchVer int) string {
	verArray := strings.Split(version, ".")
	return fmt.Sprintf("%s.%s.%d", verArray[0], verArray[1], patchVer)
}

func getPatchVersion(version string) int {
	verArray := strings.Split(version, ".")
	re := regexp.MustCompile(`[-]?\d[\d,]*[\.]?[\d{2}]*`)
	patchString := re.FindAllString(verArray[2], 1)[0]
	patchVer, _ := strconv.Atoi(patchString)
	return patchVer
}

func getMajorVersion(version string) string {
	verArray := strings.Split(version, ".")
	return fmt.Sprintf("%s.%s", verArray[0], verArray[1])
}

func findLatestPatchVersion(version string, versions []tagStruct) string {
	currPatchVer := getPatchVersion(version)
	majorVer := getMajorVersion(version)
	for _, ver := range versions {
		if strings.Contains(ver.Tag, majorVer) && !ver.IsPrelease {
			loopPatchVer := getPatchVersion(ver.Tag)
			if loopPatchVer > currPatchVer {
				currPatchVer = loopPatchVer
			}
		}
	}
	return updatePatchVersion(version, currPatchVer)
}

func getReleaseTags() ([]tagStruct, error) {
	jsonBytes, err := getDataBytes()
	if err != nil {
		log.Printf("Failed to get JSON: %v", err)
		return make([]tagStruct, 0), err // skip patch update
	}
	v := []tagStruct{}
	err = json.Unmarshal(jsonBytes, &v)
	return v, err
}

func SetToLatestPatchVersion(version string) string {
	versions, _ := getReleaseTags()
	return findLatestPatchVersion(version, versions)
}
