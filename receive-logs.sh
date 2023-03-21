#!/bin/bash

ufw allow 514/udp

go run ./cmd/udp/record/main.go

ufw delete allow 514/udp