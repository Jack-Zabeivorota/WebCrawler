#!/bin/bash

ROOT_DIR=$(pwd)

cd "$ROOT_DIR/aggregator"
bash builder.sh

cd "$ROOT_DIR/controller"
bash builder.sh

cd "$ROOT_DIR/main"
bash builder.sh

cd "$ROOT_DIR/worker"
bash builder.sh