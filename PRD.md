# PRD: Multi-user Data Handling & Isolation

## Overview

MakoClaw is designed to be a multi-user, apex-efficiency AI agent. This document outlines the requirements and implementation details for user-specific data isolation.

## Objectives

- Ensure complete data isolation between different users.
- Provide a seamless transition between global defaults and personal configurations.
- Securely store user-specific credentials (tokens, secrets).

## Technical Implementation

### 1. Configuration Hierarchy

- **System Defaults**: Built-in into the binary.
- **Global Configuration**: `~/.makoclaw/config.json`.
- **User Overrides**: `~/.makoclaw/users/<user-uuid>/config.json`.
- **Merge Logic**: User settings override global settings, which override system defaults.

### 2. Physical Isolation

- Each user has a dedicated directory at `~/.makoclaw/users/<user-uuid>/`.
- **Databases**: Each user has their own `makoclaw.db` for tasks, chat history, and long-term memory.
- **Workspaces**: User file operations are restricted to `~/.makoclaw/users/<user-uuid>/workspace/`.

### 3. Identity and Authentication

- Users are identified by a unique UUID.
- JWT tokens are used for session management.
- Backend handlers must always resolve the user identity through the `request` context.

## User Experience (UX)

- Configuring a channel (like Telegram) in the user's settings only affects that specific user.
- Agents created by a user are not visible to other non-admin users.
- Admin users can manage global settings and user accounts but cannot see user-specific sensitive data by default.
