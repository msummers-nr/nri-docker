FROM newrelic/infrastructure

COPY ./nri-docker-config.yml /etc/newrelic-infra/integrations.d/
COPY ./nri-docker-def-nix.yml /var/db/newrelic-infra/custom-integrations/
COPY ./bin/nri-docker /var/db/newrelic-infra/custom-integrations/