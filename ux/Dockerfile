FROM nginx:stable-alpine

COPY ./dist/ux /usr/share/nginx/html
COPY ./etc/nginx/nginx.conf /etc/nginx/nginx.conf
COPY ./etc/nginx/server.conf /etc/nginx/conf.d/default.conf

EXPOSE 4200
