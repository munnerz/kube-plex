#!/bin/bash

mkdir -p "$1"
 
cd "$1"

echo "CD'd to directory: $1"

echo "Executing with parameters: ${@:2}"

/plexmediaserver/Resources/Plex\ New\ Transcoder "${@:2}"