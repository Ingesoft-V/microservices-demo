# Terraform Bootstrap - Azure AKS

Infraestructura mínima para empezar en Azure con Kubernetes distribuido en más de una VM.

## Qué crea
- Resource Group
- Azure Container Registry (ACR)
- AKS (Azure Kubernetes Service) con `node_count = 2`
- Asignación de rol `AcrPull` para que AKS pueda descargar imágenes de ACR

## Prerrequisitos (CLI)
- `az` (Azure CLI)
- `terraform`
- sesión iniciada: `az login`
- suscripción seleccionada: `az account set --subscription <SUBSCRIPTION_ID>`

## Inicializar y validar
```bash
cd infra/terraform
terraform init
terraform fmt -recursive
terraform validate
```

> Nota: el provider `azurerm` está configurado con `resource_provider_registrations = "none"` para evitar fallos intermitentes `409 ConflictingConcurrentWriteNotAllowed` en CI cuando Azure intenta registrar múltiples Resource Providers en paralelo.

## Plan y Apply preprod
```bash
terraform plan -var-file=preprod.tfvars
terraform apply -var-file=preprod.tfvars
```

## Plan y Apply prod
```bash
terraform plan -var-file=prod.tfvars
terraform apply -var-file=prod.tfvars
```

## Obtener credenciales de AKS
```bash
az aks get-credentials \
  --resource-group $(terraform output -raw resource_group_name) \
  --name $(terraform output -raw aks_cluster_name)
```

## Ver nodos (VMs del cluster)
```bash
kubectl get nodes
```
