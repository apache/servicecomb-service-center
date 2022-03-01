# Service-Center Release

#### Release Notes

- [Service-Center-2.0.0 Release Notes](releaseNotes-2.0.0.md)
- [Service-Center-1.1.0 Release Notes](releaseNotes-1.1.0.md)
- [Service-Center-1.0.0 Release Notes](releaseNotes-1.0.0.md)
- [Service-Center-1.0.0-m1 Release Notes](releaseNotes-1.0.0-m1.md)
- [Service-Center-1.0.0-m2 Release Notes](releaseNotes-1.0.0-m2.md)

#### Running Apache Rat tool

This guide will help you to run the [Apache Rat](http://creadur.apache.org/rat/index.html) tool on service-center source
code. For running the tool please follow the below guidelines.

##### Step 1

Clone the Servcice-Center code and download Apache Rat tool.

```
git clone https://github.com/apache/servicecomb-service-center
```

```
wget http://mirrors.tuna.tsinghua.edu.cn/apache/creadur/apache-rat-0.13/apache-rat-0.13-bin.tar.gz

# Untar the release
tar -xvf apache-rat-0.13-bin.tar.gz

# Copy the jar in the root directory
cp  apache-rat-0.13/apache-rat-0.13.jar ./
```

##### Step 2

Run the Rat tool using the below command

```
java -jar apache-rat-0.13.jar -a -d servicecomb-service-center/ -e '(.+(\.svg|\.md|\.MD|\.cer|\.tpl|\.json|\.yaml|\.proto|\.pb.go))|(.gitignore|.gitmodules|ux|docs|vendor|licenses|bower.json|cert_pwd|glide.yaml|go.mod|go.sum)'
```

Below is the list of the files which has been excluded from the list of RAT tool.

- *.md  *.MD *.html:  Skip all the Readme and Documentation file like Api Docs.
- .gitignore .gitmodules .travis.yml : Skip the git files and travis file.
- manifest **vendor : Skip manifest and all the files under vendor.
- bower.json :  Skip bower installation file
- cert_pwd server.cer trust.cer :  Skip ssl files
- *.tpl : Ignore template files
- glide.yaml go.mod go.sum : Skip dependency config files
- docs : Skip document files
- .yaml : Skip configuration files
- ux : Skip foreground files
- .proto .pb.go : Skip proto files

You can access the latest RAT report [here](rat-report)  

#### Make a release

See [here](https://github.com/apache/servicecomb-service-center/blob/master/scripts/release/README.md)

#### Archive

##### Step 1

> If you are doing release for the first time, you can read this [document](https://doris.apache.org/branch-0.14/zh-CN/community/release-process.html#%E5%87%86%E5%A4%87%E7%8E%AF%E5%A2%83).

Execute script, archive source code and generate summary and signature
```
bash scripts/release/archive.sh apache-servicecomb-service-center 2.0.0 littlecui@apache.org
```

list current directory

```
-rw-rw-r--  1 ubuntu ubuntu 3.1M Jun  8 20:35 apache-servicecomb-service-center-2.0.0-src.tar.gz
-rw-rw-r--  1 ubuntu ubuntu  862 Jun  8 20:35 apache-servicecomb-service-center-2.0.0-src.tar.gz.asc
-rw-rw-r--  1 ubuntu ubuntu  181 Jun  8 20:35 apache-servicecomb-service-center-2.0.0-src.tar.gz.sha512
```

##### Step 2

PUSH to apache dev repo

```
svn co https://dist.apache.org/repos/dist/dev/servicecomb/
cd servicecomb/
mkdir -p 2.0.0
cp apache-servicecomb-service-center-* 2.0.0/
svn add .
svn ci --username xxx --password xxx -m "Add the Service-Center 2.0.0 version"
```
