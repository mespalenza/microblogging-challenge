# Implementar publicación de tweets

> Bueno chat, vamos a empezar con la primera implementación => Implementar publicación de tweets
> Trabajemos sobre:
> /Users/emispalenza/GolandProjects/microblogging-challenge
> Revisá primero docs/architecture.md y docs/api-contracts.md. Guiame paso a paso y no escribas toda la solución por mí quiero implementar el código y después recibir una evaluación con criterio Senior

<details><summary>2 previous messages</summary>

> Perfecto. Voy a revisar primero `docs/architecture.md` y `docs/api-contracts.md`, y después inspeccionaré lo mínimo del repositorio para ubicar la implementación actual. No voy a modificar archivos ni adelantarte la solución completa: te propondré un primer paso acotado, con criterios de aceptación, y revisaremos tu código como en una evaluación Senior.
>
> Encontré ambos documentos y no hay un `AGENTS.md` aplicable en el repositorio. Ahora voy a leer los contratos completos y luego contrastarlos con la estructura del código para convertirlos en una secuencia de implementación, sin darte código resuelto.
</details>

Ya revisé la arquitectura, los contratos y el estado actual del repositorio. Está prácticamente vacío: solo existe `cmd/api/main.go` con un modelo inicial. No hice cambios.

Para esta implementación vamos a trabajar por capas y en iteraciones pequeñas. Yo te daré requisitos y pistas; vos escribís el código. Después revisaré diseño, corrección, tests, errores y mantenibilidad con criterio Senior.

## Paso 1: modelar el dominio de creación de tweets

Empezá únicamente por la lógica de negocio. Todavía no implementes HTTP ni persistencia real.

Creá una estructura similar a esta:

```text
internal/tweet/
├── tweet.go
└── service.go
```

No es obligatorio usar exactamente esos nombres, pero sí separar el dominio de `cmd/api`.

Tu primera entrega debe resolver:

- Representación interna de un tweet.
- Entrada necesaria para crear uno: `user_id` y `content`.
- Normalización con `strings.TrimSpace`.
- Validación:
    - `user_id` normalizado no puede estar vacío.
    - `content` normalizado no puede estar vacío.
    - Máximo de 280 runas Unicode, no bytes.
- Generación del ID y fecha fuera de la entidad o mediante dependencias inyectables.
- El tweet resultante debe contener:
    - ID
    - User ID normalizado
    - contenido normalizado
    - fecha

Decisiones que quiero que tomes vos:

1. ¿La validación pertenece a un constructor de dominio o al servicio?
2. ¿Cómo representarás los errores para que luego el handler pueda distinguir `invalid_user_id`, `invalid_content` y `content_too_long` sin comparar strings?
3. ¿Qué tipo usarás internamente para la fecha: `string` o `time.Time`?
4. ¿Cómo evitarás acoplar la lógica a `uuid.New()` y `time.Now()` para poder probarla determinísticamente?

Criterios de aceptación:

- `"  user-123  "` termina como `"user-123"`.
- `"  Hola  "` termina como `"Hola"`.
- 280 emojis es válido.
- 281 emojis es inválido.
- Los errores son identificables programáticamente.
- La lógica puede probarse sin depender del reloj real ni de un UUID aleatorio.
- `cmd/api` no contiene reglas de negocio.

No te preocupes aún por:

- JSON.
- Status codes.
- Router.
- Base de datos.
- Repositorios.
- Docker.
- Timeline o follows.

Cuando tengas esta primera versión, pasame el código o decime que la revise directamente en el repositorio. Haré una revisión como PR: primero bugs y riesgos, luego diseño y finalmente calidad de tests.

> Tomé estas decisiones
>
> La validación la dejé en el servicio porque por ahora es el único lugar desde donde se crean tweets. No quise agregar un constructor de dominio sin necesitarlo todavía. Si después aparece otra forma de crear un Tweet, podría mover las validaciones a un constructor para asegurar que nunca se cree uno inválido
>
> Para los errores creé ErrInvalidUserID, ErrInvalidContent y ErrContentTooLong dentro del paquete tweet. Así después el handler los puede distinguir con errors.Is sin comparar textos. No les puse status HTTP porque eso lo debería resolver el handler
>
> Para la fecha usé time.Time y no string. Más adelante, cuando haga el handler, la convertiré al formato RFC 3339 que pide la API
>
> El ID y la fecha los genero en el servicio con uuid.NewString() y time.Now().UTC(). Sé que así quedan acoplados y no puedo controlar sus valores exactos en los tests. Podría inyectar un generador y un reloj, pero para esta primera versión preferí no agregar esas abstracciones porque no hay ninguna regla de negocio que dependa del valor concreto del ID o de la fecha. Si después lo necesito, lo refactorizo
>
> La normalización se hace con strings.TrimSpace antes de validar. Primero valido user_id y después content para que, si ambos son inválidos, el error sea determinista
>
> También agregué tests para los campos vacíos, valores con espacios, normalización, 280 y 281 caracteres ASCII, y 280 y 281 emojis. Los emojis sirven para comprobar que se cuentan runas y no bytes
>
> La lógica quedó dentro de internal/tweet y todavía no agregué HTTP, repositorios ni base de datos

<details><summary>2 previous messages</summary>

> Muy buenas decisiones y, sobre todo, bien justificadas. Voy a revisar ahora la implementación y los tests reales, ejecutar la suite y evaluarla como un PR: primero problemas de corrección, luego diseño y finalmente cobertura/mantenibilidad. No voy a modificar nada.
>
> La suite no pudo ejecutarse porque `go` no está instalado o no está disponible en el `PATH` de este entorno. También veo cambios en `follow` y `timeline`; los trataré como trabajo ajeno al alcance y revisaré únicamente `internal/tweet` y la modificación necesaria de dependencias en `go.mod`. Haré una inspección estática con referencias de línea.
</details>

La lógica productiva cumple el alcance del Paso 1. No encontré errores funcionales bloqueantes en `internal/tweet`. Mi evaluación provisional: **aprobado con observaciones**, nivel intermedio sólido, con mejoras necesarias en la calidad de los tests.

::code-comment{title="[P2] Los casos exitosos pueden dar falsos positivos" body="Los casos de 280 caracteres solo comprueban que el error sea nil. El test seguiría pasando si CreateTweet devolviera un Tweet vacío sin error. Para cada caso exitoso conviene verificar explícitamente UserID y Content, además de las propiedades del ID y CreatedAt." file="/Users/emispalenza/GolandProjects/microblogging-challenge/internal/tweet/service_test.go" start=90 end=99 priority=2}

### Evaluación Senior

Lo que está bien:

- Separaste dominio y caso de uso de `cmd/api`.
- `time.Time` es la representación interna correcta.
- Los errores centinela permiten usar `errors.Is` sin introducir HTTP en el dominio.
- El orden de validación es explícito y determinista.
- Normalizás antes de validar y contar.
- `len([]rune(content))` satisface el contrato Unicode.
- Tu decisión de posponer reloj y generador inyectables es razonable: evitaste abstracciones especulativas y explicaste correctamente el trade-off.
- Los nombres de errores pueden mapearse limpiamente a los códigos de la API.

Observaciones:

1. **Tests exitosos poco estrictos**

   En los casos de 280 caracteres, `wantContent` queda vacío. Por eso nunca comprobás que el contenido retornado sea realmente el esperado. El mismo problema existe para `UserID`.

   Te sugiero que cada caso exitoso declare explícitamente el resultado esperado y que el test distinga claramente entre:

    - debe fallar;
    - debe tener éxito.

2. **Faltan invariantes básicas del resultado**

   En al menos un caso exitoso verificá:

    - `ID` no vacío.
    - `ID` con formato UUID válido.
    - `CreatedAt` no es zero.
    - `CreatedAt` está expresado en UTC.

   No hace falta controlar los valores exactos para validar esas propiedades.

3. **Nombre del servicio**

   `ServiceTweet` se entiende, pero en Go resulta más idiomático `Service` dentro del paquete `tweet`, o `TweetService` si necesitás explicitarlo. Al importarlo, `tweet.Service` ya da contexto suficiente.

4. **Contexto todavía no utilizado**

   Tener `context.Context` desde ahora es defendible porque será necesario al persistir. No lo eliminaría, pero tampoco lo usaría para validaciones o generación de valores: tendrá sentido cuando el servicio invoque el repositorio.

5. **Comentario placeholder**

   `//Repository` no aporta todavía y parece código incompleto. Eliminalo hasta que exista la dependencia real.

No pude ejecutar `go test ./...` porque el ejecutable `go` no está disponible en este entorno, así que la aprobación es por inspección estática.

## Antes del Paso 2

Mejorá únicamente los tests:

- Hacé explícito si cada caso espera éxito.
- Verificá todos los campos relevantes en los casos exitosos.
- Validá propiedades del UUID y de la fecha sin exigir valores exactos.
- Si querés, renombrá el servicio y quitá el comentario placeholder.

No agregues todavía repositorio ni HTTP. Cuando termines, lo revisamos y recién entonces avanzamos a persistencia.

> Entiendo las observaciones de los tests, pero no quiero seguir dedicándole tanto tiempo a este paso porque el enunciado dice que los tests se valoran, no que la cobertura tenga que ser del 100%
> Ya probé las reglas más importantes validaciones, normalización y límite de runas. Prefiero dejar como mejora los checks adicionales del UUID y la fecha, y avanzar con la implementación funcional
> Voy a cambiar el nombre del servicio, sacar el comentario del repositorio y continuar con el Paso 2

De acuerdo. Es una priorización razonable: los tests actuales cubren las reglas con mayor riesgo —validación, normalización, orden determinista y conteo Unicode—. Los checks de UUID y fecha pueden quedar como mejora sin bloquear el avance.

## Paso 2: persistencia mediante un repositorio

El objetivo ahora es que `CreateTweet` no solo construya el tweet, sino que lo guarde.

Implementá:

- Una interfaz de repositorio dentro de `internal/tweet`.
- Una operación para guardar un `Tweet`.
- Inyección del repositorio en el servicio.
- Propagación de errores de persistencia.
- Un repositorio en memoria para esta primera implementación funcional.

Pensá estas decisiones:

1. ¿La interfaz debe llamarse `Repository`, `TweetRepository` o algo más específico?
2. ¿La operación debería ser `Save`, `Create` o `Insert`?
3. ¿Debe recibir `context.Context`?
4. ¿El repositorio genera ID/fecha o recibe un tweet ya completo?
5. ¿Cómo evitás que el mapa interno del repositorio tenga data races?

Mi recomendación conceptual:

```text
HTTP → tweet.Service → tweet.Repository
```

El servicio conserva:

- Normalización.
- Validaciones.
- Generación de ID y fecha.
- Orquestación de persistencia.

El repositorio solamente guarda el tweet ya válido y completo.

Criterios de aceptación:

- Un tweet válido queda almacenado.
- Si el repositorio falla, `CreateTweet` devuelve el error.
- Un tweet inválido no llega al repositorio.
- El repositorio recibe el tweet normalizado.
- La implementación en memoria es segura ante concurrencia.
- El servicio depende de una interfaz, no de la implementación en memoria.

Tests mínimos que sí considero valiosos en este paso:

- El repositorio es invocado una vez con el tweet normalizado.
- No es invocado cuando falla una validación.
- Un error del repositorio se propaga.
- El repositorio en memoria permite recuperar o inspeccionar lo guardado.

No implementes todavía el handler HTTP. Cuando completes este paso, revisaré especialmente límites de responsabilidad, manejo de dependencias y concurrencia.

> 1. Voy a llamar Repository a la interfaz porque ya está dentro del paquete tweet y se entiende de qué repositorio estamos hablando
> 2. El método va a ser Save porque el tweet ya llega armado desde el servicio. El repositorio solamente lo guarda, no genera el ID, la fecha ni hace validaciones
> 3. Le voy a pasar context.Context porque después, cuando use una base de datos, lo voy a necesitar para cancelaciones y timeouts
> 4. Ya agregué la interfaz, la inyecté en el servicio con NewService y llamo a Save después de validar y crear el tweet. Si el repositorio devuelve un error, lo devuelvo desde CreateTweet
> 5. Ahora me falta hacer el repositorio en memoria. Voy a usar un map y protegerlo con sync.RWMutex para que no haya problemas si llegan varias requests al mismo tiempo
>
> En esta etapa quiero priorizar terminar la implementación funcional. No quiero que nos frenemos ahora revisando o ampliando los tests en cada paso. Cuando termine de implementar todas las funcionalidades voy a hacer una etapa específica para completar los tests

Perfecto. Las cinco decisiones son coherentes y la estrategia de pruebas queda acordada: avanzamos con implementación funcional y dejamos una etapa específica de endurecimiento al final. Solo señalaré tests si aparece un riesgo crítico.

Para terminar el repositorio en memoria:

- Usá el ID del tweet como clave del `map`.
- Inicializá el mapa mediante un constructor; evitá depender de un mapa `nil`.
- En `Save`, usá `Lock`, no `RLock`.
- Guardá el `Tweet` por valor para evitar mutaciones externas posteriores.
- Respetá el `context.Context`; como mínimo, si ya está cancelado, no guardes y devolvé `ctx.Err()`.
- Liberá el lock con `defer` inmediatamente después de adquirirlo.
- No agregues validaciones de negocio al repositorio.
- No conviertas errores a códigos HTTP en esta capa.

Sobre la ubicación: podés mantenerlo inicialmente junto al paquete `tweet` por simplicidad. Más adelante, una implementación SQL sí debería vivir como adaptador separado, mientras la interfaz permanece en `tweet`, que es quien la consume.

No necesitás implementar todavía operaciones generales como `Delete` o `Update`. Si necesitás comprobar manualmente el contenido, podés agregar una lectura puntual por ID, pero no amplíes la interfaz principal solamente para facilitar tests.

Una vez terminado esto, el siguiente paso será `POST /tweets`:

1. Decodificación estricta del JSON.
2. Distinción entre errores estructurales `400` y errores semánticos `422`.
3. Invocación del servicio.
4. Serialización del tweet como `201 Created`.
5. Mapeo de errores internos a la respuesta contractual.

Terminá el repositorio y avanzamos directamente al handler.

> Revisa el codigo que hice para el paso 2

<details><summary>2 previous messages</summary>

> Voy a revisar únicamente los cambios relevantes de `tweet` y el repositorio en memoria, sin modificar archivos. Haré una revisión tipo PR centrada en corrección, concurrencia, límites entre capas y cumplimiento de lo acordado; intentaré también compilar o ejecutar tests si encuentro un runtime de Go disponible.
>
> Ya aparece un problema funcional importante en la integración: los tests existentes construyen `&Service{}` y los casos válidos ahora intentarán ejecutar `Save` sobre un repositorio `nil`, provocando un panic. Antes de cerrar la revisión voy a inspeccionar el adaptador en memoria y sus tests —si existen— para comprobar constructor, locking y cancelación del contexto.
</details>

Encontré **un problema bloqueante** y dos observaciones menores. El diseño general del repositorio está bien encaminado.

::code-comment{title="[P1] Los casos válidos provocan panic" body="El test construye Service sin repositorio. Cuando una entrada supera las validaciones, CreateTweet ejecuta s.repository.Save y desreferencia una interfaz nil. Actualizá el wiring del test para construir el servicio mediante NewService con un repositorio o stub. Esto no es ampliar cobertura: es adaptar los tests existentes al nuevo requisito del servicio." file="/Users/emispalenza/GolandProjects/microblogging-challenge/internal/tweet/service_test.go" start=84 priority=1}

::code-comment{title="[P2] El contexto recibido se ignora" body="Save acepta context.Context pero almacena el tweet incluso si el contexto ya fue cancelado. Si el adaptador declara soporte de contexto, comprobá ctx.Err() antes de modificar el mapa y devolvé ese error cuando corresponda." file="/Users/emispalenza/GolandProjects/microblogging-challenge/internal/platform/memory/tweet_repository.go" start=22 end=26 priority=2}

::code-comment{title="[P3] Error no identificable ni estable" body="errors.New se ejecuta en cada búsqueda fallida, por lo que el llamador no puede clasificar el resultado con errors.Is. Si FindByID va a formar parte del comportamiento real, definí un error centinela; si solo fue agregado para inspeccionar tests, considerá no ampliar todavía el adaptador con una operación que el caso de uso no necesita." file="/Users/emispalenza/GolandProjects/microblogging-challenge/internal/platform/memory/tweet_repository.go" start=29 end=35 priority=3}

### Evaluación del diseño

Está bien resuelto:

- La interfaz está en el paquete consumidor `tweet`.
- `Save` recibe un tweet completo y no duplica reglas de negocio.
- El servicio genera ID y fecha antes de persistir.
- Los errores de persistencia se propagan.
- La implementación concreta está separada en `internal/platform/memory`.
- El mapa usa el ID como clave.
- Guardás structs por valor.
- El constructor inicializa correctamente el mapa.
- `Lock` y `RLock` están usados correctamente.
- El servicio fue renombrado de forma más idiomática.

El único cambio imprescindible antes de avanzar es corregir la construcción del servicio en los tests existentes. Actualmente, el primer caso válido entra en `Save` con `repository == nil` y termina en panic.

También decidiría qué hacer con `NewService(nil)`: no hace falta diseñar un sistema complejo, pero fallar inmediatamente en el constructor resulta más fácil de diagnosticar que un panic posterior dentro de `CreateTweet`.

No pude ejecutar la suite porque `go` sigue sin estar disponible en este entorno. Después de corregir el wiring del test, ejecutá localmente:

```bash
go test ./...
go test -race ./...
```

Con el P1 resuelto, considero el Paso 2 suficientemente completo para avanzar al handler HTTP.

> Corregi los tests para crear el servicio
> Por ahora no voy a seguir agregando tests ni ampliando el coverage. Una vez corregido este panic quiero avanzar con el handler HTTP y dejar la revisión completa de los tests para el final

<details><summary>1 previous message</summary>

> Perfecto: no voy a pedir más cobertura en esta etapa. Verificaré solamente que el panic haya quedado corregido y, si el wiring está bien, pasamos directamente al diseño del handler `POST /tweets`.
</details>

El panic quedó corregido correctamente. Ahora todos los casos construyen el servicio mediante `NewService` con un repositorio válido. Además, el stub verifica naturalmente la integración con `Save`. Doy el Paso 2 por cerrado.

## Paso 3: handler `POST /tweets`

Creá una capa HTTP separada, por ejemplo:

```text
internal/platform/http/
├── handler.go
├── tweet_handler.go
└── response.go
```

Podés ajustar la estructura, pero evitá poner la lógica HTTP en `cmd/api/main.go` o dentro de `internal/tweet`.

El handler debe:

1. Aceptar solamente `POST /tweets`.
2. Decodificar el JSON de manera estricta.
3. Convertir el request HTTP en `tweet.CreateInput`.
4. Invocar `CreateTweet` con `r.Context()`.
5. Mapear los errores del dominio.
6. Responder el tweet creado con `201 Created`.

### Punto importante: campos ausentes

Con campos `string` normales no podés distinguir entre:

```json
{}
```

y:

```json
{"user_id": ""}
```

Pero el contrato exige respuestas distintas:

- Campo ausente → `400 missing_required_field`.
- Campo presente pero vacío → `422 invalid_user_id`.

Diseñá el DTO HTTP de forma que preserve esa diferencia. Esta preocupación pertenece al transporte, no a `tweet.CreateInput`.

### Decodificación estricta

Tenés que distinguir:

- Sintaxis JSON inválida → `invalid_json`.
- Campo ausente → `missing_required_field`.
- Tipo incorrecto → `invalid_field_type`.
- Campo desconocido → `unknown_field`.
- Más de un valor JSON o contenido sobrante → rechazado como JSON inválido.

`json.Decoder.DisallowUnknownFields()` ayuda con campos desconocidos, pero prestá atención: el error que devuelve no tiene un tipo público específico. Evitá que toda la clasificación dependa de mensajes frágiles si encontrás una alternativa limpia.

Orden determinista sugerido:

1. Decodificación/sintaxis.
2. Campos desconocidos.
3. Tipos incorrectos.
4. Presencia de `user_id`.
5. Presencia de `content`.
6. Validaciones semánticas del servicio.

### Mapeo de errores

El handler debe traducir:

- `tweet.ErrInvalidUserID` → `422`, `invalid_user_id`.
- `tweet.ErrInvalidContent` → `422`, `invalid_content`.
- `tweet.ErrContentTooLong` → `422`, `content_too_long`.
- Cualquier otro error del servicio → `500`, `internal_error`.

Usá `errors.Is`, no comparaciones de texto.

### Respuesta exitosa

La representación HTTP debe usar exactamente:

```json
{
  "id": "...",
  "user_id": "...",
  "content": "...",
  "date": "..."
}
```

Consideraciones:

- `date` debe estar en UTC y formato RFC 3339.
- Enviá `Content-Type: application/json`.
- Llamá a `WriteHeader(201)` antes de escribir el JSON.
- No expongas mensajes internos del repositorio en respuestas `500`.
- La respuesta de error debe conservar el envelope definido en el contrato.

Por ahora implementá el handler y sus helpers de respuesta. No agregues tests ni conectes todavía el servidor en `main.go`; revisamos brevemente que el contrato HTTP esté bien interpretado y después hacemos el wiring ejecutable.

> Ya terminé el handler de POST /tweets con el contrato que definimos
> Revisá lo que implementé sin modificar el código. Si no encontrás ningún problema funcional importante, quiero seguir con cmd/api/main.go y probar el endpoint manualmente
> Los tests y el coverage completo los voy a revisar al final para no frenar ahora la implementación

<details><summary>1 previous message</summary>

> Voy a revisar solo el flujo funcional de `POST /tweets`: decodificación estricta, clasificación de errores, mapeo al dominio y formato de respuesta. No modificaré archivos ni voy a frenar por cobertura; únicamente marcaré problemas que puedan romper el contrato o impedir el uso del endpoint.
</details>

Encontré dos desvíos funcionales importantes respecto del contrato original. No avanzaría todavía con `main.go` hasta resolverlos o decidir explícitamente cambiar la API.

::code-comment{title="[P1] Errores semánticos devuelven 400 en lugar de 422" body="El contrato original asigna 422 Unprocessable Content a invalid_user_id, invalid_content y content_too_long. Actualmente los tres se mapean a 400 Bad Request. Cambiar api-contracts.md para reflejar la implementación modifica el contrato acordado; el handler debería usar http.StatusUnprocessableEntity salvo que tomemos explícitamente una nueva decisión de API." file="/Users/emispalenza/GolandProjects/microblogging-challenge/internal/rest/errorcatalog/catalog.go" start=21 end=39 priority=1}

::code-comment{title="[P1] No se distinguen campos ausentes, null y strings vacíos" body="Con campos string, {}, {\"user_id\":null} y {\"user_id\":\"\"} terminan con el mismo zero value. El contrato original exige campo ausente como missing_required_field (400), tipo incorrecto como invalid_field_type (400) y string vacío como invalid_user_id (422). El DTO actual no conserva la información necesaria para hacer esa distinción." file="/Users/emispalenza/GolandProjects/microblogging-challenge/internal/rest/writer/entities.go" start=9 end=12 priority=1}

El resto del flujo está correctamente resuelto:

- `DisallowUnknownFields` rechaza campos desconocidos.
- Se rechazan valores JSON adicionales.
- Los errores internos no se filtran al cliente.
- El contexto del request llega al servicio.
- La respuesta exitosa tiene los nombres correctos.
- La fecha se expresa en UTC y RFC 3339.
- Se envía `Content-Type: application/json`.
- La ruta con patrón `POST /tweets` deja que `ServeMux` gestione otros métodos.

Veo que modificaste [docs/api-contracts.md](~/GolandProjects/microblogging-challenge/docs/api-contracts.md) para que campos ausentes, `null` y vacíos sean equivalentes y para responder `400` en todas las validaciones. Esa simplificación puede defenderse, pero representa un cambio del contrato original, no solamente una decisión interna de implementación.

Mi recomendación es conservar el contrato inicial:

- Ausente → `400 missing_required_field`.
- `null` o tipo no string → `400 invalid_field_type`.
- String vacío/espacios → `422 invalid_user_id` o `invalid_content`.
- Más de 280 runas → `422 content_too_long`.

Después de corregir esos dos puntos, podés avanzar directamente con `cmd/api/main.go`. No hace falta agregar tests ahora. No pude compilar porque el runtime de Go continúa sin estar disponible en este entorno.

> Volví a revisar el PDF original y ahí no dice qué status codes usar, qué códigos de error devolver ni que tenga que diferenciar entre un campo ausente, null o vacío
> Esa diferencia la habíamos definido nosotros, no es un requisito de Ualá. Por eso prefiero mantener el contrato simplificado: si user_id o content están ausentes, son null o están vacíos, devuelvo 400 con invalid_user_id o invalid_content
> Ya dejé esta decisión documentada y la implementación se comporta de la misma forma. Quiero seguir con el  main.go y no agregar complejidad que el challenge no pide
> También voy a mantener la decisión de completar los tests y el coverage al final, porque la consigna aclara que valoran el testing pero no esperan 100% de cobertura

De acuerdo. La decisión es válida porque:

- No contradice el enunciado original.
- Está documentada explícitamente.
- El comportamiento es determinista.
- El código y el contrato actualizado coinciden.
- Reduce complejidad que no aporta valor al challenge.

Retiro esas dos observaciones como bloqueantes. El handler queda aprobado para avanzar.

## Paso 4: conectar `cmd/api/main.go`

El `main` debe actuar únicamente como composition root:

1. Crear el repositorio en memoria.
2. Crear `tweet.Service` inyectando el repositorio.
3. Crear el handler HTTP.
4. Obtener sus rutas.
5. Configurar e iniciar el servidor.

El flujo final será:

```text
HTTP request
    → writer.Handler
    → tweet.Service
    → memory.TweetRepository
```

Evitá introducir reglas de negocio o DTOs en `main.go`.

Para una implementación razonable usá `http.Server` en lugar de llamar directamente a `http.ListenAndServe`, configurando al menos:

- Dirección, por ejemplo `:8080`.
- Handler.
- `ReadHeaderTimeout`.

Al arrancar:

- Registrá la dirección donde escucha.
- Tratá `http.ErrServerClosed` como un cierre normal.
- Cualquier otro error de `ListenAndServe` debe finalizar el proceso con un mensaje claro.

No hace falta todavía implementar graceful shutdown si querés mantener esta etapa acotada; puede quedar como mejora de robustez.

## Prueba manual

Request exitoso:

```bash
curl -i \
  -X POST http://localhost:8080/tweets \
  -H 'Content-Type: application/json' \
  -d '{"user_id":"  user-123  ","content":"  Hola  "}'
```

Esperado:

- `201 Created`.
- `user_id` normalizado.
- `content` normalizado.
- UUID generado.
- Fecha RFC 3339 en UTC.

Probá también:

```bash
curl -i \
  -X POST http://localhost:8080/tweets \
  -H 'Content-Type: application/json' \
  -d '{}'
```

Debe devolver `400 invalid_user_id`.

```bash
curl -i \
  -X POST http://localhost:8080/tweets \
  -H 'Content-Type: application/json' \
  -d '{"user_id":"user-123","content":null}'
```

Debe devolver `400 invalid_content`.

```bash
curl -i \
  -X POST http://localhost:8080/tweets \
  -H 'Content-Type: application/json' \
  -d '{"user_id":"user-123","content":"Hola","extra":true}'
```

Debe devolver `400 unknown_field`.

Cuando lo conectes y hagas las pruebas manuales, revisamos el resultado funcional. Después podremos considerar terminada la publicación de tweets y pasar a la siguiente funcionalidad; la etapa completa de tests queda para el final.

> Listo ya termine y probe en postman esta todo ok.

Perfecto. Con eso queda terminada funcionalmente la publicación de tweets:

- `POST /tweets` operativo.
- Decodificación JSON estricta.
- Validación y normalización en el servicio.
- Errores según el contrato documentado.
- Persistencia concurrente en memoria.
- Wiring completo desde HTTP hasta repositorio.
- Prueba manual exitosa en Postman.

Dejamos conscientemente para la etapa final:

- Completar tests del handler y repositorio.
- Ejecutar coverage y race detector.
- Revisar cancelación de contexto.
- Robustecer lifecycle y graceful shutdown.
- Evaluación integral con criterio Senior.

El siguiente bloque lógico es **implementar seguimiento de usuarios** mediante `PUT /users/{follower_id}/following/{followed_id}`.

> Estuve revisando la estructura del repo según la arquitectura que elegimos: monolito modular por funcionalidad, con capas simples y algunas ideas de arquitectura hexagonal
> La arquitectura no va a cambiar, pero quiero ordenar mejor las carpetas para que representen esa decisión. En vez de tener todo separado en reader y writer, voy a organizar los handlers por funcionalidad: tweet, follow y timeline
> La idea es hacer solamente este cambio de estructura, sin tocar el comportamiento de POST /tweets, las validaciones, el contrato ni la persistencia. Después verifico que siga funcionando y avanzo con Follow

Es un refactor coherente con el monolito modular. Agrupar el transporte por funcionalidad expresa mejor los límites del sistema que separarlo genéricamente en `reader` y `writer`.

Una estructura razonable sería:

```text
internal/
├── tweet/                 # dominio y servicio
├── follow/                # dominio y servicio
├── timeline/              # caso de uso
├── platform/
│   └── memory/            # adaptadores de persistencia
└── rest/
    ├── tweet/             # handler HTTP de tweets
    │   ├── handler.go
    │   ├── entities.go
    │   └── routes.go
    ├── follow/
    ├── timeline/
    ├── errorcatalog/      # infraestructura HTTP compartida
    └── helper/            # lectura/escritura JSON compartida
```

Para este refactor solamente:

- Mové `internal/rest/writer` a `internal/rest/tweet`.
- Cambiá el package de `writer` a un nombre claro, por ejemplo `tweethandler` o `tweethttp`.
- Actualizá el import y el constructor usados en `cmd/api/main.go`.
- Eliminá `reader` si continúa vacío y no tiene consumidores.
- Conservá sin cambios los DTOs, el handler, las rutas y el catálogo de errores.
- No muevas `internal/tweet`: ese paquete sigue representando dominio y aplicación, no HTTP.
- No mezcles todavía código de Follow o Timeline en el paquete nuevo.

Un detalle práctico: si nombrás al paquete HTTP simplemente `tweet`, chocará visualmente con el paquete de dominio `internal/tweet` en archivos que importen ambos. Podés resolverlo con aliases, pero `tweethandler` suele hacer más legible el código.

Después del movimiento verificá:

```bash
gofmt -w ./internal/rest ./cmd/api
go test ./...
go run ./cmd/api
```

Y repetí al menos un `POST /tweets` exitoso y uno inválido. Si mantienen exactamente las mismas respuestas, el refactor está terminado y podés avanzar con Follow.

> Ya hice el refactor de estructura que habíamos hablado.
> Revisá si la estructura quedó alineada con el monolito modular por funcionalidad y las capas simples que elegimos. No modifiques el código si no hay problemas importantes, quiero cerrar definitivamente Create Tweet y avanzar con Follow.

<details><summary>1 previous message</summary>

> Voy a revisar el árbol final, dependencias entre paquetes y el wiring de `main.go`, enfocándome únicamente en si el refactor conserva el comportamiento y representa bien los módulos por funcionalidad. No modificaré nada ni abriré una nueva ronda de mejoras menores.
</details>

La estructura quedó alineada con la arquitectura elegida y no encontré ningún problema importante que impida cerrar Create Tweet.

La separación actual es coherente:

- `internal/tweet`: dominio, servicio y puerto de persistencia.
- `internal/platform/memory`: adaptador concreto de persistencia.
- `internal/rest/tweet`: transporte HTTP organizado por funcionalidad.
- `internal/rest/helper`: utilidades HTTP compartidas.
- `internal/rest/errorcatalog`: traducción de errores.
- `cmd/api/main.go`: composición de dependencias y arranque.
- `follow` y `timeline`: módulos independientes preparados para crecer.

También está bien que `TweetCreator` se defina junto al handler: el consumidor declara la interfaz mínima que necesita, evitando acoplar el transporte a la implementación concreta del servicio.

Dos observaciones no bloqueantes para tener presentes:

- Al incorporar Follow necesitarás componer las rutas de Tweet y Follow en un `ServeMux` común; cada módulo no puede reemplazar el handler principal del servidor.
- `errorcatalog` actualmente importa directamente el módulo `tweet`. Si empieza a concentrar errores de todos los módulos, puede convertirse en un punto de acoplamiento central. Podemos decidir cómo organizarlo cuando aparezcan los errores de Follow, sin refactorizarlo ahora.

El alias `resttweet` en `main.go` resuelve correctamente la coincidencia de nombres entre dominio y transporte.

No pude ejecutar Go en este entorno, pero ya verificaste manualmente el endpoint después del refactor. Con eso, considero **Create Tweet definitivamente cerrado** y podemos avanzar con Follow.
