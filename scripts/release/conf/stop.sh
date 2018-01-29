kill -9 $(ps aux | grep 'incubator-servicecomb-service-center' | awk '{print $2}')

