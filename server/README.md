
### Флаги сервиса.

 - `sdf`. Значение по умолчанию - `files` . Переменная окружения -`SERVER_SUBDIR_FILES`.

 - `sdb`. Значение по умолчанию - `backup`. Переменная окружение - `SERVER_SUBDIR_BACKUP`. 

 - `d`. Значение по умалчанию - `file:remote.db?cache=shared&foreign_keys=on&mode=rwc`. Переменная окружение - `SERVER_DB_DSN`.
    Поддерживается только SQlite3.

 - `p`. Значение по умолчанию - `50100`. Переменная окружение - `SERVER_PORT`. 

 ### Как собрать приложение.

Для Linux:   `go build -ldflags "-X main.buildVersion=1.0.0 -X main.buildDate=$(date +%Y-%m-%d)" -o <имя>`
Для Windows: `GOOS=windows GOARCH=386 go build -ldflags "-X main.buildVersion=1.0.0 -X main.buildDate=$(date +%Y-%m-%d)" -o <имя>.exe`
Для macOS:   `GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.buildVersion=1.0.0 -X main.buildDate=$(date +%Y-%m-%d)" -o <имя>`
