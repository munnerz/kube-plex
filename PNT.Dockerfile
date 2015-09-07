FROM scratch

ADD "Plex New Transcoder" "/Plex New Transcoder"
ADD "libraries/" "/libraries"

ENV LD_LIBRARY_PATH "/libraries"

ENTRYPOINT ["/Plex New Transcoder"]