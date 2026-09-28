# Use Case 7: Size-Based Request Filtering

## User Story

"As a platform operator, I want to limit request body sizes on specific paths to prevent resource exhaustion attacks, block large file uploads on sensitive endpoints, and apply different size limits for different API operations based on their legitimate use cases."

## Requirements

1. Limit request body size globally or per-path
2. Limit specific header sizes
3. Limit query string length
4. Limit URI path length
5. Different size thresholds for different paths/methods
6. Block or log oversized requests

## Real-World Scenarios

### Scenario A: Prevent Upload Bombs
```yaml
# Normal API endpoints: 100 KB max body
# File upload endpoint: 10 MB max body
# GraphQL endpoint: 1 MB max body
paths:
  - /api: 100 KB
  - /api/upload: 10 MB
  - /api/graphql: 1 MB
reason: "Prevent resource exhaustion from large payloads"
```

### Scenario B: Admin Panel Protection
```yaml
# Admin operations should be small commands
path: /admin
body_size: 10 KB
reason: "Admin commands are small; large bodies indicate attack"
```

### Scenario C: Query String Attack Prevention
```yaml
# Prevent SQL injection via long query strings
query_string_length: 2048
reason: "Legitimate queries are short; long strings indicate injection attempts"
```

### Scenario D: Header Size Limits
```yaml
# Prevent header-based DoS attacks
header_size: 8 KB per header
total_headers_size: 64 KB
reason: "Prevent slowloris and header smuggling attacks"
```

## API Design

Size-based filtering rules go in a `WafPolicy` ConfigMap — the user provides a complete provider policy. Provider support is uneven — see the provider capability check below.

## Provider Capability Check

### AWS WAFv2

**Status:** ✅ **Excellent Support**

**How it works:**
- **SizeConstraintStatement**: Match based on size of request components
- **Field types**: BODY, QUERY_STRING, URI_PATH, HEADER, SINGLE_HEADER, ALL_QUERY_ARGUMENTS
- **Comparison operators**: EQ, NE, LE, LT, GE, GT
- **Text transformations**: NONE, COMPRESS_WHITE_SPACE, HTML_ENTITY_DECODE, LOWERCASE, CMD_LINE, URL_DECODE

**Native AWS Translation:**
```json
{
  "Name": "size-based-webacl",
  "Scope": "REGIONAL",
  "DefaultAction": {"Allow": {}},
  "Rules": [
    {
      "Name": "HealthCheckNoBody",
      "Priority": 50,
      "Action": {"Block": {}},
      "Statement": {
        "AndStatement": {
          "Statements": [
            {
              "ByteMatchStatement": {
                "SearchString": "/health",
                "FieldToMatch": {"UriPath": {}},
                "PositionalConstraint": "EXACTLY",
                "TextTransformations": [{"Priority": 0, "Type": "NONE"}]
              }
            },
            {
              "SizeConstraintStatement": {
                "FieldToMatch": {"Body": {}},
                "ComparisonOperator": "GT",
                "Size": 1024,
                "TextTransformations": [{"Priority": 0, "Type": "NONE"}]
              }
            }
          ]
        }
      },
      "VisibilityConfig": {
        "SampledRequestsEnabled": false,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "HealthCheckNoBody"
      }
    },
    {
      "Name": "AdminSmallBody",
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
              "SizeConstraintStatement": {
                "FieldToMatch": {"Body": {}},
                "ComparisonOperator": "GT",
                "Size": 10240,
                "TextTransformations": [{"Priority": 0, "Type": "NONE"}]
              }
            }
          ]
        }
      },
      "VisibilityConfig": {
        "SampledRequestsEnabled": false,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "AdminSmallBody"
      }
    },
    {
      "Name": "GraphQLModerateBody",
      "Priority": 110,
      "Action": {"Block": {}},
      "Statement": {
        "AndStatement": {
          "Statements": [
            {
              "ByteMatchStatement": {
                "SearchString": "/api/graphql",
                "FieldToMatch": {"UriPath": {}},
                "PositionalConstraint": "EXACTLY",
                "TextTransformations": [{"Priority": 0, "Type": "NONE"}]
              }
            },
            {
              "SizeConstraintStatement": {
                "FieldToMatch": {"Body": {}},
                "ComparisonOperator": "GT",
                "Size": 1048576,
                "TextTransformations": [{"Priority": 0, "Type": "NONE"}]
              }
            }
          ]
        }
      },
      "VisibilityConfig": {
        "SampledRequestsEnabled": false,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "GraphQLModerateBody"
      }
    },
    {
      "Name": "BlockLongQueryStrings",
      "Priority": 200,
      "Action": {"Block": {}},
      "Statement": {
        "SizeConstraintStatement": {
          "FieldToMatch": {"QueryString": {}},
          "ComparisonOperator": "GT",
          "Size": 2048,
          "TextTransformations": [{"Priority": 0, "Type": "NONE"}]
        }
      },
      "VisibilityConfig": {
        "SampledRequestsEnabled": false,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "BlockLongQueryStrings"
      }
    },
    {
      "Name": "BlockLongURIs",
      "Priority": 210,
      "Action": {"Block": {}},
      "Statement": {
        "SizeConstraintStatement": {
          "FieldToMatch": {"UriPath": {}},
          "ComparisonOperator": "GT",
          "Size": 8192,
          "TextTransformations": [{"Priority": 0, "Type": "NONE"}]
        }
      },
      "VisibilityConfig": {
        "SampledRequestsEnabled": false,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "BlockLongURIs"
      }
    }
  ]
}
```

**Key Features:**
- `SizeConstraintStatement` with multiple field types
- Comparison operators: GT, LT, GE, LE, EQ, NE
- Can check BODY, QUERY_STRING, URI_PATH, HEADER sizes
- Combine with path conditions via AndStatement

**Complexity:** Low - Excellent native support

---

### Azure Application Gateway WAF

**Status:** ⚠️ **Limited Support - Policy-Level Only**

**How it works:**
- **Global size limits** at WAF policy level (not per-rule)
- `maxRequestBodySizeInKb` (default 128 KB, max 128 KB for WAF v2)
- `requestBodyCheck` (enable/disable body inspection)
- **No per-path or per-rule size constraints**

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
    },
    "policySettings": {
      "state": "Enabled",
      "mode": "Prevention",
      "requestBodyCheck": true,
      "maxRequestBodySizeInKb": 128,
      "fileUploadLimitInMb": 100
    }
  }
}
```

**Key Limitations:**
- ❌ **No per-path body size limits** (only global)
- ❌ **No query string length constraints**
- ❌ **No URI length constraints**
- ❌ **No header size constraints** in custom rules
- ✅ Only global `maxRequestBodySizeInKb` setting

**Complexity:** High - Cannot implement path-specific size rules

**Workaround:**
- Use Azure API Management for per-API size policies
- Implement size validation at application level

---

### GCP Cloud Armor

**Status:** ⚠️ **Partial Support - Limited Size Checks**

**How it works:**
- **No direct size constraint matching in CEL expressions**
- Body size limits configured at backend service level (not in Cloud Armor)
- Can use `request.path.size()` and `request.query.size()` in CEL but not body size
- **No built-in body size checking in security policy rules**

**Native GCP Translation:**
```json
{
  "name": "size-based-security-policy",
  "rules": [
    {
      "priority": 200,
      "description": "Block long query strings",
      "action": "deny(413)",
      "match": {
        "expr": {
          "expression": "size(request.query) > 2048"
        }
      }
    },
    {
      "priority": 210,
      "description": "Block long URIs",
      "action": "deny(413)",
      "match": {
        "expr": {
          "expression": "size(request.path) > 8192"
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

**Key Limitations:**
- ❌ **Cannot check request body size** in Cloud Armor rules
- ✅ Can check query string size via `size(request.query)`
- ✅ Can check URI path size via `size(request.path)`
- ❌ Cannot check header sizes
- ⚠️ Body size limits configured at backend service, not in security policy

**Complexity:** Medium - Limited to query/URI size checks

**Backend Service Configuration (separate from Cloud Armor):**
```yaml
# Backend service configuration (not Cloud Armor)
maxStreamDuration: 60s
# Body size limits configured here, not in security policy
```

---

## Design Decision

Size-based filtering is not exposed as typed fields. It belongs in a `WafPolicy` ConfigMap because:

- Provider support is too uneven for a meaningful portable typed field
- Azure supports only global body size limits, not per-path
- GCP cannot check request body size in Cloud Armor rules at all
- AWS's `SizeConstraintStatement` has no equivalent on other providers

### Portability validation result

| Feature | AWS WAFv2 | Azure WAF | GCP Cloud Armor |
|---------|-----------|-----------|-----------------|
| **Request body size check** | ✅ Per-rule | ⚠️ Global only (128 KB max) | ❌ Backend service only |
| **Query string size check** | ✅ Per-rule | ❌ No | ✅ CEL `size(request.query)` |
| **URI path size check** | ✅ Per-rule | ❌ No | ✅ CEL `size(request.path)` |
| **Per-path size limits** | ✅ AndStatement | ❌ Global only | ⚠️ Query/URI only |

## Validation Matrix

| Test Case | AWS | Azure | GCP | Expected Behavior |
|-----------|-----|-------|-----|-------------------|
| POST /health with 2 KB body | ✅ Blocked | ⚠️ Allowed (no per-path) | ❌ Not checked | Health check should have no body |
| POST /admin with 20 KB body | ✅ Blocked | ⚠️ Allowed (global 128 KB) | ❌ Not checked | Admin body too large |
| POST /api/graphql with 2 MB body | ✅ Blocked | ✅ Blocked (global 128 KB) | ❌ Not checked | GraphQL body too large |
| GET /api with 3 KB query string | ✅ Blocked | ❌ Not checked | ✅ Blocked | Query string too long |
| GET /somepath with 10 KB URI | ✅ Blocked | ❌ Not checked | ✅ Blocked | URI too long |

## Conclusion

**Size-based filtering has uneven provider support:**
- **AWS**: Full per-rule size constraints via `SizeConstraintStatement`
- **Azure**: Global body size limit only — per-path constraints require Azure API Management
- **GCP**: Query/URI size checks only — body size limits configured at backend service level, not in Cloud Armor

Users who need size-based filtering supply a `WafPolicy` referencing a ConfigMap with provider-specific JSON. The provider capability sections above serve as the reference for what to put in those ConfigMaps.
