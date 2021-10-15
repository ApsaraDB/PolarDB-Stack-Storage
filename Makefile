

build-agent:
	env GOOS=linux CGO_ENABLED=0 go build -o bin/sms-agent -ldflags '-s -w' -v cmd/agent/main.go

run-agent: build-agent
	ansible-playbook -vvvi scripts/agent.ini scripts/agent-sync.yml

clean:
	rm -rf bin/*