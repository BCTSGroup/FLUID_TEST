#!/bin/bash

./terminate.sh > /dev/null

./build_linux.sh || { echo -e "\033[31mError: build failed!\033[0m"; exit 1; }

cd test && ./test_linux.sh

