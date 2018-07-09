### Guide to build the Docker Image

###### Step 1: 
Clone the Dependency
```
glide install
```

###### Step 2:
Make the release

```
bash -x scripts/release/make_release.sh linux 1.0.0 latest
```

###### Step 3:
Make the Docker Image
```
docker build -t service-center:dev .
```
