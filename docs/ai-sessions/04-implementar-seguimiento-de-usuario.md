# Implementar seguimiento de usuarios

> Bueno chat, vamos a implementar el siguiente requisito del challenge: seguimiento de usuarios
> PUT /users/{follower_id}/following/{followed_id}
> Trabajemos sobre => /Users/emispalenza/GolandProjects/microblogging-challenge
> Revisá el PDF original, docs/architecture.md y docs/api-contracts.md
> Quiero mantener la implementación acotada a los requisitos funcionales
> Guiame paso a paso. Yo voy a escribir el código y después quiero una revisión con criterio Senior. Los tests , coverage quedan para la etapa final

<details><summary>2 previous messages</summary>

> Voy a revisar primero el PDF original y los dos documentos de arquitectura/contratos. Como el PDF forma parte de la fuente de requisitos, usaré la skill de PDF para extraerlo y contrastarlo; después te propongo un primer paso pequeño, sin escribir código ni adelantarnos con tests.
>
> No aparece ningún PDF dentro del repositorio; sí hay un `business.txt`, que probablemente sea la extracción del enunciado original. Voy a verificar su contenido junto con los documentos y el estado actual del módulo `follow`, sin modificar nada.
</details>

Revisé [architecture.md](~/GolandProjects/microblogging-challenge/docs/architecture.md), [api-contracts.md](~/GolandProjects/microblogging-challenge/docs/api-contracts.md) y `business.txt`. No encontré el PDF original en el repositorio ni dentro de `/Users/emispalenza/GolandProjects`; si está en otra ubicación, después pasame la ruta.

El alcance funcional de este endpoint es:

- Crear una relación dirigida `follower_id → followed_id`.
- Impedir el auto-seguimiento.
- No duplicar relaciones.
- Informar si fue creada (`201`) o ya existía (`204`).
- Convertir errores inesperados en `500`.
- Garantizar atomicidad ante concurrencia.
- No validar existencia de usuarios, autenticar, dejar de seguir ni añadir funcionalidades extra.

También detecté cambios tuyos sin confirmar en `internal/follow`; los preservaré.

## Paso 1: terminar el servicio de dominio

La estructura elegida es correcta:

- `Follow` representa la relación.
- `Repository.SaveFollow(...) (bool, error)` comunica si la relación fue creada.
- `Service.FollowUser(...)` aplica la regla de auto-seguimiento y delega la persistencia.

Hay un error puntual en [service.go](~/GolandProjects/microblogging-challenge/internal/follow/service.go:16):

```go
if value.FollowedID == value.FollowedID {
```

Esa condición siempre es verdadera. Cambiala para comparar los dos participantes:

```go
if value.FollowerID == value.FollowedID {
	return false, ErrCannotFollowSelf
}
```

No agregaría validaciones para identificadores vacíos, normalización ni comprobación de usuarios: no están requeridas y los documentos dicen que los identificadores recibidos se consideran válidos.

Hacé solamente ese cambio. Cuando me digas que está listo, reviso tu implementación y avanzamos al paso 2: el repositorio en memoria, especialmente su idempotencia y seguridad ante concurrencia.

> Ya esta el primer paso, si esta ok voy por el repository

<details><summary>1 previous message</summary>

> Voy a revisar únicamente el cambio del servicio y su coherencia con el contrato, sin ejecutar todavía tests ni coverage.
</details>

El paso 1 está correcto. La validación compara ambos usuarios, evita llamar al repositorio en caso de auto-seguimiento y propaga correctamente los errores.

## Paso 2: repositorio en memoria

Creá:

`internal/platform/memory/follow_repository.go`

Implementá un repositorio con estas características:

- Un `map[follow.Follow]struct{}` para almacenar relaciones únicas.
- Un `sync.Mutex` para proteger el mapa.
- Un constructor `NewFollowRepository()`.
- El método:

```go
SaveFollow(ctx context.Context, value follow.Follow) (bool, error)
```

Comportamiento:

1. Adquirir el lock.
2. Consultar si la relación ya existe.
3. Si existe, devolver `false, nil`.
4. Si no existe, guardarla y devolver `true, nil`.
5. Liberar el lock mediante `defer`.

La comprobación y la inserción deben ocurrir bajo el mismo lock. Eso garantiza que dos requests concurrentes no puedan obtener ambos `true`.

Para mantenernos acotados:

- No hace falta validar auto-seguimiento aquí; es responsabilidad del servicio.
- No hace falta generar IDs ni timestamps.
- No hace falta modelar usuarios.
- No hace falta agregar todavía métodos de lectura para el timeline.
- No agregues lógica especial para `context.Context`; mantenelo porque forma parte del contrato del repositorio.

Cuando lo tengas, lo reviso antes de pasar al handler HTTP.

> Revisa el paso dos, ya lo complete. Voy haciendo el paso 3 del handler.

<details><summary>1 previous message</summary>

> Reviso el repositorio con foco en idempotencia, atomicidad y ajuste al alcance. Después te dejo el paso 3 delimitado para que implementes el handler sin mezclar aún el routing ni el `main`.
</details>

El repositorio está correcto y cumple el paso 2:

- La clave del mapa representa exactamente `(follower_id, followed_id)`.
- La repetición devuelve `false, nil`.
- El lock cubre consulta e inserción, por lo que la operación es atómica.
- No incorporaste responsabilidades fuera de alcance.

`relation` podría llamarse `relations`, pero es solamente estilo; no hace falta cambiarlo.

## Paso 3: handler HTTP

Veo que empezaste copiando la estructura del handler de tweets. En este endpoint hay una diferencia importante: no existe JSON de entrada ni salida.

### 1. No necesitás DTO

`internal/rest/follow/dto.go` no aporta nada. Los dos valores vienen directamente del path y pueden convertirse al dominio dentro del handler:

```go
value := domainfollow.Follow{
	FollowerID: r.PathValue("follower_id"),
	FollowedID: r.PathValue("followed_id"),
}
```

Podés eliminar el `dto.go` vacío.

### 2. Definí el puerto del servicio

Como el paquete HTTP y el dominio se llaman `follow`, usá un alias:

```go
import domainfollow "github.com/mespalenza/microblogging-challenge/internal/follow"
```

La interfaz del handler debería recibir el tipo del dominio:

```go
type FollowService interface {
	FollowUser(context.Context, domainfollow.Follow) (bool, error)
}
```

`FollowService` resulta más claro que `FollowUser`, porque evita que el nombre de la interfaz parezca una acción.

### 3. Flujo del método

`Handler.FollowUser` debe:

1. Obtener ambos path parameters.
2. Construir `domainfollow.Follow`.
3. Ejecutar `h.service.FollowUser(r.Context(), value)`.
4. Mapear cualquier error con `errorcatalog.From`.
5. Si `created == true`, responder `201 Created`.
6. Si `created == false`, responder `204 No Content`.

En ambos casos exitosos:

- No codifiques JSON.
- No escribas response body.
- No hace falta establecer `Content-Type`.

La decisión de estado puede quedar así:

```go
if created {
	w.WriteHeader(http.StatusCreated)
	return
}

w.WriteHeader(http.StatusNoContent)
```

### 4. Agregá el error al catálogo

`errorcatalog.From` debe reconocer `domainfollow.ErrCannotFollowSelf` y producir:

```go
Status:  http.StatusUnprocessableEntity,
Code:    "cannot_follow_self",
Message: "a user cannot follow themselves",
```

Mantené el caso por defecto como `500 internal_error`.

Por ahora no implementes `routes.go` ni conectes el handler en `main.go`. Terminá solamente handler y catálogo de errores; después los revisamos como paso 3.

> Listo

<details><summary>1 previous message</summary>

> Voy a revisar el handler y el mapeo de errores contra el contrato. Me voy a fijar también en detalles HTTP sutiles —cuerpo vacío, status y separación de capas— sin entrar todavía en tests.
</details>

El paso 3 está funcionalmente correcto:

- Lee ambos parámetros desde el path.
- Construye la relación de dominio.
- Delega al servicio.
- Responde `201` si fue creada.
- Responde `204` si ya existía.
- No escribe cuerpo en respuestas exitosas.
- Mapea auto-seguimiento a `422` con el código y mensaje exactos.
- Conserva el fallback `500`.

También adelantaste correctamente `routes.go`; el patrón coincide con el contrato.

Como mejora menor, había sugerido `FollowService` para la interfaz y un alias como `domainfollow` para distinguir dominio de transporte. Tu versión compila conceptualmente y no tiene un problema funcional, así que no es necesario frenarnos por eso.

## Paso 4: conectar el endpoint

Ahora modificá `cmd/api/main.go`.

### Construí las dependencias

Además de las dependencias de tweets, necesitás una única instancia de:

```text
FollowRepository → follow.Service → rest/follow.Handler
```

Usá aliases para diferenciar dominio y HTTP, por ejemplo:

```go
domainfollow "github.com/mespalenza/microblogging-challenge/internal/follow"
restfollow "github.com/mespalenza/microblogging-challenge/internal/rest/follow"
```

### Usá un mux raíz compartido

No pongas directamente `followHandler.Routes()` como handler del servidor, porque perderías la ruta de creación de tweets. Creá un `http.NewServeMux()` y registrá ambos métodos:

```go
router := http.NewServeMux()

router.HandleFunc("POST /tweets", tweetHandler.CreateTweet)
router.HandleFunc(
	"PUT /users/{follower_id}/following/{followed_id}",
	followHandler.FollowUser,
)
```

Después:

```go
server := &http.Server{
	Addr:              ":8080",
	Handler:           router,
	ReadHeaderTimeout: 5 * time.Second,
}
```

La instancia de `FollowRepository` debe mantenerse viva durante toda la ejecución; no la construyas dentro del handler, porque se perderían las relaciones entre requests.

Hacé solamente el wiring de `main.go`. Después hacemos una revisión funcional completa y una prueba manual mínima; los tests automatizados y coverage siguen reservados para la etapa final.

> Listo, implementé el seguimiento de usuarios.
> Agregué el modelo y el servicio de follow, con la validación para que un usuario no pueda seguirse a sí mismo.
> También hice el repositorio en memoria usando un map y un mutex, para evitar relaciones duplicadas y asegurar que el check y la inserción sean atómicos.
> Implementé el endpoint:
> PUT /users/{follower_id}/following/{followed_id}
> Devuelve 201 cuando crea la relación, 204 cuando ya existía y 422 si el usuario intenta seguirse a sí mismo.
> Finalmente conecté el handler al router principal. Probé los tres casos manualmente y también pasan go vet ./... y go test ./....
> Dejé los tests más completos, coverage para la etapa final, como quedamos
> Me revisás la implementación antes de que avance con el timeline?

<details><summary>1 previous message</summary>

> Sí. Voy a hacer una revisión tipo PR con criterio Senior: primero buscaré defectos funcionales o de concurrencia, después coherencia arquitectónica y finalmente detalles de mantenibilidad. No voy a modificar archivos ni sumar tests en esta etapa.
</details>

No encontré defectos bloqueantes ni desviaciones del contrato. La implementación está en condiciones de avanzar al timeline.

### Revisión Senior

- Dominio: `Follow` modela correctamente una relación dirigida.
- Servicio: la regla de auto-seguimiento está en la capa adecuada.
- Repositorio: el `Mutex` cubre conjuntamente check e inserción, garantizando atomicidad.
- Idempotencia: la primera llamada devuelve `true`; las siguientes, `false`.
- HTTP: los estados `201`, `204`, `422` y `500` están correctamente resueltos.
- Respuestas exitosas: no incluyen cuerpo, como exige el contrato.
- Routing: `RegisterRoutes` permite compartir un único `ServeMux` sin perder el endpoint de tweets.
- Wiring: el repositorio se instancia una sola vez y conserva las relaciones entre requests.
- Alcance: no se agregaron validaciones de usuarios ni funcionalidades innecesarias.
- El working tree está limpio y la implementación quedó en el commit `bd96ac0`.

Detalles menores, no bloqueantes:

- `FollowUser` como nombre de interfaz puede confundirse con el método; `FollowService` sería algo más expresivo.
- `relation` podría ser `relations` porque contiene una colección.
- `context.Context` todavía no se usa en el repositorio en memoria, pero conservarlo en la interfaz es correcto.

Intenté repetir `go vet ./...` y `go test ./...`, pero este entorno aislado tiene una instalación de Go/GOROOT inconsistente y además bloquea su caché externa; los errores aparecen incluso al localizar paquetes estándar como `context` y `errors`, por lo que no señalan un problema de tu código. Tomo como válida tu ejecución local satisfactoria.

Los tests finales deberían cubrir especialmente la carrera entre dos inserciones iguales, además de servicio y handler. Eso puede esperar a la etapa acordada.

Podés avanzar con el timeline.
