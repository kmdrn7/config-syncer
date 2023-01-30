# Config Syncer

Config Syncer is a utility to synchronize specified Kubernetes secrets between different namespaces.

## Features
* Syncs specified secrets in real-time between specified namespaces
* Supports one-way sync

## Usage
* Deploy the config-syncer deployment in your Kubernetes cluster.
* Configure the namespaces, secrets to be synced.
* Start the sync process by running the config-syncer command.

## Configuration
The configuration file is a YAML file with the following fields:

* `secrets`: A list of secrets to be synced, their source namespace, name, and destination namespaces and names.

Example configuration:

```yaml
secrets: 
  - namespace: common
    name: common-secret
    destinations:
      - namespace: default
        name: default-secret
  - namespace: dev
    name: dev-secret
    destinations:
      - namespace: prod
        name: prod-secret
```

## Development
To contribute to the development of config-syncer, follow these steps:

1. Fork the repository.
2. Clone the forked repository to your local machine.
3. Create a new branch for your changes.
4. Make the desired changes and commit them to the branch.
5. Push the branch to your forked repository.
6. Submit a pull request to the original repository.

## License
This project is licensed under the Apache 2.0 license.