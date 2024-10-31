# KMS Plugin for Infisical

Enables [encryption at rest](https://kubernetes.io/docs/tasks/administer-cluster/encrypt-data/#providers) of your Kubernetes data in etcd using Infisical. With this plugin, you can use a key in Infisical for etcd encryption.

💡 **NOTE**: Currently, this KMS plugin only supports Kubernetes KMS v2.

## Getting Started

### Prerequisites
- Your Kubernetes cluster must use etcd v3 or later.
- Your Kubernetes cluster must be at least version 1.28.
- You will need to have an existing KMS key on Infisical. To create one, refer to the documentation [here](https://infisical.com/docs/documentation/platform/kms#key-management-service-kms).

### Authentication
There are multiple ways to authenticate the KMS plugin with Infisical:
- [GCP Auth](https://infisical.com/docs/documentation/platform/identities/gcp-auth): recommended if you're running the cluster on GCP
- [Azure Auth](https://infisical.com/docs/documentation/platform/identities/azure-auth): recommended if you're running the cluster on Azure
- [AWS Auth](https://infisical.com/docs/documentation/platform/identities/aws-auth): recommended if you're running the cluster on AWS
- [Universal Auth](https://infisical.com/docs/documentation/platform/identities/universal-auth): recommended if you're running the cluster on-premise.

## Installation
The recommended cluster installation method of the KMS plugin is via [static pods](https://kubernetes.io/docs/tasks/configure-pod-container/static-pod). Thus, you will have to configure the following steps for each of the control plane nodes of your cluster.

### Add the Infisical KMS plugin
Create the resource definition file for the Infisical KMS plugin. You can refer to the example [here](/). The following are supported flags:
| Flag | Default Value | Description |
|------|--------------|-------------|
| `--host-url` | `https://app.infisical.com` | URL of Infisical instance |
| `--listen-addr` | `/opt/infisicalkms.socket` | gRPC socket address for the plugin to listen on |
| `--kms-key` |  | Infisical KMS key ID (required) |
| `--ca-certificate` |  | SSL/TLS certificate for the Infisical instance |
| `--identity-id` | | Machine identity ID for authentication |
| `--ua-client-id` |  | Universal Auth client ID |
| `--ua-client-secret` |  | Universal Auth client secret |
| `--azure-resource` |  | Azure resource identifier |
| `--service-account-keyfile-path` |  | File path to the service account credentials |
| `--healthz-port` | `8787` | Port number for health check endpoint |
| `--healthz-path` | `/healthz` | URL path for health check endpoint |
| `--healthz-timeout` | `20s` | Timeout duration for health check RPC calls |





