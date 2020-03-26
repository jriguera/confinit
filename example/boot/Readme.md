# Raspbian-cloud Parameters

Please define the configuration settings for the Raspberry Pi in the
file `parameters.yml`. Those values will be used in the templates to
render the configuration files for each service.

The folders defined here contain those template files which will be
processed by https://github.com/jriguera/confinit at every boot. There
are two processing units:

1. At early boot (just after the local fs are mounted)
2. At the end of the startup process, before docker-compose and monit

Each unit has its own configuration file.


# Docker-compose

You can define the services to run at startup via `docker-compose.yml`
By default, only Portainer is defined, so you can use it
to manage your own containers.
