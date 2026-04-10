environment        = "preprod"
location           = "eastus"
node_count         = 2
node_vm_size       = "Standard_B2s"
kubernetes_version = null

tags = {
  managed_by  = "terraform"
  project     = "microservices-demo"
  environment = "preprod"
}
