## Description: <br>
Captures learnings, errors, feature requests, and corrections so agents can preserve useful lessons across sessions. <br>

This skill is ready for commercial/non-commercial use. <br>

## Publisher: <br>
[pskoett](https://clawhub.ai/user/pskoett) <br>

### License/Terms of Use: <br>
MIT-0 <br>


## Use Case: <br>
Developers and agent operators use this skill to log corrections, command failures, knowledge gaps, and requested capabilities into local learning files. It also provides optional OpenClaw and hook workflows for reminding agents to review and capture useful lessons. <br>

### Deployment Geography for Use: <br>
Global <br>

## Known Risks and Mitigations: <br>
Risk: Local learning logs can accidentally preserve secrets, private source, raw transcripts, or command output containing credentials. <br>
Mitigation: Log short summaries or redacted excerpts, and avoid storing secrets, raw transcripts, command output with credentials, or private source dumps in .learnings/. <br>
Risk: Broad hook configuration can inject reminders across unintended workspaces or inspect command output more often than expected. <br>
Mitigation: Use project-level hook configuration with a narrow matcher, review script paths before enabling hooks, and avoid global every-prompt hooks unless that behavior is intentional. <br>
Risk: Cross-session learning workflows may expose sensitive context if raw transcripts or command output are forwarded. <br>
Mitigation: Use cross-session sharing only in trusted environments and send concise sanitized summaries with relevant file paths instead of raw transcripts or secret-bearing output. <br>


## Reference(s): <br>
- [OpenClaw Integration](references/openclaw-integration.md) <br>
- [Hook Setup Guide](references/hooks-setup.md) <br>
- [Entry Examples](references/examples.md) <br>
- [Agent Skills Specification](https://agentskills.io/specification) <br>
- [ClawHub Skill Page](https://clawhub.ai/pskoett/self-improving-agent) <br>


## Skill Output: <br>
**Output Type(s):** [Markdown, Shell commands, Configuration instructions, Code, Guidance] <br>
**Output Format:** [Markdown with inline shell, JSON, and code blocks] <br>
**Output Parameters:** [1D] <br>
**Other Properties Related to Output:** [May create or update local .learnings markdown files and optional hook or skill scaffold files when the user enables those workflows.] <br>

## Skill Version(s): <br>
3.0.23 (source: server release evidence) <br>

## Ethical Considerations: <br>
Users should evaluate whether this skill is appropriate for their environment, review any generated or modified files before relying on them, and apply their organization's safety, security, and compliance requirements before deployment. <br>
