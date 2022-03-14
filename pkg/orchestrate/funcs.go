package orchestrate

import (
	hivev1 "github.com/openshift/hive/apis/hive/v1"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	runtimec "sigs.k8s.io/controller-runtime/pkg/client"
)

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

func (c ClusterClaimDeleteFlagSetNameFlagEmptyError) Error() string {
	return "--name flag set to an existing ClusterClaim required to perform a deletion."
}

func (c ClusterClaimNameLengthIncorrectError) Error() string {
	return "--name length is incorrect; must be at least 8 but not more than 16 ASCII alphanumeric characters."
}

func (c ClusterClaimNameHasInvalidCharactersError) Error() string {
	return "--name contains invalid characters; ASCII alphanumeric characters only permitted."
}
