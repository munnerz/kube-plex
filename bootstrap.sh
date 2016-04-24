#!/bin/bash

cd "$1"

echo "CD'd to directory: $1"

echo "Executing with parameters: ${@:2}"

export LD_LIBRARY_PATH=/usr/lib/plexmediaserver

/plexmediaserver/Resources/Plex\ New\ Transcoder "${@:2}"