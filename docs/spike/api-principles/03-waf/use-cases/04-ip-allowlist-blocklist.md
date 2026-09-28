# Use Case 5: IP Allowlist and Blocklist

## User Story

"As a platform operator, I want to block known malicious IP addresses and ranges from accessing my application, while explicitly allowing trusted internal networks and partner IPs, to reduce attack surface and prevent abuse."

## Requirements

1. Block specific IP addresses or CIDR ranges (blocklist)
2. Allow only specific IP addresses or CIDR ranges (allowlist)
3. Combine IP rules with path conditions (e.g., block IPs only on /admin)
4. Priority ordering (allowlist before blocklist)
5. Support for both IPv4 and IPv6

## Real-World Scenarios

### Scenario A: Block Known Malicious IPs
```yaml
# Block IPs from threat intelligence feeds
# Block entire countries with high attack rates
blocklist:
  - 203.0.113.0/24      # Known botnet range
  - 198.51.100.42       # Specific attacker IP
  - 192.0.2.0/24        # Tor exit nodes
```

### Scenario B: Admin Panel IP Allowlist
```yaml
# Only allow admin access from office and VPN
path: /admin
allowlist:
  - 10.0.0.0/8          # Internal network
  - 172.16.0.0/12       # Office VPN
  - 203.0.113.50        # Admin home IP
```

### Scenario C: Partner API Access Control
```yaml
# Different partners have access to different APIs
/api/partner-a:
  allow: 198.51.100.0/24
/api/partner-b:
  allow: 203.0.113.0/24
/api/internal:
  allow: 10.0.0.0/8
```

### Scenario D: Defense in Depth
```yaml
# Priority order:
1. Allow internal network (10.0.0.0/8) everywhere
2. Block known malicious IPs everywhere
3. Allow specific partner IPs on /api/partner
4. Block all other IPs on /admin
5. Allow everyone else on public paths
```

## API Design

IP allowlist/blocklist rules go in a `WafPolicy` ConfigMap — the user provides a complete provider policy. See [implementation-examples.md](../implementation-examples.md) Use Case 3 for ConfigMap content per provider.

## Provider Capability Check

### AWS WAFv2

**Status:** ✅ **Perfect Support**

**How it works:**
- **IPSet resources**: Create reusable IP sets, reference in rules
- **IPSetReferenceStatement**: Match against IP set
- **Supports IPv4 and IPv6**
- **Up to 10,000 IP addresses per IP set**
- **Can update IP sets independently from WebACL**

**Native AWS Translation:**
```json
{
  "IPSets": [
    {
      "Name": "InternalNetwork",
      "Scope": "REGIONAL",
      "IPAddressVersion": "IPV4",
      "Addresses": [
        "10.0.0.0/8",
        "172.16.0.0/12"
      ]
    },
    {
      "Name": "KnownThreats",
      "Scope": "REGIONAL",
      "IPAddressVersion": "IPV4",
      "Addresses": [
        "203.0.113.0/24",
        "192.0.2.0/24",
        "198.51.100.42/32"
      ]
    }
  ],
  "WebACL": {
    "Name": "ip-access-control-webacl",
    "Scope": "REGIONAL",
    "DefaultAction": {"Allow": {}},
    "Rules": [
      {
        "Name": "AllowInternalNetwork",
        "Priority": 10,
        "Action": {"Allow": {}},
        "Statement": {
          "IPSetReferenceStatement": {
            "Arn": "arn:aws:wafv2:region:account:regional/ipset/InternalNetwork/..."
          }
        },
        "VisibilityConfig": {
          "SampledRequestsEnabled": false,
          "CloudWatchMetricsEnabled": true,
          "MetricName": "AllowInternalNetwork"
        }
      },
      {
        "Name": "BlockKnownThreats",
        "Priority": 100,
        "Action": {"Block": {}},
        "Statement": {
          "IPSetReferenceStatement": {
            "Arn": "arn:aws:wafv2:region:account:regional/ipset/KnownThreats/..."
          }
        },
        "VisibilityConfig": {
          "SampledRequestsEnabled": false,
          "CloudWatchMetricsEnabled": true,
          "MetricName": "BlockKnownThreats"
        }
      },
      {
        "Name": "AdminAccessControl",
        "Priority": 200,
        "Action": {"Allow": {}},
        "Statement": {
          "AndStatement": {
            "Statements": [
              {
                "ByteMatchStatement": {
                  "SearchString": "/admin",
                  "FieldToMatch": {"UriPath": {}},
                  "PositionalConstraint": "STARTS_WITH",
                  "TextTransformations": [{"Priority": 0, "Type": "NONE"}]
                }
              },
              {
                "IPSetReferenceStatement": {
                  "Arn": "arn:aws:wafv2:region:account:regional/ipset/InternalNetwork/..."
                }
              }
            ]
          }
        },
        "VisibilityConfig": {
          "SampledRequestsEnabled": false,
          "CloudWatchMetricsEnabled": true,
          "MetricName": "AdminAccessControl"
        }
      },
      {
        "Name": "BlockAdminFromPublic",
        "Priority": 300,
        "Action": {"Block": {}},
        "Statement": {
          "ByteMatchStatement": {
            "SearchString": "/admin",
            "FieldToMatch": {"UriPath": {}},
            "PositionalConstraint": "STARTS_WITH",
            "TextTransformations": [{"Priority": 0, "Type": "NONE"}]
          }
        },
        "VisibilityConfig": {
          "SampledRequestsEnabled": false,
          "CloudWatchMetricsEnabled": true,
          "MetricName": "BlockAdminFromPublic"
        }
      }
    ]
  }
}
```

**Key Features:**
- Separate IPSet resources (lifecycle independent from WebACL)
- Can update IP sets without modifying WebACL
- Support for very large IP lists (10K addresses)
- IPv4 and IPv6 support

**Complexity:** Low - Excellent native support

---

### Azure Application Gateway WAF

**Status:** ✅ **Perfect Support**

**How it works:**
- **IPMatch operator** in custom rules
- **Inline IP addresses/CIDRs** in matchValues
- **Supports IPv4 and IPv6**
- **No separate IP set resource** (IPs inline in rules)

**Native Azure Translation:**
```json
{
  "location": "eastus",
  "properties": {
    "customRules": [
      {
        "name": "AllowInternalNetwork",
        "priority": 10,
        "ruleType": "MatchRule",
        "action": "Allow",
        "matchConditions": [
          {
            "matchVariables": [
              {"variableName": "RemoteAddr"}
            ],
            "operator": "IPMatch",
            "matchValues": [
              "10.0.0.0/8",
              "172.16.0.0/12"
            ],
            "negationConditon": false
          }
        ]
      },
      {
        "name": "BlockKnownThreats",
        "priority": 100,
        "ruleType": "MatchRule",
        "action": "Block",
        "matchConditions": [
          {
            "matchVariables": [
              {"variableName": "RemoteAddr"}
            ],
            "operator": "IPMatch",
            "matchValues": [
              "203.0.113.0/24",
              "192.0.2.0/24",
              "198.51.100.42/32"
            ],
            "negationConditon": false
          }
        ]
      },
      {
        "name": "AdminAccessControl",
        "priority": 200,
        "ruleType": "MatchRule",
        "action": "Allow",
        "matchConditions": [
          {
            "matchVariables": [
              {"variableName": "RequestUri"}
            ],
            "operator": "BeginsWith",
            "matchValues": ["/admin"],
            "negationConditon": false
          },
          {
            "matchVariables": [
              {"variableName": "RemoteAddr"}
            ],
            "operator": "IPMatch",
            "matchValues": [
              "10.0.0.0/8",
              "172.16.0.0/12",
              "203.0.113.50/32"
            ],
            "negationConditon": false
          }
        ]
      },
      {
        "name": "BlockAdminFromPublic",
        "priority": 300,
        "ruleType": "MatchRule",
        "action": "Block",
        "matchConditions": [
          {
            "matchVariables": [
              {"variableName": "RequestUri"}
            ],
            "operator": "BeginsWith",
            "matchValues": ["/admin"],
            "negationConditon": false
          }
        ]
      }
    ],
    "managedRules": {
      "managedRuleSets": [
        {
          "ruleSetType": "Microsoft_DefaultRuleSet",
          "ruleSetVersion": "2.1"
        }
      ]
    },
    "policySettings": {
      "state": "Enabled",
      "mode": "Prevention"
    }
  }
}
```

**Key Features:**
- Simple IPMatch operator
- Multiple IPs/CIDRs in single matchCondition (OR logic)
- Can combine with path, header, etc. (AND logic)
- IPv4 and IPv6 support

**Limitation:**
- No separate IP set resource (IPs duplicated across rules)
- Updating IP lists requires updating policy

**Complexity:** Low - Simple and direct

---

### GCP Cloud Armor

**Status:** ✅ **Perfect Support**

**How it works:**
- **CEL expressions**: `inIpRange(origin.ip, 'CIDR')` or `origin.ip == 'IP'`
- **Multiple IPs via OR**: `inIpRange(origin.ip, 'CIDR1') || inIpRange(origin.ip, 'CIDR2')`
- **Supports IPv4 and IPv6**
- **Preconfigured IP lists**: Regional IP lists available

**Native GCP Translation:**
```json
{
  "name": "ip-access-control-security-policy",
  "rules": [
    {
      "priority": 10,
      "description": "Allow internal network",
      "action": "allow",
      "match": {
        "expr": {
          "expression": "inIpRange(origin.ip, '10.0.0.0/8') || inIpRange(origin.ip, '172.16.0.0/12')"
        }
      }
    },
    {
      "priority": 100,
      "description": "Block known threats",
      "action": "deny(403)",
      "match": {
        "expr": {
          "expression": "inIpRange(origin.ip, '203.0.113.0/24') || inIpRange(origin.ip, '192.0.2.0/24') || origin.ip == '198.51.100.42'"
        }
      }
    },
    {
      "priority": 200,
      "description": "Admin access control - only internal network",
      "action": "allow",
      "match": {
        "expr": {
          "expression": "request.path.matches('/admin.*') && (inIpRange(origin.ip, '10.0.0.0/8') || inIpRange(origin.ip, '172.16.0.0/12') || origin.ip == '203.0.113.50')"
        }
      }
    },
    {
      "priority": 300,
      "description": "Block admin from public",
      "action": "deny(403)",
      "match": {
        "expr": {
          "expression": "request.path.matches('/admin.*')"
        }
      }
    },
    {
      "priority": 1000,
      "description": "OWASP ModSecurity CRS",
      "action": "deny(403)",
      "match": {
        "expr": {
          "expression": "evaluatePreconfiguredWaf('owasp-crs-v030301-id', {'sensitivity': 1})"
        }
      }
    },
    {
      "priority": 2147483647,
      "description": "Default rule - allow",
      "action": "allow",
      "match": {
        "expr": {
          "expression": "true"
        }
      }
    }
  ]
}
```

**Key Features:**
- Very flexible CEL expressions
- Easy to combine multiple IP ranges with `||`
- Can reference preconfigured regional IP lists
- Clean integration with other conditions

**Complexity:** Low - CEL is very expressive

---

## Design Decision

IP allowlist/blocklist rules are not exposed as typed fields. They belong in a `WafPolicy` ConfigMap because:

- AWS requires pre-created IPSet resources (ARNs) — not inline CIDRs at policy level; this cannot be abstracted transparently
- Provider JSON structures differ enough that a typed field would require provider knowledge anyway

### Portability validation result

| Feature | AWS WAFv2 | Azure WAF | GCP Cloud Armor |
|---------|-----------|-----------|-----------------|
| **IP allowlist/blocklist** | ✅ Perfect | ✅ Perfect | ✅ Perfect |
| **CIDR support** | ✅ Yes | ✅ Yes | ✅ Yes |
| **Individual IP support** | ✅ Yes | ✅ Yes | ✅ Yes |
| **IPv4 and IPv6** | ✅ Yes | ✅ Yes | ✅ Yes |
| **Reusable IP sets** | ✅ IPSet resource | ❌ Inline only | ❌ Inline only |
| **Path + IP combination** | ✅ AndStatement | ✅ Multiple matchConditions | ✅ CEL && |

## Validation Matrix

| Test Case | AWS | Azure | GCP | Expected Behavior |
|-----------|-----|-------|-----|-------------------|
| Request from 10.0.0.5 | ✅ Allowed | ✅ Allowed | ✅ Allowed | Internal network allowed globally |
| Request from 198.51.100.42 | ✅ Blocked | ✅ Blocked | ✅ Blocked | Known threat blocked |
| Request from 203.0.113.50 to /admin | ✅ Allowed | ✅ Allowed | ✅ Allowed | Trusted admin IP |
| Request from public IP to /admin | ✅ Blocked | ✅ Blocked | ✅ Blocked | Admin protected |
| Request from 198.51.100.0/24 to /api/partner-a | ✅ Allowed | ✅ Allowed | ✅ Allowed | Partner access granted |

## Conclusion

**IP allowlist/blocklist works perfectly across all three providers and is delivered via a `WafPolicy` ConfigMap.**

See `implementation-examples.md` Use Case 3 for complete provider-specific ConfigMap content for this use case.

