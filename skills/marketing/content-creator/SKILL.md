---
name: content-creator
description: Create platform-optimized social media content with character limits, hashtag strategies, and image sizing
metadata: {"nanobot":{"emoji":"✍️"}}
---

# Content Creator

Reference guide for writing platform-optimized social media content. Covers character limits, hashtag strategy, image sizing, copywriting formulas, and cross-posting rules.

## Platform Character Limits

Respect these limits strictly. Truncated posts look unprofessional.

| Platform    | Max Characters | Visible Before Fold       | Ideal Length for Engagement |
|-------------|----------------|---------------------------|-----------------------------|
| Twitter/X   | 280            | All visible               | 71-100 characters           |
| LinkedIn    | 3,000          | First 140 before "see more" | 1,300-2,000 characters     |
| Instagram   | 2,200          | First 125 before "more"   | 138-150 characters          |
| Facebook    | 63,206         | Varies by device           | 40-80 characters            |
| TikTok      | 4,000          | First ~100 in feed        | 100-150 characters          |

### Twitter/X Threading

For content that exceeds 280 characters, use threads:
- First tweet must stand alone and hook the reader
- Number tweets (1/7, 2/7, etc.) for long threads
- End with a CTA or summary tweet
- Pin the first tweet if it is a key message

## Hashtag Best Practices

### Per-Platform Strategy

**Twitter/X**: 2-3 hashtags maximum
- Mix one trending hashtag with one or two niche hashtags
- Place at the end of the tweet or inline if natural
- Research trending tags with `web_search query="trending twitter hashtags [topic] 2026"`

**Instagram**: 5-10 relevant hashtags (up to 30 allowed)
- Mix: 2-3 popular (100K+ posts), 3-4 medium (10K-100K), 2-3 niche (<10K)
- Place in first comment or at the end of caption after line breaks
- Create a branded hashtag for the campaign

**LinkedIn**: 3-5 professional hashtags
- Use industry-standard tags (#SaaS, #DevOps, #AI)
- Include one branded hashtag
- Place at the end of the post

**Facebook**: 1-2 hashtags or none at all
- Hashtags are less effective on Facebook
- Only use if running a branded campaign with a specific hashtag
- Over-tagging looks spammy here

**TikTok**: 3-5 hashtags
- Research trending hashtags on the platform before posting
- Mix trending with niche for discoverability
- #FYP and #ForYouPage are overused -- focus on topic-specific tags

## Image Size Guide

Use these sizes with the `image_generate` tool. The tool supports three sizes: `1024x1024`, `1792x1024`, and `1024x1792`.

| Platform         | Optimal Size | image_generate Size | Orientation |
|------------------|-------------|---------------------|-------------|
| Twitter post     | 1200x675    | `size="1792x1024"`  | Landscape   |
| Instagram feed   | 1080x1080   | `size="1024x1024"`  | Square      |
| Instagram story  | 1080x1920   | `size="1024x1792"`  | Portrait    |
| LinkedIn post    | 1200x627    | `size="1792x1024"`  | Landscape   |
| Facebook post    | 1200x630    | `size="1792x1024"`  | Landscape   |
| TikTok cover     | 1080x1920   | `size="1024x1792"`  | Portrait    |

### Example: Generate a Twitter Banner

```
image_generate prompt="Clean tech product announcement, dark blue gradient, subtle grid pattern, space for headline text on the left, abstract geometric shapes on the right" size="1792x1024"
```

### Example: Generate an Instagram Square

```
image_generate prompt="Minimalist quote card, white background, bold sans-serif typography, accent color coral, clean and modern" size="1024x1024"
```

## Copywriting Formulas

Use these proven structures to write compelling copy.

### AIDA: Attention, Interest, Desire, Action

Best for: product launches, announcements, promotional posts.

```
Attention: "Your CI pipeline is burning 40% of your cloud budget."
Interest: "We analyzed 500 pipelines and found 3 patterns that waste resources."
Desire: "Teams that fixed these cut build costs by half in one sprint."
Action: "Read the full breakdown -- link in bio."
```

### PAS: Problem, Agitate, Solve

Best for: pain point-driven content, B2B messaging.

```
Problem: "Most Go agents eat 500MB+ of RAM before they do anything useful."
Agitate: "That rules out edge devices, Raspberry Pis, and half the cloud instances people actually use."
Solve: "MakoClaw runs on <10MB. Same capabilities, 50x lighter."
```

### BAB: Before, After, Bridge

Best for: transformation stories, case studies.

```
Before: "We were managing 5 Slack bots, 3 cron jobs, and a custom API gateway."
After: "Now it is one MakoClaw agent with 9 channels, built-in scheduling, and a web UI."
Bridge: "Here is how we migrated in a weekend."
```

### 4U: Useful, Urgent, Unique, Ultra-Specific

Best for: headlines and hook lines.

- **Useful**: Does this help the reader?
- **Urgent**: Is there a reason to act now?
- **Unique**: Is this different from what everyone else says?
- **Ultra-specific**: Does it include concrete numbers or details?

```
Weak: "Build better AI agents"
Strong: "Ship a Go AI agent to a Raspberry Pi in under 5 minutes -- no GPU required"
```

## CTA Patterns by Platform

Every post should have a clear call-to-action adapted to the platform.

**Twitter/X**:
- "RT if you agree"
- "Drop your hot take below"
- "Thread below (bookmark for later)"
- "Link in bio" or direct URL

**LinkedIn**:
- "What has your experience been with...?"
- "Share your thoughts in the comments"
- "Follow for more on [topic]"
- "Agree or disagree? Let me know why"

**Instagram**:
- "Save this for later"
- "Tag someone who needs to see this"
- "Double tap if you agree"
- "Link in bio for the full guide"

**Facebook**:
- "Share with your team"
- "What would you add to this list?"
- "Join our community [link]"

**TikTok**:
- "Follow for part 2"
- "Stitch this with your experience"
- "Comment what you want to see next"

## Cross-Posting Rules

NEVER copy-paste the same content across platforms. Each platform has a different audience, tone, and format.

### Adapt Tone

| Platform  | Tone              | Example Opening                                    |
|-----------|-------------------|----------------------------------------------------|
| Twitter   | Casual, direct    | "Hot take: most agent frameworks are bloated."     |
| LinkedIn  | Professional      | "After 3 years building agent systems, here is what I have learned about resource efficiency." |
| Instagram | Visual, lifestyle | (Lead with a striking image, short punchy caption) |
| Facebook  | Community         | "Who else has run into this problem?"              |

### Adapt Length

- Twitter: get to the point in 1-2 sentences
- LinkedIn: open with a hook, develop the argument, close with a question
- Instagram: short caption that complements the visual
- Facebook: medium length, conversational

### Adapt Media

- Twitter/LinkedIn: landscape images (1792x1024), GIFs, link previews
- Instagram: square (1024x1024) for feed, portrait (1024x1792) for stories
- TikTok: portrait video (1024x1792)

### Example: Same Message, Three Platforms

**Core message**: MakoClaw ships a new multi-agent orchestration feature.

**Twitter**:
```
social_post action=preview platforms=["twitter"] content="MakoClaw now supports multi-agent orchestration. Define specialists, set keywords, and let the orchestrator route conversations automatically. One config file. Zero boilerplate." hashtags=["golang","AI","agents"]
```

**LinkedIn**:
```
social_post action=preview platforms=["linkedin"] content="We just shipped multi-agent orchestration in MakoClaw.\n\nThe problem: as AI agent use cases grow, a single agent with dozens of tools becomes slow and unfocused.\n\nOur solution: define specialist agents with their own tools, models, and system prompts. The orchestrator analyzes incoming messages and routes them to the right specialist automatically.\n\nConfiguration is a single JSON file. No custom code required.\n\nWhat patterns are you seeing in multi-agent systems? I would love to hear what has worked for your teams." hashtags=["AI","AgentFrameworks","Golang"]
```

**Instagram**:
```
social_post action=preview platforms=["instagram"] content="Multi-agent orchestration just dropped.\n\nOne orchestrator. Multiple specialists. Zero boilerplate.\n\nLink in bio for the docs." hashtags=["golang","ai","coding","opensourceai","devtools"]
```
