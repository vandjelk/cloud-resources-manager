# WAF Use Cases - Provider Capability Validation

This directory contains use case validations that test specific WAF scenarios across AWS, Azure, and GCP.

## Purpose

**These are NOT API design documents.**  
The API design is in [../api-specification.md](../api-specification.md).

These files validate:
- ✅ Which providers support specific use cases natively
- ✅ What workarounds are needed per provider
- ✅ How use cases map to `WafPolicy` ConfigMap content

---

## Use Case Index

| Use Case | AWS | Azure | GCP |
|----------|-----|-------|-----|
| [00. Specify Managed Rules](00-specify-managed-rules.md) | ✅ Perfect | ✅ Perfect | ✅ Perfect |
| [01. Managed Rules Override](01-managed-rules-override.md) | ✅ Perfect | ✅ Perfect | ⚠️ Limited |
| [02. Custom Rule with Conditions](02-custom-rule-with-conditions.md) | ✅ Native | ✅ Native | ✅ Native |
| [03. Rate Limiting with Path Thresholds](03-rate-limiting-with-path-thresholds.md) | ✅ Native | ❌ Not supported | ✅ Native |
| [04. IP Allowlist/Blocklist](04-ip-allowlist-blocklist.md) | ✅ Native | ✅ Native | ✅ Native |
| [05. Geographic Blocking](05-geographic-blocking.md) | ✅ Native | ❌ Limited | ✅ Native |
| [06. Size-Based Filtering](06-size-based-filtering.md) | ✅ Native | ⚠️ Limited | ⚠️ Limited |
| [07. Bot Protection](07-bot-protection.md) | ✅ Native | ⚠️ Limited | ✅ Native |

---

## Validation Matrix

### Provider Capability Summary

| Feature | AWS | Azure | GCP | Via WafPolicy ConfigMap? |
|---------|-----|-------|-----|----------------|
| Managed rule groups | ✅ | ✅ | ✅ | ✅ Complete policy per provider |
| Managed rule override | ✅ | ✅ | ⚠️ Degrades entire ruleset | ✅ Provider-specific override JSON |
| Custom rules (path/header/IP) | ✅ | ✅ | ✅ | ✅ Complete policy with custom rules |
| Rate limiting | ✅ | ❌ Not supported | ✅ | ✅ AWS/GCP only |
| IP allowlist/blocklist | ✅ | ✅ | ✅ | ✅ Complete policy per provider |
| Geographic blocking | ✅ | ❌ WAF level unsupported | ✅ | ✅ AWS/GCP only |
| Size-based filtering | ✅ | ⚠️ Global only | ⚠️ Limited | ✅ Provider-specific JSON |
| Bot protection | ✅ | ⚠️ Limited | ✅ | ✅ Complete policy per provider |

**Implementation notes:**
- AWS has best native support (nested conditions, full boolean logic)
- Azure supports AND-only conditions (flat structure)
- GCP uses CEL expressions (most flexible syntax)
- Azure applies WAF mode globally (all rule sets in Prevention or Detection)

---

## How to Read Use Case Files

Each use case file follows this structure:

```markdown
# Use Case N: [Name]

## User Story
"As a [role], I want [goal]..."

## Requirements
1. Requirement 1
2. Requirement 2
...

## Provider Capability Check

### AWS WAFv2
**Status:** ✅/⚠️/❌
**How it works:** [Native mechanism or workaround]
[Example AWS JSON]

### Azure WAF
**Status:** ✅/⚠️/❌
**How it works:** [Native mechanism or workaround]
[Example Azure JSON]

### GCP Cloud Armor
**Status:** ✅/⚠️/❌
**How it works:** [Native mechanism or workaround]
[Example GCP JSON]

## Validation Result
Summary of provider support for this use case.
Links to api-specification.md and implementation-examples.md for how it maps to WafPolicy.
```

---

## Use Cases vs API Design

**Use cases validate provider capabilities.**  
**API design (in api-specification.md) decides how to expose those capabilities.**

Example:
- **Use Case 02** validates that all providers support custom rules with conditions natively
- **API Design** routes these to a `WafPolicy` ConfigMap — users supply a complete provider policy containing the rule
- **Implementation** (controller) reads the ConfigMap and provisions the cloud WAF resource

---

## Cross-References

- See [../api-specification.md](../api-specification.md) for WafPolicy API design
- See [../waf-configuration-design.md](../waf-configuration-design.md) for the deferred WafConfiguration exploration
- See [../research/](../research/) for detailed cross-provider analysis
- See [../implementation-examples.md](../implementation-examples.md) for complete working examples

---

## Adding New Use Cases

To add a new use case:

1. **Describe the user need** (not the API)
2. **List requirements** (what must work)
3. **Validate on each provider** (AWS, Azure, GCP)
4. **Show provider-specific examples** (actual JSON/YAML that works)
5. **Summarize feasibility** (can this be handled via a `WafPolicy` ConfigMap? which providers?)
6. **Link to API design** (don't design API here, reference main spec)

Remember: Use cases validate **provider capabilities**, not **API design**.
