FROM scratch

ADD "Plex New Transcoder" "/Plex New Transcoder"

ENTRYPOINT ["/Plex New Transcoder"]