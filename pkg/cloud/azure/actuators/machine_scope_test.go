/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package actuators

import (
	"testing"

	"github.com/ghodss/yaml"
	clusterv1 "github.com/openshift/cluster-api/pkg/apis/cluster/v1alpha1"
	machinev1 "github.com/openshift/cluster-api/pkg/apis/machine/v1beta1"
	"github.com/openshift/cluster-api/pkg/client/clientset_generated/clientset/fake"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/pointer"
	clusterproviderv1 "sigs.k8s.io/cluster-api-provider-azure/pkg/apis/azureprovider/v1alpha1"
	machineproviderv1 "sigs.k8s.io/cluster-api-provider-azure/pkg/apis/azureprovider/v1beta1"
	controllerfake "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func providerSpecFromMachine(in *machineproviderv1.AzureMachineProviderSpec) (*machinev1.ProviderSpec, error) {
	bytes, err := yaml.Marshal(in)
	if err != nil {
		return nil, err
	}
	return &machinev1.ProviderSpec{
		Value: &runtime.RawExtension{Raw: bytes},
	}, nil
}

func newMachine(t *testing.T) *machinev1.Machine {
	machineConfig := machineproviderv1.AzureMachineProviderSpec{}
	providerSpec, err := providerSpecFromMachine(&machineConfig)
	if err != nil {
		t.Fatalf("error encoding provider config: %v", err)
	}
	return &machinev1.Machine{
		ObjectMeta: metav1.ObjectMeta{
			Name: "machine-test",
		},
		Spec: machinev1.MachineSpec{
			ProviderSpec: *providerSpec,
		},
	}
}

func TestNilClusterScope(t *testing.T) {
	m := newMachine(t)
	params := MachineScopeParams{
		AzureClients: AzureClients{},
		Cluster:      nil,
		CoreClient:   nil,
		Machine:      m,
		Client:       fake.NewSimpleClientset(m).MachineV1beta1(),
	}
	_, err := NewMachineScope(params)
	if err != nil {
		t.Errorf("Expected New machine scope to succeed with nil cluster: %v", err)
	}
}

func TestCredentialsSecretSuccess(t *testing.T) {
	credentialsSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testCredentials",
			Namespace: "dummyNamespace",
		},
		Data: map[string][]byte{
			"azure_subscription_id": []byte("dummySubID"),
			"azure_client_id":       []byte("dummyClientID"),
			"azure_client_secret":   []byte("dummyClientSecret"),
			"azure_tenant_id":       []byte("dummyTenantID"),
			"azure_resourcegroup":   []byte("dummyResourceGroup"),
			"azure_region":          []byte("dummyRegion"),
			"azure_resource_prefix": []byte("dummyClusterName"),
		},
	}
	scope := &Scope{Cluster: &clusterv1.Cluster{}, ClusterConfig: &clusterproviderv1.AzureClusterProviderSpec{}}
	err := updateScope(
		controllerfake.NewFakeClient(credentialsSecret),
		&corev1.SecretReference{Name: "testCredentials", Namespace: "dummyNamespace"},
		scope)
	if err != nil {
		t.Errorf("Expected New credentials secrets to succeed: %v", err)
	}

	if scope.SubscriptionID != "dummySubID" {
		t.Errorf("Expected subscriptionID to be dummySubID but found %s", scope.SubscriptionID)
	}

	if scope.Location() != "dummyRegion" {
		t.Errorf("Expected location to be dummyRegion but found %s", scope.Location())
	}

	if scope.Cluster.Name != "dummyClusterName" {
		t.Errorf("Expected cluster name to be dummyClusterName but found %s", scope.Cluster.Name)
	}

	if scope.ClusterConfig.ResourceGroup != "dummyResourceGroup" {
		t.Errorf("Expected resourcegroup to be dummyResourceGroup but found %s", scope.ClusterConfig.ResourceGroup)
	}
}

func testCredentialFields(credentialsSecret *corev1.Secret) error {
	scope := &Scope{Cluster: &clusterv1.Cluster{}, ClusterConfig: &clusterproviderv1.AzureClusterProviderSpec{}}
	return updateScope(
		controllerfake.NewFakeClient(credentialsSecret),
		&corev1.SecretReference{Name: "testCredentials", Namespace: "dummyNamespace"},
		scope)
}

func TestCredentialsSecretFailures(t *testing.T) {
	credentialsSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testCredentials",
			Namespace: "dummyNamespace",
		},
		Data: map[string][]byte{},
	}

	if err := testCredentialFields(credentialsSecret); err == nil {
		t.Errorf("Expected New credentials secrets to fail")
	}

	credentialsSecret.Data["azure_subscription_id"] = []byte("dummyValue")
	if err := testCredentialFields(credentialsSecret); err == nil {
		t.Errorf("Expected New credentials secrets to fail")
	}

	credentialsSecret.Data["azure_client_id"] = []byte("dummyValue")
	if err := testCredentialFields(credentialsSecret); err == nil {
		t.Errorf("Expected New credentials secrets to fail")
	}

	credentialsSecret.Data["azure_client_secret"] = []byte("dummyValue")
	if err := testCredentialFields(credentialsSecret); err == nil {
		t.Errorf("Expected New credentials secrets to fail")
	}

	credentialsSecret.Data["azure_tenant_id"] = []byte("dummyValue")
	if err := testCredentialFields(credentialsSecret); err == nil {
		t.Errorf("Expected New credentials secrets to fail")
	}

	credentialsSecret.Data["azure_resourcegroup"] = []byte("dummyValue")
	if err := testCredentialFields(credentialsSecret); err == nil {
		t.Errorf("Expected New credentials secrets to fail")
	}

	credentialsSecret.Data["azure_region"] = []byte("dummyValue")
	if err := testCredentialFields(credentialsSecret); err == nil {
		t.Errorf("Expected New credentials secrets to fail")
	}

	credentialsSecret.Data["azure_resource_prefix"] = []byte("dummyValue")
	if err := testCredentialFields(credentialsSecret); err != nil {
		t.Errorf("Expected New credentials secrets to succeed but found : %v", err)
	}
}

func testCredentialSecret() *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testCredentials",
			Namespace: "dummyNamespace",
		},
		Data: map[string][]byte{
			"azure_subscription_id": []byte("dummySubID"),
			"azure_client_id":       []byte("dummyClientID"),
			"azure_client_secret":   []byte("dummyClientSecret"),
			"azure_tenant_id":       []byte("dummyTenantID"),
			"azure_resourcegroup":   []byte("dummyResourceGroup"),
			"azure_region":          []byte("dummyRegion"),
			"azure_resource_prefix": []byte("dummyClusterName"),
		},
	}
}

func testProviderSpec() *machineproviderv1.AzureMachineProviderSpec {
	return &machineproviderv1.AzureMachineProviderSpec{
		Location:          "test",
		ResourceGroup:     "test",
		CredentialsSecret: &corev1.SecretReference{Name: "testCredentials", Namespace: "dummyNamespace"},
	}
}

func testMachineWithProviderSpec(t *testing.T, providerSpec *machineproviderv1.AzureMachineProviderSpec) *machinev1.Machine {
	providerSpecWithValues, err := providerSpecFromMachine(providerSpec)
	if err != nil {
		t.Fatalf("error encoding provider config: %v", err)
	}

	return &machinev1.Machine{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test",
			Annotations: map[string]string{},
		},
		Spec: machinev1.MachineSpec{
			ProviderSpec: *providerSpecWithValues,
		},
	}
}

func testMachine(t *testing.T) *machinev1.Machine {
	return testMachineWithProviderSpec(t, testProviderSpec())
}

func TestPersistMachineScope(t *testing.T) {
	machine := testMachine(t)

	params := MachineScopeParams{
		Machine:    machine,
		Cluster:    nil,
		Client:     fake.NewSimpleClientset(machine).MachineV1beta1(),
		CoreClient: controllerfake.NewFakeClientWithScheme(scheme.Scheme, testCredentialSecret()),
	}

	scope, err := NewMachineScope(params)
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}

	nodeAddresses := []corev1.NodeAddress{
		{
			Type:    corev1.NodeHostName,
			Address: "hostname",
		},
	}

	scope.Machine.Annotations["test"] = "testValue"
	scope.MachineStatus.VMID = pointer.StringPtr("vmid")
	scope.Machine.Status.Addresses = make([]corev1.NodeAddress, len(nodeAddresses))
	copy(nodeAddresses, scope.Machine.Status.Addresses)

	if err = scope.Persist(); err != nil {
		t.Errorf("Expected MachineScope.Persist to success, got error: %v", err)
	}

	updatedMachine, err := params.Client.Machines(params.Machine.Namespace).Get(params.Machine.Name, metav1.GetOptions{})
	if err != nil {
		t.Errorf("Unable to get updated machine: %v", err)
	}

	if !equality.Semantic.DeepEqual(updatedMachine.Status.Addresses, nodeAddresses) {
		t.Errorf("Expected node addresses to equal, updated addresses %#v, expected addresses: %#v", updatedMachine.Status.Addresses, nodeAddresses)
	}

	if updatedMachine.Annotations["test"] != "testValue" {
		t.Errorf("Expected annotation 'test' to equal 'testValue', got %q instead", updatedMachine.Annotations["test"])
	}

	machineStatus, err := machineproviderv1.MachineStatusFromProviderStatus(updatedMachine.Status.ProviderStatus)
	if err != nil {
		t.Errorf("failed to get machine provider status: %v", err)
	}

	if machineStatus.VMID == nil {
		t.Errorf("Expected VMID to be 'vmid', got nil instead")
	} else if *machineStatus.VMID != "vmid" {
		t.Errorf("Expected VMID to be 'vmid', got %q instead", *machineStatus.VMID)
	}
}

func TestNewMachineScope(t *testing.T) {
	machineConfigNoValues := &machineproviderv1.AzureMachineProviderSpec{
		CredentialsSecret: &corev1.SecretReference{Name: "testCredentials", Namespace: "dummyNamespace"},
	}

	testCases := []struct {
		machine               *machinev1.Machine
		secret                *corev1.Secret
		expectedLocation      string
		expectedResourceGroup string
	}{
		{
			machine:               testMachine(t),
			secret:                testCredentialSecret(),
			expectedLocation:      "test",
			expectedResourceGroup: "test",
		},
		{
			machine:               testMachineWithProviderSpec(t, machineConfigNoValues),
			secret:                testCredentialSecret(),
			expectedLocation:      "dummyRegion",
			expectedResourceGroup: "dummyResourceGroup",
		},
	}

	for _, tc := range testCases {
		scope, err := NewMachineScope(MachineScopeParams{
			Machine:    tc.machine,
			Cluster:    nil,
			Client:     fake.NewSimpleClientset(tc.machine).MachineV1beta1(),
			CoreClient: controllerfake.NewFakeClientWithScheme(scheme.Scheme, tc.secret),
		})
		if err != nil {
			t.Fatalf("Unexpected error %v", err)
		}

		if scope.ClusterConfig.Location != tc.expectedLocation {
			t.Errorf("Expected %v, got: %v", tc.expectedLocation, scope.ClusterConfig.Location)
		}
		if scope.ClusterConfig.ResourceGroup != tc.expectedResourceGroup {
			t.Errorf("Expected %v, got: %v", tc.expectedResourceGroup, scope.ClusterConfig.ResourceGroup)
		}
	}
}
