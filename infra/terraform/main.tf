locals {
  suffix           = "${var.project_name}-${var.environment}"
  resource_group   = "rg-${local.suffix}"
  acr_name         = replace("acr${var.project_name}${var.environment}", "-", "")
  aks_cluster_name = "aks-${local.suffix}"
}

resource "azurerm_resource_group" "this" {
  name     = local.resource_group
  location = var.location
  tags     = var.tags
}

resource "azurerm_container_registry" "this" {
  name                = substr(local.acr_name, 0, 50)
  resource_group_name = azurerm_resource_group.this.name
  location            = azurerm_resource_group.this.location
  sku                 = "Basic"
  admin_enabled       = false
  tags                = var.tags
}

resource "azurerm_kubernetes_cluster" "this" {
  name                = local.aks_cluster_name
  resource_group_name = azurerm_resource_group.this.name
  location            = azurerm_resource_group.this.location
  dns_prefix          = "dns-${local.suffix}"

  kubernetes_version = var.kubernetes_version

  default_node_pool {
    name       = "default"
    node_count = var.node_count
    vm_size    = var.node_vm_size
  }

  identity {
    type = "SystemAssigned"
  }

  role_based_access_control_enabled = true

  network_profile {
    network_plugin = "kubenet"
    load_balancer_sku = "standard"
  }

  tags = var.tags
}

resource "azurerm_role_assignment" "aks_acr_pull" {
  scope                = azurerm_container_registry.this.id
  role_definition_name = "AcrPull"
  principal_id         = azurerm_kubernetes_cluster.this.kubelet_identity[0].object_id
}
