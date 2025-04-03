sed -i "s/LISTEN_PORT/${LISTEN_PORT}/" /etc/nginx/nginx.conf
sed -i "s/GATEWAY_PORT/${GATEWAY_PORT}/" /etc/nginx/nginx.conf
sed -i "s|TARGET_ADDR|${TARGET_ADDR}|" /etc/nginx/nginx.conf
cat /etc/nginx/nginx.conf
rm /etc/nginx/conf.d/default.conf

cd /gateway
./test &
exec nginx -g 'daemon off;'