<a href="https://opensource.newrelic.com/oss-category/#new-relic-experimental"><picture><source media="(prefers-color-scheme: dark)" srcset="https://github.com/newrelic/opensource-website/raw/main/src/images/categories/dark/Experimental.png"><source media="(prefers-color-scheme: light)" srcset="https://github.com/newrelic/opensource-website/raw/main/src/images/categories/Experimental.png"><img alt="New Relic Open Source experimental project banner." src="https://github.com/newrelic/opensource-website/raw/main/src/images/categories/Experimental.png"></picture></a>

# NRI - Docker

#### Enhanced Docker Monitoring
#### Uses V3 Golang Infra SDK & Docker Golang SDK

- Collects a large range of additional metrics from stats, inspect, service, task, node etc.
- Supports Docker on Windows, and Linux
- Supports any orchestrator including Docker Swarm
- Able to run within a container or on host
- Detect and set NEW_RELIC_APP_NAME environment variable if available
- Add custom attributes via NRDI_* environment variables

<!-- <img src="./images/ss1.png" alt="ss1"> -->

### Setup
- Download latest compiled release [here](https://source.datanerd.us/FIT/nri-docker/releases/) (Do not download the source code!)
- Extract package
- Run included installer script for your desired OS `install_linux.sh` or `install_win.bat`;
- Else follow container setup below

### Containerized
```
docker build -t nri-docker .

docker run -d --name nri-docker --network=host --cap-add=SYS_PTRACE -v "/:/host:ro" -v "/var/run/docker.sock:/var/run/docker.sock" -e NRIA_LICENSE_KEY="newrelicInfrastructureKEY" nri-docker:latest

```

### Decorating With Additional Custom Attributes via Environment Variables
```
Multiple custom attributes can be set via NRDI_ attributes (New Relic Docker Integration) eg.
NRDI_OWNER=nr.expert.services
NRDI_TEAM=loud
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
- - Ensure/Update "os" to "windows" and check that under command section it is docker-ohi.exe && NOT "docker-ohi" without the extension
- docker-ohi.exe

Program Files/newrelic-infra/integrations.d/
- nri-docker-config.yml
```

### Compiling
```
Set GOOS as required before compilation and run 'make'
```
Tests will fail if compiling for different platform.

## Support

New Relic hosts and moderates an online forum where customers can interact with New Relic employees as well as other customers to get help and share best practices. Like all official New Relic open source projects, there's a related Community topic in the New Relic Explorers Hub. You can find this project's topic/threads here:

## Contributing
We encourage your contributions to improve `nri-docker`! Keep in mind when you submit your pull request, you'll need to sign the CLA via the click-through using CLA-Assistant. You only have to sign the CLA one time per project.
If you have any questions, or to execute our corporate CLA, required if your contribution is on behalf of a company,  please drop us an email at opensource@newrelic.com.

**A note about vulnerabilities**

As noted in our [security policy](../../security/policy), New Relic is committed to the privacy and security of our customers and their data. We believe that providing coordinated disclosure by security researchers and engaging with the security community are important means to achieve our security goals.

If you believe you have found a security vulnerability in this project or any of New Relic's products or websites, we welcome and greatly appreciate you reporting it to New Relic through [HackerOne](https://hackerone.com/newrelic).

## License
`nri-docker` is licensed under the [Apache 2.0](http://apache.org/licenses/LICENSE-2.0.txt) License.