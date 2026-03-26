---
name: campaign-planner
description: Guide agents through marketing campaign planning with briefs, audience targeting, KPIs, and content calendars
metadata: {"nanobot":{"emoji":"📋"}}
---

# Campaign Planner

Structured workflow for planning and executing marketing campaigns across social media platforms using MakoClaw tools.

## Campaign Brief Template

Before creating any content, complete the campaign brief. This is the foundation for every campaign.

### Required Fields

1. **Objective** -- What are we trying to achieve?
   - Brand awareness (reach new audiences)
   - Lead generation (drive signups, downloads)
   - Engagement (build community, increase interactions)
   - Conversions (sales, subscriptions)
   - Thought leadership (establish authority in a space)

2. **Target Audience**
   - Demographics: age range, location, job title/industry
   - Interests: topics they follow, communities they belong to
   - Pain points: problems they need solved
   - Online behavior: which platforms they use, when they are active

3. **Platforms** -- Select based on audience match (see Platform Selection Matrix below)

4. **Key Messages** -- 3 to 5 core messages that every piece of content should reinforce. Keep them concise and consistent.

5. **KPIs** -- Measurable targets for the campaign:
   - Engagement rate (likes + comments + shares / impressions)
   - Click-through rate (CTR)
   - Conversions (signups, purchases)
   - Reach (unique accounts that see the content)
   - Impressions (total times content is displayed)

6. **Timeline**
   - Start date
   - End date
   - Key milestones (content drops, launches, events)

7. **Budget** -- If applicable, define spend per platform or total.

8. **Brand Voice/Tone** -- Define how the brand should sound:
   - Professional vs. casual
   - Technical vs. accessible
   - Authoritative vs. conversational
   - Any words or phrases to always use or always avoid

## Platform Selection Matrix

Choose platforms based on where the target audience spends time, not on personal preference.

| Platform    | Best For                                          | Audience            | Character Limit | Content Style              |
|-------------|---------------------------------------------------|---------------------|-----------------|----------------------------|
| Twitter/X   | Real-time engagement, news, tech communities      | Tech, media, B2B    | 280             | Short, punchy, threaded    |
| LinkedIn    | B2B marketing, professional networking, recruiting | Professionals, B2B  | 3,000           | Thought leadership, formal |
| Instagram   | Visual storytelling, lifestyle, product showcase   | 18-34, lifestyle    | 2,200 caption   | Visual-first, casual       |
| Facebook    | Broad reach, community building, events            | 25-55, general      | 63,206          | Community, conversational  |
| TikTok      | Short video, viral content, trend riding           | Gen Z, Millennial   | 4,000           | Authentic, trend-driven    |

### Platform Decision Guide

- **B2B product launch**: LinkedIn (primary) + Twitter/X (secondary)
- **Consumer product**: Instagram (primary) + TikTok (secondary) + Facebook (retargeting)
- **Developer tools**: Twitter/X (primary) + LinkedIn (secondary)
- **Local business**: Facebook (primary) + Instagram (secondary)
- **Youth/entertainment**: TikTok (primary) + Instagram (secondary)

## Campaign Workflow

Follow these steps in order. Do not skip ahead.

### Step 1: Define Objective and Audience

Fill out the Campaign Brief Template above. Be specific -- "increase brand awareness" is too vague. "Reach 10,000 developers in the Go community within 30 days" is actionable.

### Step 2: Select Platforms

Use the Platform Selection Matrix. Pick 2-3 platforms maximum for a focused campaign. Spreading across too many platforms dilutes effort.

### Step 3: Create Content Calendar

Plan content by date, platform, and content type. Example:

| Date       | Platform  | Content Type | Topic                | Status  |
|------------|-----------|--------------|----------------------|---------|
| 2026-04-01 | Twitter   | Thread       | Product announcement | Draft   |
| 2026-04-01 | LinkedIn  | Article      | Deep dive on feature | Draft   |
| 2026-04-02 | Instagram | Carousel     | Feature highlights   | Planned |
| 2026-04-03 | Twitter   | Poll         | Audience feedback    | Planned |

Use `task_manager` to track content items:

```
task_manager action=create title="Write Twitter thread: product announcement" description="Part of Q2 launch campaign. Target: Go developers. Deadline: April 1." status=todo
```

### Step 4: Generate Visual Assets

Use `image_generate` for campaign imagery:

```
image_generate prompt="Professional tech product announcement banner, clean design, blue gradient background, minimal text space, modern typography" size="1792x1024"
```

Generate platform-specific sizes as needed (see content-creator skill for size guide).

### Step 5: Write Platform-Specific Copy

Adapt content for each platform. NEVER copy-paste the same text across platforms. Each platform has different character limits, audience expectations, and content styles.

Use the content-creator skill for copywriting formulas and platform-specific guidelines.

### Step 6: Preview All Posts

Always preview before publishing:

```
social_post action=preview platforms=["twitter","linkedin"] content="We just shipped the fastest Go agent framework. 10x lighter than alternatives, built for edge hardware." hashtags=["golang","opensource","AI"]
```

Review the preview for:
- Character count within limits
- Hashtags are appropriate for each platform
- Media attached correctly
- No typos or formatting issues

### Step 7: Get Approval and Publish

After review, publish with confirmation:

```
social_post action=post platforms=["twitter"] content="We just shipped the fastest Go agent framework. 10x lighter than alternatives, built for edge hardware." hashtags=["golang","opensource"] confirmed=true
```

For scheduled posts:

```
social_post action=schedule platforms=["linkedin"] content="..." schedule_time="2026-04-01T09:00:00Z" confirmed=true
```

### Step 8: Track Performance

Monitor results with `social_analytics`:

```
social_analytics post_id="abc123" platform="twitter"
social_analytics platform="twitter" timeframe="7d"
```

### Step 9: Optimize

Based on analytics data:
- Double down on content types with high engagement
- Adjust posting times based on when audience is active
- Refine messaging based on what resonates
- Reallocate budget to best-performing platforms

## Example Specialist Configuration

Users who want a dedicated marketing agent can add this specialist to their config:

```json
{
  "name": "marketing_manager",
  "description": "Plans and executes marketing campaigns across social media",
  "keywords": ["campaign", "marketing", "promote", "social media", "post"],
  "tools": ["image_generate", "social_post", "social_analytics", "web_search", "web_fetch", "task_manager"],
  "skills": ["campaign-planner", "content-creator", "social-media"]
}
```

This agent will have access to all marketing tools and skills, and can be invoked with `@marketing_manager` or triggered automatically by the orchestrator when marketing-related keywords are detected.
