## Install server api
`./install.sh`
****

## Gateway build
`./proto.sh`
****

## Docker
`docker-compose up -d`
****

| Type       | Supported          |
|------------|--------------------|
| 0 - Spot   | Yes                |
| 1 - Stock  | In developing      |
| 2 - Margin | No                 |


## Информация по структуре
****
## Install server api
`./install.sh`
****

## Gateway build
`./proto.sh`
****

## Docker
`docker-compose up -d`
****

| Type       | Supported          |
|------------|--------------------|
| 0 - Spot   | Yes                |
| 1 - Stock  | In developing      |
| 2 - Margin | No                 |


## Информация по структуре
****
1. install.sh - нативное разворачивания server api.
2. proto.sh - генерация/обновления файлов pb.gw.go, pb.go, файлы генерируются в папки каждому модулю server/proto/*, обязательно папка должна быть с префиксов pb.
3. assets/assets.go - файл содержит в себе структурные данные по конфигурации config.json, а также дополнительные функции чтения и обработчики.
4. assets/common - содержит набор дополнительных инструментов.
5. assets/blockchain - содержит функции механизмов чтения блокчейн данных.
6. server/server.go - содержит функции и набор параметров по запуску протокола GRPC Microservices на локальном порту 3081 tcp/udp.
7. server/gateway/gateway.go - содержит функции и набор параметров для запуска протокола http/https server api, порт 3082.
8. server/service - содержит набор интерфейсов и функций микросервисов которые были сгенерированы в пункте 2.
9. создания нового микросервиса:
    * создать папку в каталоге server/proto например pbtest
    * создать файл pbtest.proto в папке pbtest
    * прописать в файле pbtest.proto - syntax, package, option, import. Пример смотреть в файле server/proto/pbspot/spot.proto
    * прописать нужные для нас rpc microservices, сформировать структуру message для наших rpc microservices
    * сгенерировать интерфейсы rpc microservices, смотреть пункт 2
10. создания нового клиента/компонента для взаимодействия с rpc microservices.
    * создать папку в каталоге server/service например test, так как у нас интерфейс rpc с именем pbtest
    * создать файл test.go в папке test
    * прописать в файле test.go type struct с именем Service прописать поданного Context *assets.Context, пример смотреть в файле server/service/spot/spot.go
    * прописать в файле server/server.go, строку регистратора api сервера - pbtest.RegisterApiServer(srv, &test.Service{Context: option}), пример server.go:190 - строка
    * прописать в файле server/gateway/gateway.go строку api обработчика pbtest.RegisterApiHandler, пример gateway.go:358 - строка
    * прописываем интерфейсы в файл spot.go, которые мы прописали в качестве имен в файле spot.proto, например rpc GetName(..., пример смотреть в файле server/service/spot/spot.grpc.go:64-211 - строка
11. Swagger для интеграции смотрим в папке static/swagger, они генерируются вместе с rpc microservices.
