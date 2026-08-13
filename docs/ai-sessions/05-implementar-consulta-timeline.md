# Implementar consulta del timeline

> En esta sesión voy a implementar la consulta del timeline mediante GET /users/{user_id}/timeline. La idea es avanzar de forma incremental, apoyándome en la arquitectura y los contratos definidos, y reutilizando lo que ya existe para tweets y seguimiento de usuarios. Voy a priorizar una solución simple y acotada a los requisitos funcionales del challenge
> Primero voy a intentar resolver cada etapa por mi cuenta y después revisaré las decisiones con criterio Senior. Los tests , coverage quedan para el final
> Trabajemos sobre:
> /Users/emispalenza/GolandProjects/microblogging-challenge
> Revisá el PDF original, docs/architecture.md y docs/api-contracts.md

<details><summary>3 previous messages</summary>

> Voy a empezar leyendo el PDF con el flujo de revisión visual indicado para PDFs y, en paralelo conceptual, contrastarlo con `docs/architecture.md` y `docs/api-contracts.md`. Por ahora haré solo inspección: requisitos del timeline, contratos existentes y piezas reutilizables de tweets/seguimientos.
>
> El skill de PDF exige inspeccionar todas las páginas relevantes renderizadas, no basarse únicamente en texto extraído. Ahora ubico el documento original y hago inventario del repositorio, sin modificar nada.
>
> Ya aparece una conclusión importante: el contrato del timeline está bien definido, pero su implementación actual es solo scaffolding (`package timeline` y un router vacío). Además, las interfaces existentes de tweets y follows todavía exponen únicamente escritura, así que el caso de uso necesitará ampliar puertos de lectura sin acoplar el dominio a memoria.
>
> Sigo con el PDF original y después te dejo una lectura Senior de la primera etapa, sin adelantarnos a tests/coverage.
</details>

Ya revisé el estado actual del proyecto y tomé como fuentes:

- [architecture.md](~/GolandProjects/microblogging-challenge/docs/architecture.md)
- [api-contracts.md](~/GolandProjects/microblogging-challenge/docs/api-contracts.md)
- [business.txt](~/GolandProjects/microblogging-challenge/business.txt)
- La implementación existente de tweets, follows y memoria.

No encontré ningún PDF dentro de `/Users/emispalenza/GolandProjects` ni en tu directorio de usuario. Necesito que lo adjuntes o me indiques su ruta para revisarlo visualmente y contrastarlo con la documentación.

La dirección definida para el timeline es coherente:

- Fan-out en lectura.
- Solo tweets de usuarios seguidos; no incluye los propios.
- Incluye publicaciones anteriores al follow.
- Orden global por `created_at DESC, tweet_id DESC`.
- Paginación por cursor con límite predeterminado `20` y máximo `100`.
- Consulta de `limit + 1` elementos para determinar `next_cursor`.
- Cursor opaco que representa fecha e ID del último tweet entregado.

Estado de implementación:

- [internal/timeline/service.go](~/GolandProjects/microblogging-challenge/internal/timeline/service.go) está vacío.
- [internal/rest/timeline/handler.go](~/GolandProjects/microblogging-challenge/internal/rest/timeline/handler.go) está vacío.
- [internal/rest/timeline/routes.go](~/GolandProjects/microblogging-challenge/internal/rest/timeline/routes.go) solo tiene un `Router()` vacío.
- Los repositorios actuales únicamente permiten guardar; todavía faltan operaciones para obtener seguidos y consultar tweets.
- El timeline aún no está cableado en [cmd/api/main.go](~/GolandProjects/microblogging-challenge/cmd/api/main.go).

Para avanzar incrementalmente, propondría este orden:

1. Definir los modelos de entrada, resultado y cursor del caso de uso.
2. Definir los puertos de lectura que necesita el timeline.
3. Implementar esas operaciones en memoria.
4. Implementar el servicio y la paginación.
5. Agregar DTO, parsing HTTP y representación del cursor.
6. Implementar handler, ruta y composición en `main.go`.
7. Al final, tests y coverage, como planteaste.

Como criterio Senior inicial: conviene que el servicio de timeline dependa de interfaces pequeñas propias de sus necesidades de lectura. No necesita conocer repositorios concretos ni reutilizar obligatoriamente las interfaces de escritura actuales. Esto conserva el alcance simple sin mezclar responsabilidades.

> Dale, avancemos paso a paso en ese orden. Por ahora quiero definir solamente los modelos de entrada, respuesta y cursor, junto con las interfaces de lectura del timeline. Yo voy escribiendo el código y después te aviso para que lo revises con criterio Senior. No avancemos todavía con memoria, HTTP ni el main. Por el momento usemos los documentos del proyecto

Perfecto. En esta etapa el alcance queda limitado a `internal/timeline`:

- Modelo de entrada del caso de uso.
- Modelo de respuesta/página.
- Modelo interno del cursor.
- Interfaces de lectura requeridas por el servicio.

Sin implementar todavía:

- Repositorios en memoria.
- Codificación o decodificación del cursor.
- Handler, DTO ni validaciones HTTP.
- Rutas y `main.go`.
- Tests y coverage.

Como guía de diseño, el servicio necesitará expresar conceptualmente:

- Una entrada con `user_id`, `limit` y cursor opcional.
- Una respuesta con los tweets de la página y un cursor siguiente opcional.
- Un cursor interno compuesto por `created_at` y `tweet_id`.
- Un lector de follows que devuelva los IDs seguidos por un usuario.
- Un lector de tweets que reciba autores, cursor y cantidad máxima, y devuelva tweets ya ordenados por `created_at DESC, tweet_id DESC`.

Criterios para la revisión Senior:

- Las interfaces deben definirse cerca del consumidor, en `timeline`, y ser pequeñas.
- Deben aceptar `context.Context`.
- No deberían mencionar memoria, HTTP, JSON ni strings codificados del cursor.
- El cursor de dominio debería representar datos, no su serialización opaca.
- La ausencia de cursor debe poder expresarse sin valores ambiguos.
- La consulta debe permitir pedir `limit + 1`.
- El contrato debe dejar claro quién es responsable del filtrado, orden y límite.
- Conviene reutilizar `tweet.Tweet` en la respuesta si no existe una necesidad real de duplicar el modelo.
- No agregaría todavía abstracciones para una futura base de datos ni una entidad `Timeline` persistente.

Cuando tengas esa primera versión, la revisamos sin ampliar el alcance.

> Podes revisar la primera parte

<details><summary>1 previous message</summary>

> Sí. Voy a revisar únicamente los modelos y las interfaces de lectura del timeline, contrastándolos con los contratos del proyecto. No voy a modificar código ni evaluar todavía memoria, HTTP, `main`, tests o coverage.
</details>

La base está bien orientada: interfaces pequeñas en `timeline`, uso de `context.Context`, reutilización de `tweet.Tweet` y ninguna dependencia de memoria o HTTP. Hay un ajuste de diseño importante antes de continuar.

::code-comment{title="[P1] El input expone el cursor serializado" body="`Cursor string` hace que el caso de uso reciba la representación opaca de HTTP y obliga al servicio a decodificarla o depender de un codec. Para mantener separado el transporte, el input debería contener `Cursor *CursorPosition`; el handler transformará posteriormente el query parameter opaco en esa posición." file="/Users/emispalenza/GolandProjects/microblogging-challenge/internal/timeline/timeline.go" start=9 end=13 priority=1}

::code-comment{title="[P1] La respuesta expone una representación HTTP" body="`NextCursor *string` vuelve a trasladar la serialización del cursor al dominio. La página debería devolver `*CursorPosition`; más adelante la capa REST lo codificará como string opaco y representará su ausencia como `null`." file="/Users/emispalenza/GolandProjects/microblogging-challenge/internal/timeline/timeline.go" start=15 end=18 priority=1}

Mi recomendación concreta sería:

```go
type TimelineInput struct {
	UserID string
	Limit  int
	Cursor *CursorPosition
}

type TimelinePage struct {
	Tweets     []tweet.Tweet
	NextCursor *CursorPosition
}
```

La interfaz `TweetReader` está correctamente encaminada:

```go
FindByAuthors(
    ctx context.Context,
    authorIDs []string,
    position *CursorPosition,
    limit int,
) ([]tweet.Tweet, error)
```

Solo cambiaría `authorsID` por `authorIDs`, porque contiene múltiples identificadores.

Como contrato semántico, esa operación debe:

- Considerar únicamente los autores recibidos.
- Aplicar la posición de forma exclusiva.
- Ordenar por `created_at DESC, tweet_id DESC`.
- Devolver como máximo `limit` elementos.

`FollowReader` también está bien. Podrías llamar al argumento `followerID` para expresar mejor su papel, aunque es una mejora menor.

El constructor del servicio es correcto y no hace falta agregar una interfaz combinada de repositorio. Con el cambio de cursor, esta primera parte queda sólida y preparada para implementar el caso de uso sin filtrar detalles HTTP al dominio.

> Dale, ya hice esos cambios. Ahora TimelineInput y TimelinePage trabajan con *CursorPosition, dejando la codificación del cursor para la capa rest. También cambié authorsID por authorIDs y userID por followerID. Podemos seguir con la implementación de la lectura de usuarios seguidos

<details><summary>1 previous message</summary>

> Primero verifico que los contratos hayan quedado desacoplados de REST y después acotamos la siguiente etapa exclusivamente a la lectura de IDs seguidos desde el repositorio en memoria. No voy a tocar aún tweets, servicio de timeline, HTTP ni `main`.
</details>

Podemos seguir. Los cambios del cursor quedaron bien. Antes, hay un detalle pendiente: en el código actual figura `authorID []string`; debería quedar `authorIDs []string`.

Para la lectura de usuarios seguidos, implementaría:

```go
func (r *FollowRepository) FindFollowedIDs(
	ctx context.Context,
	followerID string,
) ([]string, error)
```

Como el timeline está orientado a lecturas, conviene ajustar la estructura interna del repositorio a un índice por follower:

```go
map[string]map[string]struct{}
```

Conceptualmente:

```text
followerID -> conjunto de followedID
```

Esto permite:

- Obtener seguidos en O(cantidad de usuarios seguidos).
- Mantener la idempotencia del follow.
- Evitar recorrer todas las relaciones existentes.
- Reflejar el índice `(follower_id, followed_id)` documentado en la arquitectura.

También cambiaría `sync.Mutex` por `sync.RWMutex`:

- `SaveFollow` utiliza `Lock`.
- `FindFollowedIDs` utiliza `RLock`.

Comportamiento esperado:

- Si el follower no sigue a nadie, devolver un slice vacío y `nil`.
- No devolver el mapa interno ni exponer estado mutable.
- No hace falta ordenar los IDs: el orden final se define sobre los tweets.
- No hace falta validar la existencia del usuario.
- No hace falta importar `timeline`; el repositorio satisface `FollowReader` estructuralmente.
- Para esta operación en memoria no agregaría manejo especial de cancelación del contexto.

Mantendría el cambio limitado a [follow_repository.go](~/GolandProjects/microblogging-challenge/internal/platform/memory/follow_repository.go). La interfaz de escritura de `internal/follow` no necesita ampliarse, porque la lectura es consumida por `timeline`. Cuando lo implementes, revisamos concurrencia, complejidad y preservación de la idempotencia.

> Dale, tiene sentido usar el mapa anidado porque el ejercicio pide optimizar para lectura. Así podemos buscar directamente los usuarios seguidos sin recorrer todas las relaciones.
> Entiendo que el trade off es usar un poco más de memoria y agregar algo de complejidad al guardar, a cambio de hacer más eficiente la consulta del timeline. Voy a implementarlo así y después te aviso para que lo revisemos

Exactamente. El trade-off está bien entendido.

El mapa anidado replica en memoria el acceso que tendría una clave o índice compuesto por `(follower_id, followed_id)`:

- Lectura directa por `follower_id`.
- Unicidad e idempotencia mediante el conjunto interno.
- Escritura ligeramente más compleja porque debe inicializarse el conjunto del follower.
- Algo más de overhead de memoria por cada mapa interno.

Al implementarlo, cuidá especialmente:

- Inicializar el conjunto cuando aparece un follower nuevo.
- Conservar el retorno `true` al crear y `false` cuando ya existe.
- Usar `Lock` para escritura y `RLock` para lectura.
- Devolver un slice nuevo y vacío —no `nil`— cuando no haya seguidos.
- No depender del orden de iteración del mapa.

Quedo atento para revisarlo antes de pasar a la lectura de tweets.

> Revisa el avance

<details><summary>1 previous message</summary>

> Voy a revisar solamente el refactor del repositorio de follows y `FindFollowedIDs`: estructura de datos, idempotencia, concurrencia y contrato de colección vacía. No modificaré el código ni avanzaré a tweets.
</details>

El avance está correcto; no encontré problemas funcionales ni de diseño en esta etapa.

Puntos bien resueltos:

- `map[string]map[string]struct{}` permite acceso directo por `followerID`.
- `SaveFollow` conserva la idempotencia.
- La creación del conjunto interno ocurre bajo el lock de escritura.
- `sync.RWMutex` separa correctamente lecturas y escrituras.
- `FindFollowedIDs` devuelve una copia en forma de slice, sin exponer el mapa interno.
- Un usuario sin relaciones obtiene un slice vacío no nulo.
- No se introduce dependencia desde memoria hacia `timeline`.
- `authorIDs` ya quedó corregido.
- El orden indeterminado del mapa es aceptable porque no forma parte del contrato; el orden relevante será el de los tweets.

Como observación puramente nominal, `relation` podría llamarse `followingByFollower` o `followedByFollower` para describir mejor la estructura, pero no justifica frenar ni rehacer esta etapa.

La lectura de usuarios seguidos queda aprobada. Podemos pasar a diseñar e implementar `FindByAuthors` en el repositorio de tweets.

> Si los id de los tweets son UUID no se puede saber cual fue primero o segundo o no ?

Correcto: con los UUID actuales (`uuid.NewString()`, que genera UUID v4) no se puede inferir cuál tweet fue creado primero.

Un UUID v4 es aleatorio. Compararlo lexicográficamente solo proporciona un orden determinista, no cronológico.

En este proyecto el orden es:

```text
created_at DESC, tweet_id DESC
```

Eso significa:

- `created_at` establece el orden cronológico.
- `tweet_id` se utiliza únicamente cuando dos tweets tienen exactamente el mismo `created_at`.
- En ese empate, el UUID no indica cuál fue creado primero; simplemente define un orden estable y repetible.

Por lo tanto, la documentación que dice “usar el ID como desempate” es válida. Lo que no debería afirmarse es que el ID permite conocer el orden real de creación.

Hay además un detalle práctico: `time.Now().UTC()` normalmente tiene precisión suficiente para diferenciar tweets, pero dos creaciones aún podrían compartir el mismo timestamp. Si el requisito exigiera preservar estrictamente el orden real de creación, necesitaríamos otro mecanismo, por ejemplo:

- Una secuencia monotónica.
- UUID v7, que incorpora tiempo.
- Un contador interno.
- Que la base de datos asigne una clave ordenable.

Para el alcance del challenge, `created_at DESC, id DESC` con UUID v4 es una solución simple, determinista y consistente con los documentos. Solo hay que entender que, ante timestamps idénticos, el desempate es arbitrario pero estable.

> Cual de esos es el mas prolijo ?

Para este challenge, lo más prolijo es mantener UUID y usar UUID v7.

UUID v7:

- Sigue siendo un UUID.
- Incorpora información temporal.
- Es ordenable cronológicamente en su representación estándar.
- Funciona mejor como clave de índice que un UUID v4 aleatorio.
- Permite conservar el contrato actual, donde `id` es un UUID.

El orden seguiría siendo:

```text
created_at DESC, tweet_id DESC
```

Así, `created_at` expresa la fecha de negocio y el UUID v7 aporta un desempate temporalmente coherente.

Pero hay una distinción importante: si necesitás garantizar de forma absoluta el orden entre operaciones concurrentes, lo más sólido es una secuencia monotónica asignada por la persistencia. UUID v7 mejora el orden temporal, pero no reemplaza necesariamente una secuencia transaccional para requisitos estrictos.

Mi recomendación por alcance:

- Challenge actual: UUID v4 está bien; el desempate solo necesita ser determinista.
- Mejora prolija sin cambiar el contrato: UUID v7.
- Sistema real con orden estricto garantizado: secuencia de base de datos.

No cambiaría ahora UUID v4 por UUID v7 salvo que quieras que el orden del desempate represente también el orden aproximado de creación. No es necesario para implementar correctamente el contrato actual.

> Entiendo, mi duda era si con UUID v4 podíamos interpretar el desempate como orden de creación, pero veo que no hace falta. created_at define el orden temporal y el ID solamente hace que el resultado sea determinista cuando dos fechas coinciden
> Dejo UUID v4 para no agregar un cambio que el requisito no necesita. UUID v7 podría quedar como una mejora futura si quisiéramos que el desempate también tenga una relación temporal aproximada

Exactamente. Esa decisión es correcta y está bien justificada.

Con UUID v4:

- `created_at` determina el orden temporal.
- `tweet_id` no representa el orden de creación.
- Ante fechas iguales, `tweet_id DESC` establece un orden arbitrario pero estable.
- El mismo conjunto de datos produce siempre el mismo orden.
- El cursor puede usar `(created_at, tweet_id)` sin omitir ni repetir tweets por empates de fecha.

UUID v7 queda como posible mejora futura, no como requisito para que la solución actual sea correcta. Podemos continuar con `FindByAuthors` manteniendo el criterio documentado:

```text
created_at DESC, tweet_id DESC
```

> Ya terminé la parte de memoria. Ahora voy a hacer el servicio para buscar primero a quién sigue el usuario y después traer los tweets de esos usuarios. Voy a pedir limit + 1 para saber si hay otra página y, si la hay, generar el próximo cursor con el último tweet que devuelvo.
> Queres revisar a ver si ves algo

<details><summary>1 previous message</summary>

> Sí. Voy a revisar la implementación en memoria de `FindByAuthors` contra cuatro invariantes: filtrado por autores, posición exclusiva del cursor, orden compuesto descendente y respeto del límite. También miraré que no haya efectos laterales ni problemas de concurrencia; no avanzaré con el servicio.
</details>

La implementación de memoria respeta correctamente los requisitos principales:

- Filtra solamente por los autores recibidos.
- Aplica el cursor de manera exclusiva.
- Usa la comparación correcta:
  `created_at < cursor.created_at` o misma fecha con `id < cursor.id`.
- Ordena por `created_at DESC, id DESC`.
- Aplica el límite después del orden global.
- Devuelve un slice vacío cuando no encuentra resultados.

Veo una mejora concreta de concurrencia:

::code-comment{title="[P2] El read lock se mantiene durante el ordenamiento" body="El lock solamente es necesario mientras se recorren los slices almacenados en el repositorio. `tweets` ya es un slice local con valores copiados, por lo que puede liberarse el `RLock` antes de ejecutar `sort.Slice`. Mantenerlo durante el ordenamiento bloquea innecesariamente `Save` cuando hay muchos candidatos." file="/Users/emispalenza/GolandProjects/microblogging-challenge/internal/platform/memory/tweet_repository.go" start=43 end=72 priority=2}

Para resolverlo, no usaría `defer` en este método:

```go
r.mu.RLock()

// Copiar los tweets candidatos al slice local.

r.mu.RUnlock()

// Ordenar y limitar sin mantener el lock.
```

También hay un supuesto que conviene reconocer: `Save` escribe en `tweets` y siempre agrega en `tweetsByAuthor`. Si se guardara dos veces el mismo ID, `FindByID` tendría un único tweet pero el índice por autor contendría duplicados. En el flujo actual los IDs son UUID nuevos y `Save` representa una inserción, así que no lo considero bloqueante para el challenge. No agregaría complejidad si mantenemos explícito ese supuesto.

Respecto al servicio que vas a implementar, el flujo planteado es correcto:

1. Obtener los IDs seguidos.
2. Si no hay seguidos, devolver una página vacía sin consultar tweets.
3. Solicitar `input.Limit + 1` tweets.
4. Si llegaron como máximo `input.Limit`, devolverlos todos y cursor `nil`.
5. Si llegó el elemento adicional, descartarlo.
6. Crear `NextCursor` usando el último tweet que efectivamente se devuelve, no el elemento adicional.

Ese último detalle es esencial: el cursor debe representar la posición del último tweet visible para el cliente. Con el ajuste del lock, la parte de memoria queda lista para seguir.

> Dale voy a cambiar el lock para usarlo solamente mientras copio los tweets del mapa. Después lo libero y ordeno el slice sin bloquear el repositorio. Con el Save voy a asumir que siempre llegan IDs nuevos, porque los genera el servidor, así que no voy a agregar más lógica para ese caso. Después sigo con el servicio.

Perfecto. Ambas decisiones son proporcionales al alcance.

Liberar el `RLock` después de copiar los candidatos reduce la sección crítica sin perder seguridad: los elementos de `tweet.Tweet` se copian por valor al slice local.

Y asumir IDs nuevos en `Save` es válido porque:

- Los genera el servidor.
- El caso de uso crea tweets nuevos.
- No existen edición ni reintentos idempotentes de creación en el contrato.
- Manejar colisiones o actualizaciones agregaría complejidad sin cubrir un requisito.

Para el servicio, cuidá que cualquier error de `FindFollowedIDs` termine inmediatamente la operación y que ocurra lo mismo con `FindByAuthors`. La página vacía debería contener un slice no nulo y `NextCursor: nil`. Quedo atento para revisar esa implementación.

> Ya terminé el servicio. Primero busca a quién sigue el usuario y, si no sigue a nadie, devuelve el timeline vacío sin consultar tweets. Si tiene seguidos, pide limit + 1, recorta el tweet extra y arma el próximo cursor con el último tweet que realmente devuelve. Ahora voy a agregar la validación para que el límite esté entre 1 y 100. El valor por defecto de 20 lo voy a manejar después en la capa rest
> Queres revisar el service ?

<details><summary>1 previous message</summary>

> Sí. Voy a revisar el servicio contra el flujo acordado, especialmente propagación de errores, página vacía, `limit + 1`, recorte y posición exacta del cursor. También verificaré que la validación del límite quede ubicada en la capa correcta, sin entrar todavía en REST.
</details>

El servicio está bien construido. El flujo principal coincide con la arquitectura y el contrato:

- Propaga errores del lector de follows.
- Evita consultar tweets si no hay usuarios seguidos.
- Propaga errores del lector de tweets.
- Solicita `limit + 1`.
- Detecta correctamente si existe otra página.
- Descarta el elemento adicional.
- Construye el cursor desde el último tweet entregado.
- Devuelve `NextCursor: nil` en la última página.

Hay un punto pendiente que actualmente puede producir un panic:

::code-comment{title="[P1] Validar el límite antes de consultar repositorios" body="Con `Limit == 0`, si el lector devuelve un elemento, el código recorta a `tweets[:0]` y luego intenta acceder a `tweets[-1]`. Los valores negativos también pueden provocar un slice bounds panic. Validá el rango 1–100 al comienzo del método, antes de llamar a `FindFollowedIDs`." file="/Users/emispalenza/GolandProjects/microblogging-challenge/internal/timeline/service.go" start=21 end=22 priority=1}

La validación debería ser la primera operación:

```go
if input.Limit < 1 || input.Limit > 100 {
	return TimelinePage{}, ErrLimitOutOfRange
}
```

Estoy de acuerdo con manejar el valor predeterminado `20` en REST:

- REST interpreta la ausencia de `limit`.
- El servicio siempre recibe un límite explícito.
- El servicio protege su invariante `1–100`, independientemente del caller.

Crearía un error de dominio estable en `internal/timeline/errors.go`, siguiendo el estilo existente:

```go
var ErrLimitOutOfRange = errors.New("limit out of range")
```

Más adelante `errorcatalog` lo traducirá al `422` definido en el contrato.

No validaría `UserID` aquí porque el challenge considera válidos los identificadores recibidos. Tampoco agregaría validación del contenido de `CursorPosition` todavía: el cursor opaco será decodificado y validado en la frontera REST.

Una vez agregada la validación inicial, el servicio queda listo desde el punto de vista funcional.

> Ya terminé la parte funcional del timeline y conecté todo en el main. También comprobé que compile y que pasen los tests que ya estaban. Me falta correr gofmt en el main y después podemos hacer la revisión completa. Luego seguimos con los tests del timeline, coverage

Perfecto. Corré `gofmt` sobre el `main` y cualquier otro archivo modificado; después hacemos una revisión integral antes de escribir tests.

La revisión cubrirá:

- Contrato de `GET /users/{user_id}/timeline`.
- Separación entre dominio, memoria y REST.
- Validación de `limit` y errores.
- Codificación y decodificación del cursor.
- Orden, filtrado y paginación.
- Respuestas vacías y `next_cursor: null`.
- Registro de rutas y composición en `main`.
- Concurrencia y posibles panics.
- Consistencia con `architecture.md` y `api-contracts.md`.

Luego corregimos lo necesario y recién ahí avanzamos con tests del timeline y coverage.

> Listo

<details><summary>4 previous messages</summary>

> Hago ahora la revisión completa de la implementación funcional, sin modificar archivos. Primero voy a reconstruir el flujo entero —ruta, parsing, cursor, servicio, repositorios y composición— y después verificaré compilación/tests como control de regresión.
>
> El flujo principal está bien armado. Estoy profundizando ahora en los bordes del contrato HTTP, porque ahí es donde suelen esconderse diferencias observables aunque el happy path compile: parámetro presente pero vacío, cursores estructuralmente inválidos y forma exacta del JSON vacío.
>
> La verificación automática chocó con las restricciones del entorno sobre el toolchain y la caché de Go —no con errores atribuibles al código—. Voy a repetirla con acceso fuera del sandbox para distinguir correctamente problemas ambientales de regresiones reales.
>
> La ejecución autorizada sigue en curso. Mientras termina, el único desvío contractual concreto que veo hasta ahora es el tratamiento de `?cursor=`: hoy se interpreta como “sin cursor”, aunque un cursor explícitamente presente pero vacío debería considerarse inválido. Estoy esperando el resultado final de `go test` y `go vet` para cerrar la revisión con evidencia.
</details>

La implementación funcional está bien resuelta y respeta la arquitectura. Encontré un desvío contractual y un detalle de calidad en tests.

::code-comment{title="[P2] Un cursor explícitamente vacío se acepta como primera página" body="Cuando el request contiene `?cursor=`, el parámetro está presente pero `DecodeCursor(\"\")` devuelve `(nil, nil)`, por lo que se interpreta igual que un cursor ausente. Según el contrato, el primer request no envía cursor y un cursor proporcionado pero inválido debe responder `400 invalid_cursor`. Como el handler solo llama a esta función cuando el parámetro está presente, el string vacío debería devolver `ErrInvalidCursor`." file="/Users/emispalenza/GolandProjects/microblogging-challenge/internal/rest/timeline/cursor.go" start=39 end=42 priority=2}

::code-comment{title="[P3] go vet detecta literales posicionales" body="Estos valores de `follow.Follow` usan campos sin nombre, lo que acopla el test al orden de la estructura y hace fallar `go vet`. Usá literales con `FollowerID:` y `FollowedID:` tanto en la tabla como en el caso duplicado." file="/Users/emispalenza/GolandProjects/microblogging-challenge/internal/platform/memory/follow_repository_test.go" start=17 end=23 priority=3}

La corrección del cursor sería simple:

```go
if value == "" {
	return nil, ErrInvalidCursor
}
```

La ausencia ya está representada correctamente porque el handler no llama a `DecodeCursor` cuando `cursor` no aparece en la query.

Evaluación general:

- El dominio no conoce base64, JSON ni detalles HTTP.
- `limit` predeterminado se resuelve en REST y el rango se protege en el servicio.
- Los códigos `400`, `422` y `500` coinciden con el contrato.
- El cursor conserva nanosegundos, se codifica como base64 URL-safe y usa UTC.
- La respuesta reutiliza correctamente la representación HTTP del tweet.
- El timeline vacío produce `"tweets":[]` y `"next_cursor":null`.
- La posición del cursor y el orden del repositorio son consistentes.
- `main` comparte los mismos repositorios entre escritura, follows y timeline.
- La ruta está registrada con el método y path correctos.
- `gofmt -l` no reportó archivos.
- `go test ./...` pasó en todos los paquetes.
- `go vet ./...` solo reportó los literales posicionales señalados arriba.

Con esas dos correcciones, considero cerrada la parte funcional y podemos pasar a completar sistemáticamente los tests del timeline y medir coverage.

> Ahí realice las dos correcciones que mencionaste, ves algún otros aspecto a considerar ?

<details><summary>2 previous messages</summary>

> Verifico ambas correcciones y hago una última pasada de bordes funcionales antes de cerrar: query parameters, cursores alterados, página vacía, límites y consistencia del índice en memoria. Después te digo si queda algo bloqueante o solo mejoras opcionales.
>
> Las dos correcciones quedaron bien aplicadas y todo sigue formateado. Haré una verificación final de `test`, `vet` y detector de carreras; esta última es especialmente relevante porque agregamos índices en memoria protegidos por `RWMutex`.
</details>

No veo más aspectos bloqueantes. Las dos correcciones quedaron bien y la implementación funcional puede considerarse cerrada.

Verificación final:

- `gofmt`: sin archivos pendientes.
- `go test ./...`: todos los paquetes pasan.
- `go vet ./...`: sin observaciones.
- `go test -race ./...`: sin carreras detectadas.

Aspectos opcionales, no necesarios para el challenge:

- El cursor es opaco, pero no está firmado. Un cliente podría construir uno válido manualmente. Para este sistema sin autenticación es aceptable; en producción podría firmarse.
- `json.Unmarshal` acepta campos adicionales dentro del cursor. Podría aplicarse decodificación estricta, pero el contrato no lo exige.
- La consulta en memoria reúne y ordena todos los tweets candidatos. Es coherente con el fan-out en lectura simple documentado; una combinación incremental por autor sería una optimización futura.
- La paginación no representa una fotografía inmutable si cambian los follows durante el recorrido. Esa limitación ya está documentada.

Para la siguiente etapa, enfocaría los tests faltantes en:

- Límites exactos `1` y `100`, además de `0` y `101`.
- `cursor=` vacío.
- Empate de `created_at` resuelto por ID descendente.
- Exclusión exacta del tweet indicado por el cursor.
- Tweets propios excluidos.
- Tweets anteriores al follow incluidos.
- Página vacía con `tweets: []` y `next_cursor: null`.
- Primera página, página intermedia y última página.
- Errores de ambos lectores.
- Flujo integrado HTTP → servicio → memoria.

La implementación está lista para entrar en fase de tests y coverage.
