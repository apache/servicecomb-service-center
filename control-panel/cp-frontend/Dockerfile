FROM node:alpine as builder
WORKDIR /app
COPY ./package.json ./package-lock.json ./
RUN npm install
COPY . .
RUN npm run-script ng build -- --prod && npm run-script gzipper c ./dist
FROM nginx:alpine as runner
COPY --from=builder /app/dist/cp-frontend /usr/share/nginx/html
COPY ./nginx-backend.conf /etc/nginx/templates/default.conf.template
