# Taller 1 - Construcción de Pipelines en Cloud
**Curso:** Ingeniería de Software V  
**Fecha de Presentación:** 13 de Abril  
**Proyecto Base:** [Microservices Demo (Okteto)](https://github.com/okteto/microservices-demo)  
**Repositorio del equipo:** https://github.com/Ingesoft-V/microservices-demo

---

## Metodología Ágil Seleccionada

* **Metodología:** Scrum
* **Justificación:** El proyecto debe poder ser utilizado por un equipo ágil, por lo que se escoge Scrum. Esto permite alta adaptabilidad a cambios, entregas tempranas y frecuentes de valor funcional, y una rápida retroalimentación.

---

## 1. Estrategias de Branching Para Desarrolladores

* **Modelo:** GitHub Flow
* **Descripción:** 
    * **Rama `main`:** Siempre debe estar en un estado desplegable. No se toca directamente.
    * **Ramas `feature/nombre-tarea`:** Cada desarrollador crea una rama corta para una funcionalidad o corrección específica.
    * **Pull Requests (PR):** Antes de integrar a `main`, se requiere revisión de código y que las pruebas automatizadas pasen en el pipeline.
    * **Fusión (Merge):** Una vez aprobado, se fusiona a `main` y se dispara el despliegue automático.
* **Justificación:** 
    * Esta estrategia garantiza que cualquier código integrado haya pasado ya por un proceso de validación, manteniendo la estabilidad de la rama de producción de manera constante.
    * Al realizar cada modificación (nueva funcionalidad o corrección de errores) en una rama independiente derivada de la principal, se facilita el seguimiento del proceso de desarrollo por medio de los nombres descriptivos de cada rama y el desarrollo paralelo (no dependo de los otros miembros del equipo de desarrollo).
    * Imponer un flujo basado en pull requests permite que el código sea evaluado y validado por otros miembros del equipo antes de su integración.
    * La validación y el testing se realizan en la rama de la funcionalidad antes de la fusión, lo que mitiga el riesgo de introducir código inestable en la rama principal.
    * Facilita la incorporación rápida de cambios al CI/CD pipeline dado a su simplicidad.


---

## 2. Estrategias de Branching Para Operaciones

* **Modelo:** GitOps (Modelo Branch-per-Environment)
* **Descripción:** Su objetivo principal es que el repositorio sea la única fuente de verdad del estado de tus entornos.
    * Rama `main`: Representa el entorno de pre-producción.
    * Rama `production`: Representa el estado real de lo que ven los usuarios finales. Nadie hace cambios directos aquí.
    * Para pasar un cambio de Pre-Producción a Producción se abre un Pull Request de `main` hacia `production`.
    * Promoción controlada: una vez el PR se aprueba y se hace merge, se dispara el despliegue del entorno correspondiente (CD) y la infraestructura se gestiona por un workflow separado.
* **Justificación:** 
    *  **Fuente Única de Verdad:** El estado de la infraestructura está totalmente definido y versionado en el repositorio.
    * **Control de Promoción:** Los cambios pasan de `main` (Pre-producción) a `production` solo mediante Pull Requests aprobados, evitando errores manuales.
    * **Automatización:** El despliegue se activa automáticamente tras el merge, garantizando que el entorno refleje exactamente el código.
    * **Estabilidad y Auditoría:** Ofrece un rastro claro de quién cambió la infraestructura y permite reversiones (rollbacks) rápidas ante fallos.

---

## 3. Patrones de Diseño de Nube
*Se han implementado al menos dos patrones basándonos en los temas expuestos en clase para garantizar la escalabilidad y resiliencia del proyecto.*

1. **Patrón 1: Producer-Consumer (con Competing Consumers)**
        * **Propósito:** Separa productores y consumidores para procesar mensajes de forma desacoplada y escalable.
        * **Implementación en el proyecto:**
            - **Productor:** `vote` publica eventos de voto en Kafka.
            - **Consumidor:** `worker` consume del topic `votes`.
            - **Competing Consumers:** múltiples instancias de `worker` pueden ejecutar en paralelo usando consumer group `voting-group` y estrategia `RoundRobin`.
            - **Consistencia:** `worker` persiste con `ON CONFLICT`, usando `msg.Key` como ID único de votante para evitar duplicados.

2. **Patrón 2: Publisher-Subscriber (Pub/Sub) / Comunicación Asíncrona**
    * **Propósito:** Desacopla las partes de un sistema que producen eventos (publicadores) de aquellas que los procesan (suscriptores). El componente que emite la información no necesita esperar la respuesta, mejorando la disponibilidad y la respuesta inmediata al usuario final.
    * **Implementación en el proyecto:** 
      - **`vote` (Productor):** Publica voto en topic Kafka `votes` (mediante `kafkaTemplate.send()`) y retorna éxito al usuario en <100ms.
      - **`worker` (Consumidor):** Asíncrono y desacoplado, procesa en background sin afectar la UX del frontend.
      - **`result` (Lector):** Lee datos consolidados de PostgreSQL (escritos por worker) para mostrar resultados en tiempo real.
      - **Ventaja:** El servicio `vote` nunca bloquea esperando confirmación; el `worker` procesa cuando puede sin presión de tiempo.

3. **Patrón 3: External Configuration Store**
        * **Propósito:** Centraliza configuración operativa fuera del código para facilitar cambios por entorno sin recompilar.
        * **Implementación en el proyecto:**
            - Se creó `.env.example` como plantilla de configuración centralizada.
            - `docker-compose.yml` consume variables (`POSTGRES_*`, `KAFKA_*`, puertos, topic, polling interval).
            - `vote`, `worker` y `result` leen endpoints y parámetros desde variables de entorno con valores por defecto.
            - Permite promover de dev a preprod/prod cambiando configuración, no código.

4. **Patrón 4: Retry**
        * **Propósito:** Reintenta operaciones transitorias cuando un recurso externo aún no está listo o falla temporalmente.
        * **Implementación en el proyecto:**
            - `result/server.js` usa `async.retry` para reconectar a PostgreSQL antes de comenzar a emitir resultados.
            - Los tiempos y cantidad de reintentos se pueden ajustar con `RESULT_DB_RETRY_TIMES` y `RESULT_DB_RETRY_INTERVAL_MS`.
            - Esto evita que el servicio falle si la base de datos tarda unos segundos más en arrancar.

## 4. Diagrama de Arquitectura (15.0%)
A continuación se presenta el flujo de la aplicación *Docker Voting App* y su interacción a nivel de servicios y datos e infraestructura abstraída en red.

![Diagrama de arquitectura](diagram_archictecture.png)

**Explicación del flujo:**
1. Los **usuarios** interactúan a través de un *Load Balancer/Ingress* en la nube que enruta el tráfico a los microservicios.
2. El servicio **`vote`** es una interfaz web ligera (Spring Boot) que acepta el voto y lo publica directamente en el topic Kafka `votes` (Producer Pattern).
3. Múltiples instancias del **`worker`** (Go) consumen competitivamente del topic Kafka usando consumer group `voting-group`, garantizando escalabilidad horizontal (Competing Consumers Pattern).
4. Cada **`worker`** recibe votos con `msg.Key = ID_votante`, insertándolos en PostgreSQL con identificación única (evita duplicados).
5. El servicio **`result`** (Node.js) lee datos consolidados directamente de PostgreSQL y los sirve en tiempo real al usuario observador.


---

## 5. Pipelines de Desarrollo (15.0%)
*Detalle de la automatización del ciclo de vida de la aplicación.*

* **Herramienta:** GitHub Actions
* **Workflow:** `.github/workflows/ci.yml`
* **Disparadores:** `push` y `pull_request` a `main` y `production`.
* **Tareas incluidas (por servicio):**
    - **vote (Java/Maven):** build con `mvn -DskipTests package`.
    - **worker (Go):** `go mod tidy` + `go build ./...`.
    - **result (Node.js):** `npm install` + `node --check server.js`.
* **Objetivo:** evitar que cambios incompatibles se integren a ramas protegidas.

---

## 6. Pipelines de Infraestructura (5.0%)
*Automatización del despliegue de recursos.*

* **Herramienta:** GitHub Actions + `doctl`
* **Descripción:**
    1. Workflow de infraestructura: `.github/workflows/infra-digitalocean.yml`.
    2. En **Pull Request** a `main` / `production`, ejecuta el job `infra-check` (validación rápida de archivos).
    3. Para cambios reales de infraestructura, se ejecuta **manual** (`workflow_dispatch`) con acciones:
        - `status`: consulta existencia e IP del droplet.
        - `apply`: crea/reutiliza un droplet por entorno (`microservices-demo-preprod` / `microservices-demo-prod`) e instala Docker/Compose vía `cloud-init`.
        - `destroy`: elimina el droplet del entorno.
    4. Tras `apply`, se toma la IP y se actualiza el secret `VM_HOST`.

---

## 7. Implementación de la Infraestructura (20.0%)
* **Proveedor Cloud:** DigitalOcean
* **Componentes:**
    - 1 Droplet Linux (VM)
    - Docker + Docker Compose en la VM
    - GitHub Actions (CI + CD + Infra)
    - SSH para despliegue remoto

* **Bootstrap:** `infra/digitalocean/cloud-init.yaml` instala Docker y `docker compose`.
* **Secrets (GitHub Actions):**
    - Infra: `DIGITALOCEAN_ACCESS_TOKEN`, `VM_SSH_PUBLIC_KEY`
    - Deploy app: `VM_HOST`, `VM_USER`, `VM_SSH_PRIVATE_KEY`

---

## 8. Guía para Demostración en Vivo (15.0%)
*Pasos sugeridos para demostrar pipelines durante la presentación (8 min):*

1. **Demostración Infra (PR):**
    - Crear una rama y modificar un archivo dentro de `infra/digitalocean/**`.
    - Abrir una PR hacia `main`.
    - Mostrar que corre automáticamente el workflow **Infra - DigitalOcean (doctl)** con el job `infra-check`.

2. **Demostración Infra (manual):**
    - Ejecutar el workflow manual con `action=apply` y `environment=preprod`.
    - Mostrar la IP resultante y actualizar el secret `VM_HOST`.

3. **Demostración CD (Deploy App):**
    - Hacer un cambio pequeño en alguno de los servicios o en `docker-compose.yml`.
    - Hacer merge a `main`.
    - Mostrar que corre el workflow **Deploy App - VM (Docker Compose)** y actualiza los contenedores en la VM.

4. **Validación:**
    - Abrir en navegador los endpoints del despliegue (puertos `5000` y `5001`).

---

## 9. Documentación y Resultados (10.0%)
* **Enlace al Repositorio:** https://github.com/Ingesoft-V/microservices-demo
* **Evidencias (para anexar):**
    - Screenshot de la ejecución de CI en verde (PR a `main`).
    - Screenshot de Infra `infra-check` corriendo por PR.
    - Logs/screenshot de `workflow_dispatch` con `action=apply` mostrando IP.
    - Screenshot del deploy (SSH) finalizando en verde.
    - Evidencia de endpoints en ejecución (HTTP 200 en `:5000` y `:5001`).