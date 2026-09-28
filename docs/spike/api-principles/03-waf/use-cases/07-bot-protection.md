# Use Case 8: Bot Protection and Management

## User Story

"As a platform operator, I want to detect and mitigate malicious bot traffic (scrapers, DDoS bots, credential stuffing bots) while allowing legitimate bots (search engines, monitoring tools, verified partners), and apply different protection levels to different paths based on sensitivity."

## Requirements

1. Enable bot detection and mitigation
2. Allow verified good bots (Googlebot, Bingbot, etc.)
3. Block known bad bots (scrapers, attack tools)
4. Challenge suspicious bot-like behavior
5. Different bot protection levels per path
6. Custom bot detection rules (User-Agent, behavior patterns)

## Real-World Scenarios

### Scenario A: E-Commerce Site Protection
```yaml
# Product pages: Allow search engine bots, block scrapers
# Checkout: Block all bots
# API: Allow verified partner bots with API keys
paths:
  - /products: Allow Google/Bing, block others
  - /checkout: Block all bots
  - /api: Allow bots with valid API key header
```

### Scenario B: Content Site with Ads
```yaml
# Articles: Allow search engines (SEO), block ad fraud bots
# Admin: Block all bots
level:
  - articles: Low protection (SEO-friendly)
  - admin: Maximum protection
```

### Scenario C: API Rate Limiting by Bot Type
```yaml
# Verified bots: Higher rate limits
# Unknown bots: Strict rate limits
# Malicious bots: Blocked
bot_type:
  - verified: 1000 req/min
  - unknown: 10 req/min
  - malicious: blocked
```

### Scenario D: Credential Stuffing Prevention
```yaml
# Login endpoint: Maximum bot protection
# Challenge suspicious login patterns
path: /login
protection: Maximum
actions: CAPTCHA challenge on bot detection
```

## API Design

Bot protection rules go in a `WafPolicy` ConfigMap — the user provides a complete provider policy. For common use cases, the policy enables the provider's managed bot protection rule group. For advanced rules (custom User-Agent blocking, CAPTCHA), the policy includes those rules explicitly. See the provider capability check below for the exact JSON per provider.

## Provider Capability Check

### AWS WAFv2

**Status:** ✅ **Excellent Support**

**How it works:**
- **AWS Managed Rules Bot Control**: `AWSManagedRulesBotControlRuleSet`
- **Bot detection levels**: Common, Targeted, Targeted_Extra
- **Bot categories**: CategorySearchEngine, CategoryMonitoring, CategoryAdvertising, CategorySocialMedia, CategoryScrapingFramework, etc.
- **Bot verification**: Verifies search engine bots (Googlebot, Bingbot)
- **Challenge actions**: CAPTCHA, Count, Block
- **Bot score**: 0-100 (0 = human, 100 = bot)

**Native AWS Translation:**
```json
{
  "Name": "bot-protection-webacl",
  "Scope": "REGIONAL",
  "DefaultAction": {"Allow": {}},
  "Rules": [
    {
      "Name": "AllowSearchEngineBots",
      "Priority": 10,
      "Action": {"Allow": {}},
      "Statement": {
        "OrStatement": {
          "Statements": [
            {
              "ByteMatchStatement": {
                "SearchString": "Googlebot",
                "FieldToMatch": {"SingleHeader": {"Name": "user-agent"}},
                "PositionalConstraint": "CONTAINS",
                "TextTransformations": [{"Priority": 0, "Type": "LOWERCASE"}]
              }
            },
            {
              "ByteMatchStatement": {
                "SearchString": "bingbot",
                "FieldToMatch": {"SingleHeader": {"Name": "user-agent"}},
                "PositionalConstraint": "CONTAINS",
                "TextTransformations": [{"Priority": 0, "Type": "LOWERCASE"}]
              }
            }
          ]
        }
      },
      "VisibilityConfig": {
        "SampledRequestsEnabled": false,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "AllowSearchEngineBots"
      }
    },
    {
      "Name": "BlockScraperBots",
      "Priority": 20,
      "Action": {"Block": {}},
      "Statement": {
        "OrStatement": {
          "Statements": [
            {
              "ByteMatchStatement": {
                "SearchString": "scrapy",
                "FieldToMatch": {"SingleHeader": {"Name": "user-agent"}},
                "PositionalConstraint": "CONTAINS",
                "TextTransformations": [{"Priority": 0, "Type": "LOWERCASE"}]
              }
            },
            {
              "ByteMatchStatement": {
                "SearchString": "python-requests",
                "FieldToMatch": {"SingleHeader": {"Name": "user-agent"}},
                "PositionalConstraint": "CONTAINS",
                "TextTransformations": [{"Priority": 0, "Type": "LOWERCASE"}]
              }
            }
          ]
        }
      },
      "VisibilityConfig": {
        "SampledRequestsEnabled": false,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "BlockScraperBots"
      }
    },
    {
      "Name": "AWSManagedRulesBotControlRuleSet",
      "Priority": 100,
      "Statement": {
        "ManagedRuleGroupStatement": {
          "VendorName": "AWS",
          "Name": "AWSManagedRulesBotControlRuleSet",
          "ManagedRuleGroupConfigs": [
            {
              "AWSManagedRulesBotControlRuleSet": {
                "InspectionLevel": "TARGETED"
              }
            }
          ],
          "RuleActionOverrides": [
            {
              "Name": "CategorySearchEngine",
              "ActionToUse": {"Count": {}}
            }
          ]
        }
      },
      "OverrideAction": {"None": {}},
      "VisibilityConfig": {
        "SampledRequestsEnabled": false,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "BotControlRuleSet"
      }
    },
    {
      "Name": "LoginBotChallenge",
      "Priority": 110,
      "Action": {
        "Captcha": {
          "CustomRequestHandling": {
            "InsertHeaders": [
              {
                "Name": "x-bot-challenge",
                "Value": "login"
              }
            ]
          }
        }
      },
      "Statement": {
        "AndStatement": {
          "Statements": [
            {
              "ByteMatchStatement": {
                "SearchString": "/login",
                "FieldToMatch": {"UriPath": {}},
                "PositionalConstraint": "EXACTLY",
                "TextTransformations": [{"Priority": 0, "Type": "NONE"}]
              }
            }
          ]
        }
      },
      "CaptchaConfig": {
        "ImmunityTimeProperty": {
          "ImmunityTime": 300
        }
      },
      "VisibilityConfig": {
        "SampledRequestsEnabled": false,
        "CloudWatchMetricsEnabled": true,
        "MetricName": "LoginBotChallenge"
      }
    }
  ],
  "CaptchaConfig": {
    "ImmunityTimeProperty": {
      "ImmunityTime": 300
    }
  }
}
```

**Key Features:**
- AWS Bot Control managed rule group (Common/Targeted/Targeted_Extra levels)
- Bot verification for major search engines
- CAPTCHA and Challenge actions
- Bot categories and labels
- RuleActionOverrides for specific bot categories

**Complexity:** Low - Excellent managed bot protection

**Cost:** ⚠️ Bot Control is a paid add-on (~$10/month + $1/million requests)

---

### Azure Application Gateway WAF

**Status:** ⚠️ **Limited Support - Bot Manager Rule Set**

**How it works:**
- **Microsoft Bot Manager Rule Set**: Basic bot detection
- **Rule group overrides**: Can disable/enable specific bot rules
- **No bot score or challenge actions**
- **No verified bot allowlisting**
- **Limited compared to AWS**

**Native Azure Translation:**
```json
{
  "location": "eastus",
  "properties": {
    "customRules": [
      {
        "name": "AllowSearchEngineBots",
        "priority": 10,
        "ruleType": "MatchRule",
        "action": "Allow",
        "matchConditions": [
          {
            "matchVariables": [
              {"variableName": "RequestHeaders"}
            ],
            "selector": "User-Agent",
            "operator": "Contains",
            "matchValues": ["Googlebot", "Bingbot"],
            "negationConditon": false
          }
        ]
      },
      {
        "name": "BlockScraperBots",
        "priority": 20,
        "ruleType": "MatchRule",
        "action": "Block",
        "matchConditions": [
          {
            "matchVariables": [
              {"variableName": "RequestHeaders"}
            ],
            "selector": "User-Agent",
            "operator": "Contains",
            "matchValues": ["scrapy", "python-requests", "curl"],
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
          "ruleSetType": "Microsoft_BotManagerRuleSet",
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

**Key Limitations:**
- ❌ No bot score or bot likelihood
- ❌ No CAPTCHA or challenge actions
- ❌ No verified bot list
- ✅ Basic bot detection via Bot Manager rule set
- ⚠️ Must use custom rules for User-Agent blocking

**Complexity:** Medium - Limited bot protection capabilities

---

### GCP Cloud Armor

**Status:** ✅ **Good Support - Adaptive Protection & reCAPTCHA**

**How it works:**
- **Preconfigured bot defense**: `evaluatePreconfiguredWaf('bot-defense')`
- **reCAPTCHA Enterprise integration**: Challenge suspicious requests
- **Adaptive Protection**: ML-based anomaly detection
- **Rate limiting per bot type**
- **Session affinity** for bot tracking

**Native GCP Translation:**
```json
{
  "name": "bot-protection-security-policy",
  "adaptiveProtectionConfig": {
    "layer7DdosDefenseConfig": {
      "enable": true,
      "ruleVisibility": "STANDARD"
    }
  },
  "rules": [
    {
      "priority": 10,
      "description": "Allow search engine bots",
      "action": "allow",
      "match": {
        "expr": {
          "expression": "request.headers['user-agent'].lower().contains('googlebot') || request.headers['user-agent'].lower().contains('bingbot')"
        }
      }
    },
    {
      "priority": 20,
      "description": "Block scraper bots",
      "action": "deny(403)",
      "match": {
        "expr": {
          "expression": "request.headers['user-agent'].lower().contains('scrapy') || request.headers['user-agent'].lower().contains('python-requests')"
        }
      }
    },
    {
      "priority": 100,
      "description": "reCAPTCHA challenge for login",
      "action": "redirect",
      "match": {
        "expr": {
          "expression": "request.path == '/login' && request.method == 'POST'"
        }
      },
      "redirectOptions": {
        "type": "GOOGLE_RECAPTCHA"
      },
      "recaptchaOptionsPath": {
        "recaptchaOptions": {
          "action": "login",
          "sessionTokenSiteKey": "projects/PROJECT_ID/keys/KEY_ID"
        }
      }
    },
    {
      "priority": 1000,
      "description": "Preconfigured bot defense",
      "action": "deny(403)",
      "match": {
        "expr": {
          "expression": "evaluatePreconfiguredWaf('bot-defense', {'sensitivity': 1})"
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
  }
}
```

**Key Features:**
- Preconfigured bot defense WAF rule
- reCAPTCHA Enterprise integration (challenge action)
- Adaptive Protection (ML-based)
- Custom User-Agent blocking via CEL
- Rate limiting by bot patterns

**Complexity:** Medium - Good bot protection with reCAPTCHA integration

**Cost:** ⚠️ reCAPTCHA Enterprise is paid (~$1/1000 assessments)

---

## Design Decision

Bot protection is delivered via a `WafPolicy` ConfigMap — the user provides a complete provider policy. For common use cases, the policy enables the provider's managed bot protection rule group. For advanced rules (custom User-Agent blocking, CAPTCHA challenge configuration), the policy includes those rules explicitly.

Typed bot-specific fields (`botScore`, `action: challenge`) are not part of any currently specified resource because:

- `botScore` is AWS-specific — Azure has no bot score, GCP does not expose it in CEL
- `action: challenge` has no equivalent on Azure
- These fields would only work on a subset of providers

### Portability validation result

| Feature | AWS WAFv2 | Azure WAF | GCP Cloud Armor |
|---------|-----------|-----------|-----------------|
| **Managed bot protection** | ✅ Bot Control (excellent) | ⚠️ Bot Manager (basic) | ✅ Bot defense (good) |
| **Custom User-Agent rules** | ✅ Yes | ✅ Yes | ✅ Yes (CEL) |
| **Challenge / CAPTCHA actions** | ✅ CAPTCHA, JS challenge | ❌ No | ✅ reCAPTCHA Enterprise |
| **Bot score** | ✅ 0-100 (AWS-specific) | ❌ No | ❌ Not exposed in CEL |
| **Per-path bot policies** | ✅ ScopeDownStatement | ⚠️ Custom rules only | ✅ CEL conditions |

## Validation Matrix

| Test Case | AWS | Azure | GCP | Expected Behavior |
|-----------|-----|-------|-----|-------------------|
| User-Agent: Googlebot | ✅ Allowed | ✅ Allowed | ✅ Allowed | Good bot allowed |
| User-Agent: scrapy | ✅ Blocked | ✅ Blocked | ✅ Blocked | Bad bot blocked |
| POST /login by bot | ✅ CAPTCHA challenge | ❌ Not supported | ✅ reCAPTCHA challenge | Bot challenged on login |
| Bot traffic to /admin | ✅ Blocked (Bot Control) | ⚠️ Basic detection only | ✅ Blocked (bot-defense) | Admin protected |

## Conclusion

**Bot protection is delivered via a `WafPolicy` ConfigMap with provider-specific JSON.**

The provider capability sections above serve as the reference for what to put in the ConfigMap.

