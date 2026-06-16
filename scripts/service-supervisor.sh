#!/bin/bash

set -uo pipefail

NAME="$1"
WORKDIR="$2"
LOG_FILE="$3"
SUPERVISOR_PID_FILE="$4"
CHILD_PID_FILE="$5"
COMMAND="$6"

mkdir -p "$(dirname "$SUPERVISOR_PID_FILE")"
mkdir -p "$(dirname "$LOG_FILE")"

STOP_REQUESTED=0
CHILD_PID=""

cleanup() {
    STOP_REQUESTED=1
    if [[ -n "${CHILD_PID:-}" ]] && kill -0 "$CHILD_PID" 2>/dev/null; then
        kill -TERM "$CHILD_PID" 2>/dev/null || true
        wait "$CHILD_PID" 2>/dev/null || true
    fi
    rm -f "$CHILD_PID_FILE" "$SUPERVISOR_PID_FILE"
    exit 0
}

trap cleanup TERM INT HUP

echo "$$" > "$SUPERVISOR_PID_FILE"

while true; do
    (
        cd "$WORKDIR"
        exec bash -lc "$COMMAND"
    ) >> "$LOG_FILE" 2>&1 &
    CHILD_PID=$!
    echo "$CHILD_PID" > "$CHILD_PID_FILE"

    wait "$CHILD_PID"
    EXIT_CODE=$?
    rm -f "$CHILD_PID_FILE"

    if [[ "$STOP_REQUESTED" -eq 1 ]]; then
        break
    fi

    printf '[%s] supervisor: %s exited with code %s, restarting in 1s\n' \
        "$(date '+%Y-%m-%d %H:%M:%S')" "$NAME" "$EXIT_CODE" >> "$LOG_FILE"
    sleep 1
done

rm -f "$CHILD_PID_FILE" "$SUPERVISOR_PID_FILE"
