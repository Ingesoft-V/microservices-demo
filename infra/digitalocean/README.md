# DigitalOcean Bootstrap (doctl)

Infraestructura mínima para este proyecto usando **DigitalOcean Droplet + Docker Compose**.

## Qué hace
- Crea (o reutiliza) un Droplet por entorno (`microservices-demo-preprod` / `microservices-demo-prod`).
- Crea (o reutiliza) una llave SSH en DigitalOcean usando el secret `VM_SSH_PUBLIC_KEY`.
- Instala Docker + Docker Compose Plugin vía `cloud-init`.
- Entrega la IP pública del Droplet para usarla como secret `VM_HOST`.

## Pipeline de infraestructura
Workflow: [.github/workflows/infra-digitalocean.yml](.github/workflows/infra-digitalocean.yml)

Acciones disponibles por `workflow_dispatch`:
- `status`: consulta estado/IP de la VM.
- `apply`: crea o reutiliza VM.
- `destroy`: elimina VM.

## Secrets requeridos
- `DIGITALOCEAN_ACCESS_TOKEN`
- `VM_SSH_PUBLIC_KEY`

## Secrets usados por deploy app
Workflow: [.github/workflows/deploy-vm-compose.yml](.github/workflows/deploy-vm-compose.yml)

- `VM_HOST` (IP pública del droplet)
- `VM_USER` (normalmente `root` en Droplet)
- `VM_SSH_PRIVATE_KEY` (llave privada correspondiente a `VM_SSH_PUBLIC_KEY`)

## Flujo recomendado
1. Ejecutar `Infra - DigitalOcean (doctl)` con `action=apply` y entorno.
2. Copiar IP de salida y actualizar secret `VM_HOST`.
3. Ejecutar `Deploy App - VM (Docker Compose)`.
