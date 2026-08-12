# Microblogging Challenge

API REST de una plataforma simplificada de microblogging desarrollada en Go.

La aplicación permite:

- Publicar tweets.
- Seguir usuarios.
- Consultar un timeline paginado con los tweets de los usuarios seguidos.

## Requisitos

- Go 1.26.5 o compatible.
- No requiere una base de datos externa.

La persistencia se implementa en memoria, por lo que los datos se pierden cuando se reinicia la aplicación.

## Ejecución

Descargar las dependencias:

```bash
go mod download
```

Iniciar la API:

```bash
go run ./cmd/api
```

El servidor queda disponible en:

```text
http://localhost:8080
```

## Ejecución con Docker

Para construir la imagen desde la raíz del proyecto:

```bash
docker build -t microblogging-api:local .
```

Para ejecutar el contenedor:

```bash
docker run --rm \
  --name microblogging-api \
  -p 8080:8080 \
  microblogging-api:local
```

La API queda disponible en:

```text
http://localhost:8080
```

Para detener el contenedor, presionar `Control + C` en la terminal donde se está ejecutando.

También puede detenerse desde otra terminal:

```bash
docker stop microblogging-api
```

La persistencia es en memoria. Los tweets y las relaciones de seguimiento se pierden cuando se elimina o reinicia el contenedor.

## Endpoints

| Método | Endpoint | Descripción |
|---|---|---|
| `POST` | `/tweets` | Publica un tweet |
| `PUT` | `/users/{follower_id}/following/{followed_id}` | Sigue a otro usuario |
| `GET` | `/users/{user_id}/timeline` | Consulta el timeline del usuario |

Los requests, respuestas, validaciones y errores están detallados en [docs/api-contracts.md](docs/api-contracts.md).

## Ejemplo de uso

### Publicar un tweet

```bash
curl -i \
  -X POST http://localhost:8080/tweets \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-2",
    "content": "Mi primer tweet"
  }'
```

Respuesta esperada:

```http
HTTP/1.1 201 Created
Content-Type: application/json
```

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "user-2",
  "content": "Mi primer tweet",
  "date": "2026-08-11T18:00:00Z"
}
```

### Seguir a un usuario

Para que `user-1` siga a `user-2`:

```bash
curl -i \
  -X PUT \
  http://localhost:8080/users/user-1/following/user-2
```

La primera solicitud devuelve:

```http
HTTP/1.1 201 Created
```

Si la relación ya existe, devuelve:

```http
HTTP/1.1 204 No Content
```

### Consultar el timeline

```bash
curl -i \
  "http://localhost:8080/users/user-1/timeline?limit=20"
```

Respuesta:

```json
{
  "tweets": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "user_id": "user-2",
      "content": "Mi primer tweet",
      "date": "2026-08-11T18:00:00Z"
    }
  ],
  "next_cursor": null
}
```

El timeline:

- Incluye solamente tweets de usuarios seguidos.
- No incluye tweets propios.
- Incluye tweets publicados antes de crear la relación de seguimiento.
- Ordena por fecha descendente y utiliza el ID como desempate.
- Devuelve 20 elementos por defecto.
- Permite solicitar entre 1 y 100 elementos.

### Consultar la página siguiente

Cuando `next_cursor` contiene un string, debe enviarse sin modificar:

```bash
curl -i \
  "http://localhost:8080/users/user-1/timeline?limit=20&cursor=CURSOR_RECIBIDO"
```

El cursor es opaco para el cliente: solamente debe conservarlo y enviarlo para solicitar la página siguiente.

## Tests

Ejecutar la suite completa:

```bash
go test ./...
```

Ejecutar con race detector:

```bash
go test -race ./...
```

Consultar coverage:

```bash
go test -cover ./...
```

Ejecutar análisis estático:

```bash
go vet ./...
```

## Arquitectura

La aplicación utiliza un monolito modular con separación entre:

- Dominio y casos de uso.
- Transporte HTTP/REST.
- Adaptadores de persistencia.
- Composición de dependencias.

Los casos de uso dependen de interfaces pequeñas y no conocen las implementaciones concretas de memoria ni la capa HTTP.

El timeline utiliza fan-out en lectura: se construye cuando el usuario lo consulta combinando los tweets de las cuentas que sigue.

Para favorecer las lecturas, la implementación en memoria mantiene índices por:

- Usuario seguidor, para obtener directamente las cuentas seguidas.
- Autor del tweet, para evitar recorrer tweets de autores no relacionados.

La paginación utiliza un cursor opaco que representa la fecha y el ID del último tweet devuelto.

La arquitectura, estrategia de persistencia, índices, paginación y limitaciones están documentadas en [docs/architecture.md](docs/architecture.md).

Las reglas de negocio y los supuestos adoptados están documentados en [business.txt](business.txt).

## Decisiones y limitaciones

- Todos los identificadores de usuario recibidos se consideran válidos.
- No existe registro de usuarios, autenticación ni manejo de sesiones.
- Los tweets tienen un máximo de 280 caracteres Unicode.
- Un usuario no puede seguirse a sí mismo.
- Seguir varias veces al mismo usuario no crea relaciones duplicadas.
- Los datos se almacenan en memoria y se pierden al reiniciar el proceso.
- El timeline se construye en cada lectura y no representa una fotografía inmutable.
- El costo de consultar el timeline crece con la cantidad de cuentas seguidas y su historial de tweets.
- La implementación en memoria simplifica la ejecución del challenge y no representa la persistencia de un entorno productivo.

Para un entorno real se utilizaría una base de datos persistente con los índices descritos en la documentación de arquitectura.

## Uso de inteligencia artificial

Durante el desarrollo se utilizó IA como herramienta de apoyo para:

- Analizar la consigna y los contratos.
- Discutir decisiones de arquitectura y sus trade-offs.
- Revisar incrementalmente el código.
- Identificar casos de prueba y validar el cumplimiento funcional.

La implementación fue realizada de manera incremental, revisando y justificando las decisiones antes de incorporarlas.

Sesión de trabajo: [enlace a la sesión compartida]