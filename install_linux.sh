#!/bin/bash

cp ./linux/docker-ohi-config.yml /etc/newrelic-infra/integrations.d/
cp ./linux/docker-ohi-definition.yml /var/db/newrelic-infra/custom-integrations/
cp ./linux/docker-config.yml /var/db/newrelic-infra/custom-integrations/
cp ./linux/docker-ohi /var/db/newrelic-infra/custom-integrations/

service newrelic-infra stop
service newrelic-infra start