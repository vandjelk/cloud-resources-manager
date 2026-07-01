# Using AwsCertificate Custom Resources

The Cloud Manager module offers an AwsCertificate Custom Resource Definition (CRD). When you apply an AwsCertificate custom resource (CR), it imports your SSL/TLS certificate from a Kubernetes Secret into AWS Certificate Manager (ACM), making it available for use with AWS services.

## Prerequisites  <!-- {docsify-ignore} -->

* You have the Cloud Manager module added.
* You have a valid SSL/TLS certificate and private key in PEM format.
* Optionally, you have a certificate chain (intermediate certificates) for establishing a complete chain of trust.

## Steps  <!-- {docsify-ignore} -->

### Minimal Setup

To import a basic certificate without a certificate chain, follow these steps:

1. Create a Kubernetes Secret with your certificate data.

   > [!NOTE]
   > Replace `./certificate.crt` and `./private.key` with the paths to your actual certificate and private key files.

   ```bash
   kubectl create secret tls my-tls-secret \
     --cert=./certificate.crt \
     --key=./private.key
   ```

2. Create an AwsCertificate CR that references the Secret.

   > [!NOTE]
   > The import operation typically completes within 1-2 minutes.

   ```bash
   kubectl apply -f - <<EOF
   apiVersion: cloud-resources.kyma-project.io/v1beta1
   kind: AwsCertificate
   metadata:
     name: my-certificate
   spec:
     secretRef:
       name: my-tls-secret
   EOF
   ```

3. Wait for the AwsCertificate to reach the `Ready` state.

   ```bash
   kubectl wait --for=condition=Ready awscertificate/my-certificate --timeout=5m
   ```

4. Verify that the certificate was imported successfully.

   ```bash
   kubectl get awscertificate my-certificate
   ```

   Expected output:

   ```console
   NAME             STATE   ARN                                                                      EXPIRATION
   my-certificate   Ready   arn:aws:acm:us-east-1:123456789012:certificate/abc123def-4567-890a-...  2025-12-31T23:59:59Z
   ```

   The ARN shown in the output can now be used to configure AWS services such as Application Load Balancers.

### Advanced Setup with Certificate Chain

To import a certificate with a complete certificate chain (including intermediate certificates), follow these steps:

1. Create a Kubernetes Secret with your certificate, private key, and certificate chain.

   > [!NOTE]
   > The certificate chain file should contain all intermediate and root certificates concatenated in PEM format.

   ```bash
   kubectl create secret generic my-tls-secret-with-chain \
     --from-file=tls.crt=./certificate.crt \
     --from-file=tls.key=./private.key \
     --from-file=ca.crt=./ca-bundle.crt
   ```

2. Create an AwsCertificate CR that references the Secret.

   > [!NOTE]
   > The import operation typically completes within 1-2 minutes.

   ```bash
   kubectl apply -f - <<EOF
   apiVersion: cloud-resources.kyma-project.io/v1beta1
   kind: AwsCertificate
   metadata:
     name: my-certificate-with-chain
   spec:
     secretRef:
       name: my-tls-secret-with-chain
   EOF
   ```

3. Wait for the AwsCertificate to reach the `Ready` state.

   ```bash
   kubectl wait --for=condition=Ready awscertificate/my-certificate-with-chain --timeout=5m
   ```

4. Inspect the certificate details including the expiration date.

   ```bash
   kubectl get awscertificate my-certificate-with-chain -o yaml
   ```

   The status section shows the ARN and expiration date:

   ```yaml
   status:
     arn: arn:aws:acm:us-east-1:123456789012:certificate/abc123def-4567-890a-bcde-123456789012
     expirationDate: "2025-12-31T23:59:59Z"
     state: Ready
     conditions:
     - type: Ready
       status: "True"
       reason: Ready
       message: Certificate imported successfully
   ```

### Renewing a Certificate

When your certificate needs to be renewed:

1. Update the Secret with the new certificate data.

   ```bash
   kubectl create secret tls my-tls-secret \
     --cert=./new-certificate.crt \
     --key=./new-private.key \
     --dry-run=client -o yaml | kubectl apply -f -
   ```

2. Cloud Manager detects the Secret change and automatically reimports the certificate to ACM.

   > [!NOTE]
   > The reimport operation typically completes within 1-2 minutes. The ARN remains the same.

3. Verify that the certificate was updated.

   ```bash
   kubectl get awscertificate my-certificate -o jsonpath='{.status.expirationDate}'
   ```

   The expiration date should reflect the new certificate's validity period.

### Deleting a Certificate

To remove a certificate from AWS Certificate Manager:

1. Delete the AwsCertificate CR.

   ```bash
   kubectl delete awscertificate my-certificate
   ```

   > [!NOTE]
   > If the certificate is still in use by AWS resources (such as a load balancer), the deletion will be blocked. You must first remove the certificate from all AWS resources before the AwsCertificate can be deleted.

2. Optionally, delete the Secret if it's no longer needed.

   ```bash
   kubectl delete secret my-tls-secret -n kyma-system
   ```
