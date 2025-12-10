#!/bin/bash

BASE="${PWD}/scripts/systemd"

bash "$BASE/backend.sh"
bash "$BASE/ui.sh"
bash "$BASE/ocserv.sh"