output "resource_group_name" {
  value = azurerm_resource_group.this.name
}

output "vm_name" {
  value = azurerm_linux_virtual_machine.this.name
}

output "vm_public_ip" {
  value = azurerm_public_ip.this.ip_address
}

output "admin_username" {
  value = azurerm_linux_virtual_machine.this.admin_username
}
