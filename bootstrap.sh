#!/bin/bash

mkdir -p "$2"
 
cd "$2"

echo "CD'd to directory: $2"

echo "Executing with parameters: ${@:3}"

/plexmediaserver/Resources/Plex\ New\ Transcoder "${@:3}"