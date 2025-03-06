# Калькулятор на языке Golang

Калькулятор основан на взаимодействии двух серверов: оркестратора и агента, где оркестратор делит входное математическое выражение на отдельные задачи, а агент эти задачи вычисляет.


## Запуск

1. Клонируйте библиотеку через git clone 

2. Перейдите в папку с программой (calc_go_2.0) :
```
cd calc_go_2.0
```

3.  и выполните команду:
```
go run cmd/main.go
```

4. Откройте cmd-терминал и введите команду:
```
curl -X POST http://localhost:8081/api/v1/calculate -H "Content-Type:application/json" -d "{\"expression\":\"1+2+3+4+5\"}"
```

5. В другом cmd-терминале введите команду:
```
curl -X POST http://localhost:8080/internal/calculator
```

6. Для того, чтобы вычислять одновременно два и более выражений, каждое новое следует запускать командой в отдельном терминале:
```
curl -X POST http://localhost:8081/api/v1/calculate -H "Content-Type:application/json" -d "{\"expression\":\"5-4-3-2-1\"}"
```

7. Чтобы получить все выражения, используйте в отдельном терминале команду:
```
curl -X http://localhost:8081/api/v1/expressions
```

8. Чтобы получить выражение по его ID, используйте команду (вместо id_0 можно вписать любой из имеющихся ID):
```
curl -X http://localhost:8081/api/v1/expression/id_0
```

## Принцип работы

В калькуляторе взаимодействуют пользователь, оркестратор и агент.
фотка

Оркестратор принимает от пользователя выражение, добавляет его в хранилище, присваивая ID, представляет в виде обратной польской записи. Затем начинается деление выражения на простейшие задачи (задача представлена структурой Task, имеющей поля ID, Status, Arg1, Arg2, Operation, Operation_time, Result, Error). Только что сформированный task отправляется в стек задач. Функция деления на задачи уходит в ожидание до тех пор, пока статус у отправленной им задачи не поменяется на "done".

Всё это время, сразу после запуска сервера агента, он отправляет оркестратору GET-запросы каждые 100 ms, желая получить задачу. При каждом запросе идёт проверка на наличие нерешённых задач (с Status == "need to calculate") в стеке задач. Если такие имеются, то агенту выдаётся первая нерешенная задача из стека задач. 

Попав к агенту, задача решается определённым количеством параллельно запущенных воркеров. Их количество задаётся переменной среды COMPUTING_POWER. В калькуляторе имитируются долгие вычисления, поэтому переменными среды задано время выполнения для каждой математической операции. Как только один из воркеров завершит вычисление задачи, останавливается работа всех остальных, а результат (либо ошибка, если таковая обнаружилась в процессе вычисления) отправляется обратно оркестратору. Следующим GET-запросом агент берет следующую нерешённую задачу.

Получив результат, оркестратор изменяет вычисленную задачу в стеке, добавляя в соответствующие поля результат/ошибку, а также меняет статус на "done". Функция деления на задачи выходит из режима ожидания и анализирует результаты: если ошибки не возникло - подставляем результат вычислений в следующую задачу, если ошибка всё же возникла - возвращаем её пользователю.

Как только всё выражение будет вычислено без ошибок - выводим результат пользователю.

В процессе работы сервера можно также просмотреть хранилище выражений, и найти выражение по ID с помощью запросов end-поинтами "/api/v1/expressions" и "/api/v1/expression/:id" соответственно.


## Структура проекта
```
calc_go_2.0 
	├── cmd 
	│    └── main.go
	├── env
	│    └── env_vars.go
	├── internal 
	│    ├── calculator 
	│    │    ├── calculator.go 
	│    │    └── calculator_test.go 
	│    └── orkestrator 
	│         ├── orkestrator.go
	│         ├── orkestrator_test.go 
	│         ├── o_server.go 
	│         ├── storage.go 
	│         └── errors.go 
	├── go.mod 
	└── go.sum
```

#### cmd
- main.go - отсюда запускаются два сервера: для оркестратора и для агента

#### internal/calculator - файлы агента
- calculator.go - вся реализация агента находится здесь

#### internal/orkestrator
- orkestrator.go - функции преобразования выражения в обратную польскую запись и деления на задачи
- o_server.go - реализация сервера для оркестратора (все хендлеры)
- storage.go - внутренние функции добавления и удаления выражений и задач в хранилища и стеки, объявление структур
- errors.go - объявление ошибок

#### env
- env_vars.go - переменные среды


## Переменные среды

Переменные среды и их значения по умолчанию (значения можно изменить в env_vars.env):
```
TIME_ADDITION_MS=5000          //время выполнения сложения в мс
TIME_SUBTRACTION_MS=6000       //время выполнения вычитания в мс
TIME_MULTIPLICATIONS_MS=7000   //время выполнения умножения в мс
TIME_DIVISIONS_MS=8000         //время выполнения деления в мс

COMPUTING_POWER=3              //количество запускаемых воркеров
```