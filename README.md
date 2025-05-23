# Calc service

## Описание проекта
<b>Calc service</b> — это проект, реализован на языке программирования Golang, предназначен для вычисления арифметических выражений, таких как "(6+(2+2)*2)/10". Главная цель этого сервиса заключается в том, чтобы предоставить возможность быстро и точно обрабатывать математические выражения, используя параллельые вычисления. Калькулятор автоматически разбивает выражение на задачи и параллельно производит вычисления. Это позволяет значительно ускорить процесс вычислений и сделать их более эффективными.

> [!WARNING] 
<b> P.S. Если запускаете проект не через docker меняйте параметр host в конфигурациях сервисов обязательно</b>

## Схема работы проекта

Прикладываю для вас схему, для наглядного описания, как происходит взаимодействие.

![](orchestrator/docs/diagramm/diagramm.svg)

## Настройка
1. Клонируйте репозиторий:

```sh
git clone https://github.com/DobryySoul/Calc-service.git
```
2. Перейдите в корневую папку проекта, если это не было еще сделано:
```sh
cd Сalc-service
```

3. Установите зависимости:
```sh
cd agent
go mod tidy

cd orchestrator
go mod tidy
```

Отлично, это уже успех! Если вы хотите изменить параметры конфигурации проекта, то продолжите чтение, если же нет, то переходите на следующий [этап](#Запуск).

## Конфигурация 
Сервис можно настроить с помощью переменных среды. Со списком и описанием доступных параметров к настройке можно ознакомиться [здесь](#Список-параметров-конфигурации). 

Чтобы указать параметры переменной среды, который вы хотите использовать, необходимо изменить файл [.env.example](https://github.com/DobryySoul/Calc-service/blob/main/orchestrator/.env.example) на .env и изменить переменные, иначе сервер будет запущен на дефолтных значениях, которые указаны в этом файле.

Также есть возможность указать значения переменных среды через команды в терминале, пример:
### Windows

```sh
$env:PORT=порт
```

### Linux и macOS

```sh
export PORT=порт
```

C GRPC-контрактами можно ознакомиться [контракт](https://github.com/DobryySoul/Calc-service/blob/main/api/v1/proto/calculator/service.proto).

## Endpoints

Сервис имеет следующий эндпоинты:
- `/api/v1/register` - получить страницу регистрации.
- `/api/v1/login` - получить страницу логина.
- `/api/v1/register` - отправить запрос на регистрацию.
- `/api/v1/login` - отправить запрос на авторизацию.
- `/api/v1/calculate` - отправить новое выражение для вычисления.
- `/api/v1/expressions` - получить список всех выражений.
- `/api/v1/expression/:id` - получить выражение по идентификатору id.
- `/internal/task` - получить задачу для обработки/отправить результат.

    - GET: отдает задачу на выполнение.

    - POST: отправляет результат выполнения. 
    ```json
    {
        "id": уникальный идентификатор,
        "result": результат вычислений
    }
    ```

## Запуск

Проект готов к запуску. P.S. не забудте поменять пароль от базы данных в makefile, docker-compose.yml и в файле .env.

Бэкенд делиться на 2 части:
1. Оркестратор

2. Агент

Команды, чтобы запустить их:
1. ```sh
    cd orchestrator
    go run cmd/main.go
    ```

2. ```sh
    cd agent
    go run cmd/main.go
    ```
3. При условии что у вас запущен postgresql, то можно запустить миграции из корня проекта:

```sh
make migrations-up
```

Для более изянтного запуска, можно использовать makefile и docker-compose:
Самый простой способ запустить всё вместе использовать команду makefile из корня проекта:

```sh
make all
```

Команда прогонит все тесты, запустит все сервисы(если у вас установлен docker) и запустит миграции.
Или, если вы сторонник docker-compose, то можно запустить сервисы с помощью команды:

```sh
docker-compose up -d --build
```
А затем запустить миграции командой:

```sh
make migrations-docker-up
```

Если что-то пошло не так, то можно откатить миграции:

```sh
make migrations-docker-down
```

![](orchestrator/docs/starts/start-orchestrator.png)

![](orchestrator/docs/starts/start-agent.png)

Мои поздравления! Сервис успешно запущен и готов к функционированию.


## Список параметров конфигурации

#### `host`

*(адрес)* адрес для запуска приложения

- Эквивалент env: `HOST`.

#### `port`

*(номер)* порт для запуска приложения

- Эквивалент env: `PORT`.

#### `grpc_port`

*(номер)* порт для запуска gRPC-сервера

- Эквивалент env: `GRPCPort`.

#### `time_addition_ms`
*(продолжительность)* время выполнения операции сложения в миллисекундах

- Эквивалент env: `TIME_ADDITION_MS`.

#### `time_subtraction_ms`
*(продолжительность)* время выполнения операции вычитания в миллисекундах

- Эквивалент env: `TIME_SUBTRACTION_MS`.

#### `time_multiplications_ms`
*(продолжительность)* время выполнения операции умножения в миллисекундах

- Эквивалент env: `TIME_MULTIPLICATIONS_MS`.

#### `time_divisions_ms`
*(продолжительность)* время выполнения операции деления в миллисекундах

- Эквивалент env: `TIME_DIVISIONS_MS`.

#### `postgres_username`
*(имя)* имя пользователя базы данных

- Эквивалент env: `POSTGRES_USERNAME`.

#### `postgres_password`
*(пароль)* пароль от сервера постгрес

- Эквивалент env: `POSTGRES_PASSWORD`.

#### `postgres_host`
*(адрес)* адрес сервера базы данных

- Эквивалент env: `POSTGRES_HOST`.

#### `postgres_port`
*(номер)* порт сервера базы данных

- Эквивалент env: `POSTGRES_HOST`.

#### `postgres_database`
*(название)* имя базы данных

- Эквивалент env: `POSTGRES_DATABASE`.

#### `postgres_max_conn`
*(количество)* максимальное количество соединений с сервером базы данных

- Эквивалент env: `POSTGRES_MAX_CONN`.

#### `postgres_min_conn`
*(количество)* минимальное количество соединений с сервером базы данных

- Эквивалент env: `POSTGRES_MIN_CONN`.

#### `jwt_secret`
*(название)* секретный ключ для генерации токена

- Эквивалент env: `JWT_SECRET`.

#### `postgres_min_conn`
*(продолжительность)* время жизни токена

- Эквивалент env: `JWT_TTL`.

Также вы можете настроить параметры конфигурации агента, переименовав файл [config.example.yaml](https://github.com/DobryySoul/Calc-service/blob/main/agent/config/config.example.yaml) на config.yaml и изменить параметры, порт должен совпадать с grpc-портом, который указан в файле [.env.example](https://github.com/DobryySoul/Calc-service/blob/main/orchestrator/.env.example) или в вашем аналоге .env

##


## Обработка запросов и ошибок

### Статус коды
В зависимости от типа запроса, а также корректности выражения, сервер дает различные ответы, с соответствующими статус кодами:

> [!IMPORTANT]
> #### `/api/v1/register`

- `200`: Ответ на успешную регистрацию:

```json
{
    "email": "ktototakoi@proton.me",
    "password": "123qweQWE!@#"
}
```

> [!WARNING] 
Предупреждаю, что на сервере стоит строгая валидация, поэтому у вас не получится зарегистрироваться с невалидным email или паролем.
Email должен включать символ `@` и точку, а пароль должен содержать минимум 1 цифру, 1 букву, 1 спецсимвол, и быть не короче 8 символов.

> [!IMPORTANT]
> #### `/api/v1/login`

- `200`: Ответ на успешную авторизацию, в ответ вы также получите токен, который вам нужно будет использовать для авторизации в других запросах, но если же вы используете веб-версию, то об этом вам не стоит беспокоиться:

```json
{
    "email": "ktototakoi@proton.me",
    "password": "123qweQWE!@#"
}
```


> [!IMPORTANT]
> #### `/api/v1/calculate`

- `201`: Ответ на добавление выражения для вычисления в верном формате:

```json
{
    "expression": "2 + 2 * 15" // аримфметическое выражение верного формата -> string
}
```
  
![](orchestrator/docs/POST/api/v1/calculate/status201.png)

- `422`: Ошибка в арифметическом выражении, невалидный формат, пример: 

```json
{
    "expression": "2 + 2 * 15
}
```

![](orchestrator/docs/POST/api/v1/calculate/status422.png)

> [!IMPORTANT]
> #### `/api/v1/expressions`

- `200`: Успешно получен список выражений:

![](orchestrator/docs/GET/api/v1/expressions/status200.png)


> [!IMPORTANT]
> #### `/api/v1/expressions/:id`


- `200`: Успешно полученное выражение по идентификатору id:

![](orchestrator/docs/GET/api/v1/expression/id/status200.png)


- `404`: Выражение не было найдено по id:

![](orchestrator/docs/GET/api/v1/expression/id/status404.png)


> [!IMPORTANT]
> #### `/internal/task`

- `200`: Задача успешно получена:
  
![](orchestrator/docs/GET/internal/task/status200.png)


- `404`: Нет задач для выполнения:
  
![](orchestrator/docs/GET/internal/task/status404.png)


- `200`: Успешно записан результат задачи в формате 

```json
{
    "id": 0, // идентификатор задачи -> int
    "result": 30 // валидный формат ответа-> int
}
```
  
![](orchestrator/docs/POST/internal/task/status200.png)
 

- `404`: По данному id не было найдено задачи.

![](orchestrator/docs/POST/internal/task/status404.png)


- `422`: Невалидный формат введенных данных, пример:

```json
{
    "id": 0, // идентификатор задачи -> int
    "result": "30" // невалидный формат ответа -> string
}
```

![](orchestrator/docs/POST/internal/task/status422.png)


- `500`: Случай внутренней ошибки сервера. Данная ошибка не возникает, так как сервер работает полностью исправно, но все же данная ошибка должна обрабатываться, на случай, когда сервер не сможет обработать запрос к сайту или дать ответ.

## Frontend

Перейдя в браузере по адресу http://localhost:9090/api/v1/register (если запускали на дефолтном значении переменных окружения), вы попадете на страницу регистрации, далее пройдя и авторизацию вы попадете на внешний интерфейс сервиса.

![](orchestrator/docs/frontend/pre.png)


1. Форма для отправки нового выражения на сервер.
2. Кнопка `Вычислить` отправляет выражение на бэкенд, для дальнейшего взаимодействия с ним.

![](orchestrator/docs/frontend/success_shipped.png)

3. Форма для ввода id выражения, информацию о котором вы хотите получить.
4. Кнопка `Получить` непостредственно запрашивает информацию о выражении с введенным id, если такое выражение существует, то оно будет выведено.

![](orchestrator/docs/frontend/success_get.png)

5. Кнопка для обновления списка всех выражений. Несмотря на то, что сайт сам обновляется при получении нового выражения на вычисление, если выражения будут посчитаны кем-то, например, агентами, то необходимо будет нажать `Обновить список`, для вывода актульной информации.

![](orchestrator/docs/frontend/list_of_all.png)

6. Извлекает задачу из выражения и отправляет ее вам для подсчета.

![](orchestrator/docs/frontend/expression_task.png)

7. Сюда необходимо вставить id актуальной задачи.
8. Результат ваших вычислений.
9. Кнопка взаимодействия(отправки результата на сервер).

![](orchestrator/docs/frontend/result_sent.png)

10. Форма отображения статистики, статистика ведется по количеству выполняемых операций, информация собирается в момент отправки выражений на сервер, когда они поступают на бэкенда, для отображения информации нажмите `F5`. При необходимости ее можно скрыть нажав на `Количество операций`.

![](orchestrator/docs/frontend/statistics.png)


> [!TIP]
> ### Как это все может выглядеть в совокупности

![](orchestrator/docs/frontend/summary.png)
