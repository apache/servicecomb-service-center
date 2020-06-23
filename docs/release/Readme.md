# Service-Center Release

#### Release Notes
 - [Service-Center-1.1.0 Release Notes](releaseNotes-1.1.0.md)
 - [Service-Center-1.0.0 Release Notes](releaseNotes-1.0.0.md)
 - [Service-Center-1.0.0-m1 Release Notes](releaseNotes-1.0.0-m1.md)
 - [Service-Center-1.0.0-m2 Release Notes](releaseNotes-1.0.0-m2.md)
 

#### Running Apache Rat tool
This guide will help you to run the [Apache Rat](http://creadur.apache.org/rat/index.html) tool on service-center source code.
For running the tool please follow the below guidelines.

##### Step 1
Clone the Servcice-Center code and download Apache Rat tool.
```
git clone https://github.com/apache/servicecomb-service-center
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
java -jar apache-rat-0.12.jar -a -d servicecomb-service-center/ -e *.md *.MD .gitignore .gitmodules .travis.yml manifest **vendor** **licenses** bower.json cert_pwd *.cer *.tpl glide.yaml go.mod go.sum
```

Below is the list of the files which has been excluded from the list of RAT tool.
 - *.md  *.MD *.html:  Skip all the Readme and Documentation file like Api Docs.
 - .gitignore .gitmodules .travis.yml : Skip the git files and travis file.
 - manifest **vendor : Skip manifest and all the files under vendor.
 - bower.json :  Skip bower installation file
 - cert_pwd server.cer trust.cer :  Skip ssl files
 - *.tpl : Ignore template files
 - glide.yaml go.mod go.sum : Skip dependency config files 
You can access the latest RAT report [here](/rat-report)  
 
 
