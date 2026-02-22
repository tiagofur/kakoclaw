# Workflow UI Quick Reference

Visual guide to using the workflow builder interface.

---

## Workflow Views

The workflow interface has **two main views**:

### 1. List View (Default)
Shows all your existing workflows.

**Available actions:**
- ✅ Create new workflow (click "New Workflow")
- ✅ Edit existing workflow (click "Edit" on workflow card)
- ✅ Run workflow (click "Run")
- ✅ View history (click "History")
- ✅ Delete workflow (click "Delete")

**NOT available:**
- ❌ Add steps (must enter editor first)
- ❌ Edit step configuration (must enter editor first)

---

### 2. Editor View (After clicking "New Workflow" or "Edit")
Visual pipeline builder for creating/editing workflow steps.

**Available actions:**
- ✅ Name workflow
- ✅ Add steps (+ Prompt, + Tool, + Condition buttons)
- ✅ Edit step configuration (click step to expand)
- ✅ Reorder steps (drag & drop)
- ✅ Remove steps (click X icon)
- ✅ Save workflow
- ✅ Test run workflow
- ✅ Cancel editing

---

## Common UI Confusion

### "I don't see the + Prompt button"

**Problem**: You're in List View, not Editor View.

**Solution**: 
1. Click **"New Workflow"** (green button, top-right)
2. Or click **"Edit"** on an existing workflow
3. Now you'll see the `+ Prompt`, `+ Tool`, `+ Condition` buttons

---

### "I clicked + Prompt but nothing happened"

**Possible causes:**

1. **JavaScript error** (check browser console)
   - Open DevTools (F12)
   - Look for red error messages
   - Refresh page and try again

2. **Step was added but collapsed**
   - Look for a new item in the steps list
   - Click it to expand and see configuration

3. **Browser compatibility**
   - Use modern browser (Chrome, Firefox, Safari, Edge)
   - Update browser to latest version

---

## Step by Step: Creating Your First Workflow

### Visual Guide

```
┌─────────────────────────────────────────────────┐
│  Workflows                    [New Workflow]    │  ← Click this button
├─────────────────────────────────────────────────┤
│                                                 │
│  [Empty state or list of workflows]            │
│                                                 │
└─────────────────────────────────────────────────┘
```

**After clicking "New Workflow":**

```
┌─────────────────────────────────────────────────┐
│  New Workflow            [Cancel]  [Save]       │
├─────────────────────────────────────────────────┤
│  Name: [________________]                       │
│  Description: [_________________________]       │
│                                                 │
│  Pipeline Steps                                 │
│  ┌─────────────────────────────────────┐       │
│  │  No steps yet. Add a step below.   │       │
│  └─────────────────────────────────────┘       │
│                                                 │
│  [+ Prompt]  [+ Tool]  [+ Condition]           │  ← These buttons
│                                                 │
│  Test Run                                       │
│  [▶ Test Run]                                  │
└─────────────────────────────────────────────────┘
```

### Workflow Creation Steps

**1. Enter Editor**
- Click "New Workflow" OR "Edit" on existing workflow
- You should see: Name field, Description field, and Add Step buttons

**2. Fill Basic Info**
- Enter workflow name (required)
- Enter description (optional)

**3. Add First Step**
- Click one of: `+ Prompt`, `+ Tool`, or `+ Condition`
- Step card appears (expanded by default)

**4. Configure Step**
- Fill in step label
- Configure type-specific fields:
  - **Prompt**: Message, optional model
  - **Tool**: Tool name, arguments JSON
  - **Condition**: Reference, operator, value
- Set error handling policy

**5. Add More Steps**
- Click `+ Prompt/Tool/Condition` again
- Repeat configuration
- Use drag handles (⠿) to reorder

**6. Save**
- Click "Save" button (top-right)
- Wait for success confirmation

**7. Test**
- Click "Test Run" button (below steps)
- Watch results appear in real-time

---

## Step Interface

### When Collapsed
```
┌────────────────────────────────────────────┐
│ ⠿  [1] Search web (tool)              [X] │
└────────────────────────────────────────────┘
```
- `⠿` = Drag handle (reorder)
- `[1]` = Step number
- `Search web` = Label
- `(tool)` = Type
- `[X]` = Delete button

**Click anywhere to expand**

---

### When Expanded
```
┌────────────────────────────────────────────┐
│ ⠿  [1] Search web (tool)              [X] │
│                                            │
│ Label: [Search web_________________]      │
│ On Error: [Stop workflow ▼]               │
│                                            │
│ Tool Name: [web_search ▼]                 │
│ Arguments (JSON):                          │
│ ┌────────────────────────────────────────┐ │
│ │ {                                      │ │
│ │   "query": "AI news 2026"              │ │
│ │ }                                      │ │
│ └────────────────────────────────────────┘ │
└────────────────────────────────────────────┘
```

**Click step header again to collapse**

---

## Keyboard Shortcuts

Currently no keyboard shortcuts implemented. All actions require mouse/touch.

---

## Browser Requirements

**Supported:**
- ✅ Chrome/Edge 90+
- ✅ Firefox 88+
- ✅ Safari 14+

**Features used:**
- Vue 3 Composition API
- ES6+ JavaScript
- CSS Grid/Flexbox
- Fetch API
- WebSockets (for real-time updates)

**Not supported:**
- ❌ Internet Explorer
- ❌ Very old mobile browsers

---

## Troubleshooting UI Issues

### Buttons Not Clickable

**Check:**
1. Are you in Editor View? (should see workflow name fields)
2. Is page fully loaded? (no loading spinner)
3. Browser console errors? (F12 → Console tab)

**Fix:**
- Refresh page (Cmd+R / Ctrl+R)
- Clear cache (Cmd+Shift+R / Ctrl+Shift+R)
- Try incognito/private window

---

### Steps Not Saving

**Check:**
1. Did you click "Save" button?
2. Is workflow name filled in?
3. Are all step configurations valid?

**Common validation errors:**
- Empty step label
- Invalid JSON in tool arguments
- Empty tool name

**Fix:**
- Fill all required fields
- Validate JSON (use jsonlint.com)
- Check browser console for specific error

---

### Drag-and-Drop Not Working

**Check:**
1. Are you trying to drag by the drag handle (⠿)?
2. Is step expanded? (collapse first if needed)

**Fix:**
- Use the drag handle icon specifically
- Try on desktop (touch drag may be limited)
- Refresh page

---

### Test Run Button Disabled

**Reason**: Workflow must be saved before testing.

**Fix:**
1. Click "Save" button first
2. Wait for success message
3. "Test Run" button becomes enabled

---

## Getting Help

- **Documentation**: [Workflows Guide](../examples/workflows.md)
- **API Reference**: [Workflows API](../api-reference/workflows.md)
- **Issues**: https://github.com/sipeed/kakoclaw/issues

---

**Last Updated**: February 2026
