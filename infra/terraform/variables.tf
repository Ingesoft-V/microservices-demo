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

variable "node_count" {
  description = "Number of AKS nodes (VMs)"
  type        = number
  default     = 2
}

variable "node_vm_size" {
  description = "AKS node VM size"
  type        = string
  default     = "Standard_B2s"
}

variable "kubernetes_version" {
  description = "Optional AKS version. Null uses default supported version"
  type        = string
  default     = null
}

variable "tags" {
  description = "Tags for Azure resources"
  type        = map(string)
  default = {
    managed_by = "terraform"
    project    = "microservices-demo"
  }
}
