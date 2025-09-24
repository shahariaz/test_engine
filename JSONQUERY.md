# JSON Query Test Cases

This file contains 10 comprehensive JSON queries to test the Chorki JSON-to-DQL converter API manually. These queries cover various scenarios including simple filters, complex nested conditions, multi-entity relationships, and different operators.

## How to Test

Use these queries with the `/api/v1/convert` endpoint:

```bash
curl -X POST http://localhost:8090/api/v1/convert \
  -H "Content-Type: application/json" \
  -d '<JSON_QUERY_BELOW>'
```

Or in PowerShell:
```powershell
Invoke-RestMethod -Uri "http://localhost:8090/api/v1/convert" -Method Post -ContentType "application/json" -Body '<JSON_QUERY_BELOW>'
```

---

## Test Query 1: Premium Subscribers Analysis
**Scenario**: Find active Premium subscribers from specific countries with recent activity

```json
{
  "combine_with": "AND",
  "groups": [
    {
      "combine_with": "AND",
      "filters": [
        { "field": "subscription_status", "op": "=", "value": "active" },
        { "field": "subscribed_package", "op": "=", "value": "Premium" }
      ]
    },
    {
      "combine_with": "OR",
      "filters": [
        { "field": "country", "op": "IN", "value": ["USA", "UK", "Canada"] },
        { "field": "last_login_days", "op": "<=", "value": 5 }
      ]
    }
  ]
}
```

---

## Test Query 2: Young Active iOS Users
**Scenario**: Target young users on iOS devices with recent app versions

```json
{
  "combine_with": "AND",
  "groups": [
    {
      "combine_with": "AND",
      "filters": [
        { "field": "age", "op": "<", "value": 30 },
        { "field": "device", "op": "=", "value": "iOS" }
      ]
    },
    {
      "combine_with": "AND",
      "filters": [
        { "field": "app_version", "op": ">=", "value": "2.1.0" },
        { "field": "last_login_days", "op": "<=", "value": 10 }
      ]
    }
  ]
}
```

---

## Test Query 3: Content Engagement Analysis
**Scenario**: Users who watched specific movies or like certain genres

```json
{
  "combine_with": "OR",
  "groups": [
    {
      "combine_with": "AND",
      "filters": [
        { "field": "watched_content", "op": "IN", "value": { "content_type": "Movie", "ids": [111, 222] } }
      ]
    },
    {
      "combine_with": "OR",
      "filters": [
        { "field": "favorite_genres", "op": "IN", "value": ["Action", "Sci-Fi", "Thriller"] }
      ]
    }
  ]
}
```

---

## Test Query 4: Series Enthusiasts
**Scenario**: Users who prefer series content with specific viewing patterns

```json
{
  "combine_with": "AND",
  "groups": [
    {
      "combine_with": "OR",
      "filters": [
        { "field": "content_type", "op": "=", "value": "Series" },
        { "field": "favorite_genres", "op": "IN", "value": ["Mystery", "Drama", "Sci-Fi"] }
      ]
    },
    {
      "combine_with": "AND",
      "filters": [
        { "field": "age", "op": ">=", "value": 25 },
        { "field": "subscription_status", "op": "=", "value": "active" }
      ]
    }
  ]
}
```

---

## Test Query 5: Multi-Device Power Users
**Scenario**: Active users across different devices with premium features

```json
{
  "combine_with": "AND",
  "groups": [
    {
      "combine_with": "OR",
      "filters": [
        { "field": "device", "op": "IN", "value": ["iOS", "Android", "Web"] }
      ]
    },
    {
      "combine_with": "AND",
      "filters": [
        { "field": "subscription_status", "op": "=", "value": "active" },
        { "field": "last_login_days", "op": "<=", "value": 7 }
      ]
    },
    {
      "combine_with": "OR",
      "filters": [
        { "field": "age", "op": ">=", "value": 18 },
        { "field": "country", "op": "IN", "value": ["USA", "UK", "Canada", "Australia"] }
      ]
    }
  ]
}
```

---

## Test Query 6: Churned Users Analysis
**Scenario**: Users with expired subscriptions who haven't logged in recently

```json
{
  "combine_with": "AND",
  "groups": [
    {
      "combine_with": "OR",
      "filters": [
        { "field": "subscription_status", "op": "=", "value": "expired" },
        { "field": "last_login_days", "op": ">", "value": 30 }
      ]
    },
    {
      "combine_with": "AND",
      "filters": [
        { "field": "age", "op": ">=", "value": 20 },
        { "field": "subscribed_package", "op": "IN", "value": ["Premium", "Basic"] }
      ]
    }
  ]
}
```

---

## Test Query 7: Comedy Lovers Campaign
**Scenario**: Users who enjoy comedy content across different age groups

```json
{
  "combine_with": "OR",
  "groups": [
    {
      "combine_with": "AND",
      "filters": [
        { "field": "favorite_genres", "op": "IN", "value": ["Comedy"] },
        { "field": "age", "op": ">=", "value": 18 }
      ]
    },
    {
      "combine_with": "AND",
      "filters": [
        { "field": "watched_content", "op": "IN", "value": { "content_type": "Movie", "ids": [333] } },
        { "field": "last_login_days", "op": "<=", "value": 14 }
      ]
    }
  ]
}
```

---

## Test Query 8: Android Users with Older Versions
**Scenario**: Android users who might need app update prompts

```json
{
  "combine_with": "AND",
  "groups": [
    {
      "combine_with": "AND",
      "filters": [
        { "field": "device", "op": "=", "value": "Android" },
        { "field": "app_version", "op": "<", "value": "2.0.0" }
      ]
    },
    {
      "combine_with": "OR",
      "filters": [
        { "field": "subscription_status", "op": "=", "value": "active" },
        { "field": "last_login_days", "op": "<=", "value": 15 }
      ]
    }
  ]
}
```

---

## Test Query 9: Trial Users Conversion
**Scenario**: Trial subscription users who are actively engaging with content

```json
{
  "combine_with": "AND",
  "groups": [
    {
      "combine_with": "AND",
      "filters": [
        { "field": "subscription_status", "op": "=", "value": "trial" },
        { "field": "last_login_days", "op": "<=", "value": 3 }
      ]
    },
    {
      "combine_with": "OR",
      "filters": [
        { "field": "favorite_genres", "op": "IN", "value": ["Action", "Drama", "Comedy"] },
        { "field": "content_type", "op": "IN", "value": ["Movie", "Series"] }
      ]
    }
  ]
}
```

---

## Test Query 10: Comprehensive User Segmentation
**Scenario**: Complex multi-criteria user segmentation for marketing campaigns

```json
{
  "combine_with": "AND",
  "groups": [
    {
      "combine_with": "OR",
      "filters": [
        { "field": "age", "op": ">=", "value": 21 },
        { "field": "country", "op": "IN", "value": ["USA", "UK", "Canada", "Australia"] }
      ]
    },
    {
      "combine_with": "AND",
      "filters": [
        { "field": "subscription_status", "op": "=", "value": "active" },
        { "field": "last_login_days", "op": "<=", "value": 30 }
      ]
    },
    {
      "combine_with": "OR",
      "filters": [
        { "field": "watched_content", "op": "IN", "value": { "content_type": "Movie", "ids": [111, 222, 333] } },
        { "field": "favorite_genres", "op": "IN", "value": ["Action", "Drama", "Sci-Fi"] }
      ]
    },
    {
      "combine_with": "AND",
      "filters": [
        { "field": "device", "op": "IN", "value": ["iOS", "Android"] },
        { "field": "app_version", "op": ">=", "value": "2.0.0" }
      ]
    }
  ]
}
```

---

## Expected Results Summary

Each query should return a properly formatted DQL query with:

1. **Proper entity separation** - Different root queries for different entity types
2. **Correct filter syntax** - Using `eq()`, `ge()`, `le()`, etc. instead of `uid_in()` for scalar fields
3. **Relationship traversal** - Forward and reverse edge navigation using `~` notation
4. **Field selection** - Comprehensive field lists for each entity type
5. **Performance optimization** - Query complexity analysis and validation

## Testing the Generated DQL

You can test the generated DQL directly against Dgraph using:

```bash
curl -X POST http://localhost:8080/query \
  -H "Content-Type: application/json" \
  -d '{"query": "GENERATED_DQL_HERE"}'
```

## Data Available for Testing

The test database contains:
- **6 customers** with varied demographics and activity patterns
- **6 subscriptions** (Premium/Basic, Active/Trial/Expired)
- **5 content items** (Movies: Action Hero[111], Love Stories[222], Comedy Central[333]; Series: Mystery Lane[444], Sci-Fi Adventures[555])
- **6 watch history records** with different viewing patterns
- **4 device records** across iOS/Android platforms

Happy testing! 🚀