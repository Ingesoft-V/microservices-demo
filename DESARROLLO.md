# Guía de Desarrollo Local

## Descripción

Esta guía te permite ejecutar el proyecto **Voting App** completamente en tu máquina local usando Docker Compose, sin necesidad de Kubernetes u Okteto.

## Arquitectura Local

```
┌─────────────┐
│   Vote      │ ──┐
│  (Java)     │   │
└─────────────┘   │
                  ├──> Kafka Topic: votes ──> Worker (x2) ──> PostgreSQL
┌─────────────┐   │
│   Result    │ ──┘
│  (Node.js)  │
└─────────────┘
```

## Prerequisitos

- **Docker Desktop** 20.10+
- **Docker Compose** 2.0+
- **Git**
- Mínimo 4GB RAM disponibles
- Puertos 5000, 5001, 5432, 9092 libres

### Verificar instalación

```bash
docker --version
docker-compose --version
```

## Instalación y Ejecución

### 1. Clonar el repositorio

```bash
git clone https://github.com/tu-usuario/microservices-demo.git
cd microservices-demo
```

### 2. Construir e iniciar servicios

```bash
# Construir imágenes y levantar servicios en background
docker-compose up --build -d

# Esperamos ~30 segundos mientras Kafka se inicializa y el topic se crea
```

### 3. Verificar estado de servicios

```bash
# Ver estado de todos los contenedores
docker-compose ps

# Debe mostrar algo como:
# NAME                COMMAND                  SERVICE             STATUS
# voting-postgres     postgres                 postgres            Up (healthy)
# voting-kafka        /etc/confluent/...       kafka               Up (healthy)
# voting-vote         java org.springframework Vote Service        Up
# voting-result       node server.js           Result Service      Up
# voting-worker       ./worker                 Worker (x2)         Up
```

### 4. Acceder a aplicaciones

Abre en tu navegador:

- **Vote (Frontend):** http://localhost:5000
- **Results (Dashboard):** http://localhost:5001

## Flujo de Votación

1. En **http://localhost:5000**, selecciona "Burrito" o "Taco"
2. Haz click en "Vote"
3. Deberías ver un mensaje de éxito en <100ms
4. Ve a **http://localhost:5001** y verás tu voto agregado en tiempo real

## Inspeccionar Logs

### Ver logs de un servicio específico

```bash
# Logs del worker (consumidor de Kafka)
docker-compose logs -f worker

# Logs de vote service
docker-compose logs -f vote

# Logs de result service
docker-compose logs -f result

# Logs de Kafka
docker-compose logs -f kafka

# Ver últimas 50 líneas
docker-compose logs --tail=50 worker
```

### Salir de logs

Presiona `Ctrl+C`

## Ejecutar Comandos en Contenedores

### Acceder a PostgreSQL

```bash
docker-compose exec postgres psql -U okteto -d votes

# Dentro de psql:
SELECT COUNT(*) FROM votes;
SELECT * FROM votes;
\dt  # Listar tablas
\q   # Salir
```

### Ver tópico de Kafka

```bash
# Listar tópicos
docker-compose exec kafka kafka-topics --list --bootstrap-server localhost:9092

# Ver mensajes del topic "votes"
docker-compose exec kafka kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic votes \
  --from-beginning
```

### Ejecutar comando en worker

```bash
docker-compose exec worker sh -c 'env | grep KAFKA'
```

## Escalado de Workers

### Aumentar replicas en tiempo real

```bash
# Escalar a 5 instancias del worker
docker-compose up -d --scale worker=5

# Verificar
docker-compose ps | grep worker
```

### Volver a 2 replicas

```bash
docker-compose up -d --scale worker=2
```

## Detener y Limpiar

### Detener servicios (mantiene datos)

```bash
docker-compose stop
```

### Reanudar servicios

```bash
docker-compose start
```

### Detener y eliminar contenedores (elimina datos temporales)

```bash
docker-compose down
```

### Limpiar todo (imágenes, volúmenes, contenedores)

⚠️ **Esto elimina la BD de PostgreSQL**

```bash
docker-compose down -v
```

## Troubleshooting

### Error: "Port already in use"

Puertos en conflicto. Cambiar en `docker-compose.yml`:

```yaml
ports:
  - "5002:80"  # Cambiar 5000 a 5002 si está en uso
```

### Worker no consume mensajes

```bash
# Verificar que Kafka está healthy
docker-compose logs kafka | grep healthy

# Ver si el topic fue creado
docker-compose exec kafka kafka-topics --list --bootstrap-server localhost:9092
```

### PostgreSQL no conecta

```bash
# Verificar conexión
docker-compose exec postgres psql -U okteto -d votes -c "SELECT 1"

# Si falla, reconstruir:
docker-compose down -v
docker-compose up --build -d
```

### Alto uso de CPU/memoria

```bash
# Verificar estadísticas
docker stats

# Limitar recursos en docker-compose.yml bajo "worker" deployment
```

## Desarrollo Iterativo

### Cambio en código de Vote (Java)

```bash
# 1. Editar código en vote/
# 2. Reconstruir servicio
docker-compose build vote

# 3. Reiniciar
docker-compose up -d vote

# 4. Ver logs
docker-compose logs -f vote
```

### Cambio en código de Worker (Go)

```bash
docker-compose build worker
docker-compose up -d --scale worker=2
docker-compose logs -f worker
```

### Cambio en código de Result (Node.js)

```bash
docker-compose build result
docker-compose up -d result
docker-compose logs -f result
```

## Testing End-to-End

```bash
#!/bin/bash
# test-voting.sh - Script de prueba

echo "🧪 Testing Voting App..."

# 1. Verificar servicios
echo "✓ Checking services..."
curl -s http://localhost:5000 > /dev/null && echo "  ✅ Vote service OK"
curl -s http://localhost:5001 > /dev/null && echo "  ✅ Result service OK"

# 2. Enviar voto
echo "✓ Sending test vote..."
curl -s -X POST http://localhost:5000/vote -d 'vote=Burrito' > /dev/null
echo "  ✅ Vote submitted"

# 3. Esperar procesamiento
sleep 2

# 4. Verificar resultado
echo "✓ Checking results..."
curl -s http://localhost:5001 | grep -q "Burrito" && echo "  ✅ Vote registered"

echo "🎉 All tests passed!"
```

Ejecutar:

```bash
chmod +x test-voting.sh
./test-voting.sh
```

## Variables de Entorno

Personalizables en `docker-compose.yml`:

| Variable | Valor Default | Descripción |
|----------|---------------|-------------|
| `POSTGRES_USER` | `okteto` | Usuario PostgreSQL |
| `POSTGRES_PASSWORD` | `okteto` | Contraseña PostgreSQL |
| `POSTGRES_DB` | `votes` | Base de datos |
| `KAFKA_BROKERS` | `kafka:29092` | Broker de Kafka |
| `KAFKA_TOPIC` | `votes` | Topic de Kafka |
| `KAFKA_GROUP` | `voting-group` | Consumer group |

## Notas Importantes

- **Persistencia**: Los datos de PostgreSQL se guardan en volumen `postgres_data`
- **Networking**: Todos los servicios están en la red `voting-network`
- **Escalabilidad**: Workers configurados con 2 réplicas por defecto
- **Health Checks**: Vote y Result tienen checks cada 10s
- **Logs**: Ver `docker-compose logs -f` para debugging

## Comandos Útiles

```bash
# Reiniciar todo
docker-compose restart

# Logs de todos con timestamps
docker-compose logs -f --timestamps

# Ver consumo de recursos
docker stats

# Limpiar imágenes no usadas
docker image prune

# Ver volúmenes
docker volume ls
```

## Para Más Información

- [Docker Compose Docs](https://docs.docker.com/compose/)
- [Kafka Documentation](https://kafka.apache.org/documentation/)
- [PostgreSQL Docs](https://www.postgresql.org/docs/)
