#!/bin/bash
cd ./src
swag fmt
swag init -g main.go
cd ..