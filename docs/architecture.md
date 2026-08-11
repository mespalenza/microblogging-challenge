# Architecture

## Objetivo

La aplicación implementa una versión simplificada de una plataforma de
microblogging.

Permite:

- Publicar tweets
- Seguir usuarios
- Consultar un timeline con los tweets de los usuarios seguidos

La consigna indica que la aplicación debe estar optimizada para lecturas y
pensada para millones de usuarios. Sin embargo, no define volúmenes de tráfico,
cantidad de usuarios seguidos, latencia esperada ni un read/write ratio.

Por eso se elige una solución simple para el alcance del challenge y se dejan
documentadas sus limitaciones. Si existieran métricas reales, las decisiones
podrían volver a evaluarse.

## Modelo de dominio

El modelo conceptual contiene tres entidades: `User`, `Tweet` y `Follow`.

### User

Representa al usuario que publica tweets, sigue a otros usuarios y puede ser
seguido.

`User` es una entidad externa para este servicio. La aplicación recibe
identificadores de usuario válidos, pero no administra su registro ni su ciclo
de vida.

Por este motivo, `User` forma parte del modelo conceptual, pero no se guarda en
este servicio.

### Tweet

Contiene:

- `tweet_id`
- `author_id`
- `content`
- `created_at`

Cada tweet pertenece a un usuario y un usuario puede publicar muchos tweets.

`tweet_id` y `created_at` son generados por el servidor.

### Follow

Contiene:

- `follower_id`: usuario que sigue
- `followed_id`: usuario seguido

Representa una relación dirigida muchos-a-muchos entre usuarios.

## Modelo de persistencia

El servicio guarda solamente `Tweet` y `Follow`.

Los campos:

- `Tweet.author_id`
- `Follow.follower_id`
- `Follow.followed_id`

son referencias lógicas a `User`, pero no son claves foráneas hacia una tabla
local.

La aplicación no verifica la existencia de los usuarios. Esto es aceptable
porque la consigna indica que todos los identificadores recibidos se consideran
válidos.

### Restricciones de Follow

La combinación:

```text
(follower_id, followed_id)
```

debe ser única para evitar relaciones duplicadas y permitir que seguir a un
usuario sea una operación idempotente.

También debe cumplirse:

```text
follower_id != followed_id
```

para impedir que un usuario se siga a sí mismo.

## Patrones de acceso

### Publicar un tweet

Entrada recibida:

- `author_id`
- `content`

Valores generados por el servidor:

- `tweet_id`
- `created_at`

Operación:

- Insertar un nuevo `Tweet` asociado al identificador del autor.

### Seguir a un usuario

Entrada:

- `follower_id`
- `followed_id`

Operación:

- Insertar un nuevo `Follow`.
- Si la relación ya existe, la operación termina correctamente sin crear un
  duplicado.
- Si ambos identificadores son iguales, la operación se rechaza.

### Obtener el timeline

Entrada:

- `user_id`
- Límite de resultados
- Cursor para las páginas posteriores

Recorrido:

1. Buscar los `followed_id` correspondientes al usuario.
2. Buscar los tweets publicados por esos usuarios.
3. Combinar los resultados.
4. Ordenarlos por:

```text
created_at DESC, tweet_id DESC
```

5. Devolver una cantidad limitada.

## Estrategia del timeline

Se utiliza fan-out en lectura.

El timeline se construye cuando el usuario lo consulta y no se guarda como una
entidad independiente.

Esta estrategia evita agregar duplicación de datos y procesamiento adicional al
publicar un tweet.

Como trade-off, el costo de lectura aumenta cuando un usuario sigue a muchas
cuentas, porque es necesario buscar y combinar tweets de varios autores.

Si las métricas mostraran problemas de latencia, se podría evaluar fan-out en
escritura o una estrategia híbrida.

## Orden del timeline

Los tweets se ordenan por:

```text
created_at DESC, tweet_id DESC
```

`created_at` define el orden desde el tweet más reciente hasta el más antiguo.

Si dos tweets tienen el mismo `created_at`, se utiliza `tweet_id` como desempate.
Esto permite mantener un orden determinista y evita que los tweets cambien de
posición entre consultas con los mismos datos.

## Paginación

El timeline utiliza paginación por cursor, también conocida como keyset
pagination.

El tamaño predeterminado de página es 20 y el máximo permitido es 100.

El cursor representa la posición del último tweet devuelto y contiene
lógicamente:

```text
created_at + tweet_id
```

Se utilizan ambos valores porque coinciden con el orden definido para el
timeline. Usar solamente `created_at` podría hacer que se omitan tweets cuando
varios tengan la misma fecha.

### Primera página

Para la primera página no se recibe un cursor. Conceptualmente, la consulta es:

```sql
WHERE author_id IN (:followed_ids)
ORDER BY created_at DESC, tweet_id DESC
LIMIT :page_size_plus_one
```

### Páginas siguientes

Para obtener una página siguiente se buscan los tweets más antiguos que se
encuentran después de la posición indicada por el cursor:

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
LIMIT :page_size_plus_one
```

La consulta representa el comportamiento esperado. La sintaxis concreta puede
cambiar según la base de datos elegida.

### Representación del cursor

El cursor se devuelve al cliente como un valor opaco. Aunque internamente
represente `created_at` y `tweet_id`, el cliente no necesita conocer su
estructura.

Esto permite cambiar su formato interno sin modificar el contrato de la API.

### Detección de la página siguiente

Se consultan `page_size + 1` tweets:

- Se devuelven como máximo `page_size`.
- Si existe un elemento adicional, se genera `next_cursor`.
- Si no existe un elemento adicional, `next_cursor` queda vacío.

De esta forma se puede saber si existe otra página sin realizar una consulta
adicional.

### Nuevos tweets durante la paginación

Si se publica un tweet entre la primera página y la siguiente, normalmente
quedará antes del cursor y no desplazará los resultados de las páginas
posteriores.

Para ver los tweets nuevos, el cliente debe volver a solicitar la primera
página.

Esta estrategia permite una navegación estable, pero no representa una
fotografía inmutable del timeline completo.

## Índices

Los índices se definen a partir de los patrones de acceso acordados.

### Tweet

```text
PRIMARY KEY (tweet_id)
INDEX (author_id, created_at DESC, tweet_id DESC)
```

La clave primaria identifica cada tweet.

El índice compuesto permite:

- Buscar los tweets publicados por cada usuario seguido.
- Recorrer los tweets de cada autor desde el más reciente hasta el más antiguo.
- Desempatar mediante `tweet_id`.
- Aplicar la paginación por cursor.

La base de datos todavía debe combinar los tweets de los distintos autores para
construir el orden global del timeline. Esto es una consecuencia esperada del
fan-out en lectura.

No se agrega un índice independiente por `author_id`, porque ya es la primera
columna del índice compuesto.

### Follow

```text
PRIMARY KEY (follower_id, followed_id)
```

La clave primaria compuesta cumple dos objetivos:

- Evita relaciones duplicadas.
- Permite obtener eficientemente los usuarios seguidos porque `follower_id` es
  la primera columna.

No se agrega un índice por `followed_id`, porque consultar quiénes siguen a un
usuario no es un patrón de acceso requerido por el challenge.

## Persistencia para un entorno productivo

Para simplificar la ejecución del challenge, la implementación actual utiliza
repositorios en memoria.

Esta decisión permite ejecutar la aplicación sin infraestructura adicional, pero
los datos se pierden cuando el proceso se reinicia. Tampoco busca resolver
durabilidad, replicación ni escalado horizontal.

Para una primera versión productiva elegiría PostgreSQL.

### Por qué PostgreSQL

El modelo tiene relaciones y restricciones claras:

- Cada tweet pertenece a un autor.
- Un usuario puede seguir a varios usuarios.
- La relación `(follower_id, followed_id)` debe ser única.
- Un usuario no puede seguirse a sí mismo.
- El timeline consulta tweets de varios autores y necesita un orden global.

PostgreSQL se utilizaría como persistencia principal de toda la aplicación,
tanto para los tweets como para las relaciones de seguimiento.

Permite representar las reglas del modelo mediante claves primarias,
restricciones e índices compuestos. También permite mantener la idempotencia de
los follows y consultar el timeline utilizando joins y paginación por cursor.

Se elige una base relacional porque los datos tienen una estructura estable y
las relaciones forman parte importante de los patrones de acceso.

Una base NoSQL podría ser apropiada si el timeline estuviera previamente
materializado o si existieran requisitos concretos de distribución y volumen.
Sin embargo, para el fan-out en lectura actual probablemente sería necesario
consultar por separado los tweets de cada autor y combinarlos en la aplicación.

Como no se definieron volúmenes, latencia esperada ni un read/write ratio, no
agregaría esa complejidad sin métricas que la justifiquen.

### Esquema conceptual

```sql
CREATE TABLE tweets (
    tweet_id UUID PRIMARY KEY,
    author_id TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    CHECK (char_length(content) BETWEEN 1 AND 280)
);

CREATE INDEX idx_tweets_author_timeline
    ON tweets (
        author_id,
        created_at DESC,
        tweet_id DESC
    );
```

```sql
CREATE TABLE follows (
    follower_id TEXT NOT NULL,
    followed_id TEXT NOT NULL,
    PRIMARY KEY (follower_id, followed_id),
    CHECK (follower_id <> followed_id)
);
```

No se crea una tabla de usuarios porque la consigna considera válidos todos los
identificadores recibidos y deja fuera de alcance su administración.

### Consultas principales

#### Publicar un tweet

```sql
INSERT INTO tweets (
    tweet_id,
    author_id,
    content,
    created_at
)
VALUES ($1, $2, $3, $4);
```

#### Seguir a un usuario

```sql
INSERT INTO follows (
    follower_id,
    followed_id
)
VALUES ($1, $2)
ON CONFLICT (follower_id, followed_id) DO NOTHING
RETURNING follower_id;
```

Si devuelve una fila, la relación fue creada.

Si no devuelve filas, la relación ya existía. De esta manera se mantiene la
idempotencia del endpoint.

#### Obtener los usuarios seguidos

```sql
SELECT followed_id
FROM follows
WHERE follower_id = $1;
```

La clave primaria `(follower_id, followed_id)` permite buscar por
`follower_id` sin agregar otro índice.

#### Primera página del timeline

```sql
SELECT
    t.tweet_id,
    t.author_id,
    t.content,
    t.created_at
FROM follows AS f
JOIN tweets AS t
    ON t.author_id = f.followed_id
WHERE f.follower_id = $1
ORDER BY
    t.created_at DESC,
    t.tweet_id DESC
LIMIT $2;
```

El límite utilizado es `page_size + 1` para determinar si existe una página
siguiente.

#### Páginas siguientes

```sql
SELECT
    t.tweet_id,
    t.author_id,
    t.content,
    t.created_at
FROM follows AS f
JOIN tweets AS t
    ON t.author_id = f.followed_id
WHERE f.follower_id = $1
  AND (
      t.created_at < $2
      OR (
          t.created_at = $2
          AND t.tweet_id < $3
      )
  )
ORDER BY
    t.created_at DESC,
    t.tweet_id DESC
LIMIT $4;
```

Los parámetros `$2` y `$3` representan la posición incluida en el cursor. La
comparación es exclusiva para evitar repetir el último tweet de la página
anterior.

### Escalabilidad y evolución

El costo del fan-out en lectura aumenta cuando un usuario sigue a muchas
cuentas, porque PostgreSQL debe combinar tweets de varios autores para construir
el orden global.

Antes de cambiar la estrategia mediría:

- Cantidad promedio de usuarios seguidos.
- Frecuencia de lectura y escritura.
- Latencia del timeline.
- Cantidad de tweets recorridos.
- Planes de ejecución de las consultas.

Si las métricas mostraran problemas, se podría evaluar una estrategia de
fan-out en escritura o una solución híbrida. Esas alternativas agregarían
duplicación de datos y procesamiento adicional, por lo que quedan fuera del
alcance actual.

## Limitaciones conocidas

- El costo de obtener el timeline crece con la cantidad de usuarios seguidos.
- La base de datos debe combinar tweets de varios autores para construir el
  orden global.
- La existencia de los usuarios no se puede verificar localmente.
- No se definieron volúmenes de tráfico ni objetivos concretos de latencia.
- No se definió una cantidad máxima de cuentas seguidas por usuario.
- La paginación no representa una fotografía inmutable del timeline.
- Los índices deberían validarse con métricas y planes de ejecución en un
  entorno real.