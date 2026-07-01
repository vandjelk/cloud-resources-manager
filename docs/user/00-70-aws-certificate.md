# AWS Certificate

Use the Cloud Manager module to import SSL/TLS certificates into AWS Certificate Manager.

The Cloud Manager module allows you to import SSL/TLS certificates from Kubernetes Secrets into AWS Certificate Manager (ACM), making them available for use with AWS Application Load Balancer.

## Cloud Providers

The AwsCertificate feature is specific to Amazon Web Services and uses [AWS Certificate Manager (ACM)](https://aws.amazon.com/certificate-manager/). This feature is only available when your Kyma cluster is running on AWS infrastructure.

You configure certificate imports using the AwsCertificate custom resource (CR). For more information, see [Certificate Resources](./resources/README.md#certificate-resources).

## Prerequisites

To import certificates into AWS Certificate Manager, the Cloud Manager module must be enabled in your AWS-based Kyma cluster.

The certificate data must be available in a Kubernetes Secret with the following keys:
* `tls.crt` - The certificate in PEM format (required)
* `tls.key` - The private key in PEM format (required)
* `ca.crt` - The certificate chain in PEM format (optional)

## Lifecycle

AwsCertificate is a cluster-level CR. Once you create an AwsCertificate resource, Cloud Manager imports the certificate referenced in the Secret into AWS Certificate Manager.

* AwsCertificate CR
  * AwsCertificate is a cluster-level CR.
  * References a Kubernetes Secret containing the certificate data.
  * The certificate is imported into AWS ACM and kept in sync with the Secret.

* Certificate Import
  * When an AwsCertificate CR is created, Cloud Manager reads the certificate data from the referenced Secret.
  * The certificate is imported into AWS ACM in the region where your Kyma cluster is running.
  * The ARN of the imported certificate is stored in the AwsCertificate status.
  * If the Secret is updated, Cloud Manager reimports the certificate to ACM.

* Status Information
  * The AwsCertificate status includes the ACM certificate ARN.
  * The expiration date of the certificate is tracked and displayed in the status.
  * Certificate state and conditions reflect the import status.

## Related Information

* [Using AwsCertificate Custom Resources](./tutorials/01-70-10-aws-certificate.md)
* [Cloud Manager Resources: Certificate](./resources/README.md#certificate-resources)
