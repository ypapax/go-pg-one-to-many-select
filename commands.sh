#!/usr/bin/env bash
set -ex
set -o pipefail

run() {
  docker-compose build
  docker-compose up
}

rerun() {
  set +e
  docker kill postgres
  docker rm postgres
  set -e
  run
}

sql() {
  psql -h localhost -p 5439 -U postgres -d companies -c "$@"
}

all(){
  sql "select * from people;"
  sql "select * from companies;"
}

"$@"
