# Adding configuration files

Since we have different environments we need to ensure that the configuration files loaded for our services are correct for each environment.

We currently have `dev` `testnet` and `prod` as environments.

If you require configurations to be amended for the deployments make sure you update the relevant repositories.

There are 2 different repositories.

`dev` and `testnet` configuration files are located in here.
https://gitlab.com/stripsdev/rabbit-dex-config

`prod` configuration files are located here
https://gitlab.com/stripsdev/rabbit-dex-config-prod

The `Makefile` will automatically take care the loading of the correct configuration from gitlab repositories when a `build` instruction is issued.
