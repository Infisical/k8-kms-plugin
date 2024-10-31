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

### 1. Add the Infisical KMS plugin
Create the appropriate resource definition file for the Infisical KMS plugin. You can refer to the file [here](/) as the starting point. The following are the supported flags:
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

💡 **NOTE**: Ensure that you have attached the volume mount for the path `/opt` in the plugin's resource definition.

Save the Infisical KMS plugin resource definition to the `/etc/kubernetes/manifests` directory on the control plane node. This will automatically create a static pod for the KMS plugin which you can confirm by listing the pods in the `kube-system` namespace.

### 2. Create an encryption configuration resource
Create a new encryption configuration file `/etc/kubernetes/enc/encryption-config.yaml` with the appropriate properties.
```yaml
apiVersion: apiserver.config.k8s.io/v1
kind: EncryptionConfiguration
resources:
  - resources:
      - secrets
    providers:
      - kms:
          apiVersion: v2
          name: infisical-kms-plugin
          endpoint: unix:///opt/infisicalkms.socket    # This should match the listen-addr declared in the Infisical KMS plugin's static pod definition
          timeout: 20s
      - identity: {}
```

### 3. Update the kube-apiserver resource definition
In the `etc/kubernetes/manifests` directory, open the `kube-apiserver.yaml` file. 

Update the `volumes` section so that it has the following:
```yaml
  volumes:
  ...
  - hostPath:
      path: /etc/kubernetes/enc
    name: enc
  - hostPath:
      path: /opt
    name: socket
```

Consequently, update the `volumeMounts` section of the `spec.container` property so that it uses the volumes in the preceding step.
```yaml
    volumeMounts:
    ...
    - mountPath: /etc/kubernetes/enc
      name: enc
      readOnly: true
    - mountPath: /opt
      name: socket
```
Then, update the `command` section of the `spec.container` property so that it includes the `encryption-provider-config` and the `encryption-provider-config-automatic-reload` flags.
```yaml
spec:
  containers:
  - command:
    - kube-apiserver
    - --advertise-address=192.168.49.2
    ...
    - --encryption-provider-config=/etc/kubernetes/enc/encryption-config.yaml
    - --encryption-provider-config-automatic-reload=true
```

## Verification
In order to verify that Infisical KMS encryption is working, we can do the following:

1. Create a new secret:

   ```bash
   kubectl create secret generic secret1 -n default --from-literal=mykey=mysecret
   ```

2. Using `etcdctl`, read the secret from etcd:

   ```bash
   sudo ETCDCTL_API=3 etcdctl --cacert=/etc/kubernetes/certs/ca.crt --cert=/etc/kubernetes/certs/etcdclient.crt --key=/etc/kubernetes/certs/etcdclient.key get /registry/secrets/default/secret1
   ```

3. Check that the stored secret is prefixed with `k8s:enc:kms:v2:infisical-kms-plugin`. This indicates that the secret is stored as encrypted after being processed by the Infisical KMS plugin.

4. To ensure that decryption works, fetch the secret using the following:

   ```bash
   kubectl get secrets secret1 -o yaml
   ```

   The output should match `mykey: bXlzZWNyZXQ=`, which is the encoded data of `mysecret`.

### 4. Restart your Kubernetes API server


