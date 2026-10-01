FROM debian:bookworm-slim
WORKDIR /app
RUN sed -i 's|http://deb.debian.org/debian|http://mirrors.tuna.tsinghua.edu.cn/debian|g' /etc/apt/sources.list.d/debian.sources \
 && apt-get update \
 && apt-get install -y --no-install-recommends ca-certificates libmp3lame0 libopencore-amrnb0 libopencore-amrwb0 libqmi-utils libpcsclite1 libvo-amrwbenc0 psmisc tzdata wget iproute2 \
 && rm -rf /var/lib/apt/lists/* \
 && test -x /usr/libexec/qmi-proxy \
 && mkdir -p config data logs
COPY vohive-plus_linux_amd64 /usr/local/bin/vohive-plus
RUN chmod 0755 /usr/local/bin/vohive-plus
LABEL org.opencontainers.image.title="VoHive Plus Personal" \
      org.opencontainers.image.version="0.1.2-personal-classic" \
      org.opencontainers.image.source="https://github.com/yibaiba/hideck" \
      org.opencontainers.image.revision="3fd3d6aaf55924c338c2cf3f092e1af1f588d152+personal"
ENV TZ=Asia/Shanghai
EXPOSE 7575/tcp 7576/tcp 7580/udp
HEALTHCHECK --interval=30s --timeout=5s --start-period=45s --retries=3 CMD wget -q -T 3 -O /dev/null http://127.0.0.1:7575/ping || exit 1
STOPSIGNAL SIGTERM
ENTRYPOINT ["/usr/local/bin/vohive-plus"]
CMD ["-c", "/app/config/config.yaml", "--database", "/app/data/vohive-plus.db"]
