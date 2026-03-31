package workflow

import "fmt"

// StageContext holds all available context for building an AI prompt.
type StageContext struct {
	Account          string     `json:"account"`
	Campaign         string     `json:"campaign"`
	Objective        string     `json:"objective,omitempty"`
	TargetAudience   string     `json:"target_audience,omitempty"`
	Platforms        string     `json:"platforms,omitempty"`
	Tone             string     `json:"tone,omitempty"`
	PostCount        int        `json:"post_count,omitempty"`
	ImageStyle       string     `json:"image_style,omitempty"`
	Brief            string     `json:"brief,omitempty"`
	Strategy         string     `json:"strategy,omitempty"`
	CopyFiles        []CopyFile `json:"copy_files,omitempty"`
	PreviousFeedback string     `json:"previous_feedback,omitempty"`
}

// CopyFile represents a generated copy file.
type CopyFile struct {
	Name     string `json:"name"`
	Contents string `json:"contents"`
}

// buildStagePrompt constructs the AI prompt for each pipeline stage.
func buildStagePrompt(stage Stage, ctx *StageContext, req GenerateRequest) string {
	switch stage {
	case StageBrief:
		return buildBriefPrompt(ctx)
	case StageStrategy:
		return buildStrategyPrompt(ctx)
	case StageCopy:
		return buildCopyPrompt(ctx)
	case StageVisuals:
		return buildVisualsPrompt(ctx)
	case StageSchedule:
		return buildSchedulePrompt(ctx)
	case StageLaunch:
		return buildLaunchPrompt(ctx)
	default:
		return ""
	}
}

func buildBriefPrompt(ctx *StageContext) string {
	prompt := `You are a senior marketing strategist. Create a comprehensive campaign brief.` + "\n\n"
	prompt += fmt.Sprintf("**Account**: %s\n", ctx.Account)
	prompt += fmt.Sprintf("**Campaign**: %s\n", ctx.Campaign)
	if ctx.Objective != "" {
		prompt += fmt.Sprintf("**Objective**: %s\n", ctx.Objective)
	}
	if ctx.TargetAudience != "" {
		prompt += fmt.Sprintf("**Target Audience**: %s\n", ctx.TargetAudience)
	}
	if ctx.Platforms != "" {
		prompt += fmt.Sprintf("**Platforms**: %s\n", ctx.Platforms)
	}
	if ctx.Tone != "" {
		prompt += fmt.Sprintf("**Tone**: %s\n", ctx.Tone)
	}

	prompt += "\n---\n\n"
	prompt += `Write a complete campaign brief in Markdown format. Include:
1. **Campaign Overview** — What this campaign is about
2. **Objectives** — Specific, measurable goals (with KPIs)
3. **Target Audience** — Demographics, psychographics, pain points
4. **Key Messages** — 3-5 core messages to reinforce
5. **Platforms** — Which platforms and why
6. **Timeline** — Suggested duration and milestones
7. **Budget Considerations** — If applicable
8. **Brand Voice** — Tone, style, words to use/avoid

Save the brief to the campaign workspace using write_file: ` + fmt.Sprintf("`marketing/%s/%s/brief.md`\n\n", ctx.Account, ctx.Campaign)

	if ctx.PreviousFeedback != "" {
		prompt += fmt.Sprintf("**Previous feedback (address this)**:\n%s\n", ctx.PreviousFeedback)
	}

	return prompt
}

func buildStrategyPrompt(ctx *StageContext) string {
	prompt := `You are a senior content strategist. Create a comprehensive content strategy based on the campaign brief.` + "\n\n"
	prompt += fmt.Sprintf("**Account**: %s\n", ctx.Campaign)
	prompt += fmt.Sprintf("**Campaign**: %s\n", ctx.Campaign)

	if ctx.Brief != "" {
		prompt += fmt.Sprintf("\n**Existing Brief**:\n```\n%s\n```\n", truncate(ctx.Brief, 3000))
	}

	prompt += "\n---\n\n"
	prompt += `Write a complete content strategy in Markdown format. Include:
1. **Content Pillars** — 3-5 thematic pillars for all content
2. **Messaging Framework** — Primary, secondary, and supporting messages
3. **Platform Strategy** — Specific approach per platform
4. **Content Mix** — Types of content (educational, promotional, engagement, etc.)
5. **Visual Direction** — Style, color palette, mood
6. **Hashtag Strategy** — Branded + niche hashtags
7. **Content Calendar Approach** — Posting frequency per platform
8. **Success Metrics** — KPIs per platform

Save the strategy to the campaign workspace using write_file: ` + fmt.Sprintf("`marketing/%s/%s/strategy.md`\n\n", ctx.Account, ctx.Campaign)

	if ctx.PreviousFeedback != "" {
		prompt += fmt.Sprintf("**Previous feedback (address this)**:\n%s\n", ctx.PreviousFeedback)
	}

	return prompt
}

func buildCopyPrompt(ctx *StageContext) string {
	postCount := ctx.PostCount
	if postCount <= 0 {
		postCount = 5
	}
	platforms := ctx.Platforms
	if platforms == "" {
		platforms = "twitter,linkedin"
	}

	prompt := `You are an expert copywriter who deeply understands each social media platform's culture and format.` + "\n\n"
	prompt += fmt.Sprintf("**Account**: %s\n", ctx.Account)
	prompt += fmt.Sprintf("**Campaign**: %s\n", ctx.Campaign)
	prompt += fmt.Sprintf("**Platforms**: %s\n", platforms)
	prompt += fmt.Sprintf("**Posts per platform**: %d\n", postCount)
	if ctx.Tone != "" {
		prompt += fmt.Sprintf("**Tone**: %s\n", ctx.Tone)
	}

	if ctx.Strategy != "" {
		prompt += fmt.Sprintf("\n**Content Strategy**:\n```\n%s\n```\n", truncate(ctx.Strategy, 3000))
	}

	prompt += "\n---\n\n"
	prompt += fmt.Sprintf(`Write %d posts for EACH platform. Follow these platform rules:

- **Twitter/X**: Max 280 chars per tweet. Punchy, conversational, use hooks. Threads OK.
- **LinkedIn**: Max 3000 chars. Professional but human. Strong opening line. Use line breaks.
- **Facebook**: Conversational, community-focused. Max 63,206 chars.
- **Instagram**: Visual storytelling + hashtag blocks. Captions max 2,200 chars.

**IMPORTANT**: Create each post as a SEPARATE file using write_file. Use this naming pattern:
`, postCount)

	prompt += fmt.Sprintf("- `marketing/%s/%s/copy/twitter_post_1.md`\n", ctx.Account, ctx.Campaign)
	prompt += fmt.Sprintf("- `marketing/%s/%s/copy/linkedin_post_1.md`\n", ctx.Account, ctx.Campaign)
	prompt += "- etc.\n\n"
	prompt += "Each file should contain the post text ready to publish. Include relevant hashtags.\n\n"

	if ctx.PreviousFeedback != "" {
		prompt += fmt.Sprintf("**Previous feedback (address this)**:\n%s\n", ctx.PreviousFeedback)
	}

	return prompt
}

func buildVisualsPrompt(ctx *StageContext) string {
	prompt := `You are a creative director who generates compelling visual assets for marketing campaigns.` + "\n\n"
	prompt += fmt.Sprintf("**Account**: %s\n", ctx.Account)
	prompt += fmt.Sprintf("**Campaign**: %s\n", ctx.Campaign)

	if ctx.ImageStyle != "" {
		prompt += fmt.Sprintf("**Visual Style**: %s\n", ctx.ImageStyle)
	}

	// Include copy files for context
	if len(ctx.CopyFiles) > 0 {
		prompt += "\n**Existing Copy (for visual context)**:\n"
		for _, f := range ctx.CopyFiles {
			prompt += fmt.Sprintf("- %s: %s\n", f.Name, truncate(f.Contents, 200))
		}
	}

	if ctx.Strategy != "" {
		prompt += fmt.Sprintf("\n**Visual Direction from Strategy**:\n```\n%s\n```\n", truncate(ctx.Strategy, 2000))
	}

	prompt += "\n---\n\n"
	prompt += fmt.Sprintf(`Generate visual assets for this campaign:

1. **Hero Banner** — A wide landscape banner (1792x1024) for the campaign
2. **Platform-specific cards** — One image per post from the copy folder

Use image_generate for each asset. ALWAYS specify output_dir as:
`+"`marketing/%s/%s/assets/images`"+`

Image prompt best practices:
- Be specific about style: "clean minimal tech aesthetic", "vibrant startup culture"
- Specify composition: "wide landscape hero banner", "square social media card"
- Include mood: "confident, forward-looking", "warm and approachable"
- Do NOT include text in image prompts — AI struggles with text rendering

After generating all images, write an assets-manifest.md listing each image with its intended use.
`, ctx.Account, ctx.Campaign)

	if ctx.PreviousFeedback != "" {
		prompt += fmt.Sprintf("\n**Previous feedback (address this)**:\n%s\n", ctx.PreviousFeedback)
	}

	return prompt
}

func buildSchedulePrompt(ctx *StageContext) string {
	platforms := ctx.Platforms
	if platforms == "" {
		platforms = "twitter,linkedin"
	}

	prompt := `You are a social media manager who maximizes content reach and engagement.` + "\n\n"
	prompt += fmt.Sprintf("**Account**: %s\n", ctx.Account)
	prompt += fmt.Sprintf("**Campaign**: %s\n", ctx.Campaign)
	prompt += fmt.Sprintf("**Platforms**: %s\n", platforms)

	// List available copy files
	if len(ctx.CopyFiles) > 0 {
		prompt += "\n**Available Copy Files**:\n"
		for _, f := range ctx.CopyFiles {
			prompt += fmt.Sprintf("- copy/%s\n", f.Name)
		}
	}

	prompt += "\n---\n\n"
	prompt += fmt.Sprintf(`Create a 2-week posting schedule in JSON format. Save it to:
`+"`marketing/%s/%s/schedules/schedule.json`"+`

Use this format:
`+"```json"+`
{
  "campaign": "`+ctx.Campaign+`",
  "account": "`+ctx.Account+`",
  "start_date": "YYYY-MM-DD",
  "end_date": "YYYY-MM-DD",
  "posts": [
    {
      "platform": "twitter",
      "content_file": "copy/twitter_post_1.md",
      "image_file": "assets/images/twitter-card-1.png",
      "scheduled_time": "YYYY-MM-DDTHH:MM:SSZ",
      "status": "scheduled"
    }
  ]
}
`+"```"+`

**Optimal posting times (UTC)**:
- Twitter: 12:00-15:00, 18:00-20:00
- LinkedIn: 08:00-10:00 Tue-Thu
- Facebook: 13:00-16:00
- Instagram: 11:00-13:00, 19:00-21:00

Space posts across the 2 weeks. Don't cluster all posts on the same day.
`, ctx.Account, ctx.Campaign)

	if ctx.PreviousFeedback != "" {
		prompt += fmt.Sprintf("\n**Previous feedback (address this)**:\n%s\n", ctx.PreviousFeedback)
	}

	return prompt
}

func buildLaunchPrompt(ctx *StageContext) string {
	prompt := `You are a marketing analyst who turns data into actionable insights.` + "\n\n"
	prompt += fmt.Sprintf("**Account**: %s\n", ctx.Account)
	prompt += fmt.Sprintf("**Campaign**: %s\n", ctx.Campaign)

	if ctx.Brief != "" {
		prompt += fmt.Sprintf("\n**Campaign Brief**:\n```\n%s\n```\n", truncate(ctx.Brief, 2000))
	}
	if ctx.Strategy != "" {
		prompt += fmt.Sprintf("\n**Content Strategy**:\n```\n%s\n```\n", truncate(ctx.Strategy, 2000))
	}
	if len(ctx.CopyFiles) > 0 {
		prompt += "\n**Generated Copy**:\n"
		for _, f := range ctx.CopyFiles {
			prompt += fmt.Sprintf("- %s (%d chars)\n", f.Name, len(f.Contents))
		}
	}

	prompt += "\n---\n\n"
	prompt += fmt.Sprintf(`Generate a pre-launch campaign report. Save it to:
`+"`marketing/%s/%s/analytics/pre-launch-report.md`"+`

Include:
1. **Executive Summary** — Campaign overview in 3 sentences
2. **Content Inventory** — What was created (posts per platform, images, schedule)
3. **Posting Schedule Summary** — Key dates and milestones
4. **Expected KPIs** — Projected metrics with industry benchmarks:
   - Engagement rate per platform
   - Estimated reach
   - Click-through rate
   - Conversion estimates
5. **Success Metrics** — Clear targets for the campaign
6. **Risk Assessment** — Potential issues and mitigation
7. **Recommendations** — 3-5 actionable next steps

Use markdown tables for metrics. Be specific with numbers.
`, ctx.Account, ctx.Campaign)

	if ctx.PreviousFeedback != "" {
		prompt += fmt.Sprintf("\n**Previous feedback (address this)**:\n%s\n", ctx.PreviousFeedback)
	}

	return prompt
}

// truncate limits a string to maxLen characters with ellipsis.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "\n... (truncated)"
}
