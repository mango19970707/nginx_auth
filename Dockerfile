FROM docker.m.daocloud.io/nginx:alpine

COPY nginx/nginx.conf /etc/nginx/nginx.conf
COPY gateway /gateway
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["sh", "-c", "/entrypoint.sh"]