---
name: marketing
description: Full-stack marketing campaign system. Configures 5 specialist agents (content strategist, copywriter, visual designer, social media manager, analyst) and provides a complete campaign workflow template.
emoji: 📣
---

# Marketing Campaign System

This skill configures MakoClaw as a complete marketing agency, with 5 specialist agents that collaborate to create and publish end-to-end marketing campaigns.

## Campaign Workspace Structure

All campaign outputs are organized under `marketing/{account}/{campaign}/`:

```
marketing/
  {account}/
    {campaign}/
      brief.md           — Campaign brief (auto-created by marketing_init_campaign)
      strategy.md        — Content strategy (content_strategist)
      copy/
        twitter_post_1.md
        linkedin_post_1.md
        ...
      assets/
        images/          — AI-generated visuals (visual_designer)
        videos/
      schedules/
        schedule.json    — Posting schedule (social_media_manager)
      analytics/
        pre-launch-report.md
        performance-report.md
```

## Specialist Agents

Add these specialists to your `~/.MakoClaw/users/{uuid}/config.json` under `agents.specialists`:

See `specialists-config.json` in this skill folder for ready-to-use config.

### content_strategist
Creates campaign briefs, content pillars, and messaging frameworks. Entry point for any new campaign.

### copywriter
Writes platform-optimized copy: Twitter threads, LinkedIn posts, Facebook updates, Instagram captions, email copy. Understands character limits and platform culture.

### visual_designer
Generates visual assets using `image_generate`. Uses `output_dir` parameter to save directly to campaign assets folder.

### social_media_manager
Schedules and publishes content across platforms using `social_post`. Creates JSON schedules for future posts.

### campaign_analyst
Tracks performance with `social_analytics`. Generates pre-launch estimates and post-campaign ROI reports.

## Starting a Campaign

### Via Chat (Orchestrator)
```
Create a marketing campaign for Acme Corp. Campaign: product-launch-2026.
Objective: Launch new SaaS product to tech startups.
Target audience: CTOs and engineering leads at Series A-B startups.
Platforms: LinkedIn and Twitter.
Tone: technical but approachable.
Generate 5 posts per platform and visual assets.
```

### Via Tool
Use `marketing_init_campaign` to initialize workspace:
- `account`: company slug (e.g. "acme-corp")
- `campaign`: campaign slug (e.g. "product-launch-2026")
- `description`: campaign overview
- `platforms`: "twitter,linkedin"
- `objective`: "product launch"

## Image Generation

Configure your preferred image provider in config:

```json
{
  "tools": {
    "image": {
      "provider": "fal",
      "api_key": "YOUR_FAL_API_KEY",
      "model": "fal-ai/flux/dev"
    }
  }
}
```

Available providers:
- `fal` / `fal.ai` — fal.ai FLUX models (recommended: fast, high quality)
- `replicate` — Replicate.com (wide model selection)
- `together` — Together.ai FLUX (free tier available: `black-forest-labs/FLUX.1-schnell-Free`)
- `google` / `imagen` — Google Imagen via AI Studio (`gemini-2.0-flash-exp`)
- `openai` — DALL-E 3 (default if no provider specified)

## Campaign Workflow

Import `campaign-workflow.json` from this skill via Workflows → Import to run automated end-to-end campaigns.

Parameters:
- `account`: Brand/company slug
- `campaign`: Campaign identifier slug
- `objective`: What the campaign should achieve
- `target_audience`: Who you're targeting
- `platforms`: Comma-separated platforms (default: "twitter,linkedin")
- `tone`: Content tone (default: "professional")
- `post_count`: Posts per platform (default: "5")
