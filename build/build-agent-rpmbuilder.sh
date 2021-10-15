docker build -t reg.docker.alibaba-inc.com/polarbox/polardb-sms-rpmbuilder:1.0.0 \
  . -f build/builder/Dockerfile
docker push reg.docker.alibaba-inc.com/polarbox/polardb-sms-rpmbuilder:1.0.0
