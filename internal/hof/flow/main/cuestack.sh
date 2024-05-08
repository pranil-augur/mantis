#!/bin/bash

# Function to display help for the script
function show_help() {
  echo "Usage: $0 <command> [options]"
  echo ""
  echo "Commands:"
  echo "  run       Run a cue flow from a file or directory"
  echo "  artifact  Manage OCI artifacts"
  echo ""
  echo "Run Flags:"
  echo "  -h, --help     Show help for run"
  echo "Global Flags:"
  echo "  -A, --apply    Apply the proposed state"
  echo "  -D, --destroy  Destroy resources"
  echo "  -I, --init     Init modules"
  echo "  -P, --preview  Preview the changes to the state"
  echo ""
  echo "Artifact Commands:"
  echo "  list        List the tags of an artifact"
  echo "  pull        Pull an artifact from a container registry"
  echo "  push        Push a directory contents to a container registry"
  echo "  tag         Tag an OCI artifact in the upstream registry"
}

# Check if at least one argument is provided
if [ "$#" -lt 1 ]; then
  show_help
  exit 1
fi

COMMAND=$1
shift

case $COMMAND in
  run)
    # Run a cue flow from a file or directory using flowrunner
    "$(dirname "$0")/flowrunner" run "$@" ;;
  artifact)
    # Subcommands for artifact management using cues
    if [ "$#" -lt 1 ]; then
      echo "Usage: $0 artifact <subcommand> [options]"
      echo "Subcommands: list, pull, push, tag"
      exit 1
    fi
    SUBCOMMAND=$1
    shift
    case $SUBCOMMAND in
      list|pull|push|tag)
        "$(dirname "$0")/cues" artifact $SUBCOMMAND "$@" ;;
      *)
        echo "Unknown artifact subcommand: $SUBCOMMAND"
        exit 1 ;;
    esac ;;
  *)
    echo "Unknown command: $COMMAND"
    show_help
    exit 1 ;;
esac