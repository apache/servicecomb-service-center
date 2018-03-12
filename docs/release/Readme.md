## Service-Center Release

#### Release Notes
 - [Service-Center-1.0.0-m1 Release Notes](/docs/release/releaseNotes.md)
 

#### Running Apache Rat tool
This guide will help you to run the [Apache Rat](http://creadur.apache.org/rat/index.html) tool on service-center source code.
For running the tool please follow the below guidelines.

##### Step 1
Clone the Servcice-Center code and download Apache Rat tool.
```
git clone https://github.com/apache/incubator-servicecomb-service-center
```

```
wget http://mirrors.hust.edu.cn/apache//creadur/apache-rat-0.12/apache-rat-0.12-bin.tar.gz

# Untar the release
tar -xvf apache-rat-0.12-bin.tar.gz

# Copy the jar in the root directory
cp  apache-rat-0.12/apache-rat-0.12.jar ./
```
##### Step 2
Run the Rat tool using the below command

```
java -jar apache-rat-0.12.jar -a -d incubator-servicecomb-service-center/ -e *.md *.MD .gitignore .gitmodules .travis.yml manifest **vendor** **licenses**
```

Below is the list of the files which has been excluded from the list of RAT tool.
 - *.md  *.MD *.html:  Skip all the Readme and Documentation file like Api Docs.
 - .gitignore .gitmodules .travis.yml : Skip the git files and travis file.
 - manifest **vendor : Skip manifest and all the files under vendor.
 
You can access the latest RAT report [here](/docs/release/rat-report)  
 
 
