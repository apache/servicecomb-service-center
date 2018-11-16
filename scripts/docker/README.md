### Docker Images for service-center and front-end

For Making docker images for Service-Center you can refer to this [doc](/scripts/docker/build-image)

For Making the Front-end images you can refer to this [doc](/scripts/docker/build-frontend-image)

After Making the images, you can run the following command at the `PROJECT_ROOT` to startup all

```bash
# the directory of docker-compose.yml file
cd examples/infrastructures/docker
docker-compose up
```