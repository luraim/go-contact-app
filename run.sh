#!/bin/bash

set -euxo pipefail

export BIN="contactsApp"

go build -o $BIN
CONTACTS_SESSION_KEY="contacts123" ./$BIN