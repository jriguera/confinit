# Confinit Parameters

Folders defined here contain those template files which will be
processed by https://github.com/jriguera/confinit at every boot.

The contents of this folder are supposed to be in `\boot\config` in order
to be handled by `confinit-boot@boot-config.service` and
`confinit-final@boot-config.service`.

Please define the configuration settings in the file `parameters.yml`.
Those values will be used in the templates to render the configuration files
for each service.

There are two processing units:

1. At early boot (just after the local fs are mounted) with actions defined in 
`.confinit-boot.yml` handled with `confinit-boot@boot-config.service`

2. At the end of the startup process, defined in `.confinit-final.yml` and
handled with `confinit-final@boot-config.service`

All parameters for the templates of both units are defined in the same
`parameters.yml`

# Folders

* `etc` templates rendered in `/etc` to configure system services
* `scripts/boot` render these scripts and execute them in early boot
* `scripts/final` render these scripts and execute them at the end of the boot
* `local` copy everything to `/usr/local`.

