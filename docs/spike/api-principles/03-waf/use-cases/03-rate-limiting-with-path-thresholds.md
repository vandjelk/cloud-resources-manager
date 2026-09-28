# Use Case 4: Rate Limiting with Path-Based Thresholds

## User Story

"As a platform operator, I want to apply different rate limits to different paths - strict limits on expensive operations like /api/search and /api/export, but more permissive limits on cheap operations like /health and /api/users, to prevent abuse while maintaining good user experience."

## Requirements

1. Enable rate limiting for different paths
2. Apply different rate thresholds per path
3. Different actions per rate limit (block vs count)
4. Aggregate by source IP or other keys

## Real-World Scenarios

### Scenario A: Tiered API Rate Limits
- Expensive operations: 10 req/min
- Standard operations: 100 req/min  
- Health checks: No rate limit

### Scenario B: Authenticated vs Unauthenticated
- Unauthenticated: 10 req/min globally
- With API key: 100 req/min
- With JWT token: 1000 req/min

## API Design

Rate limiting rules go in a `WafPolicy` ConfigMap — the user provides a complete provider policy. See the provider capability check below for the exact JSON per provider. Note that Azure Application Gateway WAF does **not** support path-based rate limiting — this use case requires handling at a different layer (API Management, Front Door).

## Provider Capability Check

### AWS WAFv2

**Status:** ✅ **Perfect Support**

AWS supports custom rate-based rules with:
- `RateBasedStatement` with configurable limits (per 5 minutes)
- `AggregateKeyType`: IP, FORWARDED_IP, CUSTOM_KEYS
- `ScopeDownStatement` for path/condition matching

**Complexity:** Low - Native support
**Note:** AWS limits are per 5 minutes, so multiply by 5

---

### Azure Application Gateway WAF

**Status:** ❌ **Not Supported**

Azure Application Gateway WAF does **NOT** support:
- Path-based rate limiting
- Conditional rate limiting
- Per-rule rate limiting

**Workarounds:**
1. Use Azure API Management (full rate limiting support)
2. Use Azure Front Door (better rate limiting than App Gateway)
3. Implement rate limiting at application level

**Complexity:** High - Cannot implement in WAF

**Critical:** Status must report "not supported" for Azure

---

### GCP Cloud Armor

**Status:** ✅ **Perfect Support**

GCP supports rate limiting with:
- `rate_based_ban` or `throttle` actions
- `rateLimitOptions` with flexible thresholds
- `enforceOnKey`: IP, HEADER, COOKIE, PATH, etc.
- Configurable ban duration (rate_based_ban only)

**Complexity:** Low - Excellent native support

---

## Cross-Provider Comparison

| Feature | AWS WAFv2 | Azure WAF | GCP Cloud Armor |
|---------|-----------|-----------|-----------------|
| **Path-based rate limiting** | ✅ Yes | ❌ No | ✅ Yes |
| **Custom thresholds** | ✅ Per rule | ❌ Global only | ✅ Per rule |
| **Time window** | 5 minutes | N/A | Configurable |
| **Aggregate keys** | IP, FORWARDED_IP, CUSTOM | N/A | IP, HEADER, COOKIE, etc. |
| **Condition-based** | ✅ ScopeDownStatement | ❌ No | ✅ CEL expressions |
| **Fidelity** | Perfect | ❌ Not implementable | Perfect |

---

## Design Decision

Rate limiting with path thresholds is not a typed field — it belongs in a `WafPolicy` ConfigMap because:

- Azure Application Gateway WAF does not support path-based rate limiting at all
- Rate limiting configuration (time windows, aggregate keys, ban durations) is provider-specific and cannot be meaningfully abstracted

### Portability validation result

| Feature | AWS WAFv2 | Azure WAF | GCP Cloud Armor |
|---------|-----------|-----------|-----------------|
| **Path-based rate limiting** | ✅ Yes | ❌ No | ✅ Yes |
| **Custom thresholds** | ✅ Per rule | ❌ Global only | ✅ Per rule |
| **Time window** | 5 minutes | N/A | Configurable |
| **Aggregate keys** | IP, FORWARDED_IP, CUSTOM | N/A | IP, HEADER, COOKIE, etc. |
| **Condition-based** | ✅ ScopeDownStatement | ❌ No | ✅ CEL expressions |

## Validation Matrix

| Test Case | AWS | Azure | GCP | Expected Behavior |
|-----------|-----|-------|-----|-------------------|
| /api/search: 11 req/min | ✅ Blocked | ❌ Not supported | ✅ Blocked + banned | Rate limit enforced where supported |
| /api/users: 101 req/min | ✅ Blocked | ❌ Not supported | ✅ Throttled | Higher limit enforced |
| /health: unlimited | ✅ Allowed | ✅ Allowed | ✅ Allowed | Bypass works |
| Multiple IPs under limit | ✅ Allowed | N/A | ✅ Allowed | Per-IP tracking |

## Conclusion

**Path-based rate limiting works on AWS and GCP via a `WafPolicy` ConfigMap, but is not supported on Azure WAF.**

Users on Azure who need rate limiting must use Azure API Management or Azure Front Door instead of a WAF rule.
