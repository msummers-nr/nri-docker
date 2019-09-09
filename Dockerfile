#build linux binary first
FROM newrelic/infrastructure:1.5.0
ENV NRIA_IS_FORWARD_ONLY true

COPY configs/nri-docker-config.yml /etc/newrelic-infra/integrations.d/
COPY configs/nri-docker-def-linux.yml /var/db/newrelic-infra/custom-integrations/

#update binary path if required
COPY ./bin/linux/nri-docker /var/db/newrelic-infra/custom-integrations/