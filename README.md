### 简介
用于反向代理没有鉴权的服务。

### docker-compose配置示例
```azure
  nginx_auth:
    image: docker.servicewall.cn/servicewall/nginx_auth:latest
    container_name: sw_nginx_auth
    entrypoint: ["/bin/sh", "-c", "/entrypoint.sh"]
    volumes:
        - gateway_data:/gateway
    environment:
      - LISTEN_PORT=8080 # nginx监听的端口
      - GATEWAY_PORT=28080 # 网关服务监听的端口
      - CONSOLE_ADDR=http://127.0.0.1:10020 # console服务的地址
      - DECRYPT_ADDR=http://127.0.0.1:10022 # 解密配置服务的地址
    network_mode: "host"
```