FROM busybox

RUN mkdir /pms && \
	cd /pms && \
	wget -O pms.deb "https://downloads.plex.tv/plex-media-server/0.9.12.11.1406-8403350/plexmediaserver_0.9.12.11.1406-8403350_amd64.deb" && \
	ar vx pms.deb && \
	tar -zxf data.tar.gz && \
	mv usr/lib/plexmediaserver /plexmediaserver && \
	cd / && \
	rm -Rf /pms

WORKDIR /plexmediaserver

ENV LD_LIBRARY_PATH "/plexmediaserver"

ENTRYPOINT ["/plexmediaserver/Resources/Plex New Transcoder"]