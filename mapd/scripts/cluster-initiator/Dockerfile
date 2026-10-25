FROM harbor.huaruicloud.cn/map/cluster-initiator@sha256:2a81eac4071e07b67225f2d1b7d164d078f0e915abfc7ffae7518596c6238db2

COPY disable-default-velero.sh /tmp/disable-default-velero.sh
RUN /bin/bash /tmp/disable-default-velero.sh /usr/bin/entrypoint.sh \
    && rm /tmp/disable-default-velero.sh
