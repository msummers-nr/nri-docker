#build linux binary first
FROM newrelic/infrastructure

COPY configs/nri-docker-config.yml /etc/newrelic-infra/integrations.d/
COPY configs/nri-docker-def-linux.yml /var/db/newrelic-infra/custom-integrations/

#update binary path if required
COPY ./bin/linuxnri-docker /var/db/newrelic-infra/custom-integrations/