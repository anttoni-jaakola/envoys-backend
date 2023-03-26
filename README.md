****
## Install server api
`./install.sh`
****

## Gateway build
`./proto.sh`
****

## Docker
`docker-compose up --build`
****

| Type       | Supported          |
|------------|--------------------|
| 0 - Spot   | Yes                |
| 1 - Stock  | In developing      |
| 2 - Margin | No                 |

## Информация по структуре
****
1. `install.sh` - скрипт для нативного развертывания сервера API.
2. `proto.sh` -  скрипт для генерации/обновления файлов `pb.gw.go`, `pb.go`. Файлы генерируются в папках для каждого клиентского интерфейса `server/proto/*`. Обязательно добавить префикс `pb` к названию папки.
3. `assets/assets.go` - файл содержит структурные данные по конфигурации `config.json`, а также дополнительные функции чтения и обработки.
4. `assets/common` - содержит набор дополнительных инструментов.
5. `assets/blockchain` - содержит функции механизмов чтения данных блокчейна.
6. `server/server.go` - содержит функции и параметры для запуска протокола `GRPC Microservices` на локальном порту `3081` `TCP/UDP`.
7. `server/gateway/gateway.go` - содержит функции и параметры для запуска протокола `HTTP/HTTPS` Server API на порту `3082`.
8. `server/service` - содержит набор интерфейсов и функций микросервисов, которые были сгенерированы в пункте `2`.
9. для создания нового микросервиса:
    * создать папку в каталоге `server/proto`, например, `pbtest`.
    * создать файл `pbtest.proto` в папке `server/proto/pbtest`.
    * прописать в файле `server/proto/pbtest/pbtest.proto` следующие параметры: `syntax`, `package`, `option`, `import`. Пример можно посмотреть в файле `server/proto/pbspot/spot.proto`.
    * прописать необходимые для нас `RPC Microservices` и сформировать структуру сообщений для наших `RPC Microservices`.
    * сгенерировать интерфейсы `RPC Microservices`, смотреть пункт `2`.
10. для создания нового клиента/компонента для взаимодействия с `RPC Microservices`:
    * создать папку в каталоге `server/service`, например, `test`, так как у нас `proto` `package` с именем `pbtest`.
    * создать файл `test.go` в папке `server/service/test`.
    * прописать в файле `server/service/test/test.go` структуру типа с именем `Service` и прописать следующие параметры: поданного контекста `*assets.Context`. Пример можно посмотреть в файле `server/service/spot/spot.go:28`.
    * прописать в файле `server/server.go` строку регистратора API сервера - `pbtest.RegisterApiServer(srv, &test.Service{Context: option})`. Пример можно посмотреть в файле `server/server.go:190`.
    * прописать в файле `server/gateway/gateway.go` строку API обработчика `pbtest.RegisterApiHandler`. Пример можно посмотреть в файле `server/gateway/gateway.go:358`.
    * прописать интерфейсы в файле `server/service/test/test.go`, соответствующие имена, которые были прописаны в `server/proto/pbtest/test.proto`. Например, `rpc GetAnalysis(...)`. Пример можно посмотреть в файле `server/service/spot/spot.grpc.go:64-211`.
11. `Swagger` для интеграции смотрим в папке `static/swagger`.
