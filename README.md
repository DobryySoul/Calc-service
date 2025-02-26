# Calc service

## Описание проекта
<b>Calc service</b> — это проект, реализованный на языке программирования Golang, который предназначен для вычисления арифметических выражений, таких как "(6+(2+2)*2)/10". Главная цель этого сервиса заключается в том, чтобы предоставить возможность быстро и точно обрабатывать математические выражения, которые могут включать в себя различные операции, такие как сложение, вычитание, умножение и деление, а также выражения со скобками.

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
go mod tidy
```

Отлично, это уже успех! Если вы хотите изменить параметры конфигурации проекта, то продолжите чтение, если же нет, то переходите на следующий [этап](#Запуск).

## Конфигурация 
Сервис можно настроить с помощью переменных среды. Со списком и описанием доступных параметров к настройке можно ознакомиться [здесь](#Список-параметров-конфигурации). 

Чтобы указать параметры переменной среды, который вы хотите использовать, необходимо изменить файл [.env](https://github.com/DobryySoul/Calc-service/blob/main/.env), иначе сервер будет запущен на дефолтых значения, которые указаны в этом файле.

Также есть возможность указать значения переменных среды через команды в терминале, пример:
### Windows

```sh
$env:PORT=порт
```

### Linux и macOS

```sh
export PORT=порт
```

## Endpoints

Сервис имеет следующий эндпоинты:
- `/api/v1/calculate` - отправить новое выражение для вычисления
- `/api/v1/expressions` - получить список всех выражений
- `/api/v1/expression/:id` - получить выражение по идентификатору id
- `/internal/task` - получить задачу для обработки/отправить результат

    - GET: отдает задачу на выполнение

    - POST: отправляет результат выполнения 
    ```json
    {
        "id": уникальный идентификатор, 
        "result": результат вычислений
    }
    ```

## Запуск

Проект готов к запуску.

Бэкенд делиться на 2 части:
1. Оркестратор

2. Агент

Команды, чтобы запустить их:
1. ```sh
    go run cmd/orchestrator/main.go
    ```

2. ```sh
    go run cmd/agent/main.go
    ```

![](docs/starts/start-orchestrator.png)

![](docs/starts/start-agent.png)

Мои поздравления! Сервис успешно запущен и готов к функционированию.


## Список параметров конфигурации

#### `port`

*(номер)* Порт для запуска приложения.

- Эквивалент env: `PORT`.

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

#### `computing_power`
*(продолжительность)* количество независимых вычислителей(горутин)

- Эквивалент env: `COMPUTING_POWER`.

## Обработка запросов и ошибок

### Статус коды
В зависимости от типа запроса, а также корректности выражения, сервер дает различные ответы, с соответствующими статус кодами:

- `201`: Ответ на добавление задачи для выполнения в верном формате 

`{"expression": "аримфметическое выражение верного формата"}`.
  
![](docs/POST/api/v1/calculate/status201.png)

- `422`: Ошибка в арифметическом выражении, например: `{"expression": "2 + 2 * 15}`.

![](docs/POST/api/v1/calculate/status422.png)

- `200`: Успешно получен список выражений:

![](docs/GET/api/v1/expressions/status200.png)

- `200`: Успешно полученное выражение по идентификатору id:

![](docs/GET/api/v1/expression/id/status200.png)

- `404`: Выражение не было найдено по id:

![](docs/GET/api/v1/expression/id/status404.png)

- `200`: Задача успешно получена:
  
![](docs/GET/internal/task/status200.png)

- `404`: Нет задач для выполнения:
  
![](docs/GET/internal/task/status404.png)


- `200`: Успешно записан результат задачи в формате 

`{"id": идентификатор задачи, "result": валидный формат ответа}`.
  
![](docs/POST/internal/task/status200.png)
 
- `404`: По данном id не было найдено задачи.

![](docs/POST/internal/task/status404.png)

- `422`: Невалидный формат введенных данных 

`{"id": идентификатор задачи, "result": невалидный формат ответа}`

![](docs/POST/internal/task/status404.png)

- `500`: Случай внутренней ошибки сервера. Данная ошибка не возникает, так как сервер работает полностью исправно, но все же данная ошибка должна обрабатываться, на случай, когда сервер не сможет обработать запрос к сайту.
 