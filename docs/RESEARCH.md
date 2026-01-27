# Startup Validation Research Runner

Gobot includes a bonus tool: an autonomous 14-step startup validation system that runs through Claude Code. Use it to validate your SaaS idea before building.

## What It Does

The research runner automatically:

1. Validates your business idea against real market evidence
2. Discovers your target audience with specific behavior signals
3. Analyzes competitors and finds white space
4. Identifies jobs-to-be-done and "hair on fire" moments
5. Develops positioning that differentiates from competitors
6. Crafts a "Big Idea" - the belief shift that drives your marketing
7. Generates hook variations, ad copy, and pitch scripts
8. Scopes a 3-day MVP build plan
9. Writes complete landing page copy

## How to Use

### Step 1: Open Claude Code

```bash
claude
```

### Step 2: Load the Research Prompt

Copy the entire contents of `RESEARCH-PLAN.md` and paste it as the system prompt, or reference it:

```
Use the system prompt from RESEARCH-PLAN.md
```

### Step 3: Provide Your Input

The system accepts one of three starting points:

**Option A - You own a domain:**
```
I own the domain CoolSaaS.com
```

**Option B - You have expertise:**
```
My expertise is in healthcare data analytics
```

**Option C - You have a specific idea:**
```
I want to build an AI-powered code review tool for solo developers
```

### Step 4: Let It Run

The system runs autonomously through all 14 steps, writing outputs to `./plan/`:

```
plan/
├── 00-EXECUTION-COMPLETE.md (or 00-EXECUTION-STOPPED.md)
├── 01-starting-point.md
├── 02-business-idea-validation.md
├── 03-market-structure.md
├── 04-audience-discovery.md
├── 05-competitive-analysis.md
├── 06-jobs-to-be-done.md
├── 07-positioning.md
├── 08-research-summary.md
├── 09-big-idea.md
├── 10-hook-testing.md
├── 11-ad-copy-testing.md
├── 12-pitch-testing.md
├── 13-mvp-scope.md
└── 14-landing-page-copy.md
```

## Stop Conditions

The system will automatically stop if:

- **Step 2**: No evidence of complaints or existing solutions (idea abandoned)
- **Step 3**: Market dominated by single player or no profitable companies (market failed)
- **Step 8**: Any validation criterion fails (research failed)
- **Step 9**: Big Idea fails after 2 recognition attempts (positioning failed)

When stopped, it writes an explanation of what evidence was missing and suggests pivots.

## The 14 Steps Explained

| Step | File | Purpose |
|------|------|---------|
| 1 | `01-starting-point.md` | Generate business ideas from domain/expertise/idea |
| 2 | `02-business-idea-validation.md` | Find 10+ real complaints proving the problem exists |
| 3 | `03-market-structure.md` | Verify 4-10 competitors at similar scale (healthy market) |
| 4 | `04-audience-discovery.md` | Define exact target customer with behavior signals |
| 5 | `05-competitive-analysis.md` | Find competitor weaknesses and white space |
| 6 | `06-jobs-to-be-done.md` | Identify the "hair on fire" moment and switching costs |
| 7 | `07-positioning.md` | Craft "Unlike X, we Y" differentiation |
| 8 | `08-research-summary.md` | Validate all criteria, compress into 1000-word brief |
| 9 | `09-big-idea.md` | Discover the belief shift that drives all marketing |
| 10 | `10-hook-testing.md` | Generate 10 headline hooks from the Big Idea |
| 11 | `11-ad-copy-testing.md` | Write 3 complete ad variations |
| 12 | `12-pitch-testing.md` | Full pitch + objection handling |
| 13 | `13-mvp-scope.md` | 3-day build plan for the core feature |
| 14 | `14-landing-page-copy.md` | Complete landing page copy |

## The Big Idea Framework

Step 9 is the most important. A "Big Idea" is:

> The first clear articulation of a story the market already believes subconsciously, which explains past failure, names the real enemy, and reorganizes desire around a single motivating force.

Format:
```
"The reason [target audience] hasn't [achieved desired result] isn't [what they believe].
It's [the real mechanism], which means [new approach]."
```

The Big Idea must pass 11 validation tests including:
- Does it explain past failure in a new way?
- Does it create a clear villain?
- Is it unusable by competitors without changing their product?
- Does it use language that already exists in customer complaints?

## Tips for Best Results

1. **Be specific** - "Solo founders running bootstrapped SaaS" beats "entrepreneurs"
2. **Let it fail** - A stopped execution saves you from building the wrong thing
3. **Trust the process** - Don't skip steps; each builds on the previous
4. **Use the outputs** - The landing page copy and ad copy are production-ready starting points

## Example Output

After running for a code review tool idea, you might get:

**Big Idea:**
```
"The reason solo developers ship buggy code isn't lack of skill.
It's that code review requires a second brain, and hiring is
overkill for a 10-person codebase. AI can be that second brain
without the salary."
```

**Hook:**
```
"You're mass-producing bugs because you can't afford a second pair of eyes"
```

**3-Day MVP:**
- Day 1: GitHub OAuth + PR webhook listener
- Day 2: Claude API integration + comment UI
- Day 3: Stripe checkout + launch on r/SaaS

## Privacy Note

The `plan/` directory is gitignored by default. Your research outputs stay local and are never committed to version control.
