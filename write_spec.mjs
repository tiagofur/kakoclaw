
import { writeFileSync } from "fs";

const dest = "C:/Users/tfurt/source/repos/kakoclaw/pkg/web/frontend/e2e/tests/settings.spec.ts";

const lines = [];
const a = (s = "") => lines.push(s);

a("/**");
a(" * settings.spec.ts - SettingsView E2E tests");
a(" *");
a(" * Covers:");
a(" *   - Page load and tab navigation (Profile, Agents, Providers, Channels, System)");
a(" *   - AgentSettingsTab: orchestrator section, Add Specialist button, modal open/close,");
a(" *     required-field validation");
a(" *   - System tab: Backup & Restore section visibility");
a(" *   - Admin-only Users tab: visible only for admin, shows user table & Add User button");
a(" *");
a(" * Pre-condition: admin auth state restored from e2e/.auth/admin.json.");
a(" */");
a("import { test, expect } from '@playwright/test'");
a("import { goTo } from '../fixtures/helpers'");
a("");

writeFileSync(dest, lines.join("
"), "utf8");
console.log("partial write", lines.length);
