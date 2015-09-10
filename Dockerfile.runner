FROM debian

RUN apt-get update && \
	apt-get install -y curl && \
	apt-get clean && \
	rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

RUN mkdir /pms && \
	cd /pms && \
	curl -o pms.deb "https://downloads.plex.tv/plex-media-server/0.9.12.11.1406-8403350/plexmediaserver_0.9.12.11.1406-8403350_amd64.deb" && \
	dpkg-deb -x pms.deb ./ && \
	mv usr/lib/plexmediaserver /plexmediaserver && \
	cd / && \
	rm -Rf /pms


WORKDIR /plexmediaserver

ADD bootstrap.sh /plexmediaserver/bootstrap.sh

ENV LD_LIBRARY_PATH "/plexmediaserver"

ENTRYPOINT ["/plexmediaserver/bootstrap.sh"]
