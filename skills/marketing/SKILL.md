---
name: marketing
description: Full-stack marketing campaign system. Configures 5 specialist agents (content strategist, copywriter, visual designer, social media manager, analyst) and provides a complete campaign workflow, email campaign management with audience segmentation, per-campaign memory, version history, and company auto-research.
when_to_use: When planning a marketing campaign, creating content strategy, writing copy, managing social media posts, analyzing campaign performance, managing email campaigns or audience contacts, or researching a company/account profile
emoji: 📣
---

# Marketing Campaign System

This skill configures MakoClaw as a complete marketing agency, with 5 specialist agents that collaborate to create and publish end-to-end marketing campaigns. It also provides a full **Audience & Email Campaign** module accessible via the Web UI.

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

---

## Audience & Email Campaigns (Web UI)

The **Audience** tab in the Web UI provides a full email marketing suite.

### Email Campaigns

Each campaign has a name, subject, HTML body, optional template slug, and an assigned contact list.

- **Create**: POST `/api/v1/marketing/audience/email-campaigns`
- **Send** (async): POST `.../email-campaigns/{id}/send` → returns 202 immediately; delivery runs in background
- **Progress polling**: GET `.../email-campaigns/{id}/progress` → `{ status, progress: "{pct, sent, total, ...}" }`
- Statuses: `draft → sending → delivering → sent / failed`
- Archived campaigns are hidden from the default view; accessible via filter

### Per-Campaign Memory

Persistent notes and agent context attached to a specific campaign. Survives across sessions.

- **List**: GET `.../email-campaigns/{id}/memory`
- **Add entry**: POST `.../email-campaigns/{id}/memory` `{ role: "note"|"decision"|"agent", content: "..." }`
- **Delete entry**: DELETE `.../email-campaigns/{id}/memory/{entryId}`

Roles:
- `note` — free-form annotation
- `decision` — a choice made about the campaign (audience, tone, schedule)
- `agent` — context injected by or for an AI agent during a workflow run

### Version History

Snapshots of campaign content (subject + body) with auto-versioning on every save.

- **List versions**: GET `.../email-campaigns/{id}/versions` → array ordered newest-first
- **Manual snapshot**: POST `.../email-campaigns/{id}/versions` `{ note: "optional label" }`
- **Restore version**: POST `.../email-campaigns/{id}/versions/{versionId}/restore`

Auto-versioning: whenever `UpdateCampaign` detects a change in subject, body_html, or body_text, it automatically snapshots the previous content before overwriting.

Version object fields: `id`, `campaign_id`, `version_number`, `subject`, `body_html`, `body_text`, `note`, `created_at`.

### Company / Account Profiles

Research and store company context per account slug. Used for personalizing campaigns.

- **List all profiles**: GET `/api/v1/marketing/profiles`
- **Get profile**: GET `/api/v1/marketing/profiles/{account}`
- **Save / update**: PUT `/api/v1/marketing/profiles/{account}` `{ website, industry, description, target_audience, social_links, research_notes }`
- **Auto-research** (AI agent): POST `/api/v1/marketing/profiles/{account}/research` `{ website? }` → runs `web_search` + `web_fetch` and upserts the profile; returns the updated profile

The research endpoint spawns a single-turn agent that queries the web for the company, parses the JSON response, and stores it. It uses a 90s timeout. The `social_links` field stores a JSON object (`{ twitter, linkedin, instagram, ... }`).

### Audience Contacts, Lists & Segments

- **Contacts**: CRUD + CSV import/export at `/api/v1/marketing/audience/contacts`
- **Lists**: Named contact groups at `/api/v1/marketing/audience/lists`; add/remove members via `.../lists/{id}/members`
- **Segments**: Rule-based dynamic groups at `/api/v1/marketing/audience/segments`; preview with `.../segments/{id}/preview`
