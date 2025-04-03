```azure
  nginx:
    image: docker.servicewall.cn/servicewall/nginx_decryptor:latest
    container_name: sw_nginx
    entrypoint: ["/bin/sh", "-c", "/entrypoint.sh"]
    environment:
      - LISTEN_PORT=8080 # nginx监听的端口
      - GATEWAY_PORT=28080 # 网关服务监听的端口
      - LOCAL_HOST_IP=192.168.110.88 # 当前主机的ip地址
      - TARGET_ADDR=http://127.0.0.1:10022 # 网关防护的目标服务地址
    network_mode: "host"
```