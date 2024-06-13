#!/bin/bash

set -e

readonly app="$1"
readonly config_file="config.yaml"

"$app" "$config_file"