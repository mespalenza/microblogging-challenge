# HLD

> Bueno chat, ahora que ya terminé la implementación del challenge me falta armar
> el High Level Design
> Pensaba separar la arquitectura productiva en dos etapas
>
> Para la primera etapa quiero mantener una solución simple y parecida a lo que implementé:
> - La API Go en Docker
> - Un  Load Balancer
> - Tasks de la API ejecutándose en ECS Fargate
> - PostgreSQL en RDS para guardar tweets y follows
> - Fan out en lectura para construir el timeline
>
> No quiero agregar componentes que todavía no estén justificados por métricas
>
> Después la posible evolución en caso de que las lecturas del timeline se conviertan en un cuello de botella:
>
> - Mantener PostgreSQL como fuente de verdad
> - Publicar eventos mediante SQS
> - Procesarlos con workers
> - Guardar timelines preparados en DynamoDB
> - Leer esos timelines desde la API
> - Evaluar fan out en escritura o una estrategia híbrida según la cantidad de
    >   seguidores
>
> En MELI  la infra estaba bastante abstraída. Conozco conceptos como KVS y procesamiento asíncrono, pero no trabajé directamente configurando estos servicios de AWS.
>
> Me ayudás a revisar este diseño de a poco? Algnos servicios los he sacado de videos, decime si los ves ok
>
> Quiero entender qué función cumple cada componente y por qué lo usaríamos antesde agregarlo al documento.
>
> También quiero que la documentación quede escrita con mis palabras y que se
> distinga claramente:
>
> 1. Lo que implementé para el challenge.
> 2. Lo que desplegaría como primera etapa productiva.
> 3. Lo que evaluaría como evolución solamente si las métricas lo justifican.

<details><summary>4 previous messages</summary>

> Sí. Primero voy a contrastar tu propuesta con lo que realmente quedó implementado en el repositorio; así evitamos mezclar el challenge con decisiones hipotéticas de producción. Después revisamos la etapa 1 componente por componente, explicando función, motivo y qué problema todavía no justifica agregar nada más.
>
> <details><summary>Explored 2 lists, ran 2 commands</summary>
>
> - Ran `pwd && rg --files -g '!*vendor*' -g '!*.sum' | sed -n '1,160p'`
>
> <details><summary>Explored 2 lists</summary>
>
> - Listed files
> - Listed files in `.`
>
> </details>
>
> - Ran `git status --short --branch && git branch -a && git log --oneline --decorate -12`
>
> </details>
>
> Veo que el directorio de trabajo está en un `main` vacío, pero el repositorio sí contiene las ramas remotas del challenge. Voy a leerlas sin cambiar tu rama ni tocar archivos, especialmente la implementación final y la documentación existente.
>
> <details><summary>Ran 8 commands</summary>
>
> - Ran `git log --oneline --decorate remotes/origin/feature/final-fixes-and-docker -10 && git ls-tree -r --name-only remotes/origin/feature/final-fixes-and-docker | sed -n '1,200p'`
> - Ran `git show remotes/origin/feature/final-fixes-and-docker:README.md`
> - Ran `git show remotes/origin/feature/final-fixes-and-docker:docs/architecture.md`
> - Ran `git show remotes/origin/feature/final-fixes-and-docker:cmd/api/main.go`
> - Ran `git show remotes/origin/feature/final-fixes-and-docker:internal/timeline/service.go`
> - Ran `git show remotes/origin/feature/final-fixes-and-docker:internal/platform/memory/tweet_repository.go`
> - Ran `git show remotes/origin/feature/final-fixes-and-docker:internal/platform/memory/follow_repository.go`
> - Ran `for ref in remotes/origin/master remotes/origin/main remotes/origin/develop remotes/origin/feature/dockerize-application remotes/origin/feature/add-architecture-doc; do git log -1 --format='%h %ci %s' "$ref"; done && git ls-tree -r --name-only remotes/origin/feature/dockerize-application | sed -n '1,100p' && git ls-tree -r --name-only remotes/origin/develop | sed -n '1,100p'`
>
> </details>
</details>

La dirección general está bien y, sobre todo, está bien planteada la separación entre “lo que existe”, “la primera versión productiva” y “una evolución condicionada por métricas”.

Revisé el código de `origin/main`: el challenge efectivamente tiene una API Go dockerizada, repositorios en memoria y fan-out en lectura. PostgreSQL, ECS, SQS y DynamoDB no están implementados; deben aparecer explícitamente como decisiones de diseño, no como parte del challenge.

## 1. Lo implementado en el challenge

Hoy tenés:

- Una API REST en Go.
- Un monolito modular con separación entre HTTP, casos de uso, dominio y persistencia.
- Endpoints para publicar tweets, seguir usuarios y consultar el timeline.
- Persistencia en memoria:
    - Tweets indexados por autor.
    - Follows indexados por usuario seguidor.
- Timeline construido bajo demanda mediante fan-out en lectura.
- Orden por `created_at DESC, tweet_id DESC`.
- Paginación por cursor.
- Aplicación empaquetada mediante Docker.

Una forma natural de escribirlo con tus palabras sería:

> Para el challenge implementé una única API en Go, empaquetada como una imagen Docker. La aplicación utiliza repositorios en memoria para simplificar su ejecución, por lo que los datos se pierden al reiniciar el proceso. El timeline no se almacena: se construye cuando el usuario lo consulta, buscando las cuentas que sigue y combinando sus tweets.

Eso deja claro que Docker sí forma parte de lo implementado, pero ECS y PostgreSQL no.

## 2. Primera etapa productiva

Tu propuesta es razonable. Yo la representaría así:

```text
Cliente
   |
Application Load Balancer
   |
ECS Service
   |
Tasks Fargate con la API Go
   |
PostgreSQL en RDS
```

### Application Load Balancer

“Load Balancer” está bien como concepto, pero en AWS elegiría concretamente un **Application Load Balancer (ALB)** porque tu servicio expone HTTP.

Su función sería:

- Recibir las peticiones de los clientes.
- Distribuirlas entre varias instancias de la API.
- Verificar mediante health checks qué tasks están sanas.
- Dejar de enviar tráfico a una task que falla.
- Exponer un único punto de entrada aunque las tasks cambien.

No guarda estado ni ejecuta lógica del timeline. Su trabajo es solamente recibir y distribuir tráfico.

### ECS y Fargate

Acá conviene distinguir tres conceptos:

- **ECS** administra la ejecución de contenedores.
- Un **ECS Service** declara cuántas copias de la API deberían estar funcionando.
- Una **Fargate Task** es una instancia en ejecución de tu contenedor.

En vez de escribir solamente “tasks ejecutándose en ECS Fargate”, pondría:

> La API se ejecutaría como un ECS Service sobre Fargate, manteniendo una o más tasks del contenedor en funcionamiento.

Fargate evita que tengas que administrar servidores EC2. Es razonable para esta primera etapa porque te permite desplegar y escalar el contenedor sin introducir Kubernetes ni gestionar un clúster de máquinas.

El contenedor debería ser stateless. Eso significa que ninguna información importante puede quedar solamente en la memoria de una task, porque:

- Puede reiniciarse.
- ECS puede reemplazarla.
- Dos requests del mismo usuario pueden llegar a tasks diferentes.

Por eso reemplazarías los repositorios en memoria por PostgreSQL.

### PostgreSQL en RDS

RDS es el servicio administrado; PostgreSQL es el motor de base de datos.

Su función sería guardar de manera durable:

- Tweets.
- Relaciones de follow.

Es una buena elección inicial porque tu modelo ya tiene:

- Datos estructurados.
- Unicidad de `(follower_id, followed_id)`.
- Validaciones.
- Índices compuestos.
- Consultas que combinan follows y tweets.
- Orden y paginación por cursor.

Además, mantiene el modelo productivo bastante cerca de lo que diseñaste en el challenge. Cambiaría el adaptador en memoria por uno de PostgreSQL sin cambiar los casos de uso principales.

### Fan-out en lectura

También lo mantendría inicialmente.

Cuando llega `GET /users/{id}/timeline`, la API:

1. Consulta qué usuarios sigue.
2. Busca sus tweets.
3. Combina y ordena los resultados.
4. Devuelve una página.

La ventaja es que publicar un tweet sigue siendo barato: se inserta una sola fila. Tampoco duplicás el tweet en el timeline de cada seguidor.

El costo está en la lectura. Si un usuario sigue muchas cuentas, PostgreSQL debe considerar más autores y combinar más datos. Pero sin métricas que indiquen un problema, eso es una limitación conocida, no una razón suficiente para agregar otra arquitectura.

## Componentes implícitos que no conviene confundir con sobrearquitectura

Para desplegar realmente en AWS también aparecerían elementos como:

- ECR para almacenar la imagen Docker.
- VPC, subnets y security groups para conectividad.
- Secrets Manager o Parameter Store para credenciales.
- Logs y métricas, normalmente mediante CloudWatch.

No son una nueva estrategia para resolver el timeline. Son infraestructura operativa necesaria para ejecutar, conectar, proteger y observar la solución. En el diagrama principal pueden mostrarse de manera secundaria para no distraer del flujo central.

## 3. Evolución si las lecturas se vuelven un cuello de botella

La idea también es válida:

```text
Publicación de tweet
        |
    PostgreSQL
  fuente de verdad
        |
       SQS
        |
     Workers
        |
    DynamoDB
timeline preparado
        |
       API
```

Pero hay una precisión importante: **no conviene presentar “guardar en PostgreSQL y después publicar en SQS” como si ambas operaciones fueran atómicas**.

Podría ocurrir esto:

1. El tweet se guarda correctamente en PostgreSQL.
2. La aplicación falla antes de publicar el mensaje.
3. El timeline preparado nunca se actualiza.

Para una evolución real habría que evaluar un mecanismo como el **transactional outbox**: el tweet y un evento pendiente se guardan en la misma transacción de PostgreSQL; otro proceso publica luego ese evento en SQS. No hace falta incorporarlo en la primera etapa, pero sí mencionarlo si el documento profundiza en la arquitectura asíncrona.

### SQS

SQS funcionaría como una cola de trabajo:

- Desacopla la publicación del tweet de la preparación de timelines.
- Absorbe picos.
- Permite que los workers procesen a su propio ritmo.
- Facilita reintentos.

No es exactamente un “bus de eventos” completo. En este diseño se usaría principalmente para distribuir trabajo asíncrono.

Además, los workers deberían ser idempotentes: SQS puede volver a entregar un mensaje, por lo que procesarlo dos veces no debería duplicar el tweet en un timeline.

### Workers

Los workers consumirían los mensajes y actualizarían los timelines preparados.

Por ejemplo, al procesar un tweet:

1. Identifican los seguidores del autor.
2. Agregan una referencia al tweet en el timeline de cada seguidor.
3. Mantienen únicamente la ventana de contenido que el producto decida conservar.

Esos workers también podrían ejecutarse en ECS Fargate. No necesitás introducir otra plataforma de cómputo solamente porque el procesamiento sea asíncrono.

### DynamoDB

DynamoDB puede ser apropiado para guardar timelines materializados porque el patrón de acceso sería muy concreto:

> Dado un usuario, obtener sus tweets más recientes en orden y paginarlos.

Pero no lo justificaría simplemente diciendo que “NoSQL escala más”. La justificación correcta sería que el timeline preparado se modela como una colección ordenada por usuario y se consulta principalmente por clave.

Conceptualmente:

```text
partition key: user_id
sort key: created_at + tweet_id
```

La API podría recuperar una página sin volver a combinar los tweets de todos los autores.

PostgreSQL seguiría siendo la fuente de verdad. DynamoDB sería una vista derivada y eventualmente consistente: podría estar temporalmente atrasada y debería poder reconstruirse.

## Fan-out en escritura e híbrido

El fan-out en escritura mejora la lectura, pero traslada el costo a la publicación.

Si un autor tiene cinco millones de seguidores, publicar un tweet puede generar cinco millones de actualizaciones. Por eso la estrategia híbrida tiene sentido:

- Autores con una cantidad moderada de seguidores: fan-out en escritura.
- Autores con muchísimos seguidores: conservar sus tweets por autor y mezclarlos durante la lectura.
- La API combina el timeline preparado con los tweets recientes de esos autores especiales.

El umbral no debería inventarse en el documento. Se definiría observando cantidad de seguidores, ritmo de publicación, costo de procesamiento, retraso de la cola y latencia del timeline.

## Mi evaluación

Los servicios que elegiste están bien. Haría estos ajustes conceptuales:

- Nombrar específicamente **Application Load Balancer**.
- Hablar de un **ECS Service con tasks sobre Fargate**.
- Incluir ECR, red, secretos y observabilidad como infraestructura de soporte.
- Mantener PostgreSQL y fan-out en lectura como primera etapa.
- Presentar SQS, workers y DynamoDB únicamente como evolución.
- Aclarar consistencia eventual, mensajes duplicados y eventual uso de outbox.
- No asumir desde ahora que DynamoDB será obligatorio: primero hay que demostrar que el cuello de botella es la construcción del timeline.

La separación central podría resumirse así:

| Nivel | Persistencia | Construcción del timeline | Infraestructura |
|---|---|---|---|
| Challenge | Memoria | En lectura | Proceso Go en Docker |
| Primera etapa productiva | PostgreSQL en RDS | En lectura | ALB + ECS Service/Fargate |
| Evolución evaluable | PostgreSQL + vista en DynamoDB | Asíncrona o híbrida | SQS + workers |

El próximo bloque que conviene desarrollar es la primera etapa productiva: flujo de una petición, responsabilidades de ALB/ECS/RDS y qué ocurre cuando una task se reinicia. Eso da una base firme antes de entrar en SQS y DynamoDB.

> Entiendo. Para este challenge voy a mantener el diagrama principal enfocado en
> los componentes que intervienen directamente en los flujos de la aplicación.
>
> No voy a agregar por ahora ECR, networking, secretos ni transactional outbox
> porque no trabajé directamente configurando esos servicios y no quiero sumar
> conceptos que después no pueda defender.
>
> Sí voy a dejar explícita la separación entre:
>
> - La implementación actual en memoria.
> - La primera etapa productiva con ALB, ECS Fargate y PostgreSQL.
> - La evolución con SQS, workers y DynamoDB solamente si las métricas justifican
    >   materializar el timeline.
>
> También voy a documentar consistencia eventual, idempotencia y las limitaciones
> de cada estrategia.

Me parece una decisión correcta y defendible para el challenge.

El documento puede mantener tres niveles claramente separados:

1. **Implementación actual:** API Go en Docker, persistencia en memoria y fan-out en lectura.
2. **Primera etapa productiva:** ALB, ECS Service con tasks en Fargate y PostgreSQL en RDS como fuente de verdad, manteniendo fan-out en lectura.
3. **Evolución condicionada:** SQS, workers y timelines materializados en DynamoDB, evaluando fan-out en escritura o híbrido según métricas reales.

También está bien omitir del diagrama principal ECR, networking, secretos y outbox. No afirmás que no sean necesarios en un despliegue real; simplemente quedan fuera del nivel de abstracción y alcance del diseño.

Una frase útil para evitar dudas sería:

> El diagrama se concentra en los componentes que participan directamente en los flujos funcionales de publicación, seguimiento y lectura del timeline. La infraestructura transversal necesaria para operar la solución, como networking, gestión de secretos y almacenamiento de imágenes, queda fuera del alcance de este documento.

Para la evolución asíncrona, alcanza con dejar explícito que:

- PostgreSQL continúa siendo la fuente de verdad.
- DynamoDB contiene una vista derivada y reconstruible.
- El timeline puede quedar temporalmente desactualizado.
- Los workers deben tolerar reintentos y procesar eventos de forma idempotente.
- La nueva arquitectura se evaluaría únicamente si latencia, volumen o costo demuestran que el fan-out en lectura es insuficiente.

Con ese alcance, el diseño muestra criterio arquitectónico sin aparentar experiencia operativa que no tuviste.

> T

Parece que el mensaje quedó incompleto. ¿Querías continuar con el diseño?

> e

Te leo; mandame el mensaje completo cuando quieras.

> p

Parece que se está enviando letra por letra. Escribilo completo y lo retomamos.

> a

Te espero; cuando esté completo, lo revisamos.

> s

Parece que querés escribir “Te pas…”. Seguí cuando quieras.

> o

Dale, pasámelo.

> o

Cuando quieras, pegalo completo acá.

> Te mando los diagramas a ver si los ves bien
>
> User context
>
> Images:
> - `~/Desktop/Captura de pantalla 2026-08-12 a la(s) 7.58.16 p. m..png`

Sí, este primer diagrama se entiende bien y representa correctamente la primera etapa productiva. Haría solamente un ajuste conceptual importante.

Donde dice **“ECS Fargate”**, pondría:

> **ECS Service — Fargate**

Porque:

- ECS Service administra y mantiene la cantidad deseada de tasks.
- Fargate es la modalidad de cómputo con la que se ejecutan esas tasks.
- Las tasks contienen el contenedor de la API Go.

Podría quedar así:

```text
ECS Service
Launch type: Fargate
┌─────────────────┐
│ Task: API Go    │
│ Task: API Go    │
│ Task: API Go    │
└─────────────────┘
Desired count: N
```

También ajustaría estos detalles menores:

- Cambiar **“N tasks”** por **“N tasks de la API”** o **“Desired count: N”**.
- Hacer bidireccional la conexión entre la API y PostgreSQL, ya que conceptualmente hay requests y respuestas.
- Si querés evidenciar la decisión principal, agregar debajo de PostgreSQL o sobre la conexión:
  > Timeline construido mediante fan-out en lectura.
- Mantener “Tweets y follows” dentro de RDS está perfecto: deja claro qué datos persiste.

No agregaría SQS ni DynamoDB en este diagrama, porque corresponden a otra etapa.

Una posible versión de las etiquetas sería:

```text
Cliente
   ↓ HTTP
Application Load Balancer
   ↓ Distribución de requests
ECS Service — Fargate
[N tasks con la API Go]
   ↕ Lecturas y escrituras
PostgreSQL en RDS
[Tweets y follows]
```

El borde exterior con el título **“Primera etapa productiva”** funciona muy bien para evitar que se confunda con lo implementado en el challenge.

En resumen: el flujo está correcto; el único cambio realmente necesario es representar **ECS Service + Fargate** con mayor precisión. Mandame también los diagramas de implementación actual y evolución y los revisamos con el mismo criterio.

> Asi _
>
> User context
>
> Images:
> - `/var/folders/vt/4czskcg155l3r7s3y41sx3740000gn/T/TemporaryItems/NSIRD_screencaptureui_4RLF2P/Captura de pantalla 2026-08-12 a la(s) 9.42.51 p. m..png`

Sí, así está mucho más preciso. Ahora se entiende correctamente que:

- ECS Service administra las tasks.
- Fargate es el entorno donde se ejecutan.
- El service mantiene un `desired count` de `N`.
- El ALB distribuye las solicitudes.
- Las tasks leen y escriben en PostgreSQL.

Solo cambiaría la flecha entre ECS y RDS por una flecha de doble punta:

```text
ECS Service ↔ PostgreSQL en RDS
```

No porque PostgreSQL inicie llamadas hacia la API, sino para representar visualmente el intercambio request/response. Aunque también podés conservar la flecha actual: la etiqueta “Lectura/Escritura” ya expresa correctamente que la API inicia ambas operaciones.

Para cerrar la decisión del timeline, podrías agregar debajo de RDS una nota pequeña:

> Timeline construido en lectura

o:

> Fan-out en lectura

Es opcional si eso ya queda explicado en el texto del documento. De hecho, para mantener el diagrama limpio, probablemente lo dejaría como está y explicaría la estrategia debajo.

Este diagrama ya está listo para usar.

> Este seria el de la evolución
>
> User context
>
> Images:
> - `/var/folders/vt/4czskcg155l3r7s3y41sx3740000gn/T/TemporaryItems/NSIRD_screencaptureui_rTuSZe/Captura de pantalla 2026-08-12 a la(s) 9.44.03 p. m..png`

Está bien encaminado y se entiende el cambio principal: PostgreSQL continúa siendo la fuente de verdad, mientras que DynamoDB pasa a ser una vista preparada para optimizar la lectura del timeline.

Haría algunos ajustes para que sea arquitectónicamente más preciso.

### 1. Cambiar el título

Para remarcar que no es una decisión tomada desde el comienzo:

> **Posible evolución optimizada para lecturas**

o:

> **Evolución si el fan-out en lectura se convierte en un cuello de botella**

La segunda conecta directamente el diagrama con las métricas que justificarían el cambio.

### 2. Aclarar el orden de publicación

El flujo conceptual de escritura debería ser:

```text
API → Guardar tweet/follow en PostgreSQL
API → Enviar mensaje a SQS
```

Tu dibujo ya muestra ambas conexiones, pero conviene numerarlas:

1. `Guardar tweet/follow`
2. `Enviar mensaje`

Eso deja claro que PostgreSQL es la fuente de verdad y que el mensaje representa algo que ya fue persistido.

Como decidiste no profundizar en outbox, podés documentar simplemente:

> El diagrama representa el flujo conceptual. La coordinación confiable entre la escritura en PostgreSQL y el envío del mensaje debería resolverse en un diseño de producción más detallado.

No necesitás agregar otro componente al dibujo.

### 3. Cambiar “Publicar evento” por “Enviar mensaje”

SQS es una cola. Para ser más preciso, usaría:

- API → SQS: **Enviar mensaje**
- SQS → worker: **Consumir mensaje**

Podés seguir llamando “evento” al contenido:

```text
Mensaje: TweetCreated
Mensaje: FollowCreated
```

Así distinguís el evento de dominio del mecanismo utilizado para transportarlo.

### 4. Identificar dónde se ejecutan los workers

El worker aparece correctamente como un componente separado, pero quedaría la pregunta de dónde corre.

Una opción coherente sería etiquetarlo:

> **Worker — ECS Service/Fargate**

No hace falta dibujar varias tasks. Podés agregar una nota pequeña:

> Escalable según mensajes pendientes en SQS.

Esto muestra que la API y el procesamiento asíncrono pueden escalar independientemente.

### 5. La lectura del worker desde PostgreSQL está bien

La flecha:

> Leer tweets y seguidores

es correcta porque el worker necesita información distinta según el evento:

- Ante un tweet nuevo, necesita conocer los seguidores del autor.
- Ante un follow nuevo, podría necesitar tweets recientes del usuario seguido para respetar la regla actual de incluir tweets anteriores al follow.

Esto justifica que el worker consulte la fuente de verdad y no dependa exclusivamente de la información contenida en el mensaje.

### 6. Aclarar qué escribe en DynamoDB

Cambiaría:

> Escribir timeline

por:

> Actualizar timelines preparados

Es más preciso porque un tweet puede actualizar los timelines de muchos seguidores.

### 7. La flecha API → DynamoDB está correcta

Representa que las consultas del timeline dejan de construirse mediante joins en PostgreSQL:

```text
GET timeline
API → DynamoDB
```

Podrías etiquetarla como:

> Leer timeline preparado

Las publicaciones y los follows continúan escribiéndose en PostgreSQL.

## Flujo final sugerido

```text
ESCRITURA

Cliente → ALB → API
                  ├─ 1. Guardar tweet/follow → PostgreSQL
                  └─ 2. Enviar mensaje → SQS
                                            ↓
                                      Worker en Fargate
                                      ├─ Leer datos → PostgreSQL
                                      └─ Actualizar → DynamoDB

LECTURA DEL TIMELINE

Cliente → ALB → API → Leer timeline preparado → DynamoDB
```

El diagrama actual ya comunica correctamente la arquitectura. Los cambios más importantes serían:

- Título condicional.
- “Enviar/consumir mensaje” en lugar de “publicar/entregar evento”.
- Numerar persistencia y envío.
- Identificar al worker como otro ECS Service sobre Fargate.
- Cambiar “Escribir timeline” por “Actualizar timelines preparados”.

Con esos ajustes queda sólido y fácil de defender.

> Porque esta mal publicar evento ? No son eventos los que maneja SQS ?

No está mal decir “publicar evento”. SQS puede transportar mensajes que representan eventos. Mi sugerencia era solo una precisión terminológica, no una corrección funcional.

La distinción es esta:

- **Evento:** algo que ocurrió en el dominio, por ejemplo `TweetCreated` o `FollowCreated`.
- **Mensaje:** la representación de ese evento que la API envía a SQS.
- **SQS:** la infraestructura que almacena y entrega ese mensaje.
- **Worker:** consume el mensaje y reacciona al evento.

Por ejemplo:

```json
{
  "event_type": "TweetCreated",
  "tweet_id": "123",
  "author_id": "user-1",
  "occurred_at": "2026-08-12T21:00:00Z"
}
```

Todo esto sería válido:

- “La API publica un evento en SQS”.
- “La API envía a SQS un mensaje que representa el evento”.
- “El worker consume eventos desde SQS”.

La segunda formulación es técnicamente más estricta porque SQS se define como una cola de mensajes, pero para un HLD **“Publicar evento” está perfectamente bien** si explicás que los eventos se transportan como mensajes mediante SQS.

Una diferencia relevante es que SQS funciona como una cola y no como un sistema de publicación/suscripción puro:

- En una cola, un mensaje normalmente es procesado por uno de los consumidores.
- Si varios tipos de consumidores independientes necesitaran recibir el mismo evento, una sola cola no alcanzaría, porque competirían por el mensaje.
- En ese caso podría evaluarse publicar el evento en SNS o EventBridge y distribuirlo luego a diferentes colas SQS.

Pero tu diseño tiene un solo propósito: que los workers preparen timelines. Para ese patrón, **SQS sola es suficiente** y no agregaría SNS ni EventBridge.

Podés conservar sin problemas las etiquetas:

```text
API → SQS: Publicar evento
SQS → Worker: Consumir evento
```

Incluso me parecen más naturales y comprensibles para tu documento. Solo evitaría “SQS publica el evento”: quien lo publica es la API; SQS lo almacena temporalmente y lo entrega al worker.

> Pero los timelines se escriben no se actualizan, es decir, al primero hay que escribirlos previo a actualizarlos. Porque decis que se actualizan ?

Tenés razón: **“actualizar” puede sonar a modificar algo que necesariamente ya existe**, y el primer timeline todavía no existe.

Yo usé “actualizar” en un sentido amplio: hacer que el estado materializado refleje el evento, ya sea creando un registro nuevo o modificando uno existente. Pero para tu diagrama, esa palabra puede generar una duda innecesaria.

La mejor etiqueta sería:

> **Materializar timeline**

o, más explícitamente:

> **Crear o actualizar timeline preparado**

En la práctica, el worker podría hacer operaciones como:

- Crear por primera vez la entrada del timeline de un usuario.
- Agregar un nuevo tweet a un timeline existente.
- Eliminar entradas antiguas si se conserva una ventana limitada.
- Evitar duplicados si el mismo evento se procesa más de una vez.

Además, en DynamoDB probablemente no existiría “un timeline” como un único objeto que después se actualiza. Lo habitual sería almacenar múltiples ítems asociados al usuario:

```text
Partition key: user_id
Sort key: created_at#tweet_id
```

Cuando llega un `TweetCreated`, el worker inserta un nuevo ítem para cada timeline correspondiente. Por eso **“Escribir timeline” no estaba mal**, aunque podría interpretarse como reescribir el timeline completo.

Para expresar mejor lo que sucede elegiría:

> **Escribir entradas del timeline**

Esa etiqueta cubre tanto la primera escritura como las siguientes inserciones y describe mejor el posible modelo en DynamoDB.

Por lo tanto, podés dejar la flecha como:

```text
Worker → DynamoDB: Escribir entradas del timeline
```

Es más precisa que mi sugerencia anterior.

> User context
>
> Images:
> - `/var/folders/vt/4czskcg155l3r7s3y41sx3740000gn/T/TemporaryItems/NSIRD_screencaptureui_6PdLYb/Captura de pantalla 2026-08-12 a la(s) 9.52.19 p. m..png`

Ahora quedó mucho más claro y consistente. Los cambios funcionan bien:

- Se muestra el orden conceptual: primero PostgreSQL y después SQS.
- “Publicar evento” es correcto.
- El worker aparece como un proceso separado ejecutado mediante ECS/Fargate.
- “Escribir entradas del timeline” contempla tanto la creación inicial como las inserciones posteriores.
- La API lee el timeline preparado directamente desde DynamoDB.

Hay un detalle técnico importante para documentar: en esta evolución, el worker necesita consultar **quiénes siguen a un autor** cuando recibe `TweetCreated`. En la primera etapa solo habíamos definido:

```text
PRIMARY KEY (follower_id, followed_id)
```

Esa clave sirve para encontrar a quién sigue un usuario, pero no para buscar eficientemente todos los seguidores de un autor. La evolución requeriría agregar:

```sql
CREATE INDEX idx_follows_followed_id
    ON follows (followed_id, follower_id);
```

Eso no hace falta para el challenge ni para el fan-out en lectura inicial; aparece como consecuencia directa del fan-out en escritura.

También podés explicar los dos eventos así:

- `TweetCreated`: el worker busca los seguidores del autor y escribe una entrada en el timeline de cada uno.
- `FollowCreated`: el worker busca tweets recientes del usuario seguido y crea las entradas iniciales en el timeline del nuevo seguidor, porque tu regla permite ver tweets anteriores al follow.

Como retoques visuales opcionales:

- Cambiar `Worker – ECS Service/Fargate` por `Worker — ECS Service / Fargate` para mantener el mismo estilo del bloque de la API.
- Cambiar el título por **“Posible evolución optimizada para lecturas”** para reforzar que depende de métricas.
- En la flecha inferior, usar **“Leer timeline preparado”**.

Fuera de eso, el diagrama ya está bien para incorporarlo al HLD. El nuevo índice inverso de follows es la consideración técnica más importante que conviene acompañar en el texto.

> Perfecto, entiendo. En la implementación actual necesito buscar a quién sigue un
> usuario, por eso me alcanza con la clave que ya había definido.
>
> En la evolución cambia el patrón de acceso, porque el worker necesita buscar
> quiénes siguen al usuario que publicó el tweet. Por eso agregaría el índice por
> followed_id solamente para esa etapa.
>
> También voy a aclarar los dos eventos:
>
> - TweetCreated, para agregar el tweet en los timelines de los seguidores.
> - FollowCreated, para agregar una cantidad limitada de tweets anteriores al
    >   timeline del usuario que empezó a seguirlo.
>
> Entiendo también que, para los usuarios con millones de seguidores, sus tweets
> no se guardarían en todos los timelines. En ese caso, al leer habría que buscar
> también sus tweets recientes y combinarlos con el timeline preparado en
> DynamoDB.
>
> Ves alguna otra contradicción importante entre las dos etapas?

Sí: hay una contradicción funcional importante y algunas consideraciones que conviene dejar explícitas.

La principal es el historial del `FollowCreated`.

En la implementación actual, cuando un usuario empieza a seguir a otro, puede consultar **todos sus tweets anteriores**, mediante paginación. Si en la evolución el worker copia solamente una cantidad limitada de tweets recientes, el comportamiento cambia:

- Etapa inicial: historial completo disponible.
- Evolución: solamente una ventana limitada materializada.

Tenés tres alternativas:

1. Aceptar y documentar ese cambio de producto.
2. Materializar inicialmente una cantidad limitada y consultar PostgreSQL para páginas antiguas.
3. Materializar todo el historial, con un costo potencialmente alto.

La segunda suele ser un compromiso razonable, pero complica la lectura. Para el HLD podés decir:

> Ante `FollowCreated`, se materializa una cantidad limitada de tweets recientes. La disponibilidad del historial anterior y el mecanismo para consultarlo deberían definirse según los requisitos del producto.

## Otros puntos para aclarar

### Consistencia eventual

Después de publicar o comenzar a seguir a alguien:

1. PostgreSQL ya contiene el cambio.
2. El mensaje puede estar esperando en SQS.
3. DynamoDB todavía puede no reflejarlo.

Durante ese intervalo, el timeline puede estar desactualizado. No contradice la primera etapa, pero cambia su modelo de consistencia y debe quedar explícito.

### Duplicados y concurrencia

Podrían coincidir:

- El backfill provocado por `FollowCreated`.
- Un `TweetCreated` del mismo autor.

Ambos podrían intentar escribir la misma entrada. La clave de DynamoDB debería permitir que la escritura sea idempotente, por ejemplo:

```text
PK: timeline_owner_id
SK: created_at#tweet_id
```

El mismo tweet para el mismo timeline produciría la misma clave y no quedaría duplicado.

### Orden de los eventos

No conviene depender de que los mensajes se procesen exactamente en orden. El orden visible del timeline debería depender de `created_at + tweet_id`, no del momento en que el worker procesó cada evento.

### Estrategia híbrida

Tu interpretación es correcta. Para autores masivos:

- Sus tweets no se copian a millones de timelines.
- La API consulta sus tweets recientes en PostgreSQL.
- Los combina con el timeline preparado de DynamoDB.
- Ordena y pagina el resultado conjunto.

Eso introduce una nueva necesidad: durante la lectura, la API debe saber cuáles de las cuentas seguidas utilizan fan-out en escritura y cuáles se resuelven en lectura.

También complica el cursor, porque la página mezcla resultados de dos fuentes. No hace falta resolver el algoritmo en el HLD, pero sí reconocerlo como trade-off:

> La estrategia híbrida reduce el costo de fan-out para autores con muchos seguidores, a cambio de volver más compleja la lectura, combinación y paginación del timeline.

### Follows posteriores a la materialización

El índice inverso que identificaste está perfecto:

```text
(followed_id, follower_id)
```

Para la estrategia inicial no está justificado. Para fan-out en escritura pasa a ser esencial porque permite buscar eficientemente los destinatarios de `TweetCreated`.

## Conclusión

No veo otra contradicción estructural entre las etapas. La transición es coherente:

- PostgreSQL permanece como fuente de verdad.
- DynamoDB es una vista derivada.
- El costo se mueve de la lectura hacia la escritura.
- La consistencia pasa a ser eventual.
- La estrategia híbrida evita el peor caso de autores con millones de seguidores.

La única contradicción funcional real que necesitás decidir o declarar es **historial completo frente a historial limitado después de `FollowCreated`**. El resto son trade-offs esperables de la evolución.

> Voy a dejar explícito que copiar una cantidad limitada de tweets
> anteriores cambia el comportamiento respecto de la implementación actual, donde
> se puede consultar todo el historial.
>
> No voy a definir ahora cómo consultar el historial más antiguo porque dependería
> de los requisitos del producto y agregaría complejidad a una evolución que
> solamente se evaluaría si las métricas la requieren.
>
> También voy a mantener documentados como trade offs la consistencia eventual,
> la idempotencia y la mayor complejidad de lectura y paginación de la estrategia
> híbrida.
>
> Con esto cerramos el HLD

Sí, con eso el HLD queda conceptualmente cerrado y consistente.

La documentación distingue claramente:

1. **Challenge implementado:** API Go dockerizada, persistencia en memoria, fan-out en lectura y acceso al historial completo.
2. **Primera etapa productiva:** ALB, ECS Service sobre Fargate y PostgreSQL en RDS, manteniendo el comportamiento y la estrategia de lectura.
3. **Evolución condicionada por métricas:** SQS, workers y timelines preparados en DynamoDB, con posible estrategia híbrida.

También quedan reconocidos los principales trade-offs:

- Consistencia eventual.
- Procesamiento idempotente ante reintentos.
- Nuevo índice por `followed_id`.
- Historial anterior limitado después de `FollowCreated`.
- Mayor costo de escritura.
- Complejidad adicional de combinación y paginación en la estrategia híbrida.
- PostgreSQL como fuente de verdad y DynamoDB como vista derivada y reconstruible.

Es un diseño gradual, justificable y alineado con tu experiencia: no afirma que la evolución sea necesaria, sino que explica qué problema resolvería y qué costos introduciría. Con eso, el HLD está listo para redactarse en su versión final.
