# Use RWX Volumes in AWS

This tutorial explains how to create VPC peering in Azure.

## Steps <!-- {docsify-ignore} -->

1. Login to Azure and set active subscription
   ```shell
   export SUBSCRIPTION=<Name or ID of subscription.>
   az login
   az account set --subscription $SUBSCRIPTION
   ```
2. Set the region that is closest to your Kyma cluster. Use `az account list-locations` to list available locations. 
   ```shell
      export REGION=<Location>
   ```
3. Create a resource group that will be a container for related resources.
   ```shell
   export RANDOM_ID="$(openssl rand -hex 3)"
   export RESOURCE_GROUP_NAME="myResourceGroup$RANDOM_ID"
   az group create --name $RESOURCE_GROUP_NAME --location $REGION
   ```
4. Create network
   ```shell
   export VNET_NAME="myVnet$RANDOM_ID"
   export ADDRESS_PREFIX=172.0.0.0/16
   export SUBNET_PREFIX=172.0.0.0/24
   export SUBNET_NAME="MySubnet"
   
   az network vnet create -g $RESOURCE_GROUP_NAME -n $VNET_NAME --address-prefix $ADDRESS_PREFIX --subnet-name $SUBNET_NAME --subnet-prefixes $SUBNET_PREFIX
   ```
5. Create the virtual machine this resource group
   ```shell
   export VM_NAME="myVM$RANDOM_ID"
   export USERNAME=azureuser
   export VM_IMAGE="Canonical:0001-com-ubuntu-minimal-jammy:minimal-22_04-lts-gen2:latest"
   
   az vm create \
   --resource-group $RESOURCE_GROUP_NAME \
   --name $VM_NAME \
   --image $VM_IMAGE \
   --vnet-name $VNET_NAME \
   --subnet "MySubnet" \
   --public-ip-address "" \
   --nsg "" 
   ```
   
6. Tag the VNet with the Kyma name
   ```shell
   export KYMA_NAME=<Your Kyma name.>
   export VNET_ID=$(az network vnet show --name $VNET_NAME --resource-group $RESOURCE_GROUP_NAME --query id --output tsv)
   az tag update --resource-id $VNET_ID --operation Merge --tags $KYMA_NAME
   ```


7. Create an AzureVpcPeering resource.

   ```shell
   kubectl apply -f - <<EOF
   apiVersion: cloud-resources.kyma-project.io/v1beta1
   kind: AzureVpcPeering
   metadata:
     name: peering-to-my-vnet
   spec:
     allowVnetAccess: true
     remotePeeringName: peering-to-my-kyma
     remoteResourceGroup: $RESOURCE_GROUP_NAME
     remoteVnet: $VNET_ID
   EOF
   ```

8. Wait for the AzureVpcPeering to be in the `Ready` state.

   ```shell
   kubectl wait --for=condition=Ready azurevpcpeering/peering-to-my-vnet --timeout=300s
   ```

   Once the newly created AzureVpcPeering is provisioned, you should see the following message:

   ```
   azurevpcpeering.cloud-resources.kyma-project.io/peering-to-my-vnet condition met
   ```

9. Create a namespace and export its value as an environment variable. Run:

   ```shell
   export NAMESPACE={NAMESPACE_NAME}
   kubectl create ns $NAMESPACE
   ```

10. Create a workload that pings the VM in the remote network.
   ```shell
   kubectl apply -n $NAMESPACE -f - <<EOF
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: azurevpcpeering-demo
   spec:
     selector:
       matchLabels:
         app: azurevpcpeering-demo
     template:
       metadata:
         labels:
           app: azurevpcpeering-demo
       spec:
         containers:
         - name: my-container
           resources:
             limits:
               memory: 512Mi
               cpu: "1"
             requests:
               memory: 256Mi
               cpu: "0.2"
           image: ubuntu
           command:
             - "/bin/bash"
             - "-c"
             - "--"
           args:
             - "apt update; apt install iputils-ping -y; ping -c 20 $IP_ADDRESS"
   EOF
   ```

   This workload should print a sequence of 20 echo replies to stdout.

11. Print the logs of one of the workloads, run:

   ```shell
   kubectl logs -n $NAMESPACE `kubectl get pod -n $NAMESPACE -l app=azurevpcpeering-demo -o=jsonpath='{.items[0].metadata.name}'`
   ```

   The command should print something like:
   ```
   ...
   PING 172.0.0.4 (172.0.0.4) 56(84) bytes of data.
   64 bytes from 172.0.0.4: icmp_seq=1 ttl=63 time=8.10 ms
   64 bytes from 172.0.0.4: icmp_seq=2 ttl=63 time=2.01 ms
   64 bytes from 172.0.0.4: icmp_seq=3 ttl=63 time=7.02 ms
   64 bytes from 172.0.0.4: icmp_seq=4 ttl=63 time=1.87 ms
   64 bytes from 172.0.0.4: icmp_seq=5 ttl=63 time=1.89 ms
   64 bytes from 172.0.0.4: icmp_seq=6 ttl=63 time=4.75 ms
   64 bytes from 172.0.0.4: icmp_seq=7 ttl=63 time=2.01 ms
   64 bytes from 172.0.0.4: icmp_seq=8 ttl=63 time=4.26 ms
   64 bytes from 172.0.0.4: icmp_seq=9 ttl=63 time=1.89 ms
   64 bytes from 172.0.0.4: icmp_seq=10 ttl=63 time=2.08 ms
   64 bytes from 172.0.0.4: icmp_seq=11 ttl=63 time=2.01 ms
   64 bytes from 172.0.0.4: icmp_seq=12 ttl=63 time=2.24 ms
   64 bytes from 172.0.0.4: icmp_seq=13 ttl=63 time=1.80 ms
   64 bytes from 172.0.0.4: icmp_seq=14 ttl=63 time=4.32 ms
   64 bytes from 172.0.0.4: icmp_seq=15 ttl=63 time=2.03 ms
   64 bytes from 172.0.0.4: icmp_seq=16 ttl=63 time=2.03 ms
   64 bytes from 172.0.0.4: icmp_seq=17 ttl=63 time=5.19 ms
   64 bytes from 172.0.0.4: icmp_seq=18 ttl=63 time=1.86 ms
   64 bytes from 172.0.0.4: icmp_seq=19 ttl=63 time=1.92 ms
   64 bytes from 172.0.0.4: icmp_seq=20 ttl=63 time=1.92 ms
   
   --- 172.0.0.4 ping statistics ---
   20 packets transmitted, 20 received, 0% packet loss, time 19024ms
   rtt min/avg/max/mdev = 1.800/3.060/8.096/1.847 ms
   ...
   ```

12. Clean up

   * Remove the created workloads:
     ```shell
     kubectl delete -n $NAMESPACE deployment azurevpcpeering-demo
     ```

   * Remove the created azurevpcpeering:
     ```shell
     kubectl delete -n $NAMESPACE azurevpcpeering peering-to-my-vnet
     ```

   * Remove the created namespace:
     ```shell
     kubectl delete namespace $NAMESPACE
     ```
   * Remove the created azure resource group
      ```shell
      az group delete --name $RESOURCE_GROUP_NAME --yes
      ```