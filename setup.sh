#!/bin/bash

declare -Ag deps=(
  [git]="git"
  [wget]="wget"
  [go]="go"
  [go]="golang"
  [go]="golang-go"
  [postgres]="postgresql"
  [postgres]="postgres-server"
  ["redis-server"]="redis"
  ["redis-server"]="redis-server"
  [vips]="libvips"
  [vips]="vips"
)

declare -Ag package_managers=(
  [pacman]="sudo pacman -S"
  [apt]="sudo apt-get install -y"
  [yum]="sudo yum install -y epel-release"
)

install_deps() {
  for pm in ${!package_managers[@]}; do
    if which $pm &>/dev/null; then
      eval $(echo ${package_managers[$pm]} $*)
    fi
  done
}

for dep in ${!deps[@]}; do
  if ! which $dep &>/dev/null; then
    echo "Installing $dep..."
    install_deps ${deps[$dep]}
    echo
  fi
done

echo "Initializing postgresql..."
echo initdb --locale en_US.UTF-8 -D /var/lib/postgres/data | sudo su - postgres &>/dev/null

echo "Enabling postgresql and redis services..."
sudo systemctl enable postgresql redis

echo "Starting postgres and redis services..."
sudo systemctl start postgresql redis

pg_host=localhost
pg_port=5432
pg_dbname=kasen
pg_username=kasen
pg_password=kasen

echo
echo === Configuring postgres ===
echo Leave it empty to use the default value
echo

read -p "[Postgres] Host (default: $pg_host): " input
[ ! -z $input ] && pg_host=$input

read -p "[Postgres] Port (default: $pg_port): " input
[ ! -z $input ] && pg_port=$input

read -p "[Postgres] Database Name (default: $pg_dbname): " input
[ ! -z $input ] && pg_dbname=$input

read -p "[Postgres] Username (default: $pg_username): " input
[ ! -z $input ] && pg_username=$input

read -p "[Postgres] Password (default: $pg_password): " input
[ ! -z $input ] && pg_password=$input

echo
echo Creating user...
psql -h $pg_host -p $pg_port -U postgres -tc "SELECT 1 FROM pg_user WHERE usename = '$pg_username'" | grep -q 1 || psql -h $pg_host -p $pg_port -U postgres -c "CREATE USER $pg_username LOGIN SUPERUSER PASSWORD '$pg_password'" >/dev/null

echo Creating database...
psql -h $pg_host -p $pg_port -U postgres -tc "SELECT 1 FROM pg_database WHERE datname = '$pg_dbname'" | grep -q 1 || psql -h $pg_host -p $pg_port -U postgres -c "CREATE DATABASE $pg_dbname OWNER $pg_username" >/dev/null

redis_host=localhost
redis_port=6379
redis_db=0
redis_password=""

echo
echo === Configuring redis ===
echo Leave it empty to use the default value
echo

read -p "[Redis] Host (default: $redis_host): " input
[ ! -z $input ] && redis_host=$input

read -p "[Redis] Port (default: $redis_port): " input
[ ! -z $input ] && redis_port=$input

read -p "[Redis] Database index (default: $redis_db): " input
[ ! -z $input ] && redis_db=$input

read -p "[Redis] Password (default: empty): " input
[ ! -z $input ] && redis_password=$input

echo
echo Creating configuration...
cat >kasen.yaml <<EOF
database:
  host: $pg_host
  port: $pg_port
  dbname: $pg_dbname
  username: $pg_username
  password: $pg_password

directories:
  templates: ./templates
  assets: ./assets
  data: ./data

redis:
  host: $redis_host
  port: $redis_port
  db: $redis_db
  password: $redis_password

server:
  port: 42072
EOF

echo Completed!
