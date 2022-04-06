package pkg

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
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
	command := exec.Command("rm", "-rf", "/tmp/ato")
	_, _ = RunCommand(command)

	command = exec.Command("rm", "-rf", "/tmp/ato/output")
	_, _ = RunCommand(command)

	command = exec.Command("mkdir", "/tmp/ato/output")
	_, err := RunCommand(command)
	if err != nil {
		log.Fatal(err)
	}

	command = exec.Command("mkdir", "/tmp/ato")
	_, err = RunCommand(command)
	if err != nil {
		log.Fatal(err)
	}
}

func CleanupTemporaryDirs() {
	command := exec.Command("rm", "-rf", "/tmp/ato")
	_, _ = RunCommand(command)

	command = exec.Command("rm", "-rf", "/tmp/ato/output")
	_, _ = RunCommand(command)
}

// GetContainerToolFromEnvVar retrieves the value of the environment variable and defaults to docker when not set
func GetContainerToolFromEnvVar() string {
	if value, ok := os.LookupEnv("CONTAINER_ENGINE"); ok {
		return value
	}
	return DefaultContainerTool
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
