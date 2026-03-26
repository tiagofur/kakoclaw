---
name: social-media
description: Social media publishing best practices, optimal posting times, and engagement strategies
metadata: {"nanobot":{"emoji":"📱"}}
---

# Social Media

Best practices for publishing, scheduling, and managing social media content using MakoClaw tools.

## Publishing Workflow

Follow this workflow for every post. No exceptions.

### Step 1: Preview First

ALWAYS preview before publishing. This catches character overflows, formatting issues, and missing media.

```
social_post action=preview platforms=["twitter","linkedin"] content="Your post content here" hashtags=["tag1","tag2"]
```

Review the preview output for:
- Character count is within platform limits
- Content renders correctly (line breaks, special characters)
- Hashtags are appropriate and correctly formatted
- No typos

### Step 2: Attach Media (If Needed)

Add images, videos, or generated assets:

```
social_post action=preview platforms=["instagram"] content="Launch day." hashtags=["release","opensource"] media_urls=["workspace/generated-images/launch-banner.png"]
```

Use paths relative to the workspace. Images generated with `image_generate` are saved to the `generated-images/` directory by default.

### Step 3: Publish

Only when satisfied with the preview, publish with confirmation:

```
social_post action=post platforms=["twitter"] content="Your finalized content" hashtags=["tag1","tag2"] confirmed=true
```

The `confirmed=true` flag is required. Without it, the tool will not publish.

### Step 4: Schedule Future Posts

For content calendars, schedule posts in advance:

```
social_post action=schedule platforms=["linkedin"] content="Scheduled post content" hashtags=["planned","content"] schedule_time="2026-04-01T10:00:00Z" confirmed=true
```

Use ISO 8601 format for schedule times. Always specify UTC.

### Step 5: Verify Publication

After posting, check that the post went live:

```
social_analytics post_id="returned_post_id" platform="twitter"
```

## Optimal Posting Times

These are general guidelines based on aggregate engagement data. Actual optimal times vary by audience. Use analytics to refine.

All times in UTC.

| Platform  | Best Days           | Best Times (UTC) | Worst Times        |
|-----------|---------------------|-------------------|--------------------|
| Twitter   | Monday-Friday       | 9-11 AM, 1-3 PM  | Weekends, late night |
| LinkedIn  | Tuesday-Thursday    | 8-10 AM, 12 PM   | Weekends, evenings |
| Instagram | Monday-Friday       | 11 AM-1 PM, 7-9 PM | 3-5 AM            |
| Facebook  | Wednesday-Friday    | 1-4 PM            | Early morning      |
| TikTok    | Tuesday-Thursday    | 2-5 PM, 7-9 PM   | Early morning      |

### Scheduling Example: Optimal Week

```
social_post action=schedule platforms=["twitter"] content="Monday morning insight..." schedule_time="2026-04-06T09:30:00Z" confirmed=true

social_post action=schedule platforms=["linkedin"] content="Tuesday deep dive..." schedule_time="2026-04-07T08:00:00Z" confirmed=true

social_post action=schedule platforms=["instagram"] content="Wednesday visual..." media_urls=["workspace/generated-images/mid-week.png"] schedule_time="2026-04-08T11:30:00Z" confirmed=true
```

## Content Frequency

Post consistently but do not burn out the audience. Quality over quantity.

| Platform  | Recommended Frequency    | Notes                                      |
|-----------|--------------------------|--------------------------------------------|
| Twitter   | 3-5 tweets per day       | Mix original content, retweets, and replies |
| LinkedIn  | 2-5 posts per week       | Long-form performs well, space them out     |
| Instagram | 3-7 posts per week       | Add daily stories for engagement            |
| Facebook  | 3-5 posts per week       | Prioritize community posts and events       |
| TikTok    | 1-3 videos per day       | Consistency matters more than polish        |

### Frequency Management with Tasks

Track publishing cadence using the task manager:

```
task_manager action=create title="Twitter: daily post batch (Mon-Fri)" description="Prepare and schedule 5 tweets for the week. Focus on product updates and community engagement." status=todo
```

## Engagement Strategy

Publishing is half the job. Engagement is the other half.

### Response Timing

- Respond to comments within 1 hour of posting when possible
- The first hour after posting is critical for algorithmic reach
- Set up monitoring for mentions and replies

### Driving Engagement

- **Ask questions**: End posts with a question to invite responses
- **Use polls**: Twitter and LinkedIn polls drive high engagement with low effort
- **Share user-generated content**: Repost, quote-tweet, or highlight community contributions
- **Engage with industry accounts**: Comment on posts from accounts in the same space
- **Use interactive features**: Instagram stories polls, quizzes, and question stickers

### Example: Engagement-Optimized Post

```
social_post action=preview platforms=["twitter"] content="We asked 100 developers what they hate most about AI agent frameworks.\n\nThe top 3 answers:\n1. Memory usage\n2. Startup time\n3. Too many dependencies\n\nWhat would your answer be?" hashtags=["devtools","AI"]
```

This post uses a list format (easy to scan), references a concrete data point, and ends with a direct question.

## Analytics Tracking

### Checking Post Performance

After publishing, track how content performs:

```
social_analytics post_id="abc123" platform="twitter"
```

### Platform-Wide Metrics

Review overall performance for a time period:

```
social_analytics platform="twitter" timeframe="7d"
social_analytics platform="linkedin" timeframe="30d"
```

### Key Metrics to Track

| Metric           | Formula                                      | Good Benchmark    |
|------------------|----------------------------------------------|-------------------|
| Engagement rate  | (likes + comments + shares) / impressions    | >1% Twitter, >2% LinkedIn |
| Click-through rate (CTR) | clicks / impressions                  | >0.5%             |
| Reach            | unique accounts that saw the post            | Growing week over week |
| Follower growth  | new followers per period                     | Steady upward trend |

### A/B Testing

Test variations to find what works:

- Same content, different images
- Same content, different hashtags
- Same content, different posting times
- Different copy formulas (AIDA vs PAS)

Track results:

```
social_analytics post_id="version_a_id" platform="twitter"
social_analytics post_id="version_b_id" platform="twitter"
```

Compare engagement rates to determine the winner. Apply learnings to future content.

### Weekly Review Process

Every week, review the past 7 days of analytics:

1. Pull metrics: `social_analytics platform="twitter" timeframe="7d"`
2. Identify top-performing posts (highest engagement rate)
3. Identify underperformers (lowest engagement rate)
4. Look for patterns: what content types, posting times, and hashtags correlate with performance
5. Adjust the content calendar for the next week based on findings

## Content Compliance

Follow these rules for every post. Non-compliance can result in account suspension or legal issues.

### Disclosure Requirements

- **AI-generated content**: Disclose when images or text are AI-generated, where platform policies or local regulations require it
- **Sponsored content**: Always include #ad or #sponsored for paid promotions
- **Affiliate links**: Disclose affiliate relationships clearly
- **Partnerships**: Use #partner or equivalent disclosure

### Platform Terms of Service

- Do not use engagement bait tactics that violate platform rules
- Do not post misleading claims or fake testimonials
- Do not use automated tools in ways that violate rate limits
- Do not impersonate other accounts or brands

### Content Standards

- Verify claims before posting -- do not state statistics without a source
- Do not post personal or confidential information
- Respect copyright on images, quotes, and shared content
- Use `web_search` to verify facts when uncertain:

```
web_search query="latest statistics on Go language adoption 2026"
```
