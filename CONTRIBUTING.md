## Contribution guide for Service-Center

Thanks everyone for contributing to [Service-Center](https://github.com/apache/servicecomb-service-center).

This document explains the process and best practices for submitting a Pull Request to the Service-Center project. This document can be treated as a reference for all contributors and be useful to new and infrequent submitters.

#### Cloning the repo and put it in $GOPATH

Clone the repo in a proper GOPATH

```
git clone https://github.com/apache/servicecomb-service-center.git $GOPATH/src/github.com/apache/servicecomb-service-center
cd $GOPATH/src/github.com/apache/servicecomb-service-center
```

#### Download the Dependencies

We use glide for dependency management, please follow the below steps for downloading all the dependencies required for building this project.

```
curl https://glide.sh/get | sh
glide install
```

#### Make your Changes

If this is a bug or a small fix then you can directly make the changes and ensure all the steps in this documentation and raise a PR, but If it is a feature or a big design or architecture change then we recommend you to raise an issue [here](https://github.com/apache/servicecomb-service-center/issues) or discuss the same in our [mailing list](https://groups.google.com/forum/#!forum/servicecomb-developers).

#### Compile and running Test locally.

Once you are done with your changes then please follow the below checks to ensure the code quality before raising a PR.
```
go fmt ./...

go build -o service-center
```

Running UT in local env, this step assumes you have a docker running in your env.
```
bash -x scripts/ut_test_in_docker.sh 
```

Once UT has passed you can run the integration test to ensure the overall functionality is not altered.
```
bash -x scripts/integration_test.sh
```

#### Pushing the Code and Raising PR

Once you are done with compiling, UT and IT then you are good to go for raising the PR, please follow these [guidelines](https://help.github.com/articles/creating-a-pull-request/) for raising the PR.
