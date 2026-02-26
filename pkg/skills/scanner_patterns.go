package skills

import "regexp"

// DangerousCommandPatterns detects dangerous shell commands
var DangerousCommandPatterns = []PatternRule{
	// Critical: Immediate rejection - System destruction
	{
		ID:          "CMD001",
		Pattern:     regexp.MustCompile(`(?i)\brm\s+(-[rRfF]+\s+)*(/|~|\$HOME|\$\{HOME\}|/home|/root|/etc|/var|/usr)\b`),
		Severity:    SeverityCritical,
		Category:    "dangerous_command",
		Title:       "Dangerous recursive delete",
		Description: "Command could delete critical system directories",
		Suggestion:  "Restrict rm operations to specific, safe directories within the workspace",
	},
	{
		ID:          "CMD002",
		Pattern:     regexp.MustCompile(`(?i)\b(mkfs|format|diskpart|fdisk)\b`),
		Severity:    SeverityCritical,
		Category:    "dangerous_command",
		Title:       "Disk formatting command",
		Description: "Command could destroy disk data",
		Suggestion:  "Remove disk formatting commands from skills",
	},
	{
		ID:          "CMD003",
		Pattern:     regexp.MustCompile(`(?i)\bdd\s+.*of=/dev/`),
		Severity:    SeverityCritical,
		Category:    "dangerous_command",
		Title:       "Raw disk write",
		Description: "dd writing to device could destroy disk data",
		Suggestion:  "Remove raw disk write commands",
	},
	{
		ID:          "CMD004",
		Pattern:     regexp.MustCompile(`:\(\)\s*\{.*\};\s*:`),
		Severity:    SeverityCritical,
		Category:    "dangerous_command",
		Title:       "Fork bomb pattern",
		Description: "Detected potential fork bomb that could crash the system",
		Suggestion:  "Remove this pattern immediately",
	},

	// High: Remote code execution
	{
		ID:          "CMD005",
		Pattern:     regexp.MustCompile(`(?i)\b(curl|wget)\s+[^\|]*\|\s*(bash|sh|zsh|python|perl|ruby)\b`),
		Severity:    SeverityHigh,
		Category:    "dangerous_command",
		Title:       "Remote code execution pattern",
		Description: "Piping curl/wget to shell is dangerous and can execute arbitrary code",
		Suggestion:  "Download scripts first, review them, then execute separately",
	},
	{
		ID:          "CMD006",
		Pattern:     regexp.MustCompile(`(?i)\beval\s*\(`),
		Severity:    SeverityHigh,
		Category:    "dangerous_command",
		Title:       "Dynamic code evaluation",
		Description: "eval() can execute arbitrary code",
		Suggestion:  "Avoid eval; use safer alternatives",
	},
	{
		ID:          "CMD007",
		Pattern:     regexp.MustCompile(`(?i)\bchmod\s+(-R\s+)?777\s`),
		Severity:    SeverityHigh,
		Category:    "dangerous_command",
		Title:       "Insecure permissions",
		Description: "Setting 777 permissions is a security risk",
		Suggestion:  "Use more restrictive permissions (e.g., 755 or 644)",
	},
	{
		ID:          "CMD008",
		Pattern:     regexp.MustCompile(`(?i)(>\s*|>>)\s*/etc/`),
		Severity:    SeverityHigh,
		Category:    "system_modification",
		Title:       "System configuration modification",
		Description: "Writing to /etc could modify system configuration",
		Suggestion:  "Avoid modifying system configuration files",
	},

	// Medium: Elevated privileges
	{
		ID:          "CMD009",
		Pattern:     regexp.MustCompile(`(?i)\bsudo\s`),
		Severity:    SeverityMedium,
		Category:    "elevated_privileges",
		Title:       "Sudo command usage",
		Description: "Commands requiring elevated privileges should be carefully reviewed",
		Suggestion:  "Document why sudo is necessary and what permissions are required",
	},
	{
		ID:          "CMD010",
		Pattern:     regexp.MustCompile(`(?i)\b(su\s+-|su\s+root)\b`),
		Severity:    SeverityMedium,
		Category:    "elevated_privileges",
		Title:       "Switch to root user",
		Description: "Switching to root user is risky",
		Suggestion:  "Use sudo for specific commands instead of switching to root",
	},
}

// PromptInjectionPatterns detects attempts to manipulate the AI agent
var PromptInjectionPatterns = []PatternRule{
	// Critical: Direct instruction override
	{
		ID:          "INJ001",
		Pattern:     regexp.MustCompile(`(?i)(ignore|disregard|forget)\s+(all\s+)?(previous|above|prior|earlier)\s+(instructions?|rules?|constraints?|guidelines?)`),
		Severity:    SeverityCritical,
		Category:    "prompt_injection",
		Title:       "Prompt injection - instruction override",
		Description: "Detected attempt to override agent instructions",
		Suggestion:  "Remove instruction override attempts",
	},
	{
		ID:          "INJ002",
		Pattern:     regexp.MustCompile(`(?i)(bypass|circumvent|override|skip|ignore)\s+(safety|security|guard|filter|restriction|protection)`),
		Severity:    SeverityCritical,
		Category:    "prompt_injection",
		Title:       "Security bypass instruction",
		Description: "Detected attempt to bypass security measures",
		Suggestion:  "Remove security bypass instructions",
	},
	{
		ID:          "INJ003",
		Pattern:     regexp.MustCompile(`(?i)\b(jailbreak|DAN|do\s+anything\s+now)\b`),
		Severity:    SeverityCritical,
		Category:    "prompt_injection",
		Title:       "Known jailbreak pattern",
		Description: "Detected known AI jailbreak technique",
		Suggestion:  "Remove jailbreak instructions",
	},

	// High: Role manipulation
	{
		ID:          "INJ004",
		Pattern:     regexp.MustCompile(`(?i)(you\s+are|act\s+as|pretend\s+to\s+be|roleplay\s+as|imagine\s+you('re|\s+are))\s+(a\s+)?(different|new|another|evil|unrestricted)`),
		Severity:    SeverityHigh,
		Category:    "prompt_injection",
		Title:       "Role manipulation attempt",
		Description: "Detected attempt to change agent's identity or role",
		Suggestion:  "Remove role manipulation instructions",
	},
	{
		ID:          "INJ005",
		Pattern:     regexp.MustCompile(`(?i)(from\s+now\s+on|starting\s+now|henceforth)\s+(you|always|never)`),
		Severity:    SeverityHigh,
		Category:    "prompt_injection",
		Title:       "Behavioral override attempt",
		Description: "Attempting to permanently change agent behavior",
		Suggestion:  "Remove behavioral override instructions",
	},

	// Medium: Suspicious patterns
	{
		ID:          "INJ006",
		Pattern:     regexp.MustCompile(`(?i)(system\s+prompt|internal\s+instructions?|hidden\s+prompt|secret\s+rules?)`),
		Severity:    SeverityMedium,
		Category:    "prompt_injection",
		Title:       "System prompt reference",
		Description: "Skill references internal system prompts which could be manipulation",
		Suggestion:  "Remove references to system internals",
	},
	{
		ID:          "INJ007",
		Pattern:     regexp.MustCompile(`(?i)\b(IMPORTANT|CRITICAL|URGENT|OVERRIDE):\s*(ignore|override|bypass|you\s+must)`),
		Severity:    SeverityMedium,
		Category:    "prompt_injection",
		Title:       "Urgency-based manipulation",
		Description: "Using urgency to manipulate behavior",
		Suggestion:  "Remove urgency-based override attempts",
	},
}

// CredentialPatterns detects hardcoded credentials and secrets
var CredentialPatterns = []PatternRule{
	// API Keys
	{
		ID:          "CRED001",
		Pattern:     regexp.MustCompile(`(?i)(api[_-]?key|apikey|api[_-]?secret|api[_-]?token)\s*[=:]\s*["']?[a-zA-Z0-9_-]{16,}["']?`),
		Severity:    SeverityCritical,
		Category:    "credential_exposure",
		Title:       "Hardcoded API key",
		Description: "Detected potential hardcoded API key",
		Suggestion:  "Use environment variables instead of hardcoded credentials",
	},
	{
		ID:          "CRED002",
		Pattern:     regexp.MustCompile(`(?i)(password|passwd|pwd|secret)\s*[=:]\s*["'][^"']{4,}["']`),
		Severity:    SeverityCritical,
		Category:    "credential_exposure",
		Title:       "Hardcoded password",
		Description: "Detected hardcoded password or secret",
		Suggestion:  "Never hardcode passwords; use environment variables or secret management",
	},

	// Cloud provider credentials
	{
		ID:          "CRED003",
		Pattern:     regexp.MustCompile(`(?i)(aws_access_key_id|aws_secret_access_key)\s*[=:]\s*["']?[A-Z0-9]{16,}["']?`),
		Severity:    SeverityCritical,
		Category:    "credential_exposure",
		Title:       "AWS credentials exposed",
		Description: "Detected AWS credentials in skill",
		Suggestion:  "Use AWS credential files or IAM roles instead",
	},
	{
		ID:          "CRED004",
		Pattern:     regexp.MustCompile(`AKIA[0-9A-Z]{16}`),
		Severity:    SeverityCritical,
		Category:    "credential_exposure",
		Title:       "AWS Access Key ID",
		Description: "Detected AWS Access Key ID pattern",
		Suggestion:  "Remove AWS credentials; use IAM roles or credential files",
	},

	// Tokens
	{
		ID:          "CRED005",
		Pattern:     regexp.MustCompile(`ghp_[a-zA-Z0-9]{36}`),
		Severity:    SeverityCritical,
		Category:    "credential_exposure",
		Title:       "GitHub token exposed",
		Description: "Detected GitHub personal access token",
		Suggestion:  "Use GITHUB_TOKEN environment variable",
	},
	{
		ID:          "CRED006",
		Pattern:     regexp.MustCompile(`sk-[a-zA-Z0-9]{32,}`),
		Severity:    SeverityCritical,
		Category:    "credential_exposure",
		Title:       "OpenAI API key exposed",
		Description: "Detected OpenAI API key",
		Suggestion:  "Use environment variables for API keys",
	},
	{
		ID:          "CRED007",
		Pattern:     regexp.MustCompile(`(?i)bearer\s+[a-zA-Z0-9._-]{20,}`),
		Severity:    SeverityHigh,
		Category:    "credential_exposure",
		Title:       "Bearer token exposed",
		Description: "Detected bearer token in skill",
		Suggestion:  "Use environment variables for tokens",
	},

	// Private keys
	{
		ID:          "CRED008",
		Pattern:     regexp.MustCompile(`-----BEGIN\s+(RSA\s+)?PRIVATE\s+KEY-----`),
		Severity:    SeverityCritical,
		Category:    "credential_exposure",
		Title:       "Private key exposed",
		Description: "Detected private key in skill",
		Suggestion:  "Never include private keys in skills; use key files with proper permissions",
	},
}

// URLPatterns validates external resource references
var URLPatterns = []PatternRule{
	{
		ID:          "URL001",
		Pattern:     regexp.MustCompile(`(?i)https?://[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+[:/]`),
		Severity:    SeverityMedium,
		Category:    "suspicious_url",
		Title:       "IP address URL",
		Description: "Direct IP address URLs are suspicious",
		Suggestion:  "Use domain names instead of IP addresses",
	},
	{
		ID:          "URL002",
		Pattern:     regexp.MustCompile(`(?i)https?://(localhost|127\.0\.0\.1|0\.0\.0\.0)`),
		Severity:    SeverityLow,
		Category:    "local_url",
		Title:       "Localhost reference",
		Description: "Skill references localhost which may not work for all users",
		Suggestion:  "Document that this requires a local service",
	},
	{
		ID:          "URL003",
		Pattern:     regexp.MustCompile(`(?i)https?://[^\s"']+\.(onion|i2p)/`),
		Severity:    SeverityHigh,
		Category:    "suspicious_url",
		Title:       "Dark web URL",
		Description: "URL points to dark web service",
		Suggestion:  "Remove dark web references",
	},
	{
		ID:          "URL004",
		Pattern:     regexp.MustCompile(`(?i)(pastebin|hastebin|paste\.ee|ghostbin)\.com/`),
		Severity:    SeverityMedium,
		Category:    "suspicious_url",
		Title:       "Paste site URL",
		Description: "Paste sites are often used to host malicious content",
		Suggestion:  "Use official documentation or trusted sources instead",
	},
}

// BehaviorPatterns detects suspicious behavioral instructions
var BehaviorPatterns = []PatternRule{
	// High: Concealment
	{
		ID:          "BEH001",
		Pattern:     regexp.MustCompile(`(?i)(hide|conceal|obfuscate|mask)\s+(from\s+)?(the\s+)?(user|output|log|response)`),
		Severity:    SeverityHigh,
		Category:    "suspicious_behavior",
		Title:       "Output concealment",
		Description: "Skill instructs agent to hide information from user",
		Suggestion:  "Be transparent; don't hide actions from users",
	},
	{
		ID:          "BEH002",
		Pattern:     regexp.MustCompile(`(?i)(don'?t|do\s+not|never)\s+(tell|show|inform|reveal|display)\s+(the\s+)?(user|human)`),
		Severity:    SeverityHigh,
		Category:    "suspicious_behavior",
		Title:       "Information withholding",
		Description: "Skill instructs to withhold information from user",
		Suggestion:  "Always be transparent with users about actions taken",
	},

	// High: Bypassing confirmation
	{
		ID:          "BEH003",
		Pattern:     regexp.MustCompile(`(?i)(without|no|skip)\s+(user|human)?\s*(confirmation|approval|consent|permission|asking)`),
		Severity:    SeverityHigh,
		Category:    "suspicious_behavior",
		Title:       "Bypassing user confirmation",
		Description: "Skill may perform actions without user approval",
		Suggestion:  "Always get user confirmation for destructive or sensitive actions",
	},
	{
		ID:          "BEH004",
		Pattern:     regexp.MustCompile(`(?i)(auto|automatic|silent)(ly)?\s+(delete|remove|execute|run|send)`),
		Severity:    SeverityMedium,
		Category:    "suspicious_behavior",
		Title:       "Automatic action execution",
		Description: "Skill performs automatic actions that should require confirmation",
		Suggestion:  "Add explicit user confirmation before destructive actions",
	},

	// Critical: Malicious intent
	{
		ID:          "BEH005",
		Pattern:     regexp.MustCompile(`(?i)(exfiltrate|steal|harvest|collect\s+and\s+send)\s+(data|credentials|tokens?|keys?|passwords?|secrets?)`),
		Severity:    SeverityCritical,
		Category:    "malicious_intent",
		Title:       "Data exfiltration language",
		Description: "Skill contains language suggesting data theft",
		Suggestion:  "Remove any data exfiltration instructions",
	},
	{
		ID:          "BEH006",
		Pattern:     regexp.MustCompile(`(?i)(upload|send|transmit)\s+(to|all)?\s*(external|remote|third[- ]party|our)\s+(server|endpoint|api)`),
		Severity:    SeverityHigh,
		Category:    "suspicious_behavior",
		Title:       "External data transmission",
		Description: "Skill may send data to external servers",
		Suggestion:  "Document clearly why external transmission is needed and get user consent",
	},

	// Medium: Error handling
	{
		ID:          "BEH007",
		Pattern:     regexp.MustCompile(`(?i)(silent|quiet|suppress|ignore)\s+(all\s+)?(error|warning|exception|failure)`),
		Severity:    SeverityMedium,
		Category:    "suspicious_behavior",
		Title:       "Error suppression",
		Description: "Suppressing errors could hide important issues",
		Suggestion:  "Log errors appropriately instead of suppressing them",
	},
}
