# Create kind KCP if not already present
./tools/dev/kind/create-kcp.sh

# Ensure the e2e-config.yaml is created correctly
cat ./e2e-config.yaml

# Download credentials from Gardener secrets
go run ./e2e/cmd credentials download
# Verify if credentials exists
ls -la ./tmp


# Run SIM in new console window
export KUBECONFIG=./tools/dev/kind/kubeconfig-kcp.yaml
go run ./e2e/cmd sim run


# Run Cloud Manager in new console window
export CONFIG_DIR=$(PWD)/tmp
export KUBECONFIG=./tools/dev/kind/kubeconfig-kcp.yaml
export LANDSCAPE=dev
export PEERING_NETWORK_TAG=e2e
export FEATURE_FLAG_CONFIG_FILE=./pkg/feature/ff_edge.yaml
go run ./cmd

# Create shared instance in new console window
export KUBECONFIG=./tools/dev/kind/kubeconfig-kcp.yaml
export PROVIDER=aws
go run ./e2e/cmd instance create -a shared-$PROVIDER -p $PROVIDER
go run ./e2e/cmd instance modules add -m  cloud-manager -a shared-$PROVIDER

# Dump kubeconfig for shared instance
go run ./e2e/cmd instance cre dump -r 9ce93c49-3900-4131-b747-a666f9a75a38 > ./tmp/debug-kubeconfig.yaml
go run ./e2e/cmd instance cre dump -a shared-$PROVIDER > ./tmp/kubeconfig.yaml

# Verify if Cloud Manager installed CRDs
export KUBECONFIG=./tmp/kubeconfig.yaml
kubectl get crds | grep cloud-resources

# Run E2E tests
export KUBECONFIG=./tools/dev/kind/kubeconfig-kcp.yaml
export PROJECTROOT=$(PWD)
export RUN_E2E_TESTS=1
go test ./e2e/tests/skr/all/... -timeout 0 -v -godog.tags "@skr && @$PROVIDER"
