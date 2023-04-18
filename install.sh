#!/bin/bash

sudo apt update
sudo apt install -y build-essential git curl libkrb5-dev cmake postgresql-14 postgresql-client-14 postgresql-server-dev-14 rabbitmq-server redis-server
sudo rabbitmq-plugins enable --offline rabbitmq_management rabbitmq_web_stomp rabbitmq_web_mqtt rabbitmq_shovel rabbitmq_shovel_management

sudo touch rabbitmq.conf
sudo chmod 0777 rabbitmq.conf
sudo echo "loopback_users.guest = false" >> rabbitmq.conf
if [ -f "/etc/rabbitmq/rabbitmq.conf" ]; then
    sudo rm /etc/rabbitmq/rabbitmq.conf
fi
sudo mv rabbitmq.conf /etc/rabbitmq/
sudo service rabbitmq-server restart

git clone https://github.com/timescale/timescaledb/

cd timescaledb/ && ./bootstrap
cd ./build && sudo make && sudo make install -j

cd ..
cd ..

sudo rm -r timescaledb

sudo -i -u postgres pg_dropcluster 14 main --stop
sudo -i -u postgres pg_createcluster 14 main -- --auth-host=scram-sha-256 --auth-local=peer --encoding=utf8 --pwprompt

sudo service postgresql start

sudo -i -u postgres psql -c "alter user postgres with password '**************';"
sudo -i -u postgres psql -c "alter system set listen_addresses to '*';"
sudo -i -u postgres psql -c "alter system set shared_preload_libraries to 'timescaledb';"
sudo -i -u postgres psql -c "create database envoys;"
sudo -i -u postgres psql -c "create user envoys with encrypted password '**************';"
sudo -i -u postgres psql -c "grant all privileges on database envoys to envoys;"

sudo sed -i "s|# host    .*|host all all all scram-sha-256|g" /etc/postgresql/14/main/pg_hba.conf
sudo service postgresql restart

sudo -i -u postgres psql -X -c "create extension if not exists timescaledb cascade;"
for index in db/* ; do
  sudo -i -u postgres psql envoys < "${index}"
done

exit