#!/usr/bin/env zsh

ROOT_DIR="$(dirname "$0")"

VARIABLES=("NODE" "REACT_NATIVE" "PYTHON" "GO")

load_or_create_env() {
  ENV_FILE="$ROOT_DIR/.env"
  if [[ -f "$ENV_FILE" ]]; then
    set -a; source "$ENV_FILE"; set +a
  else
    touch "$ENV_FILE"
  fi
}

set_missing_env_variables() {
  for VAR in "${VARIABLES[@]}"; do
    current_value="${(P)VAR}"
    if [[ -z "$current_value" || ( "$current_value" != "true" && "$current_value" != "false" ) ]]; then
      input=""
      while [[ "$input" != "true" && "$input" != "false" ]]; do
        echo -n "Enter the value for $VAR (true/false): "
        read input
      done

      # Remove the previous value if it was defined
      grep -v "^$VAR=" "$ENV_FILE" > "$ENV_FILE.tmp" && mv "$ENV_FILE.tmp" "$ENV_FILE"

      # Add the new value
      echo "$VAR=$input" >> "$ENV_FILE"
    fi
  done
}

# set_os_env() {
#   # Remove the previous value if it was defined
#   grep -v "^OS=" "$ENV_FILE" > "$ENV_FILE.tmp" && mv "$ENV_FILE.tmp" "$ENV_FILE"

#   if [[ "$OSTYPE" == "darwin"* ]]; then
#     echo "OS=osx" >> "$ENV_FILE"
#   elif [[ "$OSTYPE" == "linux-gnu" ]]; then
#     echo "OS=linux" >> "$ENV_FILE"
#   fi
# }

load_or_create_env
set_missing_env_variables
# set_os_env

set -a; source "$ENV_FILE"; set +a