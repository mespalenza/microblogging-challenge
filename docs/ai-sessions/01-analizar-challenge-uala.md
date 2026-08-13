# Analiza challenge de Ualá

> Analisis  del challenge de Ualá
> Recibí un enunciado para desarrollar una plataforma similar a Twitter. Te lo voy a adjuntar en un pdf en conjunto al analisis que hice previo, comentame si ves alguna mejora o algo que haya pasado por alto. No quiero ni diseño, ni implementación, solo análisis:
>
> Identifico 3 casos de uso principales:
> 1. Publicar un tweet => tiene una longitud máxima
> 2. Seguir a una persona => todos los usuarios son validos
> 3. Ver el timeline de tweets => solo el de los usuarios seguidos
>
> Los requisitos no funcionales
> 1. Que va a escalar para millones de user
> 2. Que tiene que estar optimizado para lecturas
>
> En  cuanto a assumptions veo algunas importantes que deberia definir para que limitar el alcance del challenge
> 1. Pueden haber tweets vacíos ?
> 2. El user se puede seguir a el mismo?
> 3. Orden cronologico para el timeline ?
>
> Este es el mail, con algunas aclaraciones extra. Como el lenguaje a utilizar que lo establece en go, no es libre como dice el PDF.
>
> Hola Maria Emilia! ¿Cómo estás?
>
> Te escribo porque nos gustaría avanzar con el proceso de selección para nuestra búsqueda de Software Engineer. El próximo paso es un ejercicio técnico el cual encontrarás adjunto en este mail con las consignas.
>
> Vas a tener un plazo de 7 días para poder realizarlo por lo que tendrá validez hasta el próximo **13/08.** Algunas aclaraciones:
> La solución debe ser implementada en **Go**.
> La entrega de la resolución debe hacerse a través de un repositorio personal sobre este mismo mail (importante dejar el acceso publico para su posterior revisión).
> Se puede usar la wiki del repositorio para subir la documentación.
> Fomentamos el uso de IA como multiplicador de productividad. Si utilizas alguna para este ejercicio, te pedimos que **compartas la sesión de la interacción junto con tu entrega**. Evaluaremos no solo el código final, sino tu capacidad para guiar a la IA y tomar decisiones arquitectónicas sobre el código generado.
>
> User context
>
> Attachments:
> - Challenge Backend..pdf: `~/Downloads/Challenge Backend..pdf`

<details><summary>2 previous messages</summary>

> Voy a usar la habilidad de PDF para leer el enunciado completo y contrastarlo con tu análisis, manteniéndome estrictamente en alcance, requisitos, supuestos y riesgos; sin proponer diseño ni implementación.
>
> El PDF tiene dos páginas, pero las herramientas PDF del sistema no están completas. Voy a usar el runtime documental incluido para extraer el texto y renderizar ambas páginas, así también verifico si hay información visual que la extracción pueda omitir.
</details>

Tu análisis identifica correctamente el núcleo funcional, pero todavía faltan varias ambigüedades de negocio y algunos requisitos explícitos de evaluación y entrega. El challenge parece deliberadamente abierto: probablemente valoren que conviertas esas ambigüedades en decisiones claras y justificadas.

Fuente revisada: :codex-file-citation{path="/Users/emispalenza/Downloads/Challenge Backend..pdf" purpose="source"}

## 1. Casos de uso

Los tres que identificaste son correctos:

1. Publicar un tweet.
2. Seguir a otro usuario.
3. Consultar el timeline.

Haría dos ajustes en su formulación:

- “Todos los usuarios son válidos” no pertenece al caso de uso de seguir; es una premisa transversal: no hay registro, autenticación, sesiones ni validación de existencia.
- “El tweet tiene una longitud máxima” es una regla del primer caso de uso, pero el máximo todavía no está cerrado. El PDF dice “por ejemplo, 280”, no necesariamente que deba ser exactamente 280.

No agregaría como casos de uso obligatorios acciones que no aparecen en el enunciado, como dejar de seguir, borrar o editar tweets, likes, retweets, comentarios o perfiles. Conviene declararlas explícitamente fuera de alcance.

## 2. Requisitos no funcionales

Los dos que marcaste son correctos:

- Escalar conceptualmente a millones de usuarios.
- Priorizar las lecturas.

Pero el PDF también contiene requisitos o atributos de calidad adicionales:

- Separación clara de responsabilidades o capas.
- Facilidad para ejecutar y evaluar el proyecto.
- Testing de los casos de uso principales.
- Documentación de alto nivel.
- Justificación de las decisiones tomadas.
- Ausencia de frontend.

No todos son requisitos del producto; algunos son criterios de evaluación. Conviene mantenerlos en secciones separadas para no mezclar comportamiento, calidad y entregables.

“Escalar a millones de usuarios” también necesita una aclaración: no significa necesariamente implementar infraestructura distribuida durante el challenge. Significa que las decisiones adoptadas no deberían impedir explicar cómo evolucionaría la solución. El propio PDF permite una persistencia en memoria, siempre que se documente qué se usaría en un escenario real y por qué.

## 3. Supuestos que ya identificaste

Los tres son buenos y deberían resolverse, no solamente enumerarse:

- ¿Se permiten tweets vacíos o compuestos solo por espacios?
- ¿Puede un usuario seguirse a sí mismo?
- ¿Cuál es el orden del timeline?

Mi opción razonable sería:

- Rechazar tweets vacíos o con solamente espacios.
- No permitir seguirse a uno mismo.
- Mostrar el timeline en orden cronológico inverso, del tweet más reciente al más antiguo.

Pero hay una consecuencia importante: si no permitís el auto-follow y el timeline contiene exclusivamente tweets de usuarios seguidos, entonces los tweets propios no aparecen. Eso debería quedar explícito. No asumiría que se incluyen automáticamente porque el enunciado dice “de los usuarios a los que siguen”.

## 4. Ambigüedades que faltaría resolver

### Tweets

- ¿El límite es realmente 280 o debe ser configurable?
- ¿Cómo se cuentan los caracteres? En Go, bytes, runas Unicode y caracteres visibles no son equivalentes.
- ¿Se conservan los espacios iniciales y finales o se normalizan?
- ¿Se aceptan saltos de línea?
- ¿El tweet debe registrar una fecha de publicación?
- ¿Qué sucede cuando dos tweets tienen el mismo timestamp?
- ¿Una publicación repetida es válida?
- ¿Editar y eliminar quedan expresamente fuera de alcance?

La definición de “carácter” es especialmente importante: un límite aplicado por bytes puede rechazar antes un texto con emojis o caracteres acentuados.

### Follow

- ¿Seguir dos veces a la misma persona produce un error o se considera una operación idempotente?
- ¿Se permite dejar de seguir? No está solicitado; yo lo declararía fuera de alcance.
- Aunque “todos los usuarios son válidos”, ¿un identificador vacío o mal formado también se considera válido? Probablemente no: una cosa es no comprobar existencia y otra aceptar un ID vacío.
- ¿Existe alguna diferencia entre usuario inexistente y usuario que todavía no publicó?
- ¿Las relaciones de seguimiento creadas después de una publicación incluyen tweets históricos o solamente publicaciones futuras?

El último punto afecta directamente el significado del timeline y conviene definirlo como regla de negocio.

### Timeline

- ¿Incluye todo el historial de los usuarios seguidos?
- ¿Cuál es el tamaño máximo de una respuesta?
- ¿Cómo se recorren resultados adicionales?
- ¿Qué devuelve cuando el usuario no sigue a nadie?
- ¿Qué sucede si dos publicaciones tienen la misma fecha?
- ¿Los cambios de follow deben reflejarse inmediatamente?
- ¿El timeline debe ser fuertemente consistente o se acepta una demora breve?

La paginación es probablemente lo más importante que falta en tu análisis. Un timeline sin límite contradice la expectativa de millones de usuarios y de optimización de lecturas. No hace falta diseñar cómo implementarla todavía, pero sí reconocerla como comportamiento necesario.

## 5. Identidad y seguridad

El PDF elimina autenticación y sesiones, pero eso no significa que el sistema tenga usuarios reales previamente registrados. La premisa más precisa sería:

> El identificador recibido se considera confiable y no se comprueba contra un catálogo de usuarios. No se implementan registro, login, autorización ni administración de usuarios.

También conviene decidir si el identificador del usuario debe venir siempre por el mismo lugar. El PDF permite header, parámetro o body, pero la solución debería elegir una convención consistente.

## 6. Conflicto entre el PDF y el correo

El correo debe considerarse una aclaración posterior y, por tanto, prevalece sobre el PDF:

- El lenguaje no es libre: debe utilizarse Go.
- La fecha límite es el 13/08.
- El repositorio debe ser personal y público.
- Si se utiliza IA, debe compartirse la sesión.
- Evaluarán no solo el resultado, sino también cómo guiaste la IA y qué decisiones tomaste sobre sus propuestas.

Este último requisito merece aparecer en tu checklist de entrega. Además, dado que compartís la sesión, es valioso que quede visible que cuestionaste supuestos, rechazaste complejidad innecesaria y tomaste personalmente las decisiones.

## 7. Checklist de alcance recomendado

Antes de comenzar, dejaría definido en `business.txt`:

- Máximo exacto y criterio para contar caracteres.
- Tratamiento de tweets vacíos y espacios.
- Auto-follow.
- Follow duplicado.
- Inclusión o exclusión de tweets propios.
- Inclusión de tweets históricos al comenzar a seguir.
- Orden del timeline y desempate.
- Tamaño máximo y forma de paginación.
- Comportamiento de un timeline vacío.
- Consistencia esperada.
- Formato mínimo válido de los identificadores.
- Funcionalidades fuera de alcance.

Y separaría claramente:

- Requisitos funcionales.
- Reglas de negocio y supuestos.
- Requisitos no funcionales.
- Funcionalidades fuera de alcance.
- Criterios de evaluación.
- Entregables.

En síntesis: tu análisis inicial está bien orientado, pero la mayor omisión es la definición precisa del timeline —paginación, tweets propios, historial, orden y consistencia—. Es el caso de uso central del challenge y el que más se relaciona con los requisitos de escalabilidad y optimización para lectura.

> Claro chat pero mi idea era ahora solo analizar el negocio, no criterios de evaluación, entregables
> Poder refinar mejor el análisis que hice

Tenés razón. Si nos limitamos al negocio, tu análisis está bien encaminado. Lo refinaría así:

## Casos de uso

### 1. Publicar un tweet

Un usuario puede publicar un mensaje corto asociado a su identificador.

Reglas por definir:

- Longitud máxima exacta: el PDF menciona 280 solo como ejemplo.
- No permitir contenido vacío ni compuesto únicamente por espacios.
- Definir qué significa “carácter”, especialmente para emojis y Unicode.
- Determinar si se preservan o eliminan espacios iniciales y finales.
- Cada tweet necesita una fecha de publicación para poder ordenar el timeline.
- Edición y eliminación quedan fuera de alcance.

### 2. Seguir a otro usuario

Un usuario puede comenzar a seguir a otro para ver sus tweets en el timeline.

Reglas por definir:

- No permitir que un usuario se siga a sí mismo.
- Seguir nuevamente a alguien ya seguido debería ser una operación idempotente: el estado no cambia.
- “Todos los usuarios son válidos” significa que no se valida su existencia, no que se acepten identificadores vacíos.
- Dejar de seguir queda fuera de alcance porque no aparece en el enunciado.
- Debe definirse si al seguir a alguien aparecen también sus tweets anteriores. Lo más natural es que sí, porque el timeline consulta los tweets de los usuarios actualmente seguidos.

### 3. Consultar el timeline

Un usuario puede ver los tweets publicados por los usuarios que sigue.

Reglas por definir:

- Orden cronológico inverso: primero los tweets más recientes.
- En caso de empate temporal, debe existir algún criterio estable de desempate.
- El timeline no incluye automáticamente los tweets propios.
- Si se quiere incluirlos, tendría que definirse expresamente, ya que el enunciado solo menciona usuarios seguidos.
- Si no sigue a nadie o los seguidos no publicaron, devuelve un timeline vacío.
- Debe tener una cantidad limitada de resultados y alguna forma de consultar los siguientes. Esto es parte del comportamiento del negocio, aunque después tenga implicancias técnicas.
- Si empieza a seguir a alguien, aparecen sus publicaciones históricas.
- Si dejar de seguir está fuera de alcance, no hace falta definir qué ocurre en ese caso.

## Premisas generales

Reformularía “todos los usuarios son válidos” de esta manera:

> No existe gestión de usuarios. Los identificadores recibidos se consideran pertenecientes a usuarios válidos y no se comprueba su existencia.

Esto excluye del negocio:

- Registro de usuarios.
- Inicio de sesión.
- Sesiones.
- Perfiles.
- Autorización.
- Bloqueos o cuentas privadas.

Aun así, los identificadores deberían cumplir una condición mínima: no ser vacíos.

## Requisitos no funcionales

Tus dos puntos están bien identificados:

- El modelo debe contemplar millones de usuarios.
- El caso de uso prioritario es la lectura del timeline.

No profundizaría todavía en cómo se logra. En esta etapa basta con reconocer una consecuencia funcional: el timeline no puede devolver una cantidad ilimitada de tweets.

## Versión refinada de tus supuestos

Podrías dejar una primera propuesta como esta:

1. Un tweet no puede estar vacío ni contener solamente espacios.
2. El límite será de 280 caracteres Unicode.
3. Un usuario no puede seguirse a sí mismo.
4. Seguir a un usuario ya seguido no modifica el estado.
5. No se valida la existencia de los usuarios, pero sus identificadores no pueden ser vacíos.
6. El timeline contiene exclusivamente tweets de los usuarios seguidos.
7. Los tweets propios no aparecen, salvo que se cambie expresamente esa regla.
8. El timeline se ordena del tweet más reciente al más antiguo.
9. Al comenzar a seguir a un usuario se incluyen también sus tweets anteriores.
10. El timeline entrega resultados de forma limitada y paginada.
11. Crear, editar o eliminar usuarios y tweets, y dejar de seguir, quedan fuera del alcance.

La pregunta de negocio más importante que todavía requiere una decisión consciente es si el timeline incluye los tweets propios. Incluirlos suele coincidir con la experiencia de Twitter, pero la lectura literal del challenge los excluye. Para este ejercicio, seguiría la lectura literal y documentaría la decisión.

> Separá lo explícitamente solicitado de tus inferencias, porque la parte de definir las reglas las quiero en otro item no mezcladas con lo que sí esta en el enunciado. Por eso había ordenado la info con lo que si salia en el pdf

Entiendo. Tu estructura original era correcta: primero registrar fielmente el enunciado y después, en una sección separada, identificar decisiones pendientes. No conviene mezclar ambas cosas porque parecería que el PDF exige reglas que en realidad estamos proponiendo.

Lo organizaría así:

## 1. Requisitos funcionales explícitos

### Publicar un tweet

- Los usuarios pueden publicar mensajes cortos.
- Los tweets tienen un límite máximo de caracteres.
- El enunciado menciona 280 caracteres como ejemplo, pero no establece explícitamente que ese sea el límite definitivo.

### Seguir usuarios

- Un usuario puede seguir a otros usuarios.

### Consultar el timeline

- Un usuario puede consultar un timeline.
- El timeline muestra los tweets de los usuarios a los que sigue.

## 2. Requisitos no funcionales explícitos

- La solución debe pensarse para escalar a millones de usuarios.
- La aplicación debe estar optimizada para lecturas.

## 3. Supuestos explícitos del enunciado

- Todos los usuarios se consideran válidos.
- No es necesario implementar registro o inicio de sesión.
- No es necesario administrar sesiones.
- El identificador del usuario puede recibirse mediante header, parámetro, body u otro mecanismo.

El último punto define una libertad sobre cómo identificar al usuario, pero no establece todavía una regla concreta.

## 4. Preguntas de negocio no resueltas por el enunciado

Estas no son requisitos ni decisiones tomadas: son ambigüedades que habría que resolver y posteriormente documentar como assumptions.

### Sobre los tweets

- ¿Cuál es el límite exacto de caracteres?
- ¿Se permiten tweets vacíos?
- ¿Se permiten tweets compuestos únicamente por espacios?
- ¿Cómo se cuentan los caracteres cuando existen emojis o caracteres Unicode?
- ¿Se permiten saltos de línea?
- ¿Se pueden editar o eliminar tweets?

### Sobre el seguimiento

- ¿Un usuario puede seguirse a sí mismo?
- ¿Qué ocurre si intenta seguir dos veces al mismo usuario?
- ¿Se puede dejar de seguir a un usuario?
- ¿Es necesario validar que los identificadores no estén vacíos, aunque todos los usuarios se consideren válidos?

### Sobre el timeline

- ¿Los tweets se muestran en algún orden determinado?
- Si el orden es cronológico, ¿es ascendente o descendente?
- ¿Cómo se resuelven dos tweets con la misma fecha?
- ¿El timeline incluye los tweets del propio usuario?
- ¿Al seguir a un usuario aparecen sus tweets históricos o solamente los publicados desde ese momento?
- ¿Existe un límite de tweets por consulta?
- ¿Debe existir paginación?
- ¿Qué se muestra si el usuario no sigue a nadie?

## 5. Funcionalidades no mencionadas

Esto tampoco significa que estén excluidas todavía, sino que el enunciado no las solicita:

- Dejar de seguir.
- Editar tweets.
- Eliminar tweets.
- Likes.
- Retweets.
- Comentarios.
- Perfiles.
- Búsqueda.
- Usuarios privados o bloqueados.
- Notificaciones.

Para limitar el alcance, después podrías declarar explícitamente cuáles de estas funcionalidades quedan fuera.

Así queda clara la diferencia:

- **Requisito:** está escrito en el enunciado.
- **Supuesto explícito:** el propio enunciado pide tomarlo como verdadero.
- **Pregunta abierta:** el enunciado no da suficiente información.
- **Assumption propia:** respuesta que vos elegís para cerrar una pregunta abierta.
- **Fuera de alcance:** funcionalidad que decidís no cubrir porque no fue solicitada.

Tu análisis original podría mejorarse principalmente agregando los supuestos que el PDF ya declara —ausencia de login y sesiones— y ampliando las preguntas abiertas del timeline. Pero todavía sin responderlas ni convertirlas en reglas.

> Cuales de estas ambigüedades requieren una aclaración para poder avanzar y cuales pueden resolverse mediante supuestos razonables documentados? Por ej editar o eliminar tweets no está solicitado explícitamente, así que lo pondría fuera del alcance y no como una pregunta pendiente. Priorizaría implementar los casos de uso solicitados y solo agregaría funcionalidades extra si el tiempo lo permite

Coincido con tu criterio. En este challenge, prácticamente ninguna ambigüedad bloquea el avance: el PDF permite agregar assumptions, por lo que conviene tomar decisiones razonables, documentarlas y concentrarse en los tres casos de uso solicitados.

Solo pediría una aclaración si una interpretación pudiera cambiar sustancialmente el comportamiento central y no hubiera una opción claramente razonable.

## Aclaraciones potencialmente valiosas, pero no bloqueantes

### Límite exacto del tweet

El enunciado dice “por ejemplo, 280 caracteres”, por lo que técnicamente no fija el valor.

Podrías preguntar cuál debe ser el límite exacto, pero no es necesario para avanzar. La suposición más natural es:

> El límite máximo es de 280 caracteres.

Además, conviene documentar cómo se cuentan los caracteres. Una interpretación razonable es contar caracteres Unicode, no bytes.

### Inclusión de tweets propios en el timeline

El texto dice que el timeline muestra tweets de los usuarios seguidos. No aclara si incluye también publicaciones propias.

Esta duda afecta directamente al caso de uso principal, pero puede resolverse con una lectura literal:

> El timeline contiene exclusivamente tweets de usuarios seguidos y no incluye automáticamente los tweets propios.

Si quisieras consultar algo a Ualá, esta sería una de las preguntas con mayor valor porque la experiencia habitual de Twitter puede sugerir lo contrario. Aun así, no bloquea el desarrollo si la decisión queda documentada.

## Ambigüedades resolubles mediante assumptions

Estas pueden resolverse sin consultar porque existe una respuesta convencional, simple y coherente con el alcance.

### Publicación de tweets

- No se aceptan tweets vacíos.
- No se aceptan tweets compuestos únicamente por espacios.
- El límite se aplica sobre caracteres Unicode.
- El contenido puede normalizarse eliminando espacios al principio y al final.
- Cada tweet registra su momento de publicación.

### Seguimiento

- Un usuario no puede seguirse a sí mismo.
- Seguir nuevamente al mismo usuario no crea una relación duplicada.
- Los identificadores no pueden estar vacíos.
- No se valida la existencia del usuario porque el enunciado indica que todos son válidos.

Sobre el follow duplicado, puede elegirse entre devolver éxito o informar que la relación ya existe. Lo importante para el negocio es que no se duplique.

### Timeline

- Se ordena cronológicamente desde el tweet más reciente al más antiguo.
- Si no existen tweets de usuarios seguidos, se devuelve un timeline vacío.
- Al comenzar a seguir a un usuario, se incluyen sus tweets históricos.
- La cantidad de resultados por consulta es limitada.
- Se permite recorrer el timeline mediante paginación.
- Ante timestamps iguales, se utiliza un criterio estable de desempate.

La paginación no es una funcionalidad adicional comparable con likes o retweets: es una delimitación necesaria del caso de uso “consultar timeline”, especialmente porque el enunciado exige contemplar millones de usuarios.

## Funcionalidades que pondría directamente fuera de alcance

No las presentaría como preguntas pendientes porque no son necesarias para implementar los casos de uso solicitados:

- Editar tweets.
- Eliminar tweets.
- Dejar de seguir.
- Eliminar relaciones de seguimiento.
- Likes.
- Retweets.
- Respuestas o comentarios.
- Menciones y hashtags.
- Adjuntar imágenes o videos.
- Perfiles de usuario.
- Registro, login y sesiones.
- Cuentas privadas.
- Bloqueos.
- Notificaciones.
- Búsqueda.
- Listados de seguidores y seguidos.

La única salvedad es “dejar de seguir”: está relacionado con follow, pero el enunciado dice únicamente que los usuarios deben poder seguir a otros. Por lo tanto, implementarlo sería una extensión, no una obligación.

## Clasificación recomendada

| Tema | Tratamiento |
|---|---|
| Límite exacto del tweet | Asumir 280; aclaración opcional |
| Tweet vacío | Resolver con assumption |
| Conteo Unicode | Resolver con assumption |
| Auto-follow | Resolver con assumption |
| Follow duplicado | Resolver con assumption |
| Orden del timeline | Resolver con assumption |
| Tweets propios en timeline | Assumption importante; aclaración opcional |
| Tweets históricos al seguir | Resolver con assumption |
| Timeline vacío | Resolver con assumption |
| Límite y paginación | Definir como parte del caso de uso |
| Editar o borrar tweets | Fuera de alcance |
| Unfollow | Fuera de alcance |
| Funcionalidades sociales adicionales | Fuera de alcance |

En resumen: no veo una pregunta que necesariamente debas enviar a Ualá antes de avanzar. El challenge parece evaluar justamente la capacidad de detectar ambigüedades, elegir una interpretación razonable y documentarla. Consultaría solamente si querés reducir el riesgo sobre dos puntos centrales: el límite exacto de caracteres y la inclusión de tweets propios en el timeline.

> No considero necesario consultar por estas ambigüedades. Ya checkeamos que no hay ninguna que bloquee el desarrollo y el enunciado permite deducir supuestos y documentarlos
> Para el límite del tweet voy a tomar 280 caracteres max, tomando como referencia el enunciado. Voy a definir comportamientos razonables para los demás casos ambiguos (por ej rechazar tweets vacíos) y los voy a dejar explicitos en la docu
> Las funcionalidades no solicitadas explicitamente, como editar o eliminar tweets, no son ambigüedades y van a quedar fuera del alcance. Voy priorizar completar bien los tres casos de uso requeridos antes de considerar una funcionalidad nueva
> Con esta conclusión cerramos el analisis inicial del negocio

Queda cerrado el análisis inicial del negocio con estas conclusiones:

- No existen ambigüedades que bloqueen el desarrollo.
- Las decisiones no especificadas se resolverán mediante supuestos razonables y documentados.
- Los tweets tendrán un máximo de 280 caracteres, tomando como referencia el enunciado.
- Se definirán reglas explícitas para casos ambiguos, como rechazar tweets vacíos.
- Las funcionalidades no solicitadas —editar o eliminar tweets, unfollow, likes, retweets, etc.— quedan fuera del alcance; no se consideran ambigüedades.
- Se priorizará completar correctamente los tres casos de uso requeridos:
    1. Publicar un tweet.
    2. Seguir a otro usuario.
    3. Consultar el timeline de los usuarios seguidos.
- Solo se considerarán funcionalidades adicionales si sobra tiempo, sin comprometer los casos principales.

Es una delimitación de alcance sólida y coherente con el enunciado.
