package bundles

import (
	"audit-tool-orchestrator/pkg"
	"audit-tool-orchestrator/pkg/index"
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var flags = index.BundleFlags{}

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "bundles",
		Short:   "",
		Long:    "",
		PreRunE: validation,
		RunE:    run,
	}

	cmd.Flags().StringVar(&flags.IndexImage, "index-image", "",
		"index image and tag which will be audit")
	if err := cmd.MarkFlagRequired("index-image"); err != nil {
		log.Fatalf("Failed to mark `index-image` flag for `index` sub-command as required")
	}

	cmd.Flags().StringVar(&flags.OutputPath, "output-path", "",
		"inform the path of the directory to output the report.")

	cmd.Flags().StringVar(&flags.ContainerEngine, "container-engine", pkg.Docker,
		fmt.Sprintf("specifies the container tool to use. If not set, the default value is docker. "+
			"Note that you can use the environment variable CONTAINER_ENGINE to inform this option. "+
			"[Options: %s and %s]", pkg.Docker, pkg.Podman))

	return cmd
}

func validation(cmd *cobra.Command, args []string) error {
	if len(flags.OutputPath) > 0 {
		if _, err := os.Stat(flags.OutputPath); os.IsNotExist(err) {
			return err
		}
	}

	if len(flags.ContainerEngine) == 0 {
		flags.ContainerEngine = pkg.GetContainerToolFromEnvVar()
	}

	if flags.ContainerEngine != pkg.Docker && flags.ContainerEngine != pkg.Podman {
		return fmt.Errorf("invalid value for the flag --container-engine (%s)."+
			" The valid options are %s and %s", flags.ContainerEngine, pkg.Docker, pkg.Podman)
	}

	return nil
}

func run(cmd *cobra.Command, args []string) error {
	pkg.CleanupTemporaryDirs()
	pkg.GenerateTemporaryDirs()

	if err := index.DownloadImage(flags.IndexImage, flags.ContainerEngine); err != nil {
		return err
	}

	if err := index.ExtractIndexDB(flags.IndexImage, flags.ContainerEngine); err != nil {
		return err
	}

	bundlelist := index.BundleList{}
	bundlelist, err := getDataFromIndexDB(bundlelist)
	if err != nil {
		return err
	}

	if err := bundlelist.OutputList(); err != nil {
		return err
	}

	pkg.CleanupTemporaryDirs()

	return nil
}

func getDataFromIndexDB(data index.BundleList) (index.BundleList, error) {
	// Connect to the database
	db, err := sql.Open("sqlite3", "/tmp/ato/output/index.db")
	if err != nil {
		return data, fmt.Errorf("unable to connect in to the database : %s", err)
	}

	query, err := index.BuildBundlesQuery()
	if err != nil {
		return data, err
	}

	row, err := db.Query(query)
	if err != nil {
		return data, fmt.Errorf("unable to query the index db : %s", err)
	}

	defer row.Close()
	for row.Next() {
		var bundleName string
		var bundlePath string

		err = row.Scan(&bundleName, &bundlePath)
		if err != nil {
			log.Errorf("unable to scan data from index %s\n", err.Error())
		}
		log.Infof("Generating data from the bundle (%s)", bundleName)
		bundle := index.NewBundle(bundleName, bundlePath)

		query = fmt.Sprintf("SELECT c.channel_name, c.package_name FROM channel_entry c "+
			"where c.operatorbundle_name = '%s'", bundle.Name)
		row, err := db.Query(query)
		if err != nil {
			return data, fmt.Errorf("unable to query channel entry in the index db : %s", err)
		}

		defer row.Close()
		var channelName string
		var packageName string
		for row.Next() { // Iterate and fetch the records from result cursor
			_ = row.Scan(&channelName, &packageName)
			bundle.Channels = append(bundle.Channels, channelName)
			bundle.PackageName = packageName
		}

		query = fmt.Sprintf("SELECT default_channel FROM package WHERE name = '%s'", bundle.PackageName)
		row, err = db.Query(query)
		if err != nil {
			return data, fmt.Errorf("unable to query default channel entry in the index db : %s", err)
		}

		defer row.Close()
		var defaultChannelName string
		for row.Next() { // Iterate and fetch the records from result cursor
			_ = row.Scan(&defaultChannelName)
			bundle.DefaultChannel = defaultChannelName
		}

		defer row.Close()

		data.Bundles = append(data.Bundles, *bundle)
	}

	return data, nil
}
