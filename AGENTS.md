<!-- gitnexus:start -->
# GitNexus — Code Intelligence

This project is indexed by GitNexus as **gcommon** (1345 symbols, 2725 relationships, 71 execution flows). Use the GitNexus MCP tools to understand code, assess impact, and navigate safely.

> If any GitNexus tool warns the index is stale, run `npx gitnexus analyze` in terminal first.

## Always Do

- **MUST run impact analysis before editing any symbol.** Before modifying a function, class, or method, run `gitnexus_impact({target: "symbolName", direction: "upstream"})` and report the blast radius (direct callers, affected processes, risk level) to the user.
- **MUST run `gitnexus_detect_changes()` before committing** to verify your changes only affect expected symbols and execution flows.
- **MUST warn the user** if impact analysis returns HIGH or CRITICAL risk before proceeding with edits.
- When exploring unfamiliar code, use `gitnexus_query({query: "concept"})` to find execution flows instead of grepping. It returns process-grouped results ranked by relevance.
- When you need full context on a specific symbol — callers, callees, which execution flows it participates in — use `gitnexus_context({name: "symbolName"})`.

## Never Do

- NEVER edit a function, class, or method without first running `gitnexus_impact` on it.
- NEVER ignore HIGH or CRITICAL risk warnings from impact analysis.
- NEVER rename symbols with find-and-replace — use `gitnexus_rename` which understands the call graph.
- NEVER commit changes without running `gitnexus_detect_changes()` to check affected scope.

## Resources

| Resource | Use for |
|----------|---------|
| `gitnexus://repo/gcommon/context` | Codebase overview, check index freshness |
| `gitnexus://repo/gcommon/clusters` | All functional areas |
| `gitnexus://repo/gcommon/processes` | All execution flows |
| `gitnexus://repo/gcommon/process/{name}` | Step-by-step execution trace |

## CLI

| Task | Read this skill file |
|------|---------------------|
| Understand architecture / "How does X work?" | `.claude/skills/gitnexus/gitnexus-exploring/SKILL.md` |
| Blast radius / "What breaks if I change X?" | `.claude/skills/gitnexus/gitnexus-impact-analysis/SKILL.md` |
| Trace bugs / "Why is X failing?" | `.claude/skills/gitnexus/gitnexus-debugging/SKILL.md` |
| Rename / extract / split / refactor | `.claude/skills/gitnexus/gitnexus-refactoring/SKILL.md` |
| Tools, resources, schema reference | `.claude/skills/gitnexus/gitnexus-guide/SKILL.md` |
| Index, status, clean, wiki CLI commands | `.claude/skills/gitnexus/gitnexus-cli/SKILL.md` |
| Work in the Httpx area (61 symbols) | `.claude/skills/generated/httpx/SKILL.md` |
| Work in the Ginx area (51 symbols) | `.claude/skills/generated/ginx/SKILL.md` |
| Work in the Logx area (21 symbols) | `.claude/skills/generated/logx/SKILL.md` |
| Work in the Optional area (20 symbols) | `.claude/skills/generated/optional/SKILL.md` |
| Work in the Errorx area (16 symbols) | `.claude/skills/generated/errorx/SKILL.md` |
| Work in the Vo area (15 symbols) | `.claude/skills/generated/vo/SKILL.md` |
| Work in the Server area (14 symbols) | `.claude/skills/generated/server/SKILL.md` |
| Work in the Pinger area (11 symbols) | `.claude/skills/generated/pinger/SKILL.md` |
| Work in the Tree area (10 symbols) | `.claude/skills/generated/tree/SKILL.md` |
| Work in the Consul area (7 symbols) | `.claude/skills/generated/consul/SKILL.md` |
| Work in the Test area (6 symbols) | `.claude/skills/generated/test/SKILL.md` |
| Work in the Http area (5 symbols) | `.claude/skills/generated/http/SKILL.md` |
| Work in the Data area (5 symbols) | `.claude/skills/generated/data/SKILL.md` |
| Work in the Stack area (4 symbols) | `.claude/skills/generated/stack/SKILL.md` |

<!-- gitnexus:end -->
