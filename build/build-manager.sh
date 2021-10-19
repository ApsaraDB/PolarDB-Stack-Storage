docker build -t polardb/polardb-sms-manager:1.0.0 \
  --build-arg CodeSource=$(git remote -v | head -n 1 | awk '{print $2}') \
  --build-arg CodeBranch=$(git branch --show-current) \
  --build-arg CodeVersion=$(git rev-parse HEAD) \
  --build-arg BuildDate=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
  --build-arg ssh_prv_key="$(cat ~/.ssh/id_rsa)" \
  --build-arg ssh_pub_key="$(cat ~/.ssh/id_rsa.pub)" \
  . -f build/manager/Dockerfile
