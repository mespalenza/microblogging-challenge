# API Contracts

## 1. Create Tweet

### Endpoint

```http
POST /tweets
```

### Request body

```json
{
  "user_id": "user-123",
  "content": "Este es mi primer tweet"
}
```

### Representación en Go

```go
type TweetRequest struct {
	UserID  string `json:"user_id"`
	Content string `json:"content"`
}
```

### Validaciones y normalización

- `user_id` y `content` deben contener valores válidos.
- Los valores no nulos deben ser strings.
- Un campo ausente o con valor `null` se interpreta como un string vacío y recibe el mismo tratamiento que un campo presente con valor vacío.
- Los campos desconocidos provocan el rechazo del request.
- `user_id` se normaliza quitando espacios iniciales y finales.
- Un `user_id` ausente, vacío o compuesto únicamente por espacios no es válido.
- `content` se normaliza quitando espacios iniciales y finales.
- Se guarda y devuelve el contenido normalizado.
- Un `content` ausente, vacío o compuesto únicamente por espacios no es válido.
- El límite máximo es de 280 runas después de normalizar.
- Los caracteres se cuentan como runas Unicode, no como bytes.

#### Decisión sobre campos ausentes

El DTO HTTP utiliza campos `string`.

Durante la decodificación JSON, Go representa un campo ausente, con valor `null` o presente con valor vacío mediante el zero value `""`.

La API no distingue entre ambos casos porque producen el mismo resultado desde el punto de vista de las reglas del caso de uso:

- `user_id` ausente o vacío devuelve `invalid_user_id`.
- `content` ausente o vacío devuelve `invalid_content`.

Esta decisión simplifica el DTO HTTP y evita agregar estados que no modifican el comportamiento del negocio.

#### Ejemplo de normalización

Request recibido:

```json
{
  "user_id": "  user-123  ",
  "content": "  Hola  "
}
```

Se procesa como:

```json
{
  "user_id": "user-123",
  "content": "Hola"
}
```

### Respuesta exitosa

**Estado:** `201 Created`

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "user-123",
  "content": "Hola",
  "date": "2026-08-08T23:30:00Z"
}
```

#### Decisiones

- `id` es un UUID representado como string y generado por el servidor.
- `date` es un string en formato RFC 3339 y UTC.
- La respuesta contiene el `content` y el `user_id` normalizados.

### Errores

#### `400 Bad Request`

- JSON con sintaxis inválida.
- Campo con tipo incorrecto.
- Campo desconocido.
- Más de un valor JSON o contenido sobrante.
- `user_id` ausente, `null` o vacío devuelve `invalid_user_id`.
- `content` ausente, `null` o vacío devuelve `invalid_content`.
- `content` contiene más de 280 runas después de normalizar.

#### `500 Internal Server Error`

- Error inesperado durante el procesamiento.

#### Formato de error

```json
{
  "error": {
    "code": "content_too_long",
    "message": "content must not exceed 280 characters"
  }
}
```

#### Códigos de error

- `invalid_json`: el JSON tiene sintaxis inválida, contiene más de un valor o tiene contenido sobrante.
- `invalid_field_type`: un campo tiene un tipo incorrecto.
- `unknown_field`: el request contiene un campo desconocido.
- `invalid_user_id`: `user_id` está ausente, es `null`, está vacío o queda vacío después de normalizar.
- `invalid_content`: `content` está ausente, es `null`, está vacío o queda vacío después de normalizar.
- `content_too_long`: `content` supera las 280 runas.
- `internal_error`: ocurrió un error interno inesperado.

Si existen varios errores simultáneos:

- Se devuelve solamente el primero encontrado.
- El orden de validación es determinista.
- Primero se procesan los errores de decodificación.
- Después se procesan los campos desconocidos y los tipos incorrectos.
- Luego el servicio valida `user_id`, `content` y su longitud, en ese orden.


## 2. Follow User

### Endpoint

```http
PUT /users/{follower_id}/following/{followed_id}
```

Ejemplo:

```http
PUT /users/user-123/following/user-456
```

### Request body

No lleva body. Toda la información necesaria está contenida en el path:

- `follower_id`: usuario que comienza a seguir.
- `followed_id`: usuario seguido.

Por los supuestos del challenge, todos los identificadores recibidos se consideran
existentes y válidos. No hay autenticación ni sesiones.

### Respuestas exitosas

Cuando el request crea la relación:

- **Estado:** `201 Created`
- Sin response body.

Cuando la relación ya existía:

- **Estado:** `204 No Content`
- Sin response body.

La diferencia forma parte observable del contrato:

- `201`: ese request creó la relación.
- `204`: la relación ya existía y no fue creada nuevamente.

### Validación

Un usuario no puede seguirse a sí mismo:

```text
follower_id == followed_id
```

**Estado:** `422 Unprocessable Content`

```json
{
  "error": {
    "code": "cannot_follow_self",
    "message": "a user cannot follow themselves"
  }
}
```

#### Error inesperado

**Estado:** `500 Internal Server Error`

```json
{
  "error": {
    "code": "internal_error",
    "message": "an unexpected error occurred"
  }
}
```

### Idempotencia

Aunque originalmente se había considerado `POST`, se eligió `PUT` porque:

- La relación queda completamente identificada por `follower_id` y `followed_id`.
- El servidor no genera un identificador adicional para ella.
- Repetir el mismo request no crea duplicados.
- Ejecutarlo varias veces deja el mismo estado final que ejecutarlo una vez.

### Concurrencia

La combinación `(follower_id, followed_id)` debe ser única y su creación debe ser
atómica.

Si dos requests simultáneos intentan crear la misma relación:

- El que la crea devuelve `201 Created`.
- El que encuentra que ya fue creada devuelve `204 No Content`.
- La violación interna de unicidad no se expone como error al cliente.
- El estado final contiene una sola relación.

## 3. Timeline

### Endpoint

```http
GET /users/{user_id}/timeline
```

Ejemplo del primer request:

```http
GET /users/user-123/timeline?limit=20
```

Ejemplo de página siguiente:

```http
GET /users/user-123/timeline?limit=20&cursor=opaque-cursor-value
```

No lleva request body.

### Contenido

El timeline:

- Incluye tweets de los usuarios seguidos.
- No incluye tweets propios.
- Incluye tweets publicados antes de comenzar a seguir al usuario.
- Devuelve una colección vacía cuando no hay tweets.
- Ordena los tweets desde el más reciente hasta el más antiguo.
- Desempata tweets con la misma fecha utilizando el `id` en orden descendente.
- Usa la misma representación de tweet que `createTweet`.

### Paginación

Parámetro `limit`:

- Es un entero opcional.
- Su valor por defecto es `20`.
- Su valor mínimo es `1`.
- Su valor máximo es `100`.

Parámetro `cursor`:

- Es un string opcional.
- No se envía en el primer request.
- Para la página siguiente se envía el `next_cursor` de la respuesta anterior.
- Es opaco para el cliente.

En la respuesta:

- `next_cursor` es un string cuando existe una página siguiente.
- `next_cursor` es `null` cuando no quedan más páginas.

El cursor representa:

- El `created_at` del último tweet entregado.
- El `tweet_id` del último tweet entregado.
- La posición desde la que debe continuar la página siguiente.

### Respuesta exitosa

**Estado:** `200 OK`

```json
{
  "tweets": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440002",
      "user_id": "user-789",
      "content": "Primer tweet",
      "date": "2026-08-08T23:30:00Z"
    },
    {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "user_id": "user-456",
      "content": "Segundo tweet",
      "date": "2026-08-08T22:15:00Z"
    }
  ],
  "next_cursor": "opaque-cursor-value"
}
```

#### Timeline vacío

**Estado:** `200 OK`

```json
{
  "tweets": [],
  "next_cursor": null
}
```

#### Última página

Se devuelven los tweets restantes y `next_cursor` pasa a `null`:

```json
{
  "tweets": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "user_id": "user-789",
      "content": "Último tweet",
      "date": "2026-08-08T20:00:00Z"
    }
  ],
  "next_cursor": null
}
```

### Consistencia durante la paginación

- Los tweets publicados después de solicitar la primera página normalmente quedan
  antes del cursor y no aparecen en las páginas siguientes.
- Para obtener los tweets nuevos, el cliente debe solicitar nuevamente la primera página.
- No se garantiza una fotografía inmutable de las relaciones de seguimiento.
- Si el usuario comienza a seguir a otra persona durante el recorrido, las páginas
  posteriores pueden reflejar el cambio.
- Dejar de seguir usuarios está fuera del alcance.

### Errores

#### Parámetros con formato o tipo incorrecto

- **Estado:** `400 Bad Request`
- **Código:** `invalid_pagination_parameters`
- **Mensaje:** `pagination parameters have an invalid format`

#### `limit` fuera del rango 1–100

- **Estado:** `422 Unprocessable Content`
- **Código:** `limit_out_of_range`
- **Mensaje:** `limit must be between 1 and 100`

#### Cursor inválido

- **Estado:** `400 Bad Request`
- **Código:** `invalid_cursor`
- **Mensaje:** `the provided cursor is invalid`

#### Error inesperado

- **Estado:** `500 Internal Server Error`
- **Código:** `internal_error`
- **Mensaje:** `an unexpected error occurred`

Todos los errores siguen el formato común:

```json
{
  "error": {
    "code": "stable_error_code",
    "message": "human-readable message"
  }
}
```