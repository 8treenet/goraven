## Description: <br>
Edit PDFs with natural-language instructions using the nano-pdf CLI. <br>

This skill is ready for commercial/non-commercial use. <br>

## Publisher: <br>
[steipete](https://clawhub.ai/user/steipete) <br>

### License/Terms of Use: <br>


## Use Case: <br>
Developers, operators, and other PDF authors use this skill to instruct an agent to edit a specific PDF page through the nano-pdf CLI and then review the edited output. <br>

### Deployment Geography for Use: <br>
Global <br>

## Known Risks and Mitigations: <br>
Risk: Installing or running an untrusted PDF editing CLI could execute unwanted code or process sensitive documents in ways the user did not intend. <br>
Mitigation: Install only if the nano-pdf PyPI package source is trusted, and review the edited PDF before sending, publishing, or overwriting important documents. <br>
Risk: PDF edits can target the wrong page or produce unexpected changes, especially when page numbering differs by tool version or configuration. <br>
Mitigation: Keep original PDFs or backups, sanity-check the output, and retry with the alternate page numbering convention if the edit appears off by one. <br>


## Reference(s): <br>
- [Nano Pdf on ClawHub](https://clawhub.ai/steipete/nano-pdf) <br>
- [nano-pdf PyPI project](https://pypi.org/project/nano-pdf/) <br>


## Skill Output: <br>
**Output Type(s):** [guidance, shell commands, files] <br>
**Output Format:** [Markdown with inline shell commands and PDF file output from the CLI] <br>
**Output Parameters:** [1D] <br>
**Other Properties Related to Output:** [Requires the nano-pdf CLI and review of the edited PDF before use.] <br>

## Skill Version(s): <br>
1.0.0 (source: server-resolved release metadata) <br>

## Ethical Considerations: <br>
Users should evaluate whether this skill is appropriate for their environment, review any generated or modified files before relying on them, and apply their organization's safety, security, and compliance requirements before deployment. <br>
