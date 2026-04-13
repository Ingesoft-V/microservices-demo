variable "project_name" {
  description = "Project short name used in resource naming"
  type        = string
  default     = "voting"
}

variable "environment" {
  description = "Deployment environment (preprod|prod)"
  type        = string
  default     = "preprod"
}

variable "location" {
  description = "Azure region"
  type        = string
  default     = "eastus"
}

variable "vm_size" {
  description = "Azure VM size"
  type        = string
  default     = "Standard_B2s"
}

variable "admin_username" {
  description = "Admin username for the Linux VM"
  type        = string
  default     = "azureuser"
}

variable "ssh_public_key" {
  description = "SSH public key content used to access the VM"
  type        = string
}

variable "tags" {
  description = "Tags for Azure resources"
  type        = map(string)
  default = {
    managed_by = "terraform"
    project    = "microservices-demo"
  }
}
