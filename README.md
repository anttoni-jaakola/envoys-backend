**Build tools:**

* $ sudo apt-get install build-essential
* $ sudo apt-get install curl
* $ sudo apt-get install -y libkrb5-dev // Timescaledb build
* $ sudo apt-get install -y cmake // Timescaledb build

**Grpc package:**

* $ sudo apt install protobuf-compiler && sudo apt install libprotobuf-dev
* $ go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
* $ go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
* $ go install github.com/golang/protobuf/protoc-gen-go
* $ go install github.com/golang/protobuf/{proto,protoc-gen-go}

****

**Gateway build:**
> $ protoc -I=. -I/usr/local/include -I=$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway -I=$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --grpc-gateway_out=logtostderr=true:. --go_out=plugins=grpc:. server/proto/*.proto

**Gateway+Swagger build:**
> $ protoc -I=. -I/usr/local/include -I=$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway -I=$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --grpc-gateway_out=logtostderr=true:. --go_out=plugins=grpc:. --swagger_out=logtostderr=true:./swagger server/proto/*.proto

****

**Install PostgreSQL and create new database and setting timescaledb:**

* https://packagecloud.io/timescale/timescaledb/install#bash-deb
* Build https://www.timescale.com/blog/how-to-build-timescaledb-on-docker 
* $ sudo apt update   
* $ sudo apt -y install postgresql-14 postgresql-client-14 postgresql-server-dev-14
* $ sudo timescaledb-tune  // Not for the built version  
* $ sudo service postgresql restart  
* $ sudo su - postgres  
> $ psql -c "alter user postgres with password 'envoys'"  
> $ psql  
> $ CREATE DATABASE envoys;  
> $ CREATE USER envoys WITH ENCRYPTED PASSWORD 'envoys';  
> $ GRANT ALL PRIVILEGES ON DATABASE envoys to envoys;  
> $ \c envoys  
> $ CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;  
* $ sudo service postgresql restart  
> Deploy database to server ./static/dump
> $ SELECT create_hypertable('trades', 'create_at');

********

**Restart/Start commands:**

> $ sudo service postgresql start  
> $ sudo service rabbitmq-server start  
> $ sudo service redis-server start

********

**Install Rabbitmq and setting:**

* $ sudo apt -y install rabbitmq-server
* $ sudo rabbitmq-plugins enable rabbitmq_management
* $ sudo rabbitmq-plugins enable rabbitmq_web_stomp
* $ sudo rabbitmq-plugins enable rabbitmq_web_mqtt
* $ sudo rabbitmq-plugins enable rabbitmq_shovel rabbitmq_shovel_management
* $ sudo service rabbitmq-server restart
* Delete all the queues from RabbitMQ 
* $ sudo rabbitmqctl stop_app 
* $ sudo rabbitmqctl reset    # Be sure you really want to do this!
* $ sudo rabbitmqctl start_app 
 
> Management host: http://localhost:15672  
> Login: guest  
> Password: guest

********

**Install Redis:**

* $ sudo apt install redis-server

**Rules admin:**
* ["currencies", "chains", "pairs", "accounts", "contracts", "listing", "news", "support", "advertising"].