# NRI - Docker

#### Enhanced Docker Monitoring
#### Uses V3 Golang Infra SDK & Docker Golang SDK

- Collects a large range of additional metrics from stats, inspect, service, task, node etc.
- Supports Docker on Windows, and Linux
- Supports any orchestrator including Docker Swarm
- Able to run within a container or on host

<!-- <img src="./images/ss1.png" alt="ss1"> -->

### Containerized
```
docker build -t nri-docker .

docker run -d --name nri-docker --network=host --cap-add=SYS_PTRACE -v "/:/host:ro" -v "/var/run/docker.sock:/var/run/docker.sock" -e NRIA_LICENSE_KEY="newrelicInfrastructureKEY" nri-docker:latest

```

### Linux

Download the latest release, and run install_linux.sh with Administrative permissions.
Run chmod +x install_linux.sh on the file incase of any permission issues.

Else, modify install_linux.sh script with paths to suit or copy files as below instructions:

```
Copy files into following locations:
cp ./nri-docker-config.yml /etc/newrelic-infra/integrations.d/
cp ./nri-docker-def-nix.yml /var/db/newrelic-infra/custom-integrations/
- - Ensure/Update "os" to "linux" and check that under command section it is docker-ohi && NOT docker-ohi.exe

cp ./bin/docker-ohi /var/db/newrelic-infra/custom-integrations/
```

### Windows Install

Download the latest release, and run install_win.bat with Administrative permissions.
Else, modify install_win.bat script with paths to suit or copy files as below instructions:

```
Copy files into following locations:

Program Files/newrelic-infra/custom-integrations/
- nri-docker-def-win.yml 
- - Ensure/Update "os" to "linux" and check that under command section it is docker-ohi.exe && NOT "docker-ohi" without the extension
- docker-ohi.exe

Program Files/newrelic-infra/integrations.d/
- nri-docker-config.yml
```

### Compiling
```
Set GOOS as required before compilation and run 'make'
```
Tests will fail if compiling for different platform.