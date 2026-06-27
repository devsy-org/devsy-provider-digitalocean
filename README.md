# DigitalOcean Provider for Devsy

## Getting started

The provider is available for auto-installation using

```sh
devsy provider add digitalocean
devsy provider use digitalocean
```

Follow the on-screen instructions to complete the setup.

Needed variables will be:

- TOKEN

The provider will inherit `DIGITALOCEAN_TOKEN` or `DIGITALOCEAN_ACCESS_TOKEN`
from the environment, or you can supply `TOKEN` directly via provider options.

### Creating your first devsy env with digitalocean

After the initial setup, just use:

```sh
devsy workspace up .
```

You'll need to wait for the machine and environment setup.

### Customize the VM Instance

This provider has the following options

| NAME         | REQUIRED | DESCRIPTION                                | DEFAULT      |
| ------------ | -------- | ------------------------------------------ | ------------ |
| TOKEN        | true     | The DigitalOcean API token to use.         |              |
| REGION       | true     | The DigitalOcean region (e.g. fra1).       | fra1         |
| DISK_SIZE    | false    | The disk size in GB.                       | 30           |
| DISK_IMAGE   | false    | The disk image to use.                     | docker-20-04 |
| MACHINE_TYPE | false    | The Droplet size slug.                     | s-4vcpu-8gb  |

Options can either be set in `env` or on the command line, for example:

```sh
devsy provider set-options -o MACHINE_TYPE=s-8vcpu-16gb
```

## Local Development

To build and test the provider locally, use [task](https://taskfile.dev/)
`task build:provider:dev`. The provider file is created in `./dist/provider.yaml`.
