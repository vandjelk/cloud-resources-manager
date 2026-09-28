# Use Case 6: Geographic Blocking and Allowlisting

## User Story

"As a platform operator, I want to block or allow traffic based on geographic location (country/region), to comply with data residency requirements, reduce attack surface from high-risk regions, or restrict service availability to specific markets."

## Requirements

1. Block traffic from specific countries
2. Allow traffic only from specific countries
3. Combine geographic rules with path conditions
4. Support for country codes (ISO 3166-1 alpha-2)
5. Handle unknown/unresolvable geographic locations

## Real-World Scenarios

### Scenario A: GDPR Compliance - EU Only Service
```yaml
# SaaS product only available in EU
# Block all non-EU traffic
allowlist:
  - EU countries only
  - Block all others
reason: "GDPR compliance - service only for EU residents"
```

### Scenario B: High-Risk Region Blocking
```yaml
# Block traffic from countries with high attack rates
blocklist:
  - CN (China)
  - RU (Russia)
  - KP (North Korea)
reason: "99% of attacks originate from these regions"
```

### Scenario C: Market-Specific APIs
```yaml
# Different APIs for different markets
/api/us: Allow only US
/api/eu: Allow only EU countries
/api/apac: Allow only APAC countries
```

### Scenario D: Admin Panel Geographic Lock
```yaml
# Admin panel only from headquarters country
path: /admin
allow: US only
reason: "Admin operations only from US headquarters"
```

## API Design

Geographic blocking rules go in a `WafPolicy` ConfigMap — the user provides a complete provider policy containing the geographic rule. Note that **Azure Application Gateway WAF does not support geographic filtering** — this use case only applies on AWS and GCP clusters. Azure users requiring geo-filtering must use Azure Front Door or implement it at the application layer.

## Provider Capability Check

### AWS WAFv2

**Status:** ✅ **Perfect Support**

**How it works:**
- **GeoMatchStatement**: Native geographic matching
- **ISO 3166-1 alpha-2 country codes**
- **Supports all countries and territories**
- **Can combine with ForwardedIPConfig** for X-Forwarded-For header

**Native AWS Translation:**
```json
{
  "Name": "geographic-webacl",
  "Scope": "REGIONAL",
  "DefaultAction": {"Allow": {}},
  "Rules": [
    {
      "Name": "AllowEUOnly",
      "Priority": 10,
      "Action": {"Allow": {}},
      "Statement": {
        "GeoMatchStatement": {
          "CountryCodes": [
            "AT",
            "BE",
            "DE",
            "FR",
            "IT",
            "NL"
          ]
        }
      },
      "VisibilityConfig": {
        "SampledRequestsEnabled": false,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "AllowEUOnly"
      }
    },
    {
      "Name": "BlockNonEU",
      "Priority": 20,
      "Action": {"Block": {}},
      "Statement": {
        "NotStatement": {
          "Statement": {
            "GeoMatchStatement": {
              "CountryCodes": ["AT", "BE", "DE", "FR", "IT", "NL"]
            }
          }
        }
      },
      "VisibilityConfig": {
        "SampledRequestsEnabled": false,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "BlockNonEU"
      }
    },
    {
      "Name": "BlockHighRiskCountries",
      "Priority": 100,
      "Action": {"Block": {}},
      "Statement": {
        "GeoMatchStatement": {
          "CountryCodes": ["CN", "RU", "KP"]
        }
      },
      "VisibilityConfig": {
        "SampledRequestsEnabled": false,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "BlockHighRiskCountries"
      }
    },
    {
      "Name": "USApiOnlyUS",
      "Priority": 200,
      "Action": {"Allow": {}},
      "Statement": {
        "AndStatement": {
          "Statements": [
            {
              "ByteMatchStatement": {
                "SearchString": "/api/us",
                "FieldToMatch": {"UriPath": {}},
                "PositionalConstraint": "STARTS_WITH",
                "TextTransformations": [{"Priority": 0, "Type": "NONE"}]
              }
            },
            {
              "GeoMatchStatement": {
                "CountryCodes": ["US"]
              }
            }
          ]
        }
      },
      "VisibilityConfig": {
        "SampledRequestsEnabled": false,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "USApiOnlyUS"
      }
    }
  ]
}
```

**Key Features:**
- Native `GeoMatchStatement`
- Support for `ForwardedIPConfig` (CloudFront, ALB)
- Can negate with `NotStatement`
- Combine with other conditions via `AndStatement`

**Complexity:** Low - Excellent native support

---

### Azure Application Gateway WAF

**Status:** ❌ **Not Supported**

**How it works:**
- **Azure Application Gateway WAF v2 does NOT support geographic matching**
- **Azure Front Door WAF has geo-filtering** (different product)

**Workarounds:**
1. **Use Azure Front Door** instead of Application Gateway (has geo-filtering rules)
2. **Use Azure Firewall** in front of Application Gateway (has geo-based rules)
3. **Implement at application level** with IP geolocation libraries

**Native Azure Translation:**
```json
{
  "location": "eastus",
  "properties": {
    "customRules": [],
    "managedRules": {
      "managedRuleSets": [
        {
          "ruleSetType": "Microsoft_DefaultRuleSet",
          "ruleSetVersion": "2.1"
        }
      ]
    }
  }
}
```

**Complexity:** High - **Cannot implement geographic rules in Application Gateway WAF**

**Critical Limitation:**
- ❌ No geographic matching capability
- ⚠️ Must use different Azure product (Front Door) or architecture

---

### GCP Cloud Armor

**Status:** ✅ **Perfect Support**

**How it works:**
- **CEL expression**: `origin.region_code` variable
- **ISO 3166-1 alpha-2 country codes**
- **Supports all countries**
- **Very flexible with CEL operators**

**Native GCP Translation:**
```json
{
  "name": "geographic-security-policy",
  "rules": [
    {
      "priority": 10,
      "description": "Allow EU countries only",
      "action": "allow",
      "match": {
        "expr": {
          "expression": "origin.region_code in ['AT', 'BE', 'DE', 'FR', 'IT', 'NL']"
        }
      }
    },
    {
      "priority": 20,
      "description": "Block non-EU traffic",
      "action": "deny(403)",
      "match": {
        "expr": {
          "expression": "!(origin.region_code in ['AT', 'BE', 'DE', 'FR', 'IT', 'NL'])"
        }
      }
    },
    {
      "priority": 100,
      "description": "Block high-risk countries",
      "action": "deny(403)",
      "match": {
        "expr": {
          "expression": "origin.region_code in ['CN', 'RU', 'KP']"
        }
      }
    },
    {
      "priority": 200,
      "description": "US API only from US",
      "action": "allow",
      "match": {
        "expr": {
          "expression": "request.path.matches('/api/us.*') && origin.region_code == 'US'"
        }
      }
    },
    {
      "priority": 210,
      "description": "Block non-US access to US API",
      "action": "deny(403)",
      "match": {
        "expr": {
          "expression": "request.path.matches('/api/us.*') && origin.region_code != 'US'"
        }
      }
    },
    {
      "priority": 300,
      "description": "Admin only from US",
      "action": "allow",
      "match": {
        "expr": {
          "expression": "request.path.matches('/admin.*') && origin.region_code == 'US'"
        }
      }
    },
    {
      "priority": 310,
      "description": "Block non-US access to admin",
      "action": "deny(403)",
      "match": {
        "expr": {
          "expression": "request.path.matches('/admin.*')"
        }
      }
    },
    {
      "priority": 2147483647,
      "description": "Default allow",
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
- `origin.region_code` CEL variable
- Clean `in` operator for multiple countries
- Easy negation with `!`
- Combine with path/IP conditions naturally

**Complexity:** Low - CEL is very expressive

---

## Design Decision

Geographic blocking is not exposed as a typed field. It belongs in a `WafPolicy` ConfigMap because:

- Azure Application Gateway WAF does not support geographic filtering at all — a typed portable field cannot behave consistently across all three providers
- Geographic rules (country code lists) are provider-specific in syntax

### Portability validation result

| Feature | AWS WAFv2 | Azure WAF | GCP Cloud Armor |
|---------|-----------|-----------|-----------------|
| **Geographic matching** | ✅ Perfect | ❌ Not supported | ✅ Perfect |
| **Country codes** | ISO 3166-1 alpha-2 | N/A | ISO 3166-1 alpha-2 |
| **Multiple countries** | ✅ Array in GeoMatchStatement | N/A | ✅ CEL `in` operator |
| **Negation (NOT country)** | ✅ NotStatement | N/A | ✅ CEL `!` or `!=` |
| **Path + geo combination** | ✅ AndStatement | N/A | ✅ CEL `&&` |

Azure users requiring geographic restrictions must use **Azure Front Door** (not Application Gateway WAF) or implement geo-filtering at the application level.

## Validation Matrix

| Test Case | AWS | Azure | GCP | Expected Behavior |
|-----------|-----|-------|-----|-------------------|
| Request from DE (Germany) | ✅ Allowed (EU) | ❌ Not supported | ✅ Allowed | EU traffic allowed |
| Request from US to /api/eu | ✅ Blocked | ❌ Not supported | ✅ Blocked | Non-EU to EU API blocked |
| Request from CN (China) | ✅ Blocked | ❌ Not supported | ✅ Blocked | High-risk country blocked |
| Request from US to /admin | ✅ Allowed | ❌ Not supported | ✅ Allowed | US admin access allowed |
| Request from FR to /admin | ✅ Blocked | ❌ Not supported | ✅ Blocked | Non-US admin blocked |

## Conclusion

**Geographic filtering works on AWS and GCP via a `WafPolicy` ConfigMap, but is not supported on Azure Application Gateway WAF.**

Common use cases: GDPR compliance (EU-only services), export control compliance, attack surface reduction, data residency requirements. Despite the Azure limitation, geographic filtering is often a legal requirement and is fully supported on AWS and GCP.
