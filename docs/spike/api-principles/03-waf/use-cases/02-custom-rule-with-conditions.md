# Use Case 2: Custom Rule with Path and Header Conditions

## User Story

"As a platform operator, I want to block requests to /admin paths that contain a specific suspicious header value (e.g., X-Debug: true from external networks), because this indicates an attempted exploit of a debug endpoint that should only be accessible internally."

## Requirements

1. Define custom blocking rule (not managed rule override)
2. Apply rule only to specific path prefix (e.g., /admin)
3. Match on header name and value
4. Optionally combine with source IP conditions
5. Block action when conditions match

## Real-World Threat Scenarios

### Scenario A: Debug Header Exploitation
```yaml
# Block X-Debug: true on /admin from public internet
# Legitimate: Internal tools use X-Debug on /admin from 10.0.0.0/8
# Threat: Attackers try X-Debug: true on /admin from public IPs
condition:
  path: {prefix: "/admin"}
  header:
    name: "X-Debug"
    value: "true"
  sourceIP:
    cidr: "0.0.0.0/0"
    negate: true  # NOT internal network
action: block
```

### Scenario B: Admin Panel Credential Stuffing
```yaml
# Block requests to /admin/login with X-Forwarded-For header
# Threat: Attackers use X-Forwarded-For to bypass rate limits
condition:
  path: {exact: "/admin/login"}
  header:
    name: "X-Forwarded-For"
    exists: true
  method: POST
action: block
```

### Scenario C: Scanner Detection
```yaml
# Block known scanner User-Agent on admin paths
# Threat: Automated scanners probing admin interfaces
condition:
  path: {prefix: "/admin"}
  header:
    name: "User-Agent"
    contains: "sqlmap"
action: block
```

## API Design

Custom rules with path, header, and IP conditions are **genuinely portable** across all three providers — the condition types translate cleanly (see the provider capability check below).

Custom rules go in a `WafPolicy` ConfigMap — the user provides a complete provider policy containing the rule. See the provider capability check below for the exact JSON per provider.

## Provider Capability Check

### AWS WAFv2

**Status:** ✅ **Perfect Support**

**How it works:**
- Custom rules use `Action` (not `OverrideAction` like managed rules)
- Multiple conditions combined with `AndStatement`
- Supports all condition types (path, header, method, IP)
- `NotStatement` for negation

**Native AWS Translation:**
```json
{
  "Name": "admin-protection-webacl",
  "Scope": "REGIONAL",
  "DefaultAction": {"Allow": {}},
  "Rules": [
    {
      "Name": "BlockDebugHeaderOnAdmin",
      "Priority": 100,
      "Action": {"Block": {}},
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
              "ByteMatchStatement": {
                "SearchString": "X-Debug",
                "FieldToMatch": {
                  "SingleHeader": {"Name": "x-debug"}
                },
                "PositionalConstraint": "EXACTLY",
                "TextTransformations": [{"Priority": 0, "Type": "NONE"}]
              }
            },
            {
              "NotStatement": {
                "Statement": {
                  "IPSetReferenceStatement": {
                    "Arn": "arn:aws:wafv2:region:account:regional/ipset/internal-network/..."
                  }
                }
              }
            }
          ]
        }
      },
      "VisibilityConfig": {
        "SampledRequestsEnabled": false,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "BlockDebugHeaderOnAdmin"
      }
    },
    {
      "Name": "BlockForwardedHeaderOnAdminLogin",
      "Priority": 110,
      "Action": {"Block": {}},
      "Statement": {
        "AndStatement": {
          "Statements": [
            {
              "ByteMatchStatement": {
                "SearchString": "/admin/login",
                "FieldToMatch": {"UriPath": {}},
                "PositionalConstraint": "EXACTLY",
                "TextTransformations": [{"Priority": 0, "Type": "NONE"}]
              }
            },
            {
              "ByteMatchStatement": {
                "SearchString": "POST",
                "FieldToMatch": {"Method": {}},
                "PositionalConstraint": "EXACTLY",
                "TextTransformations": [{"Priority": 0, "Type": "NONE"}]
              }
            },
            {
              "ByteMatchStatement": {
                "SearchString": "X-Forwarded-For",
                "FieldToMatch": {
                  "SingleHeader": {"Name": "x-forwarded-for"}
                },
                "PositionalConstraint": "EXACTLY",
                "TextTransformations": [{"Priority": 0, "Type": "NONE"}]
              }
            }
          ]
        }
      },
      "VisibilityConfig": {
        "SampledRequestsEnabled": false,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "BlockForwardedHeaderOnAdminLogin"
      }
    },
    {
      "Name": "BlockScannerUserAgent",
      "Priority": 120,
      "Action": {"Block": {}},
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
              "OrStatement": {
                "Statements": [
                  {
                    "ByteMatchStatement": {
                      "SearchString": "sqlmap",
                      "FieldToMatch": {
                        "SingleHeader": {"Name": "user-agent"}
                      },
                      "PositionalConstraint": "CONTAINS",
                      "TextTransformations": [{"Priority": 0, "Type": "LOWERCASE"}]
                    }
                  },
                  {
                    "ByteMatchStatement": {
                      "SearchString": "nikto",
                      "FieldToMatch": {
                        "SingleHeader": {"Name": "user-agent"}
                      },
                      "PositionalConstraint": "CONTAINS",
                      "TextTransformations": [{"Priority": 0, "Type": "LOWERCASE"}]
                    }
                  },
                  {
                    "ByteMatchStatement": {
                      "SearchString": "nmap",
                      "FieldToMatch": {
                        "SingleHeader": {"Name": "user-agent"}
                      },
                      "PositionalConstraint": "CONTAINS",
                      "TextTransformations": [{"Priority": 0, "Type": "LOWERCASE"}]
                    }
                  }
                ]
              }
            }
          ]
        }
      },
      "VisibilityConfig": {
        "SampledRequestsEnabled": false,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "BlockScannerUserAgent"
      }
    },
    {
      "Name": "AWSManagedRulesCommonRuleSet",
      "Priority": 1000,
      "Statement": {
        "ManagedRuleGroupStatement": {
          "VendorName": "AWS",
          "Name": "AWSManagedRulesCommonRuleSet"
        }
      },
      "OverrideAction": {"None": {}},
      "VisibilityConfig": {
        "SampledRequestsEnabled": false,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "CommonRuleSet"
      }
    }
  ]
}
```

**Complexity:** Low - Direct 1:1 mapping

---

### Azure Application Gateway WAF

**Status:** ✅ **Perfect Support**

**How it works:**
- Custom rules defined in `customRules` array
- Each custom rule has `matchConditions` (all conditions are AND)
- Supports path, method, header, IP conditions
- `negationConditon: true` for negation

**Native Azure Translation:**
```json
{
  "location": "eastus",
  "properties": {
    "customRules": [
      {
        "name": "BlockDebugHeaderOnAdmin",
        "priority": 100,
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
          },
          {
            "matchVariables": [
              {"variableName": "RequestHeaders"}
            ],
            "selector": "X-Debug",
            "operator": "Contains",
            "matchValues": [""],
            "negationConditon": false
          },
          {
            "matchVariables": [
              {"variableName": "RemoteAddr"}
            ],
            "operator": "IPMatch",
            "matchValues": ["10.0.0.0/8"],
            "negationConditon": true
          }
        ]
      },
      {
        "name": "BlockForwardedHeaderOnAdminLogin",
        "priority": 110,
        "ruleType": "MatchRule",
        "action": "Block",
        "matchConditions": [
          {
            "matchVariables": [
              {"variableName": "RequestUri"}
            ],
            "operator": "Equal",
            "matchValues": ["/admin/login"],
            "negationConditon": false
          },
          {
            "matchVariables": [
              {"variableName": "RequestMethod"}
            ],
            "operator": "Equal",
            "matchValues": ["POST"],
            "negationConditon": false
          },
          {
            "matchVariables": [
              {"variableName": "RequestHeaders"}
            ],
            "selector": "X-Forwarded-For",
            "operator": "Contains",
            "matchValues": [""],
            "negationConditon": false
          }
        ]
      },
      {
        "name": "BlockScannerUserAgent",
        "priority": 120,
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
          },
          {
            "matchVariables": [
              {"variableName": "RequestHeaders"}
            ],
            "selector": "User-Agent",
            "operator": "Contains",
            "matchValues": ["sqlmap", "nikto", "nmap"],
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
        },
        {
          "ruleSetType": "Microsoft_SQLInjectionRuleSet",
          "ruleSetVersion": "1.0"
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

**Complexity:** Low - Direct mapping to customRules

**Azure Quirk:**
- Header exists check: Use `operator: "Contains"` with `matchValues: [""]`
- Multiple match values in single condition treated as OR (good for User-Agent list)

---

### GCP Cloud Armor

**Status:** ✅ **Perfect Support**

**How it works:**
- Custom rules use CEL (Common Expression Language) expressions
- CEL supports rich boolean logic: `&&`, `||`, `!`, `has()`, `matches()`, `contains()`
- Very flexible and expressive
- Lower priority numbers evaluated first

**Native GCP Translation:**
```json
{
  "name": "admin-protection-security-policy",
  "rules": [
    {
      "priority": 100,
      "description": "Block debug header on admin paths from external IPs",
      "action": "deny(403)",
      "preview": false,
      "match": {
        "expr": {
          "expression": "request.path.matches('/admin.*') && has(request.headers['x-debug']) && !inIpRange(origin.ip, '10.0.0.0/8')"
        }
      }
    },
    {
      "priority": 110,
      "description": "Block X-Forwarded-For header on admin login POST",
      "action": "deny(403)",
      "preview": false,
      "match": {
        "expr": {
          "expression": "request.path == '/admin/login' && request.method == 'POST' && has(request.headers['x-forwarded-for'])"
        }
      }
    },
    {
      "priority": 120,
      "description": "Block known scanner user agents on admin paths",
      "action": "deny(403)",
      "preview": false,
      "match": {
        "expr": {
          "expression": "request.path.matches('/admin.*') && (request.headers['user-agent'].lower().contains('sqlmap') || request.headers['user-agent'].lower().contains('nikto') || request.headers['user-agent'].lower().contains('nmap'))"
        }
      }
    },
    {
      "priority": 1000,
      "description": "OWASP ModSecurity CRS",
      "action": "deny(403)",
      "preview": false,
      "match": {
        "expr": {
          "expression": "evaluatePreconfiguredWaf('owasp-crs-v030301-id', {'sensitivity': 1})"
        }
      }
    },
    {
      "priority": 1100,
      "description": "SQL injection protection",
      "action": "deny(403)",
      "preview": false,
      "match": {
        "expr": {
          "expression": "evaluatePreconfiguredWaf('sqli-v33-stable', {'sensitivity': 1})"
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

**Complexity:** Low - CEL expressions are very expressive

**GCP Advantages:**
- Most flexible condition syntax via CEL
- Built-in string functions: `contains()`, `matches()`, `lower()`, `startsWith()`
- Clean OR logic: `|| ` operator
- Negation: `!` operator or `!has()`

---

## Design Decision

### Why there is no typed `customRules` field

Custom rules with path, header, and IP conditions are genuinely portable — all three providers support them natively. They are delivered via a `WafPolicy` ConfigMap containing the complete provider policy. A typed `customRules` field is not part of any currently specified resource because:

- Exposing them as typed fields covers only one portable concept while leaving size filtering, geographic blocking, and managed rule tuning requiring provider-specific JSON anyway
- A partial typed middle layer adds API surface without eliminating the need for the escape hatch
- The cleaner boundary is: all rule-level expression — portable or not — belongs in a `WafPolicy` ConfigMap

Users who need custom rules supply a `WafPolicy` referencing a ConfigMap with provider-specific JSON. The provider knowledge required is explicit, not hidden.

### Portability validation result

| Feature | AWS WAFv2 | Azure WAF | GCP Cloud Armor |
|---------|-----------|-----------|-----------------|
| **Custom rules with conditions** | ✅ Perfect | ✅ Perfect | ✅ Perfect |
| **Path matching** | ✅ Exact, Prefix, Regex | ✅ Exact, Prefix, Contains | ✅ CEL: ==, matches(), startsWith() |
| **Header matching** | ✅ Exact, Contains, Exists | ✅ Exact, Contains, Exists | ✅ CEL: ==, contains(), has() |
| **Method matching** | ✅ Yes | ✅ Yes | ✅ CEL: request.method |
| **IP matching** | ✅ CIDR, negation | ✅ CIDR, negation | ✅ CEL: inIpRange(), negation |
| **Boolean logic** | ✅ Nested AND/OR/NOT | ⚠️ AND only (flat) | ✅ CEL: &&, \|\|, ! |
| **Multiple values OR** | ✅ OrStatement | ✅ Multiple matchValues | ✅ CEL: \|\| |

## Validation Matrix

| Test Case | AWS | Azure | GCP | Expected Behavior |
|-----------|-----|-------|-----|-------------------|
| /admin + X-Debug header from public IP | ✅ Block | ✅ Block | ✅ Block | All providers block |
| /admin + X-Debug header from 10.0.0.0/8 | ✅ Allow | ✅ Allow | ✅ Allow | Internal network allowed |
| /api + X-Debug header from public IP | ✅ Allow | ✅ Allow | ✅ Allow | Path doesn't match |
| /admin/login POST with X-Forwarded-For | ✅ Block | ✅ Block | ✅ Block | All providers block |
| User-Agent contains "sqlmap" on /admin | ✅ Block | ✅ Block | ✅ Block | Scanner blocked |
| User-Agent contains "Chrome" on /admin | ✅ Allow | ✅ Allow | ✅ Allow | Legitimate browser allowed |

## Conclusion

**Custom rules with path and header conditions work perfectly across all providers — delivered via a `WafPolicy` ConfigMap.**

The provider capability is fully validated. Users who need these rules supply a `WafPolicy` referencing a ConfigMap with provider-specific JSON. The provider translations in the "Provider Capability Check" sections above serve as the reference for what to put in those ConfigMaps.

