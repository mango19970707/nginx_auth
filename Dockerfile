############################################################
# Backend Build
############################################################
#FROM docker.servicewall.cn/golang-builder AS builder
#
#WORKDIR /app
#
#COPY ./gateway/go.mod .
#COPY ./gateway/go.sum .
#RUN go mod download
#
#COPY ./gateway .
#RUN CGO_ENABLED=0 go build -o gateway main.go

FROM docker.m.daocloud.io/nginx:alpine

COPY nginx/nginx.conf /etc/nginx/nginx.conf
COPY gateway /gateway
#COPY --from=builder /app /gateway
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["sh", "-c", "/entrypoint.sh"]