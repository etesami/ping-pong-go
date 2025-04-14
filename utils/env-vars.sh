#!/bin/bash

while [[ "$#" -gt 0 ]]; do
  case $1 in
    -p|--ping) ping=1 ;;
    -q|--pong) pong=1 ;;
    *) echo "Unknown parameter passed: $1"; exit 1 ;;
  esac
  shift
done

if [[ -z $ping && -z $pong ]]; then
  echo "No service specified. Exiting..."
  exit 1
fi

if [[ $ping -eq 1 ]]; then
  export SVC_ADDR="0.0.0.0"
  export SVC_PORT="50051"
  export METRIC_ADDR="0.0.0.0"
  export METRIC_PORT="9100"
  export FILE_SIZE="1.0"
  export UPDATE_FREQUENCY="10"
fi

if [[ $pong -eq 1 ]]; then
  export SVC_ADDR="0.0.0.0"
  export SVC_PORT="50051"
  export FILE_SIZE="1.0"
fi