docker build -t reg.docker.alibaba-inc.com/polarbox/polardb-sms-agent:1.0.0 \
  . -f build/agent/Dockerfile
docker push reg.docker.alibaba-inc.com/polarbox/polardb-sms-agent:1.0.0
