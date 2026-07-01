# AwsCertificate Custom Resource

The `awscertificate.cloud-resources.kyma-project.io` is a cluster-scoped custom resource (CR).
It imports SSL/TLS certificates from Kubernetes Secrets into AWS Certificate Manager (ACM).
The imported certificate becomes available for use with AWS services such as Application Load Balancers, API Gateway, and CloudFront.

The AwsCertificate CR references a Kubernetes Secret that must contain the certificate and private key in PEM format.
Optionally, you can include a certificate chain (intermediate certificates) to establish a complete chain of trust.

When creating an AwsCertificate, only the `secretRef` field is mandatory.
It specifies the name of the Secret containing the certificate data.

## Certificate Data Requirements

The referenced Secret must contain the following keys:

* **tls.crt** (required): The X.509 certificate in PEM-encoded format
* **tls.key** (required): The private key in PEM-encoded format
* **ca.crt** (optional): The certificate chain (intermediate and root certificates) in PEM-encoded format

You can create such a Secret using the standard `kubectl create secret tls` command or by manually defining a Secret with these keys.

## Certificate Lifecycle

When you create an AwsCertificate CR, Cloud Manager imports the certificate into AWS ACM. The certificate ARN and expiration date are stored in the CR status.

If you update the referenced Secret (for example, to renew a certificate), Cloud Manager detects the change and reimports the certificate to ACM. The ARN in the status remains the same if the certificate is successfully reimported.

When you delete an AwsCertificate CR, Cloud Manager removes the certificate from AWS ACM unless it is still in use by other AWS resources (such as load balancers). In that case, the deletion is blocked until all dependencies are removed.

## Specification

This table lists the parameters of AwsCertificate, together with their descriptions:

| Parameter          | Type   | Description                                                                                                      |
| ------------------ | ------ | ---------------------------------------------------------------------------------------------------------------- |
| **secretRef**      | object | Required. Reference to a Secret containing certificate data.                                                     |
| **secretRef.name** | string | Required. Name of the Secret containing `tls.crt`, `tls.key`, and optionally `ca.crt` keys.                     |

## Status Fields

The AwsCertificate status includes the following fields:

| Field              | Type   | Description                                                                                                      |
| ------------------ | ------ | ---------------------------------------------------------------------------------------------------------------- |
| **arn**            | string | The ARN of the imported certificate in AWS Certificate Manager.                                                  |
| **expirationDate** | time   | The expiration date of the certificate as reported by ACM.                                                       |
| **state**          | string | Current state of the certificate. Possible values: `Processing`, `Ready`, `Deleting`, `Error`.                   |
| **conditions**     | array  | Standard Kubernetes conditions array. Includes a `Ready` condition with detailed status information.             |

## Sample Custom Resource

### Minimal Configuration

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AwsCertificate
metadata:
  name: my-app-certificate
spec:
  secretRef:
    name: my-app-tls-secret
```

### Complete Example with Certificate Chain

First, create a Secret with the certificate data:

```bash
kubectl create secret generic my-app-tls-secret \
  --from-file=tls.crt=./certificate.crt \
  --from-file=tls.key=./private.key \
  --from-file=ca.crt=./ca-bundle.crt \
  -n kyma-system
```

Then, create the AwsCertificate CR:

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AwsCertificate
metadata:
  name: my-app-certificate-with-chain
spec:
  secretRef:
    name: my-app-tls-secret
```

After the certificate is imported, you can view its status:

```bash
kubectl get awscertificate my-app-certificate-with-chain
```

Expected output:

```console
NAME                           STATE   ARN                                                                      EXPIRATION
my-app-certificate-with-chain  Ready   arn:aws:acm:us-east-1:123456789012:certificate/abc123def-4567-890a-...  2025-12-31T23:59:59Z
```
