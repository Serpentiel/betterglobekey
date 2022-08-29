<!-- markdownlint-disable -->
<div id="top"></div>
<!-- markdownlint-restore -->

# Security Policy

betterglobekey has adopted this security disclosure and response policy to ensure we responsibly handle critical
issues.

<!-- markdownlint-disable -->
<details>
  <summary>Table of Contents</summary>
  <ul>
    <li>
      <a href="#supported-versions">1. Supported Versions</a>
    </li>
    <li>
      <a href="#reporting-a-vulnerabilityprivate-disclosure-process">2. Reporting a Vulnerability—Private Disclosure Process</a>
      <ul>
        <li>
          <a href="#proposed-message-content">2.1. Proposed Message Content</a>
        </li>
      </ul>
    </li>
    <li>
      <a href="#when-to-report-a-vulnerability">3. When to Report a Vulnerability</a>
    </li>
    <li>
      <a href="#patch-release-and-disclosure">4. Patch, Release, and Disclosure</a>
      <ul>
        <li>
          <a href="#public-disclosure-process">4.1. Public Disclosure Process</a>
        </li>
      </ul>
    </li>
    <li>
      <a href="#confidentiality-integrity-and-availability">5. Confidentiality, Integrity and Availability</a>
    </li>
  </ul>
</details>
<!-- markdownlint-restore -->

## Supported Versions

The betterglobekey project maintains release branches for the three most recent minor releases. Applicable fixes,
including security fixes, may be backported to those three release branches, depending on severity and feasibility.

<!-- markdownlint-disable -->
<p align="right"><a href="#top">(back to top)</a></p>
<!-- markdownlint-restore -->

## Reporting a Vulnerability—Private Disclosure Process

Security is of the highest importance and all security vulnerabilities or suspected security vulnerabilities should be
reported to betterglobekey privately, to minimize attacks against current users of betterglobekey before they are fixed.
Vulnerabilities will be investigated and patched on the next patch or minor release as soon as possible. This
information could be kept entirely internal to the project.

If you know of a publicly disclosed security vulnerability for betterglobekey, please **IMMEDIATELY** contact us via any
contact method available to you to inform our team.

> **N.B.** Do not file public issues on GitHub for security vulnerabilities.

To report a vulnerability or a security-related issue, please contact us via any contact method available to you with
the details of the vulnerability. The message will be fielded by the our team, which is made up of betterglobekey
maintainers who have committer and release permissions. Messages will be addressed within 3 business days, including a
detailed plan to investigate the issue and any potential workarounds to perform in the meantime. Do not report
non-security-impacting bugs through this channel.
Use [GitHub Issues](https://github.com/Serpentiel/betterglobekey/issues/new/choose) instead.

### Proposed Message Content

Please, include the following information to your message:

- Basic identity information, such as your name and your affiliation or company
- Detailed steps to reproduce the vulnerability, e.g. PoC, screenshots, depending on what seems relevant
- Description of the effects of the vulnerability on betterglobekey and the related hardware and software
  configurations, so that our team can reproduce it
- How the vulnerability affects betterglobekey usage and an estimation of the attack surface, if there is one
- List other projects or dependencies that were used in conjunction with betterglobekey to produce the vulnerability

<!-- markdownlint-disable -->
<p align="right"><a href="#top">(back to top)</a></p>
<!-- markdownlint-restore -->

## When to Report a Vulnerability

- When you think betterglobekey has a potential security vulnerability
- When you suspect a potential vulnerability but you are unsure that it impacts betterglobekey
- When you know of or suspect a potential vulnerability on another project that is used by betterglobekey, e.g.
  dependencies of betterglobekey

<!-- markdownlint-disable -->
<p align="right"><a href="#top">(back to top)</a></p>
<!-- markdownlint-restore -->

## Patch, Release, and Disclosure

Our team will respond to vulnerability reports as follows:

1. Our team will investigate the vulnerability and determine its effects and criticality
2. If the issue is not deemed to be a vulnerability, our team will follow up with a detailed reason for rejection
3. Our team will initiate a conversation with the reporter within 3 business days
4. If a vulnerability is acknowledged and the timeline for a fix is determined, our team will work on a plan to
   communicate with the appropriate community, including identifying mitigating steps that affected users can take to
   protect themselves until the fix is rolled out
5. Our team will also create a [CVSS](https://first.org/cvss/specification-document) using
   the [CVSS Calculator](https://first.org/cvss/calculator/3.0). Our team makes the final call on the calculated CVSS;
   it is better to move quickly than making the CVSS perfect. Issues may also be reported
   to [Mitre](https://cve.mitre.org/) using
   this [scoring calculator](https://nvd.nist.gov/vuln-metrics/cvss/v3-calculator). The CVE will initially be set to
   private
6. Our team will work on fixing the vulnerability and perform internal testing before preparing to roll out the fix
7. A public disclosure date is negotiated by our team and the bug submitter. We prefer to fully disclose the bug as soon
   as possible once a user mitigation or patch is available. It is reasonable to delay disclosure when the bug or the
   fix is not yet fully understood or the solution is not well-tested. The timeframe for disclosure is from
   immediate—especially if it’s already publicly known—to a few weeks. For a critical vulnerability with a
   straightforward mitigation, we expect report date to public disclosure date to be on the order of 14 business days.
   Our team holds the final say when setting a public disclosure date
8. Once the fix is confirmed, our team will patch the vulnerability in the next patch or minor release, and
   backport a patch release into all earlier supported releases. Upon release of the patched version of betterglobekey,
   we will follow the [Public Disclosure Process](#public-disclosure-process)

### Public Disclosure Process

Our team publishes a public [advisory](https://github.com/Serpentiel/betterglobekey/security/advisories) to the betterglobekey
community via GitHub. In most cases, additional communication via Slack, Twitter, blog and other channels will assist in
educating betterglobekey users and rolling out the patched release to affected users.

Our team will also publish any mitigating steps users can take until the fix can be applied to their betterglobekey
setup.

<!-- markdownlint-disable -->
<p align="right"><a href="#top">(back to top)</a></p>
<!-- markdownlint-restore -->

## Confidentiality, Integrity and Availability

We consider vulnerabilities leading to the compromise of data confidentiality, elevation of privilege, or integrity to
be our highest priority concerns. Availability, in particular in areas relating to DoS and resource exhaustion, is also
a serious security concern. Our team takes all vulnerabilities, potential vulnerabilities, and suspected vulnerabilities
seriously and will investigate them in an urgent and expeditious manner.

<!-- markdownlint-disable -->
<p align="right"><a href="#top">(back to top)</a></p>
<!-- markdownlint-restore -->
