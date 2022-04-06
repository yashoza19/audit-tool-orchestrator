package index

import (
	. "audit-tool-orchestrator/pkg"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

func DownloadImage(image string, containerEngine string) error {
	log.Infof("Downloading image %s to audit...", image)
	cmd := exec.Command(containerEngine, "pull", image)
	_, err := RunCommand(cmd)
	// if found an error try again
	// Sometimes it faces issues to download the image
	if err != nil {
		log.Warnf("error %s faced to downlad the image. Let's try more one time.", err)
		cmd := exec.Command(containerEngine, "pull", image)
		_, err = RunCommand(cmd)
	}
	return err
}

func ExtractIndexDB(image string, containerEngine string) error {
	log.Info("Extracting database...")
	// Remove image if exists already
	command := exec.Command(containerEngine, "rm", catalogIndex)
	_, _ = RunCommand(command)

	// Download the image
	command = exec.Command(containerEngine, "create", "--name", catalogIndex, image, "\"yes\"")
	_, err := RunCommand(command)
	if err != nil {
		return fmt.Errorf("unable to create container image %s : %s", image, err)
	}

	// Extract
	command = exec.Command(containerEngine, "cp", fmt.Sprintf("%s:/database/index.db", catalogIndex), "./output/")
	_, err = RunCommand(command)
	if err != nil {
		return fmt.Errorf("unable to extract the image for index.db %s : %s", image, err)
	}
	return nil
}

func BuildBundlesQuery() (string, error) {
	query := sq.Select("o.name, o.bundlepath").From(
		"operatorbundle o, channel c").Distinct()

	query = query.Where("c.head_operatorbundle_name == o.name")

	//query := sq.Select("o.name, o.bundlepath").From(
	//	"operatorbundle o")

	query.OrderBy("o.name")

	sql, _, err := query.ToSql()
	if err != nil {
		return "", fmt.Errorf("unable to create sql : %s", err)
	}
	return sql, nil
}

func NewBundle(bundleName, bundleImagePath string) *Bundle {
	bundle := Bundle{}
	bundle.Name = bundleName
	bundle.BundleImage = bundleImagePath
	return &bundle
}

func (b *BundleList) OutputList() error {
	b.fixPackageNameInconsistency()

	data, err := json.Marshal(b)
	if err != nil {
		return err
	}

	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, data, "", "\t")
	if err != nil {
		return err
	}

	path := "/tmp/bundlelist.json"

	_, err = ioutil.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return ioutil.WriteFile(path, prettyJSON.Bytes(), 0644)
}

// fix inconsistency in the index db
// some packages are empty then, we get them by looking for the bundles
// which are publish with the same registry path
func (b *BundleList) fixPackageNameInconsistency() {
	for _, bundle := range b.Bundles {
		if bundle.PackageName == "" {
			split := strings.Split(bundle.BundleImage, "/")
			nm := ""
			for _, v := range split {
				if strings.Contains(v, "@") {
					nm = strings.Split(v, "@")[0]
					break
				}
			}
			for _, refbundle := range b.Bundles {
				if strings.Contains(refbundle.BundleImage, nm) {
					bundle.PackageName = refbundle.PackageName
				}
			}
		}
	}
}
