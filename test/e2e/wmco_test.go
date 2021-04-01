package e2e

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	nodeCreationTime     = time.Minute * 30
	nodeRetryInterval    = time.Minute * 1
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 5
	// deploymentRetries is the amount of time to retry creating a Windows Server deployment, to compensate for the
	// time it takes to download the Server2019 image to the node
	deploymentRetries = 10
)

// TestWMCO sets up the testing suite for WMCO.
func TestWMCO(t *testing.T) {
	gc.numberOfNodes = int32(numberOfNodes)
	require.NotEmpty(t, privateKeyPath, "private-key-path is not set")
	gc.privateKeyPath = privateKeyPath
	// When the OPENSHIFT_CI env var is set to true, the test is running within CI
	if inCI := os.Getenv("OPENSHIFT_CI"); inCI == "true" {
		// In the CI container the WMCO binary will be found here
		wmcoPath = "/usr/local/bin/windows-machine-config-operator"
	}

	testCtx, err := NewTestContext()
	require.NoError(t, err)
	require.NoError(t, testCtx.ensureNamespace(testCtx.workloadNamespace), "error creating test namespace")

	// When the upgrade test is run from CI, the namespace that gets created does not have the required monitoring
	// label, so we apply that here.
	require.NoError(t, testCtx.applyMonitoringLabelToOperatorNamespace(), "error applying monitoring label")

	// test that the operator can deploy without the secret already created, we can later use a secret created by the
	// individual test suites after the operator is running
	t.Run("operator deployed without private key secret", testOperatorDeployed)
	t.Run("create", creationTestSuite)
	t.Run("network", testNetwork)
	t.Run("upgrade", upgradeTestSuite)
	t.Run("reconfigure", reconfigurationTest)
	t.Run("destroy", deletionTestSuite)
}

// testOperatorDeployed tests that the operator pod is running
func testOperatorDeployed(t *testing.T) {
	testCtx, err := NewTestContext()
	require.NoError(t, err)
	deployment, err := testCtx.client.K8s.AppsV1().Deployments(testCtx.namespace).Get(context.TODO(),
		"windows-machine-config-operator", meta.GetOptions{})
	require.NoError(t, err, "could not get WMCO deployment")
	require.NotZerof(t, deployment.Status.AvailableReplicas, "WMCO deployment has no available replicas: %v", deployment)
}
