package pkg

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	hivev1 "github.com/openshift/hive/apis/hive/v1"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"os/exec"
	"sigs.k8s.io/controller-runtime/pkg/client"
	runtimec "sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

// Run executes the provided command within this context
func RunCommand(cmd *exec.Cmd) ([]byte, error) {
	command := strings.Join(cmd.Args, " ")
	log.Infof("running: %s\n", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return output, fmt.Errorf("%s failed with error: (%v) %s", command, err, string(output))
	}
	if len(output) > 0 {
		log.Debugf("command output :%s", output)
	}
	return output, nil
}

func GenerateTemporaryDirs() {
	command := exec.Command("rm", "-rf", "tmp")
	_, _ = RunCommand(command)

	command = exec.Command("rm", "-rf", "./output/")
	_, _ = RunCommand(command)

	command = exec.Command("mkdir", "./output/")
	_, err := RunCommand(command)
	if err != nil {
		log.Fatal(err)
	}

	command = exec.Command("mkdir", "tmp")
	_, err = RunCommand(command)
	if err != nil {
		log.Fatal(err)
	}
}

func CleanupTemporaryDirs() {
	command := exec.Command("rm", "-rf", "tmp")
	_, _ = RunCommand(command)

	command = exec.Command("rm", "-rf", "./output/")
	_, _ = RunCommand(command)
}

// GetContainerToolFromEnvVar retrieves the value of the environment variable and defaults to docker when not set
func GetContainerToolFromEnvVar() string {
	if value, ok := os.LookupEnv("CONTAINER_ENGINE"); ok {
		return value
	}
	return DefaultContainerTool
}

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
		"operatorbundle o")

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

	path := "bundlelist.json"

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

/*func (b *BundleList) PrepareList() Report {
	b.fixPackageNameInconsistency()

	var allColumns []Column
	for _, v := range b.AuditBundle {
		col := NewColumn(v)

		// do not add bundle which has not the label
		if len(b.Flags.Label) > 0 && !v.FoundLabel {
			continue
		}

		allColumns = append(allColumns, *col)
	}

	sort.Slice(allColumns[:], func(i, j int) bool {
		return allColumns[i].PackageName < allColumns[j].PackageName
	})

	finalReport := Report{}
	finalReport.Flags = b.Flags
	finalReport.Columns = allColumns
	finalReport.IndexImageInspect = b.IndexImageInspect

	dt := time.Now().Format("2006-01-02")
	finalReport.GenerateAt = dt

	if len(allColumns) == 0 {
		log.Fatal("No data was found for the criteria informed. " +
			"Please, ensure that you provide valid information.")
	}

	return finalReport
}*/

/*func (b *BundleList) writeJSON() error {
	data, err := json.Marshal(b)
	if err != nil {
		return err
	}

	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, data, "", "\t")
	if err != nil {
		return err
	}

	path := filepath.Join(outputPath, GetReportName(imageName, typeName, "json"))

	_, err = ioutil.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return ioutil.WriteFile(path, prettyJSON.Bytes(), 0644)
}*/

func GetHiveClient() runtimec.Client {
	// create hive client
	cfg, err := clientcmd.BuildConfigFromFlags("", os.Getenv("OPENSHIFT_KUBECONFIG"))
	if err != nil {
		log.Printf("Unable to build config from flags: %v\n", err)
	}

	nrs := runtime.NewScheme()
	err = hivev1.AddToScheme(nrs)
	if err != nil {
		log.Printf("Unable to add Hive scheme to client: %v\n", err)
	}

	hiveclient, err := runtimec.New(cfg, client.Options{Scheme: nrs})

	return hiveclient
}
