## Description: <br>
Guide for creating effective skills. This skill should be used when users want to create a new skill (or update an existing skill) that extends Claude's capabilities with specialized knowledge, workflows, or tool integrations. <br>

This skill is ready for commercial/non-commercial use. <br>

## Publisher: <br>
[chindden](https://clawhub.ai/user/chindden) <br>

### License/Terms of Use: <br>
Apache 2.0 <br>


## Use Case: <br>
Developers and skill authors use Skill Creator to design, validate, package, and iterate on skills with reusable instructions, scripts, references, and assets. <br>

### Deployment Geography for Use: <br>
Global <br>

## Known Risks and Mitigations: <br>
Risk: Helper scripts can create or package files in unintended locations if run on the wrong path. <br>
Mitigation: Run scripts only on intentional skill directories and review generated files before using or sharing them. <br>
Risk: A packaged skill archive could include credentials, private notes, build outputs, or unrelated files left in the skill folder. <br>
Mitigation: Inspect the skill folder or .skill archive and delete unnecessary or sensitive files before sharing. <br>


## Reference(s): <br>
- [Output Patterns](references/output-patterns.md) <br>
- [Workflows](references/workflows.md) <br>


## Skill Output: <br>
**Output Type(s):** [Text, Markdown, Code, Shell commands, Configuration, Files, Guidance] <br>
**Output Format:** [Markdown with inline code and shell commands] <br>
**Output Parameters:** [1D] <br>
**Other Properties Related to Output:** [May include generated skill directories, validation results, and packaged .skill archives when helper scripts are run.] <br>

## Skill Version(s): <br>
0.1.0 (source: server release metadata) <br>

## Ethical Considerations: <br>
Users should evaluate whether this skill is appropriate for their environment, review any generated or modified files before relying on them, and apply their organization's safety, security, and compliance requirements before deployment. <br>
