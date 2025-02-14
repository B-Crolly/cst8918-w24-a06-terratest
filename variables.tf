# Define config variables
variable "labelPrefix" {
  type        = string
  description = "Your college username. This will form the beginning of various resource names."
}

variable "region" {
  description = "The Azure region to deploy to"
  type        = string
  default     = "eastus"  # or your preferred region
}

variable "admin_username" {
  description = "Username for the VM"
  type        = string
  default     = "azureuser"
}

variable "ssh_public_key" {
  description = "SSH public key for VM access"
  type        = string
}
