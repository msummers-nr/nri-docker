#!/bin/bash

cp ./linux/nri-docker-config.yml /etc/newrelic-infra/integrations.d/
cp ./linux/nri-docker-definition.yml /var/db/newrelic-infra/custom-integrations/
cp ./linux/nri-docker /var/db/newrelic-infra/custom-integrations/

service newrelic-infra stop
service newrelic-infra start