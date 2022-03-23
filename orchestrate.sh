#!/usr/bin/env bash

# TODO: maybe we can run through 30 every hour as a default with 45 minute wait for ready and 15 minutes to run?

# Get the bundlelist.json
# TODO: index-image and container-engine should be dynamic
audit-tool-orchestrator index bundles --index-image registry.redhat.io/redhat/certified-operator-index:v4.9 --container-engine podman

count=`jq '.Bundles | length' bundlelist.json`
batch=$((count / 10))
lump=$((count % 10))

# Create the ClusterPool
# TODO: we will use size of 30 and running 10 for now with other hardcoded values for flags
audit-tool-orchestrator orchestrate pool

# Ensure the ClusterPool has at least 10 ready; wait at least 45 minutes for the first 10 clusters

# Run the audit
## Create a ClusterClaim
# TODO: --name should be set from the bundle.PackageName
audit-tool-orchestrator orchestrate claim --name cloud-native-postgresql
## Ensure the cluster claimed can be accessed
## Copy kubeconfig from Hive into default namespace of claimed cluster
## Create image pull secret secret in the default namespace of the claimed cluster
## Run the audit-tool against claimed cluster for current operator
#### Put the audit-tool job in the claimed cluster; default namespace

# TODO:
#### store result of audit-tool job in the