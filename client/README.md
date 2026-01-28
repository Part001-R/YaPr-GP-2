
### Флаги сервиса.

 - `m`. Значение по умолчанию - `local` . Переменная окружения -`MODE_CLIENT`.
    Возможные значения: `local`, `remote`.

- `d`. Значение по умалчанию - `file:localStorage.db?cache=shared&foreign_keys=on&mode=rwc`. Переменная окружение - `DSN_STORAGE`.
    Поддерживается только SQlite3.

- `container`. Значение по умолчанию - `localContainer.data`. Переменная окружение - `LOCAL_NAME_CONTAINER`. 

 
### Как собрать приложение.

Для Linux:   `go build -o <имя> main.go`
Для Windows: `GOOS=windows GOARCH=386 go build -o <имя.exe> main.go`
Для macOS:   `GOOS=darwin GOARCH=amd64 go build -o <имя> main.go`

### Примечание.

- Для Windows.
Удаление данных в поле ввода, выполнять через `Delete`. Перемещение позиции через клавиши `<-` и `->`.
