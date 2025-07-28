#!/bin/bash

go build ..
./gochat authenticate -u Chek1 -p 123456
./gochat chat -u Chek1

