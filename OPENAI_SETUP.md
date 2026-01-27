# IncidentTeller OpenAI Integration Setup Guide

## Overview

This guide explains how to set up and configure OpenAI integration with your IncidentTeller system for AI-powered incident analysis.

## Prerequisites

- OpenAI API key (from https://platform.openai.com/api-keys)
- IncidentTeller system running
- Go 1.22 or later

## Configuration Steps

### 1. Obtain an OpenAI API Key

1. Visit [OpenAI API Keys page](https://platform.openai.com/api-keys)
2. Create a new API key
3. Copy the key securely

### 2. Set Environment Variable

Set your OpenAI API key as an environment variable:

```bash
# Linux/Mac
export OPENAI_API_KEY="sk-your-api-key-here"

# Windows PowerShell
$env:OPENAI_API_KEY="sk-your-api-key-here"

# Docker
docker run -e OPENAI_API_KEY="sk-your-api-key-here" incident-teller
```

### 3. Update Configuration

Edit your `config.yaml`:

```yaml
ai:
  enabled: true
  model_type: "hybrid" # local, openai, or hybrid
  confidence_threshold: 0.7
  max_predictions: 5
  
  openai:
    enabled: true
    api_key: "${OPENAI_API_KEY}" # Uses environment variable
    model: "gpt-4" # gpt-4, gpt-3.5-turbo, gpt-4o, etc.
    timeout: "30s"
    max_tokens: 2048
    temperature: 0.7 # 0.0-2.0 (higher = more creative)
    top_p: 1.0 # nucleus sampling (0.0-1.0)
```

### 4. Available Models

Choose based on your needs:

| Model | Speed | Quality | Cost |
|-------|-------|---------|------|
| gpt-3.5-turbo | Fast | Good | Low |
| gpt-4 | Slower | Excellent | Higher |
| gpt-4-turbo | Fast | Excellent | Medium |
| gpt-4o | Fast | Excellent | Medium |

## Features Enabled

### 1. AI-Powered Incident Analysis

When OpenAI is configured, the system provides:

- **Incident Summary**: Concise overview of what happened
- **Root Cause Analysis**: Identifies the likely root cause
- **Impact Assessment**: Business impact evaluation
- **Recommendations**: Actionable recommendations organized by timeframe:
  - Immediate actions (within 5 minutes)
  - Short-term actions (within 8 hours)
  - Long-term prevention strategies

### 2. Alert Grouping

The system automatically groups related alerts by:

- **Host**: Alerts from the same system
- **Time Window**: Related alerts within 15 minutes
- **Cascade Detection**: Identifies cascading failures
- **Resource Dependencies**: Detects resource-related cascades

### 3. Enhanced Timeline

Displays a detailed incident timeline showing:

- Alert triggers and escalations
- Cascade points and propagation
- Root cause identification
- Resolution events

## API Endpoints

### AI Analysis

```bash
curl -X POST http://localhost:8080/api/analyze
```

Response:
```json
{
  "summary": "High CPU usage on main-server...",
  "root_cause_text": "Primary cause: Memory leak in service X...",
  "impact_assessment": "Critical impact on user-facing services...",
  "recommendations": {
    "immediate": ["Restart service X", "Scale up..."],
    "short_term": ["Review logs", "Update monitoring..."],
    "long_term": ["Fix memory leak", "Add unit tests..."]
  },
  "generated_at": "2024-01-17T10:30:00Z",
  "alert_count": 15,
  "time_span": "45m30s"
}
```

### Alert Groups

```bash
curl http://localhost:8080/api/alert-groups
```

Response:
```json
{
  "groups": [
    {
      "id": "group-1",
      "alert_count": 5,
      "primary_host": "main-server",
      "affected_hosts": ["main-server", "worker-1"],
      "resource_types": ["CPU", "MEMORY"],
      "is_cascading": true,
      "group_type": "cascading",
      "duration": "45m30s"
    }
  ],
  "total": 1
}
```

### Enhanced Timeline

```bash
curl http://localhost:8080/api/timeline-enhanced/{incident-id}
```

## UI Features

### Analysis Page

Access at: `http://localhost:3000/analysis`

**AI Analysis Tab**:
- View AI-generated summary
- See detailed root cause analysis
- Review business impact assessment
- Get organized recommendations

**Alert Groups Tab**:
- See grouped alerts by host
- Identify cascading failures
- View affected resources
- Timeline information

## Monitoring

Monitor OpenAI usage:

1. Check [OpenAI Usage Dashboard](https://platform.openai.com/usage)
2. Set usage limits in OpenAI dashboard
3. Monitor API costs
4. Track request latency in IncidentTeller logs

## Troubleshooting

### Issue: OpenAI API errors

**Solution**:
- Verify API key is valid and active
- Check API key has correct permissions
- Ensure rate limits not exceeded
- Check internet connectivity

### Issue: Slow analysis

**Solution**:
- Use faster model (gpt-3.5-turbo)
- Reduce max_tokens
- Lower temperature value
- Check OpenAI API status

### Issue: High costs

**Solution**:
- Use gpt-3.5-turbo instead of gpt-4
- Reduce max_tokens
- Set usage limits in OpenAI dashboard
- Monitor requests in logs

## Hybrid Mode

For balanced performance and cost:

```yaml
ai:
  model_type: "hybrid"
  
  openai:
    enabled: true
    api_key: "${OPENAI_API_KEY}"
    model: "gpt-3.5-turbo"
    max_tokens: 1024
    temperature: 0.5
```

This uses OpenAI for high-complexity analysis and local models for quick predictions.

## Local-Only Mode

If you don't want to use OpenAI:

```yaml
ai:
  model_type: "local"
  
  openai:
    enabled: false
```

The system will use built-in ML models for analysis (less powerful but free).

## Advanced Configuration

### Rate Limiting

Control API usage:

```yaml
openai:
  timeout: "60s" # Timeout for API requests
  max_tokens: 4096 # Maximum tokens per request
```

### Model Selection

For different use cases:

```yaml
# For speed (free tier friendly)
model: "gpt-3.5-turbo"
max_tokens: 512
temperature: 0.3

# For quality
model: "gpt-4"
max_tokens: 2048
temperature: 0.7

# For balanced cost/quality
model: "gpt-4-turbo"
max_tokens: 1024
temperature: 0.5
```

## Best Practices

1. **Start with gpt-3.5-turbo** for cost-effectiveness
2. **Monitor API usage** regularly
3. **Set usage limits** to prevent unexpected costs
4. **Use appropriate temperatures**:
   - 0.0-0.3: Deterministic, good for analysis
   - 0.5-0.7: Balanced
   - 0.8-1.0+: Creative, less suitable for analysis
5. **Cache results** to reduce API calls
6. **Batch analysis requests** when possible

## Support

For issues with:
- **OpenAI API**: Visit [OpenAI Support](https://help.openai.com/)
- **IncidentTeller**: Check logs in `./logs` directory
- **Configuration**: Review `config.yaml` documentation

## See Also

- [OpenAI API Documentation](https://platform.openai.com/docs)
- [OpenAI Pricing](https://openai.com/pricing)
- [IncidentTeller Documentation](./README.md)
