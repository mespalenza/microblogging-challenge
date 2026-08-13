# Definir accesos y modelo de datos

> Bueno chats vamos a definir patrones de acceso y modelo de datos
>
> Quiero definir los patrones de acceso y después proponer el modelo de datos para el challenge.
> Según la consigna, las operaciones principales son:
> - Publicar un tweet
> - Seguir a otro usuario
> - Obtener el timeline con los tweets de los usuarios seguidos
>
> Mi analisis es que publicar y seguir son escrituras, mientras que consultar el timeline es la lectura. La consigna dice que la aplicación debe estar optimizada para lecturas, pero no da un read/write ratio ni datos de tráfico, así que no quiero inventar esos números
> Para buscar el timeline pienso este recorrido =>  buscar los usuarios que sigue el usuario, obtener sus tweets, mezclarlos y ordenarlos del más reciente al más antiguo. El orden es un supuesto mío porque no está definido
>
> También asumiría inicialmente:
> Un usuario no puede seguirse a sí mismo.
> Seguir dos veces al mismo usuario es una operación idempotente y no crea relaciones duplicadas.
> El timeline devuelve una cantidad limitada de tweets; la paginación sería una decisión de diseño y no un requisito explícito.
> Quiero que revisemos primero si estos patrones y supuestos tienen sentido. Después quiero derivar las entidades, relaciones e índices necesarios, sin agregar funcionalidades que el challenge no pide.
> Necesito que avancemos de forma práctica e incremental. Cuestioná mis decisiones cuando haya un problema importante o un trade-off que deba poder defender, pero evitemos extendernos indefinidamente sobre detalles menores.
> Para cada tema:
> Marcame los problemas relevantes.
> Dejame responder o elegir una alternativa.
> Cuando la decisión sea razonable para el alcance del challenge, la documentamos y la damos por cerrada.
> Si falta información de negocio, definimos un supuesto explícito y avanzamos.
> No reabras decisiones cerradas salvo que aparezca una contradicción importante.
> No busco diseñar un Twitter perfecto ni cubrir todos los escenarios de producción. Quiero entregar una solución acotada, consistente y defendible dentro del tiempo disponible. No me des una solución cerrada: revisá mi razonamiento y avancemos paso a paso, manteniendo el foco en terminar esta sección.

Tu razonamiento es consistente con la consigna y está bien no inventar volúmenes ni un ratio de lectura/escritura. Podemos avanzar con supuestos cualitativos y dejar explícito qué optimizamos.

Sobre los supuestos:

- Un usuario no puede seguirse a sí mismo: razonable. Lo documentaría y cerraría.
- Seguir dos veces al mismo usuario es idempotente: razonable y además implica una restricción de unicidad sobre la relación. Lo documentaría y cerraría.
- Timeline limitado y paginado: razonable como decisión técnica necesaria. Todavía no hace falta definir tamaño ni mecanismo exacto.
- Orden cronológico descendente: es el supuesto más simple y defendible cuando no existe una definición de relevancia. Conviene aclarar que se ordena por fecha de publicación, del más reciente al más antiguo.
- El timeline contiene tweets de los usuarios seguidos: seguiría literalmente la consigna. Como el usuario no puede seguirse a sí mismo, sus propios tweets no aparecerían. No agregaría esa funcionalidad sin pedirla.

El punto importante que cuestionaría es este recorrido:

> buscar seguidos → obtener sus tweets → mezclarlos → ordenarlos

Es una implementación de timeline mediante fan-out en lectura. Es válida y probablemente la más simple para el challenge, pero entra en tensión con “optimizada para lecturas”: el costo de cada consulta crece con la cantidad de usuarios seguidos y con los tweets candidatos que haya que combinar.

La alternativa sería fan-out en escritura: al publicar, insertar o distribuir el tweet en los timelines de los seguidores. Las lecturas serían mucho más simples, a cambio de escrituras amplificadas, datos derivados y mayor complejidad.

No necesitamos diseñar ambas. Tenemos que elegir qué interpretación defender:

1. **Simplicidad con fan-out en lectura:** aceptamos el recorrido propuesto, apoyado por índices y consultas limitadas. Explicamos que, sin datos de escala, evitamos introducir materialización y duplicación.
2. **Optimización fuerte de lecturas:** materializamos el timeline mediante fan-out en escritura, tomando literalmente el requisito de priorizar lecturas.

Para un challenge acotado, yo elegiría la primera, pero dejaría registrada su limitación. ¿Querés sostener fan-out en lectura o preferís que exploremos la materialización del timeline antes de cerrar esta decisión?

> Para este challenge voy a usar fan-out en lectura
> Sé que puede tener un costo mayor al consultar el timeline, porque primero hay que buscar los usuarios seguidos, traer sus tweets, mezclarlos y ordenarlos. Pero al no tener datos sobre tráfico, cantidad de usuarios seguidos o latencia esperada, me parece que implementar fan-out en escritura agregaría bastante complejidad sin saber si realmente es necesaria
> Para mejorar las lecturas, voy a limitar y paginar los resultados y después definir los índices necesarios para estas consultas
> Voy a documentar que esta solución puede volverse costosa si un usuario sigue a muchas cuentas. Si tuviéramos métricas que mostraran problemas de latencia, se podría evaluar fan-out en escritura
> Para el alcance del challenge quiero cerrar fan-out en lectura como decisión y avanzar con las entidades y relaciones necesarias. Dejemos los índices para después de definir el modelo

Queda cerrada la decisión de **fan-out en lectura**, con esta justificación:

> Se prioriza una solución simple y consistente con la información disponible. El timeline se construye consultando las cuentas seguidas y combinando sus tweets. Los resultados estarán limitados y paginados. Se reconoce que el costo de lectura puede crecer para usuarios que siguen muchas cuentas; si las métricas evidenciaran problemas de latencia, podría evaluarse fan-out en escritura.

No la reabrimos salvo contradicción importante.

Para el modelo lógico mínimo veo tres entidades:

- **Usuario**
    - `id`
    - No agregaría todavía nombre, email u otros atributos porque no participan en los patrones definidos.

- **Tweet**
    - `id`
    - `author_id`
    - `content`
    - `created_at`
    - Cada tweet pertenece a exactamente un usuario.
    - Un usuario puede publicar muchos tweets.

- **Follow**
    - `follower_id`: usuario que sigue.
    - `followed_id`: usuario seguido.
    - Representa una relación dirigida muchos-a-muchos entre usuarios.
    - La combinación `(follower_id, followed_id)` debe ser única para garantizar idempotencia.
    - Debe cumplirse `follower_id != followed_id`.

No propondría una entidad `Timeline`: con fan-out en lectura, el timeline es el resultado de una consulta, no información almacenada. Tampoco agregaría likes, retweets, comentarios, perfiles o relación de seguidores inversa independiente.

Hay una decisión relevante antes de cerrar este modelo: `created_at` por sí solo puede no determinar un orden estable si dos tweets tienen la misma marca de tiempo. Propongo definir el timeline mediante:

`created_at DESC, tweet_id DESC`

Así la paginación tiene un orden total y determinista. Esto no obliga todavía a elegir el tipo de ID ni el índice.

¿Cerramos estas tres entidades y ese criterio de desempate?

> Sí, cerremos el modelo con User, Tweet y Follow
> No agregaría una entidad Timeline porque como elegimos fan out en lectura, el timeline se arma cuando el usuario lo consulta y no queda guardado
> Para ordenar los tweets usaría => created_at DESC, tweet_id DESC
> Primero se ordenan del más reciente al más antiguo. Si dos tweets tienen la misma fecha, uso el ID para desempatar y asegurar que siempre se mantenga el mismo orden. Esto también nos va a servir cuando definamos la paginación.
>
> Entonces cerramos en
> - Las tres entidades con sus relaciones
> - Que no se puede repetir la combinación (follower_id, followed_id)
> - Que un usuario no se puede seguir a sí mismo
> - El orden del timeline por created_at DESC, tweet_id DESC
> - Que el timeline no se guarda, sino que se arma cuando se consulta
    > Por ahora no agregaría más atributos. Sigamos con las consultas que tiene que soportar este modelo y después definimos los índices necesarios

Queda cerrado el modelo y las decisiones enumeradas. No agregamos atributos ni una entidad `Timeline`.

El modelo debe soportar estos tres patrones de acceso:

1. **Publicar un tweet**
    - Entrada: `tweet_id`, `author_id`, `content`, `created_at`.
    - Operación: insertar un `Tweet` asociado al autor.
    - La existencia del autor puede garantizarse mediante la relación con `User`.

2. **Seguir a un usuario**
    - Entrada: `follower_id`, `followed_id`.
    - Operación: insertar un `Follow`.
    - Si la combinación ya existe, la operación termina correctamente sin duplicarla.
    - Si ambos identificadores son iguales, la operación se rechaza.
    - La existencia de ambos usuarios puede garantizarse mediante sus relaciones con `User`.

3. **Obtener el timeline**
    - Entrada: `user_id`, límite y, para páginas posteriores, una referencia de continuación.
    - Recorrido:
        1. Buscar los `followed_id` correspondientes al usuario.
        2. Buscar los tweets cuyos `author_id` pertenezcan a ese conjunto.
        3. Ordenarlos por `created_at DESC, tweet_id DESC`.
        4. Devolver una cantidad limitada.

El único punto relevante que debemos decidir ahora es el tipo de paginación:

- **Offset:** más sencillo de explicar, pero puede degradarse en páginas profundas y producir desplazamientos si se publican tweets entre consultas.
- **Cursor:** usa el último `(created_at, tweet_id)` de la página anterior. Es consistente con el orden definido y evita recorrer resultados anteriores.

Con cursor, la siguiente página buscaría tweets que cumplan conceptualmente:

```sql
created_at < :cursor_created_at
OR (
    created_at = :cursor_created_at
    AND tweet_id < :cursor_tweet_id
)
```

manteniendo:

```sql
ORDER BY created_at DESC, tweet_id DESC
LIMIT :page_size
```

Recomiendo cerrar **paginación por cursor**. Es una complejidad pequeña, aprovecha directamente el desempate ya definido y luego orientará con claridad el índice de tweets. ¿La cerramos así?

> Creo que User tiene que seguir siendo una entidad porque es el usuario el que publica tweets, sigue a otros usuarios y también puede ser seguido
> Como la consigna dice que todos los usuarios son válidos y no pide registro ni autenticación, voy a asumir que los usuarios ya existen por fuera de esta aplicación. Nosotros solamente recibimos sus ids y no validamos su existencia
> Por eso mantendría User en el modelo, pero no guardaría usuarios en una tabla ni crearía usuarios automáticamente. Tweet y Follow solamente van a usar esos ids
> Entonces quedan las tres entidades en el modelo, pero en la persistencia solamente necesitamos guardar Tweet y Follow
> Si no ves un problema importante con esto, quiero cerrar la decisión y avanzar con los índices

No veo un problema que impida avanzar, pero documentaría una precisión para que el modelo no parezca contradictorio:

- `User` es una **entidad externa del dominio**, no una entidad persistida por este servicio.
- El servicio recibe `user_id` válidos y no administra su ciclo de vida.
- `Tweet.author_id`, `Follow.follower_id` y `Follow.followed_id` son referencias lógicas, no claves foráneas hacia una tabla local.
- En consecuencia, la base de datos no puede verificar existencia ni aplicar integridad referencial o eliminaciones en cascada. Esto es aceptable por el supuesto explícito de que todos los IDs recibidos son válidos.

Con esa formulación, cerraría la decisión: **tres entidades conceptuales; solo `Tweet` y `Follow` se persisten**.

Para los índices mínimos propondría:

### Tweet

```text
PRIMARY KEY (tweet_id)
INDEX (author_id, created_at DESC, tweet_id DESC)
```

- La clave primaria identifica cada tweet.
- El índice compuesto permite buscar los tweets de cada usuario seguido y recorrerlos en el mismo orden usado por el timeline.
- `tweet_id` al final resuelve empates y permite implementar el cursor.
- No incluiría `content` para crear un índice cubriente: aumentaría su tamaño sin evidencia de que sea necesario.

### Follow

```text
PRIMARY KEY (follower_id, followed_id)
```

Esta clave compuesta cumple dos objetivos:

- Evita relaciones duplicadas y hace posible que “seguir” sea idempotente.
- Permite obtener eficientemente todos los `followed_id` de un usuario porque `follower_id` es la primera columna.

Mantendría además la restricción:

```text
CHECK (follower_id <> followed_id)
```

No agregaría un índice por `followed_id`: serviría para consultar quiénes siguen a un usuario, pero ese patrón no está pedido. Tampoco agregaría un índice independiente por `author_id`, porque ya es el prefijo del índice compuesto de `Tweet`.

Estos son los índices mínimos derivados directamente de los patrones acordados. El punto a validar antes de cerrarlos es que optimizan la obtención de tweets por cada autor, aunque la base todavía tendrá que combinar los resultados de varios autores para construir el orden global. Eso es una consecuencia esperada del fan-out en lectura, no algo que otro índice simple elimine.

> Hay una parte que no entiendo, cuando decís que no incluirías content para crear un índice cubriente que significa  que el índice sea cubriente?
> Entiendo que author_id, created_at y tweet_id se usan para buscar, ordenar y paginar

Sí, entendiste correctamente: `author_id`, `created_at` y `tweet_id` están en el índice porque participan en el filtrado, el orden y la paginación.

Un **índice cubriente** es un índice que contiene todas las columnas necesarias para responder una consulta. Así, la base puede devolver el resultado leyendo solamente el índice, sin tener que buscar después cada fila en la tabla.

Por ejemplo, el timeline necesita algo parecido a:

```sql
SELECT tweet_id, author_id, content, created_at
FROM tweet
WHERE author_id IN (...)
ORDER BY created_at DESC, tweet_id DESC
LIMIT :page_size;
```

Con este índice:

```text
(author_id, created_at DESC, tweet_id DESC)
```

la base puede localizar y ordenar eficientemente los tweets candidatos, pero `content` no está en el índice. Para devolverlo, debe usar el `tweet_id` encontrado y leer la fila correspondiente de `Tweet`.

Un índice cubriente podría ser conceptualmente:

```text
(author_id, created_at DESC, tweet_id DESC)
INCLUDE (content)
```

La ventaja es evitar esa segunda lectura. La desventaja es importante: el contenido del tweet suele ser bastante más grande que un ID o una fecha, por lo que el índice:

- ocupa más espacio;
- requiere más memoria y almacenamiento;
- aumenta el costo de publicar tweets;
- puede reducir la cantidad de entradas que caben en cada página del índice.

Para este challenge no tenemos evidencia de que esas lecturas adicionales sean un problema. Por eso mantendría el índice compacto y dejaría `content` únicamente en la tabla.

Entonces podemos documentarlo así:

> El índice de `Tweet` contiene las columnas necesarias para filtrar, ordenar y paginar. No se incluye `content` para convertirlo en un índice cubriente, porque aumentaría significativamente el tamaño y el costo de mantenimiento del índice sin métricas que justifiquen esa optimización.

> Antes de seguir quiero que veas como estoy separando la documentación
> En business.txt dejaría las reglas y supuestos sobre el comportamiento esperado y lo separe un entre supuestos y lo que dice el enunciado
> Pensaba crear un nuevo file architecture.md donde dejaría la definicion del modelo fan out, paginación e índices
> Esta sería mi propuesta
> Marques solo si algo está ubicado en el archivo incorrecto, si existe una contradicción importante o si falta alguna decisión que ya hayamos cerrado
>
> REGLAS DE NEGOCIO Y SUPUESTOS
>
> Alcance
>
> La aplicación permite:
>
> - Publicar un tweet
> - Seguir a otro usuario
> - Consultar un timeline con los tweets de los usuarios seguidos
>
> Reglas definidas por la consigna
>
> - Los tweets tienen un límite de caracteres
> - Los usuarios pueden seguir a otros usuarios
> - El timeline contiene los tweets de los usuarios seguidos
> - Todos los usuarios recibidos se consideran válidos
> - No se requiere registro, autenticación ni manejo de sesiones
>
> Supuestos adoptados para el challenge
>
> Tweets
>
> - El límite máximo de un tweet es de 280 caracteres
> - No se permiten tweets vacíos ni compuestos únicamente por espacios
> - Los espacios al comienzo y al final se ignoran al validar y guardar el tweet
>
> Seguimiento
>
> - Un usuario no puede seguirse a sí mismo
> - Seguir más de una vez al mismo usuario no crea relaciones duplicadas
> - La combinación (follower_id, followed_id) debe ser única
>
> Timeline
>
> - El timeline contiene solamente los tweets publicados por los usuarios seguidos
> - Los tweets propios no se incluyen
> - Se incluyen los tweets publicados antes de comenzar a seguir a un usuario
> - Los tweets se ordenan desde el más reciente hasta el más antiguo
> - Si dos tweets tienen la misma fecha, se ordenan por identificador descendente para mantener un orden determinista
> - Cuando no existen tweets para mostrar, se devuelve una colección vacía
> - El timeline se devuelve de forma limitada y paginada
>
> Fuera de alcance
>
> Esta versión no incluye:
>
> - Registro, autenticación o manejo de sesiones
> - Administración del ciclo de vida de los usuarios
> - Edición o eliminación de tweets
> - Dejar de seguir usuarios
> - Likes, respuestas, retweets o notificaciones
> - Perfiles de usuario
> - Desarrollo de frontend

La separación es correcta en general. Solo marcaría estos puntos:

- **Movería** `La combinación (follower_id, followed_id) debe ser única` a `architecture.md`. En `business.txt` alcanza con expresar el comportamiento: seguir dos veces no crea relaciones duplicadas. La clave o restricción única es cómo lo garantiza la persistencia.

- En el orden del timeline usaría el nombre preciso:
  > Si dos tweets tienen el mismo `created_at`, se ordenan por `tweet_id` descendente.

- `Se incluyen los tweets publicados antes de comenzar a seguir a un usuario` está bien ubicado en reglas de negocio, pero es una decisión nueva que no habíamos cerrado explícitamente. Es coherente con el modelo elegido, porque `Follow` no registra cuándo comenzó la relación. Conviene reconocerla como supuesto, tal como hiciste.

- No falta ninguna decisión de negocio que ya hayamos cerrado.

En `architecture.md` deberían quedar posteriormente las decisiones técnicas ya acordadas:

- `User` es una entidad conceptual externa y no se persiste.
- Solo se persisten `Tweet` y `Follow`.
- No existen claves foráneas locales hacia `User`.
- El timeline usa fan-out en lectura y no se almacena.
- Relaciones entre las tres entidades.
- Restricción única de `Follow`.
- Restricción para impedir el auto-seguimiento.
- Estrategia concreta de paginación, cuando la cerremos.
- Índices, cuando terminemos de validarlos.

Fuera de esos puntos, `business.txt` está bien delimitado y no presenta contradicciones importantes.

> Estoy de acuerdo. Voy a sacar de business.txt la combinación (follower_id, followed_id) como única y voy a dejar solamente que seguir dos veces al mismo usuario no crea una relación duplicada. La restricción técnica para garantizarlo la dejo en architecture.md
>
> También voy a cambiar la parte del orden para que quede más clara:
> Si dos tweets tienen el mismo created_at, se ordenan por tweet_id descendente
>
> Dejo como supuesto que el timeline incluye los tweets publicados antes de empezar a seguir a un usuario. Como no estamos guardando la fecha en la que empezó el follow, me parece coherente para este challenge.
> Con estos cambios cierro business.txt.
>
> En architecture.md voy a dejar el modelo, qué entidades se guardan, la decisión de fanout en lectura, las restricciones y las limitaciones que ya vimos
> Antes de cerrar ese documento, quiero terminar de entender y definir la paginación por cursor y los índices. Sigamos con la duda del índice cubriente y después cerramos esa parte

Perfecto, `business.txt` queda cerrado.

Volviendo al índice cubriente, la diferencia práctica es esta:

### Índice propuesto

```text
(author_id, created_at DESC, tweet_id DESC)
```

Estas columnas forman la estructura ordenada del índice:

- `author_id`: localiza los tweets de las cuentas seguidas.
- `created_at`: permite recorrerlos del más reciente al más antiguo.
- `tweet_id`: desempata y permite continuar desde un cursor.

El índice devuelve referencias a las filas encontradas. Después, la base consulta esas filas en `Tweet` para obtener `content`.

### Variante cubriente

Conceptualmente:

```text
(author_id, created_at DESC, tweet_id DESC)
INCLUDE (content)
```

`content` no se utiliza para buscar ni ordenar. Solo se copia dentro del índice para que la consulta pueda obtener todos los datos del timeline sin volver a leer la tabla.

Es importante distinguir:

- Las columnas antes de `INCLUDE` definen cómo se busca y ordena.
- Las columnas incluidas solamente están disponibles para devolver el resultado.
- La sintaxis exacta depende de la base de datos elegida.

Para este challenge, el beneficio sería evitar la lectura de las filas finales de `Tweet`. Como el timeline está limitado, probablemente estaríamos evitando una cantidad pequeña de lecturas por página. A cambio, copiaríamos el contenido de todos los tweets dentro del índice, aumentando su tamaño y el costo de cada publicación.

Mi propuesta para cerrar esta parte sería:

> Se define un índice compuesto sobre `(author_id, created_at DESC, tweet_id DESC)`. No se incluye `content` como columna cubierta porque no participa en el filtrado, ordenamiento ni paginación, y duplicarlo aumentaría el tamaño y costo de mantenimiento del índice sin métricas que justifiquen esa optimización.

No hace falta mencionar índices cubrientes en `architecture.md` salvo que quieras documentarlo como alternativa descartada. La decisión efectiva es simplemente mantener un índice compacto. ¿Te resulta clara la diferencia o querés que recorramos una consulta concreta mostrando qué lee primero del índice y qué busca después en la tabla?

> Sí entiendo la diferencia
> author_id, created_at y tweet_id forman parte del índice porque los necesitamos para buscar los tweets, ordenarlos y paginarlos.
> En cambio, content no se usaría para buscar ni ordenar. Solamente se agregaría al índice para no tener que ir después a buscarlo a la tabla
> Como el timeline va a devolver una cantidad limitada de tweets, no me parece necesario duplicar el contenido de todos los tweets dentro del índice. Eso haría que el índice ocupe más espacio y también agregaría costo al publicar
> Entonces dejo el índice así => (author_id, created_at DESC, tweet_id DESC)
> No agregaría content ni documentaría el índice cubriente porque no es la opción que vamos a usar
> Sigamos con la paginación por cursor para terminar de cerrar los índices y architecture.md

La paginación por cursor se apoya exactamente en el orden que ya definimos:

```text
created_at DESC, tweet_id DESC
```

El cursor representa la posición del último tweet devuelto, usando ambos valores.

### Primera página

La solicitud incluye solamente un límite:

```sql
WHERE author_id IN (:followed_ids)
ORDER BY created_at DESC, tweet_id DESC
LIMIT :page_size
```

Supongamos que el último tweet devuelto tiene:

```text
created_at = 2026-08-09 14:30:00
tweet_id   = 150
```

El cursor representa ese par:

```text
(2026-08-09 14:30:00, 150)
```

### Página siguiente

Se solicitan únicamente tweets posteriores a esa posición en el recorrido; es decir, tweets más antiguos:

```sql
WHERE author_id IN (:followed_ids)
  AND (
      created_at < :cursor_created_at
      OR (
          created_at = :cursor_created_at
          AND tweet_id < :cursor_tweet_id
      )
  )
ORDER BY created_at DESC, tweet_id DESC
LIMIT :page_size
```

Es necesario incluir `tweet_id` en el cursor. Si usáramos solamente `created_at`, podríamos omitir tweets cuando varios compartieran la misma fecha.

### Comportamiento ante nuevos tweets

Si se publica un tweet entre la primera y la segunda consulta, normalmente quedará antes del cursor y no desplazará ni repetirá los resultados de la página siguiente. Para verlo, el cliente deberá volver a solicitar la primera página.

Esto da una navegación estable, aunque no representa una fotografía inmutable del timeline completo. Para conseguir una fotografía fija necesitaríamos agregar más estado, algo que no parece necesario para el challenge.

### Representación del cursor

A nivel lógico contiene:

```text
created_at + tweet_id
```

Externamente conviene tratarlo como un valor opaco, por ejemplo codificando ambos campos en un string. Así el cliente no necesita conocer su estructura interna. No hace falta definir ahora un formato específico.

### Detección de otra página

Una práctica sencilla es consultar `page_size + 1` elementos:

- Se devuelven como máximo `page_size`.
- Si existe un elemento adicional, se genera `next_cursor`.
- Si no existe, `next_cursor` queda vacío.

Esto no cambia el modelo ni requiere otro índice.

La decisión para `architecture.md` podría quedar así:

> El timeline utiliza paginación por cursor o keyset pagination. El cursor identifica el último tweet devuelto mediante `(created_at, tweet_id)`, coincidiendo con el orden `created_at DESC, tweet_id DESC`. La página siguiente recupera únicamente elementos ubicados después de esa posición. El cursor se expone como un valor opaco y no garantiza una fotografía inmutable del timeline.

El índice de `Tweet` queda alineado con esta decisión:

```text
(author_id, created_at DESC, tweet_id DESC)
```

Solo quedaría confirmar si también querés cerrar la estrategia de solicitar `page_size + 1` para determinar si existe una página siguiente.

> La paginacion por cursor ya la habíamos definido. Mi idea no era volver a analizarla, sino ver como la dejamos escrita en architecture.md
> Yo pondría:
> El timeline usa paginación por cursor. El cursor guarda created_at y tweet_id del último tweet devuelto, siguiendo el mismo orden que definimos para el timeline: created_at DESC, tweet_id DESC
> Para traer la página siguiente, se buscan los tweets más antiguos que están después de esa posición. El cursor se devuelve como un valor opaco para que el cliente no necesite saber qué información contiene
> Si se publica un tweet nuevo mientras el usuario está recorriendo las páginas, ese tweet queda antes del cursor y no mueve los resultados de las páginas siguientes.
>
> Buscamos un resultado más de los que vamos a devolver
> Si existe => devolvemos next_cursor
> Si no existe => next_cursor queda vacío
>
> Con esto cierro la paginación. Quiero agregarlo junto con los índices que ya definimos en architecture.md y después revisar el documento completo, sin volver a abrir las decisiones que ya cerramos salvo veas algo importante

<details><summary>1 previous message</summary>

> Entendido: la paginación queda cerrada y ahora solo toca documentarla. Voy a ubicar `architecture.md` y `business.txt`, incorporar lo acordado si corresponde y después revisar el documento completo únicamente por contradicciones, ubicación y decisiones faltantes.
</details>

La redacción propuesta es correcta y la paginación queda cerrada. Solo ajustaría una frase para mayor precisión: en vez de “no mueve los resultados”, diría que el tweet nuevo queda fuera del recorrido en curso y aparecerá al consultar nuevamente la primera página.

No encontré todavía `architecture.md` ni `business.txt` en el workspace, así que no puedo revisar el documento completo aún. Cuando lo crees o pegues aquí, lo revisaré únicamente por:

- ubicación incorrecta;
- contradicciones importantes;
- decisiones cerradas que falten.

No volveré a discutir las decisiones ya cerradas.

> Busca los archivos y revisalos, estan en /Users/emispalenza/GolandProjects/microblogging-challenge

<details><summary>2 previous messages</summary>

> Voy a revisar primero las instrucciones del repositorio y luego leer ambos documentos completos. Haré la revisión con el alcance acordado, sin modificar archivos ni reabrir decisiones cerradas.
>
> Detecté que la salida combinada se cortó justo al final de `business.txt`, así que voy a releer ambos por separado con números de línea para no confundir un problema real del archivo con el formato de la salida.
</details>

Encontré tres observaciones relevantes:

1. **`business.txt` está truncado.** Termina en “Los tweets se ordenan desde el más re”, por lo que faltan el resto de las decisiones del timeline y la sección “Fuera de alcance”. Hay que restaurar ese contenido antes de considerarlo cerrado: [business.txt](~/GolandProjects/microblogging-challenge/business.txt:39).

2. **El `CHECK` no es un índice.** En [architecture.md](~/GolandProjects/microblogging-challenge/docs/architecture.md:292), `CHECK (follower_id <> followed_id)` aparece bajo “Índices”. Esa restricción ya está correctamente documentada en “Restricciones de Follow”; la quitaría de esta sección o la renombraría como “Índices y restricciones”.

3. **Quedó documentada la alternativa descartada de incluir `content`.** En [architecture.md](~/GolandProjects/microblogging-challenge/docs/architecture.md:284) se explica por qué no se incorpora al índice. Como decidiste no documentar el índice cubriente, eliminaría ese párrafo y la oración posterior sobre buscar el contenido en la tabla. Basta con dejar el índice elegido y su finalidad.

Fuera de eso, `architecture.md` contiene las decisiones cerradas, sin contradicciones importantes: modelo conceptual y persistente, fan-out en lectura, orden, cursor, consulta de un elemento adicional, restricciones, índices y limitaciones.

::code-comment{title="[P1] Documento truncado" body="El archivo termina a mitad de la regla de orden del timeline. Faltan las demás reglas acordadas y la sección «Fuera de alcance», por lo que business.txt todavía no refleja todas las decisiones cerradas." file="/Users/emispalenza/GolandProjects/microblogging-challenge/business.txt" start=39 priority=1}

::code-comment{title="[P2] CHECK ubicado como índice" body="CHECK (follower_id <> followed_id) es una restricción, no un índice, y ya está documentada en «Restricciones de Follow». Conviene quitarla de este bloque o renombrar la sección." file="/Users/emispalenza/GolandProjects/microblogging-challenge/docs/architecture.md" start=292 end=297 priority=2}

::code-comment{title="[P2] Alternativa descartada documentada" body="Este párrafo conserva la explicación sobre incluir content en el índice, aunque se decidió no documentar la opción de índice cubriente. Puede eliminarse junto con la oración siguiente; el índice seleccionado ya está justificado por los patrones de acceso." file="/Users/emispalenza/GolandProjects/microblogging-challenge/docs/architecture.md" start=284 end=290 priority=2}

> Ahi arregle lo mencionado

<details><summary>1 previous message</summary>

> Voy a releer los dos archivos y comprobar únicamente que quedaron resueltas las tres observaciones anteriores y que no se introdujo ninguna contradicción importante.
</details>

Quedó corregido correctamente.

- [business.txt](~/GolandProjects/microblogging-challenge/business.txt) está completo y contiene todas las reglas, supuestos y elementos fuera de alcance acordados.
- [architecture.md](~/GolandProjects/microblogging-challenge/docs/architecture.md) ubica el `CHECK` dentro de las restricciones de `Follow`, no entre los índices.
- Se eliminó la explicación del índice cubriente.
- La paginación por cursor y los índices coinciden con las decisiones cerradas.
- No encontré contradicciones importantes ni decisiones acordadas ausentes.

Podemos dar por cerrada esta versión de `architecture.md`.
