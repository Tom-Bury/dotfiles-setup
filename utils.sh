#!/usr/bin/env zsh

print_header() {
  echo "=== $1 ==="
}

print_footer() {
  echo "=== $1 ✅ ==="
  echo ""
}

safe_create_folder() {
  folder_path=$1
  if [[ ! -d $folder_path ]]; then
    mkdir -p $folder_path
    echo -e "- created $folder_path folder 🆕"
  else
    echo -e "- $folder_path folder already exists 🗂"
  fi
}

DATETIME_FORMAT="%Y-%m-%d_%H-%M-%S"

create_backup() {
  mkdir -p "$HOME/.bury-setup-backups/"
  if [ -f "$1" ]; then
    cp "$1" "$HOME/.bury-setup-backups/$(basename $1)_$(date +$DATETIME_FORMAT)"
  fi
}

setup_extra_source_scripts() {
  local dest_dir="$HOME/zshrc-scripts" 

  [ -d "$dest_dir" ] || mkdir -p "$dest_dir"

  # Define the source files
  local source_file_prefix=".zshrc_"
  local scan_dir=$1
  local source_files=($(find $scan_dir -type f -name "$source_file_prefix*"))

  # Loop over the source files
  for source_file in "${source_files[@]}"; do
    # Extract the filename from the source file path
    local file_name=$(basename "$source_file")

    # Check if the source file exists and copy it
    if [ -f "$source_file" ]; then
      cp "$source_file" "$dest_dir/$file_name"
    else
      echo "Source file $source_file does not exist."
    fi
  done
}