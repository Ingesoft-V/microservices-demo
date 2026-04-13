# Terraform Bootstrap - Azure VM (Docker Compose)

Infraestructura mínima en Azure para desplegar el proyecto en **una sola VM Linux** usando Docker Compose.

## Qué crea
- Resource Group
- Virtual Network + Subnet
- Public IP
- Network Security Group con puertos abiertos:
  - `22` (SSH)
  - `5000` (vote)
  - `5001` (result)
- 1 VM Ubuntu 22.04

La VM se aprovisiona con `cloud-init` para instalar Docker, Docker Compose Plugin y Git.

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
export TF_VAR_ssh_public_key="$(cat ~/.ssh/id_rsa.pub)"
terraform plan -var-file=preprod.tfvars
terraform apply -var-file=preprod.tfvars
```

## Plan y Apply prod
```bash
export TF_VAR_ssh_public_key="$(cat ~/.ssh/id_rsa.pub)"
terraform plan -var-file=prod.tfvars
terraform apply -var-file=prod.tfvars
```

## Outputs útiles
```bash
terraform output -raw resource_group_name
terraform output -raw vm_name
terraform output -raw vm_public_ip
terraform output -raw admin_username
```
