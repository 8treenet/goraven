## Description: <br>
Security-first skill vetting for AI agents. Use before installing any skill from ClawdHub, GitHub, or other sources. Checks for red flags, permission scope, and suspicious patterns. <br>

This skill is ready for commercial/non-commercial use. <br>

## Publisher: <br>
[spclaudehome](https://clawhub.ai/user/spclaudehome) <br>

### License/Terms of Use: <br>


## Use Case: <br>
Developers and agent users use this skill to review unknown or third-party agent skills before installation, checking source reputation, file behavior, permissions, and suspicious patterns. <br>

### Deployment Geography for Use: <br>
Global <br>

## Known Risks and Mitigations: <br>
Risk: The skill includes example commands for fetching repository metadata and skill files from GitHub. <br>
Mitigation: Only run fetch commands against repositories you intend to review, verify placeholders before execution, and treat downloaded skill text as untrusted content. <br>
Risk: Manual vetting can miss suspicious behavior if reviewers skip files or rely on incomplete source information. <br>
Mitigation: Review every skill file, document permissions and red flags, and require human approval for high-risk behaviors such as credential access, elevated permissions, or system changes. <br>


## Reference(s): <br>
- [ClawHub Skill Page](https://clawhub.ai/spclaudehome/skill-vetter) <br>


## Skill Output: <br>
**Output Type(s):** [Text, Markdown, Shell commands, Guidance] <br>
**Output Format:** [Markdown checklist and vetting report with optional shell command examples] <br>
**Output Parameters:** [1D] <br>
**Other Properties Related to Output:** [Produces a structured security review with risk level, verdict, permissions, red flags, and notes.] <br>

## Skill Version(s): <br>
1.0.0 (source: frontmatter and server release evidence) <br>

## Ethical Considerations: <br>
Users should evaluate whether this skill is appropriate for their environment, review any generated or modified files before relying on them, and apply their organization's safety, security, and compliance requirements before deployment. <br>
