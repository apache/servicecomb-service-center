# PR raising Guide

### Steps

If you want to raise a PR in this repo then you can follow the below guidelines to avoid conflicts.

1. Make your changes in your local code.
2. Once your changes are done the clone the code from ServiceComb
```
git clone http://github.com/apache/servicecomb-service-center.git
cd service-center
git remote add fork http://github.com/{YOURFORKNAME}/service-center.git
git checkout -b {YOURFEATURENAME}

#Merge your local changes in this branch.

#Once your changes are done then Push the changes to your fork

git add -A

git commit -m "{JIRA-ID YOURCOMMITMESSAGE}"

git push fork {YOURFEATURENAME}
```
3. Now go to github and browse to your branch and raise a PR from that branch.

