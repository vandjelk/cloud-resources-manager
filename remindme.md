### Common
- DNSEntry to point to Azure Application Gateway/AWS ALB

### Azure

- Dedicated subnet for Azure Application Gateway
- Public IP Address
- Azure Application Gateway with Integrated WAF
- Optional private DNS management

### AWS

- ALB enabled in shoot.spec
- Allow Istio ingress gateway to be installed as NodePort
- WebACL


### GPC
- problems with Gardener loadbalancer implamentation