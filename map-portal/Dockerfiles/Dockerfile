# stage 1: build angular application
# need docker memory resource setup > 2G. fixed https://www.ngdcorp.com/2021/02/03/docker-yarn-sigkill/
FROM node:16.15.0-alpine3.15 as build-step
ARG YARN_REGISTRY="https://registry.yarnpkg.com"
ENV NODE_OPTIONS="--max-old-space-size=8192"
RUN mkdir -p /app
WORKDIR /app
COPY ../.. /app
RUN yarn config set registry $YARN_REGISTRY && yarn && yarn build

# stage 2: add application into nginx
FROM nginx:1.21.4

ENV DEX_IDP_ENDPOINT http://dex-service:5556
ENV MAP_API_ENDPOINT http://mapd-service:9800
ENV ENABLE_SSL off
ENV TLS_CERT_PATH /etc/nginx/ssl/server.crt
ENV TLS_KEY_PATH /etc/nginx/ssl/server.key
RUN rm -f /etc/nginx/conf.d/default.conf && rm -f /etc/nginx/nginx.conf
COPY --from=build-step /app/dist/map-portal /usr/share/nginx/html/portal
COPY --from=build-step /app/Dockerfiles/nginx.conf /etc/nginx/nginx.conf
COPY --from=build-step /app/Dockerfiles/nginx.conf.template /etc/nginx/templates/nginx.conf.template
