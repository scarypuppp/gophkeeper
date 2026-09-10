# GophKeeper

## Сервер

### S3
Для локальной разработки использовался S3-совместимое хранилище minio.

Запустить:
```shell
docker compose up
```

Создание ключа
```shell
docker exec -it minio sh -c "mc alias set local http://localhost:9000 admin 12345678"
docker exec -it minio sh -c "mc admin accesskey create local"
```

### API

Создать файл с переменными окружения:
```shell
make env
```
Указать креды для PostgreSQL и S3.

Экспорт переменных окружения:
```shell
export $(cat .env | xargs)
```

Запуск dev-сервера.
```shell
make run-server
```

## Клиент

Для онлайн-работы клиента необходимо указать адрес работающего сервера.

```shell
export BASE_URL=...
```

### Авторизация

**Регистрация:**

```shell
gophkeeper register {login} {password}
```

**Логин:**

```shell
gophkeeper login {login} {password}
```

Токен и имя пользователя сохраняется в файле-дампе сессии.
Если существует файл сессии, пользователь считается авторизованным.

При логине выполняется первая синхронизация с сервером. Если сервер недоступен, вход считается
успешным, но выводится сообщение о том, что синхронизация не выполнена.

**Логаут:**

```shell
gophkeeper logout
```

Файл сессии очищается (не удаляется). Вместе с ним очищаются `data.json` и `data_base.json`,
поэтому следующий пользователь на той же машине не увидит чужих данных.
Если в `data.json` есть несинхронизированные изменения, команда предупреждает об этом
и требует подтверждения.

### Храниные данные

1. Бинарные данные (файлы)
2. Текстовые данные
3. Пары логин-пароль
4. Данные банковских карт

Для каждого типа хранимых данных дополнительно хранится произольтая текстовая метаинформация (metadata).
Каждый пользователь зожет получить доступ только к своим данным.

Изменение данных происходит локально. Отправка изменений в данных на сервер происходит в процессе **синхронизации**.

Ввод данных возможен только после логина. Команды `add`, `get`, `list`, `update` и `delete`
работают только с локальным `data.json`; сеть используется исключительно в `login` и `sync`.

Команда `update` частичная: изменяются только явно переданные флаги, остальные поля
сохраняют прежние значения.

### Бинарные данные (файлы)

Поддерживают операции:

1. Create
    ```shell
    gophkeeper file add --file={path_to_file}
    ```
   Файл копируется в каталог `downloads` из конфигурации, запись создаётся локально.
   На сервер содержимое уходит при синхронизации.
   Если запись с таким именем уже есть, содержимое перезаписывается с предупреждением.

2. List
    ```shell
    gophkeeper file list
    ```
   Вывод:
   ```text
   filename     size    stored     meta
   {filename}   {size}  {stored}   {meta}
   {filename}   {size}  {stored}   {meta}
   ```    
   stored - загружен ли файл на локальный диск (✅/❌).
   В `data.json` не хранится, вычисляется по факту наличия файла в каталоге `downloads`.

3. Update
    ```shell
    gophkeeper file update {filename} --meta=...
    ```
   Меняет только метаданные. Содержимое файла считается изменённым, если пользователь
   отредактировал файл в каталоге `downloads`; проверка выполняется по ступеням, каждая
   следующая — только если предыдущая не дала ответа:

   1. `size` отличается от сохранённого — файл изменён;
   2. время изменения файла отличается от `updated_at` записи — переходим к шагу 3;
   3. `checksum` (SHA-256) отличается от сохранённого — файл изменён.

   Полный хэш считается только на третьей ступени, поэтому неизменённые файлы не читаются с диска.

4. Download
    ```shell
    gophkeeper file download {filename}
    ```
   Вывод:
   Progress-bar скачивания файла (если не скачан). Прогресс-бар собственный, без внешних
   зависимостей.

   Файл сохраняется в каталог `downloads` из конфигурации, путь не настраивается флагом.
   После скачивания времени изменения файла присваивается значение `updated_at` записи,
   иначе только что скачанный файл выглядел бы изменённым локально.

5. Delete
    ```shell
    gophkeeper file delete {filename}
    ```
   Удаляет файл с локальной машины вместе с записью в `data.json`.
   На сервере файл удаляется при следующей синхронизации.

### Текстовые данные

Поддерживают все CRUD операции:

1. Create
    ```shell
    gophkeeper text add {name} --text={text} --meta={meta}
    ```
2. Get
    ```shell
    gophkeeper text get {name}
    ```
   Вывод:
   ```text
   name: {name}
   text: {text}
   meta: {meta}
   ```    
3. List
    ```shell
    gophkeeper text list
    ```
   Вывод:
   ```text
   name     text             meta
   {name}   {text[:20]}...   {meta[:20]}...
   {name}   {text[:20]}...   {meta[:20]}...
   ```    
4. Update
    ```shell
    gophkeeper text update {name} --text={text} --meta={meta}
    ```
5. Delete
    ```shell
    gophkeeper text delete {name}
    ```

### Пары логин-пароль (секреты)

Поддерживают все CRUD операции:

1. Create
    ```shell
    gophkeeper secret add {name} --login={login} --password={password} --meta={meta}
    ```
2. Get
    ```shell
    gophkeeper secret get {name}
    ```
   Вывод:
   ```text
   name: {name}
   login: {login}
   password: {password}
   meta: {meta}
   ```    
3. List
    ```shell
    gophkeeper secret list
    ```
   Вывод:
   ```text
   name     login      meta
   {name}   {login}    {meta}
   {name}   {login}    {meta}
   ```    
4. Update
    ```shell
    gophkeeper secret update {name} --login={login} --password={password} --meta={meta}
    ```
5. Delete
    ```shell
    gophkeeper secret delete {name}
    ```

### Данные банковских карт

1. Create
    ```shell
    gophkeeper card add {name} --number=. --holder=. --expires=. --cvv=. --meta=.
    ```
2. Get
    ```shell
    gophkeeper card get {name}
    ```
   Вывод:
   ```text
   name: {name}
   number: {number}
   holder: {holder}
   expires: {expires}
   cvv: {cvv}
   ```    
3. List
    ```shell
    gophkeeper card list
    ```
   Вывод:
    ```text
    name    number              exp     meta
    {name}  ...{number[-4:]}    {exp}   {meta[:20]}
    {name}  ...{number[-4:]}    {exp}   {meta[:20]}
    ```    
4. Update
    ```shell
    gophkeeper card update {name} --number=. --holder=. --expires=. --cvv=. --meta=.
    ```
5. Delete
    ```shell
    gophkeeper card delete {name}
    ```

### Хранение данных локально

Данные в json файле `data.json`. Шифрование не выполняется, файл хранится в открытом виде.

Формат файла:

```json
{
  "Files": [
    {
      "file_name": "",
      "file_hash": "",
      "size": 1,
      "meta": "",
      "checksum": "",
      "updated_at": ""
    }
  ],
  "Texts": [
    {
      "name": "",
      "text": "",
      "meta": "",
      "checksum": "",
      "updated_at": ""
    }
  ],
  "Cards": [
    {
      "name": "",
      "number": "",
      "holder": "",
      "expires": "",
      "cvv": "",
      "meta": "",
      "checksum": "",
      "updated_at": ""
    }
  ],
  "Secrets": [
    {
      "name": "",
      "login": "",
      "password": "",
      "meta": "",
      "checksum": "",
      "updated_at": ""
    }
  ]
}
```


### Синхронизация данных сервером

Авторизованный пользователь может синхронизировать сессию с сервером.

```shell
gophkeeper sync
```

При отсутствии конфликтов будет выведен список полученных данных:

```text
card {name} created 
text {name} updated
file {file_name} deleted
```

Порядок выполнения:

1. запросить у сервера полное состояние данных;
2. локально вычислить решения по таблице (local / base / remote);
3. разрешить все конфликты;
4. отправить изменения на сервер и скачать недостающие данные;
5. перезаписать `data_base.json`.

Изменения уходят на сервер только после того, как не осталось нерешённых конфликтов.

При вызове команды при отсутствии конфликтов создается файл data_base.json. Он хранит BASE версию для дальнейшего 
сравнения с local и server.

 Таблица принятия решений при синхронизации.
| local vs base | remote vs base | решение                               |
|---------------|----------------|---------------------------------------|
| =             | =              | ничего не делать                      |
| ≠             | =              | залить на сервер                      |
| =             | ≠              | скачать с сервера                     |
| ≠             | ≠              | конфликт                              |
| нет локально  | =              | удалёно локально → удалить на сервере |
| =             | нет на сервере | удалён на сервере → удалить локально  |


Если `data_base.json` отсутствует, все локальные и все серверные записи считаются новыми,
а при совпадении имён серверная версия побеждает без запроса.

При конфликте пользователю нужно выбрать, какие даннные выбрать (theirs), с сервера или локальные (yours).
Пользователю будут side to side версии данных. Выбор осуществляется вводом `s` (server) или `l` (local).

Флаги `--theirs` (взять серверную версию) и `--yours` (взять локальную) применяются сразу ко всем
конфликтным записям. Без флага клиент спрашивает по каждой записи отдельно.

Для файлов сравнение содержимого невозможно, поэтому выводятся `name`, `size`, `checksum`
и `updated_at` обеих версий.
