## Azure Vpc Peering


The `azurevpcpeering.cloud-resources.kyma-project.io` custom resource (CR) specifies the virtual network peering between 
Kyma and remote Azure Virtual Private Cloud (VPC) network. Virtual network peering is only possible withing the networks
of the same cloud provider.



Once an AzureVpcPeering CR is created and reconciled, the Cloud Manager controller first creates virtual network peering 
connection in the Virtual Private Cloud (VPC) network of the Kyma cluster in the underlying cloud provider and accepts
VPC peering connection in the remote cloud provider subscription.

You must authorize Cloud Manager service principal kyma-cloud-manager-ENV in remote cloud provider subscription to
accept VPC peering connection. Assign following IAM roles to Cloud Manager service principal in the remote subscription: 
* Classic Network Contributor
* Network Contributor

AzureVpcPeering can be deleted at any time but VPC peering connection in the remote subscription must be deleted
manually.

## Specification <!-- {docsify-ignore} -->


This table lists the parameters of the given resource together with their descriptions:

**Spec:**

| Parameter               | Type   | Description                                                                                                                                   |
|-------------------------|--------|-----------------------------------------------------------------------------------------------------------------------------------------------|
| **allowVnetAccess**     | bool   | Specifies whether the VMs in the local virtual network space would be able to access the VMs in remote virtual network space, and vice versa. |
| **remotePeeringName**   | string | Specifies the name of the VNet peering in the remote subscription.                                                                            |
| **remoteResourceGroup** | string | Specifies the name of the resource group in the remote subscription.                                                                          |
| **remoteVnet**          | string | Specifies the ID of the VNet in the remote subscription.                                                                                      |

**Status:**

| Parameter                         | Type       | Description                                                                                 |
|-----------------------------------|------------|---------------------------------------------------------------------------------------------|
| **id**                            | string     | Represents the VPC peering name on the Kyma cluster underlying cloud provider subscription. |
| **conditions**                    | \[\]object | Represents the current state of the CR's conditions.                                        |
| **conditions.lastTransitionTime** | string     | Defines the date of the last condition status change.                                       |
| **conditions.message**            | string     | Provides more details about the condition status change.                                    |
| **conditions.reason**             | string     | Defines the reason for the condition status change.                                         |
| **conditions.status** (required)  | string     | Represents the status of the condition. The value is either `True`, `False`, or `Unknown`.  |
| **conditions.type**               | string     | Provides a short description of the condition.                                              |


## Sample Custom Resource <!-- {docsify-ignore} -->

See an exemplary AzureVpcPeering custom resource:

```yaml
apiVersion: cloud-resources.kyma-project.io/v1beta1
kind: AzureVpcPeering
metadata:
  name: peering-to-my-vnet
spec:
  allowVnetAccess: true
  remotePeeringName: peering-to-my-kyma
  remoteResourceGroup: MyResourceGroup
  remoteVnet: /subscriptions/afdbc79f-de19-4df4-94cd-6be2739dc0e0/resourceGroups/WestEurope/providers/Microsoft.Network/virtualNetworks/MyVnet
```