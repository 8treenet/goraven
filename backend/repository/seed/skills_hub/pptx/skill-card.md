## Description: <br>
Presentation creation, editing, and analysis for creating new PowerPoint files, modifying existing .pptx files, working with layouts, and handling comments or speaker notes. <br>

This skill is ready for commercial/non-commercial use. <br>

## Publisher: <br>
[lnj22](https://clawhub.ai/user/lnj22) <br>

### License/Terms of Use: <br>
Proprietary Anthropic Terms of Service <br>


## Use Case: <br>
Developers and agent operators use this skill to inspect, create, edit, validate, and visually review PowerPoint presentations. It supports HTML-to-PPTX creation, OOXML editing workflows, template rearrangement, text replacement, inventory extraction, and thumbnail-based review. <br>

### Deployment Geography for Use: <br>
Global <br>

## Known Risks and Mitigations: <br>
Risk: PowerPoint editing workflows can rewrite, clear, or replace presentation content. <br>
Mitigation: Use copies of important decks and review generated replacement JSON before bulk replacement. <br>
Risk: Untrusted Office files or HTML may expose the agent environment to unsafe content during unpacking, rendering, or conversion. <br>
Mitigation: Avoid running the skill on untrusted Office files or HTML, and keep processing in a controlled local workspace. <br>
Risk: Optional document conversion and rendering dependencies may require local package installation. <br>
Mitigation: Install only the dependencies needed for the intended workflow and review package sources before use. <br>


## Reference(s): <br>
- [ClawHub release page](https://clawhub.ai/lnj22/organize-messy-files-pptx) <br>
- [HTML to PowerPoint Guide](html2pptx.md) <br>
- [Office Open XML Technical Reference for PowerPoint](ooxml.md) <br>


## Skill Output: <br>
**Output Type(s):** [text, markdown, code, shell commands, configuration, guidance, files] <br>
**Output Format:** [Markdown guidance with inline code blocks, JSON replacement data, local shell commands, and generated or modified presentation files.] <br>
**Output Parameters:** [1D] <br>
**Other Properties Related to Output:** [Produces local PPTX-related artifacts only when the agent runs the documented scripts or generated code.] <br>

## Skill Version(s): <br>
0.1.0 (source: server release metadata) <br>

## Ethical Considerations: <br>
Users should evaluate whether this skill is appropriate for their environment, review any generated or modified files before relying on them, and apply their organization's safety, security, and compliance requirements before deployment. <br>
