# Use Case 1: Managed Rules Override

## User Story

"As a platform operator, I want to enable OWASP protection in block mode globally, but change specific rules to count mode because they trigger false positives for our application."

## Requirements

1. Start from managed rule groups (via WafPolicy preset or custom)
2. Set default action to block
3. Override specific rule IDs to different action (e.g., count)
4. Override applies to ALL requests (no conditions)

## API Design

**Why managed rule overrides are not portable:**
- `managedRuleGroup` and `ruleId` fields use provider-specific names — AWS rule IDs do not exist on Azure or GCP
- GCP has no per-rule override mechanism; any override degrades the **entire matched ruleset** to preview mode
- A field that silently behaves differently per provider is worse than no abstraction

Users who need to override managed rules create a `WafPolicy` referencing a ConfigMap with the complete provider policy including the override configuration (see the provider capability translations below).

## Provider Capability Check

### AWS WAFv2

**Status:** ✅ **Perfect Support**

**How it works:**
- Uses `RuleActionOverrides` field in `ManagedRuleGroupStatement`
- Can override individual rule actions by rule name
- No conditions needed

**Native AWS Translation:**
```json
{
  "Name": "production-webacl",
  "Scope": "REGIONAL",
  "DefaultAction": {"Allow": {}},
  "Rules": [
    {
      "Name": "AWSManagedRulesCommonRuleSet",
      "Priority": 1000,
      "Statement": {
        "ManagedRuleGroupStatement": {
          "VendorName": "AWS",
          "Name": "AWSManagedRulesCommonRuleSet",
          "RuleActionOverrides": [
            {
              "Name": "SizeRestrictions_BODY",
              "ActionToUse": {"Count": {}}
            },
            {
              "Name": "GenericRFI_BODY",
              "ActionToUse": {"Count": {}}
            }
          ]
        }
      },
      "OverrideAction": {"None": {}},
      "VisibilityConfig": {
        "SampledRequestsEnabled": false,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "CommonRuleSet"
      }
    },
    {
      "Name": "AWSManagedRulesSQLiRuleSet",
      "Priority": 1010,
      "Statement": {
        "ManagedRuleGroupStatement": {
          "VendorName": "AWS",
          "Name": "AWSManagedRulesSQLiRuleSet"
        }
      },
      "OverrideAction": {"None": {}},
      "VisibilityConfig": {
        "SampledRequestsEnabled": false,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "SQLiRuleSet"
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
- Uses `ruleGroupOverrides` in managed rule sets
- Can override rule actions at rule group → rule level
- Supports `action` field (Log, Block, Allow)

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
          "ruleSetVersion": "2.1",
          "ruleGroupOverrides": [
            {
              "ruleGroupName": "SQLI",
              "rules": [
                {
                  "ruleId": "942100",
                  "state": "Enabled",
                  "action": "Log"
                },
                {
                  "ruleId": "920280",
                  "state": "Enabled",
                  "action": "Log"
                }
              ]
            }
          ]
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

**Complexity:** Low - Direct mapping via ruleGroupOverrides

**Mapping Notes:**
- `action: count` → `action: "Log"` (logs but doesn't block)
- `action: block` → `action: "Block"`
- `action: allow` → `action: "Allow"` or `state: "Disabled"`

---

### GCP Cloud Armor

**Status:** ⚠️ **Partial Support - Rule Set Level Only**

**How it works:**
- GCP Cloud Armor preconfigured WAF rules have limited granularity
- Cannot override individual rule IDs within a rule set
- Can only affect entire rule set via `preview: true` (equivalent to count mode)
- Uses sensitivity levels (0-4) to tune rule sets, not individual rules

**Best-Effort Translation:**
```json
{
  "name": "production-security-policy",
  "rules": [
    {
      "priority": 1000,
      "description": "OWASP ModSecurity CRS - Count mode for entire rule set",
      "action": "deny(403)",
      "preview": true,
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

**Complexity:** Medium - Degrades to rule-set level

**Degradation Strategy:**
- If ANY rule in a managed rule group has `action: count` override → entire rule set in `preview: true`
- Cannot selectively override individual rules
- Status should report: "GCP Cloud Armor does not support individual rule overrides; entire [ruleset-id] in preview mode"

---

## Design Decision

### Why `ruleOverrides` is not a typed field

- `managedRuleGroup` and `ruleId` are provider-specific identifiers — not portable
- GCP cannot override individual rules; any override degrades the entire matched ruleset to preview mode
- A field that silently behaves differently per provider is a leaky abstraction — worse than no abstraction

Provider-specific rule tuning belongs in a `WafPolicy` ConfigMap, not in a portable resource field.

### Validation result

| Feature | AWS WAFv2 | Azure WAF | GCP Cloud Armor |
|---------|-----------|-----------|-----------------|
| **Individual rule override** | ✅ Yes | ✅ Yes | ❌ No |
| **Granularity** | Rule ID level | Rule ID level | Rule set level only |
| **Override actions** | Count, Allow, Block | Log, Allow, Block, Disabled | Preview (entire rule set) |
| **Fidelity** | Perfect | Perfect | Degraded |

GCP's inability to override individual rules (it degrades the entire ruleset to preview) is the deciding factor. A portable `ruleOverrides` field cannot behave consistently across all three providers.

## Validation Matrix

| Test Case | AWS | Azure | GCP | Expected Behavior |
|-----------|-----|-------|-----|-------------------|
| Override single rule to count | ✅ Rule level | ✅ Rule level | ⚠️ Rule set preview | AWS/Azure: specific rule → count; GCP: entire set → preview |
| Override multiple rules to count | ✅ Multiple rules | ✅ Multiple rules | ⚠️ Rule set preview | AWS/Azure: each rule individually; GCP: entire set |
| Override rule to allow | ✅ Rule disabled | ✅ Rule disabled | ⚠️ Rule set preview | AWS/Azure: rule bypassed; GCP: entire set → preview |
| No overrides | ✅ Default action | ✅ Default action | ✅ Default action | All rules follow default managed rule group action |

## Conclusion

`ruleOverrides` is not part of any currently specified resource. Users who need to override managed rules supply a `WafPolicy` referencing a ConfigMap with provider-specific JSON that includes the override configuration. This makes the provider knowledge explicit rather than hiding it behind a field that silently degrades on GCP.
