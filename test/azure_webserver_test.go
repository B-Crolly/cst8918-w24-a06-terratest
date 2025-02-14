package test

import (
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/gruntwork-io/terratest/modules/azure"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

// You normally want to run this under a separate "Testing" subscription
// For lab purposes you will use your assigned subscription under the Cloud Dev/Ops program tenant
var subscriptionID string = "d0681dfb-9589-4bb4-8dbd-11883649bebf"

type TestData struct {
	terraformOptions  *terraform.Options
	resourceGroupName string
	vmName           string
}

func setupTestData(t *testing.T) *TestData {
	t.Helper()

	// The path to where our Terraform code is located
	terraformOptions := &terraform.Options{
		TerraformDir: "../",
		// Override the default terraform variables
		Vars: map[string]interface{}{
			"labelPrefix":    "crol0005",
			"region":        "westus2",
			"admin_username": "azureuser",
		},
		// Maximum time to wait for apply
		MaxRetries:         3,
		TimeBetweenRetries: 5 * time.Second,
	}

	// Run `terraform init` and `terraform apply`
	terraform.InitAndApply(t, terraformOptions)

	// Get outputs
	resourceGroupName := terraform.Output(t, terraformOptions, "resource_group_name")
	vmName := terraform.Output(t, terraformOptions, "vm_name")

	testData := &TestData{
		terraformOptions:  terraformOptions,
		resourceGroupName: resourceGroupName,
		vmName:           vmName,
	}

	// Add a cleanup function
	t.Cleanup(func() {
		terraform.Destroy(t, terraformOptions)
	})

	return testData
}

func TestAzureInfrastructure(t *testing.T) {
	testData := setupTestData(t)

	// Run subtests
	t.Run("VM Creation and Configuration", func(t *testing.T) {
		t.Log("Testing VM existence and configuration")
		
		// Get the VM instance
		vm := azure.GetVirtualMachine(t, testData.vmName, testData.resourceGroupName, subscriptionID)
		assert.NotNil(t, vm, "Virtual Machine should exist")

		// Verify VM size
		vmSize := azure.GetSizeOfVirtualMachine(t, testData.vmName, testData.resourceGroupName, subscriptionID)
		assert.Equal(t, compute.VirtualMachineSizeTypes("Standard_B1s"), vmSize, "Incorrect VM size")

		// Get and verify OS disk name
		osDiskName := azure.GetVirtualMachineOSDiskName(t, testData.vmName, testData.resourceGroupName, subscriptionID)
		assert.Contains(t, osDiskName, "OSDisk", "OS disk name should contain 'OSDisk'")
	})

	t.Run("Network Configuration", func(t *testing.T) {
		t.Log("Testing network interface configuration")
		
		// Get the list of NICs attached to the VM
		nics := azure.GetVirtualMachineNics(t, testData.vmName, testData.resourceGroupName, subscriptionID)
		assert.NotEmpty(t, nics, "VM should have at least one NIC")
		
		// Get the expected NIC name from Terraform
		expectedNicName := terraform.Output(t, testData.terraformOptions, "nic_name")
		assert.Contains(t, nics, expectedNicName, "NIC name should match the expected name")

		// Get the public IP from Terraform
		publicIP := terraform.Output(t, testData.terraformOptions, "public_ip")
		assert.NotEmpty(t, publicIP, "Network Interface should have a public IP")
	})

	t.Run("VM Image Configuration", func(t *testing.T) {
		t.Log("Testing VM image configuration")
		
		// Get the VM image details
		vmImage := azure.GetVirtualMachineImage(t, testData.vmName, testData.resourceGroupName, subscriptionID)
		
		// Verify image details
		assert.Equal(t, "Canonical", vmImage.Publisher, "Incorrect image publisher")
		assert.Equal(t, "0001-com-ubuntu-server-jammy", vmImage.Offer, "Incorrect image offer")
		assert.Equal(t, "22_04-lts-gen2", vmImage.SKU, "Incorrect image SKU")
		assert.Equal(t, "latest", vmImage.Version, "Incorrect image version")
	})
}
