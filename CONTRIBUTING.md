<!-- markdownlint-disable -->
<div id="top"></div>
<!-- markdownlint-restore -->

# Contributing to betterglobekey

First off, thanks for taking the time to contribute! ❤️

<!-- markdownlint-disable-next-line -->
All types of contributions are encouraged and valued. See the [Table of Contents](#table-of-contents) for different ways
to help and details about how this project handles them. Please make sure to read the relevant section before making
your contribution. It will make it a lot easier for us maintainers and smooth out the experience for all involved. The
community looks forward to your contributions. 🎉

If you like the project, but just do not have time to contribute, that is fine. There are other easy ways to support the
project and show your appreciation, which we would also be very happy about:

- Star the project
- Tweet about it
- Refer this project in your project's `README.md`
- Mention the project at local meet-ups and tell your friends/colleagues about it

<!-- markdownlint-disable -->
<details id="table-of-contents">
  <summary>Table of Contents</summary>
  <ul>
    <li>
      <a href="#code-of-conduct">1. Code of Conduct</a>
    </li>
    <li>
      <a href="#i-have-a-question">2. I Have a Question</a>
    </li>
    <li>
      <a href="#i-want-to-contribute">3. I Want to Contribute</a>
      <ul>
        <li>
          <a href="#reporting-bugs">3.1. Reporting Bugs</a>
          <ul>
            <li>
              <a href="#before-submitting-a-bug-report">3.1.1. Before Submitting a Bug Report</a>
            </li>
            <li>
              <a href="#how-do-i-submit-a-good-bug-report">3.1.2. How Do I Submit a Good Bug Report?</a>
            </li>
          </ul>
        </li>
        <li>
          <a href="#suggesting-enhancements">3.2. Suggesting Enhancements</a>
          <ul>
            <li>
              <a href="#before-submitting-an-enhancement">3.2.1. Before Submitting an Enhancement</a>
            </li>
            <li>
              <a href="#how-do-i-submit-a-good-enhancement-suggestion">3.2.2. How Do I Submit a Good Enhancement Suggestion?</a>
            </li>
          </ul>
        </li>
        <li>
          <a href="#your-first-code-contribution">3.3. Your First Code Contribution</a>
        </li>
        <li>
          <a href="#improving-the-documentation">3.4. Improving the Documentation</a>
        </li>
      </ul>
    </li>
    <li>
      <a href="#styleguides">4. Styleguides</a>
      <ul>
        <li>
          <a href="#commit-messages">4.1. Commit Messages</a>
        </li>
      </ul>
    </li>
    <li>
      <a href="#join-the-project-team">5. Join the Project Team</a>
    </li>
  </ul>
</details>
<!-- markdownlint-restore -->

## Code of Conduct

This project and everyone participating in it is governed by
our [Code of Conduct](https://github.com/Serpentiel/betterglobekey/blob/master/CODE_OF_CONDUCT.md). By participating,
you are expected to uphold this code. Please report unacceptable behavior via any contact method available to you.

<!-- markdownlint-disable -->
<p align="right"><a href="#top">(back to top)</a></p>
<!-- markdownlint-restore -->

## I Have a Question

> **N.B.** If you want to ask a question, we assume that you have read the
> available [documentation](https://github.com/Serpentiel/betterglobekey/wiki).

Before you ask a question, it is best to search for existing [issues](https://github.com/Serpentiel/betterglobekey/issues)
that might help you. In case you have found a suitable issue and still need clarification, you can write your question
in this issue. It is also advisable to search the Internet for answers first.

If you then still feel the need to ask a question and need clarification, we recommend the following:

- Open an [issue](https://github.com/Serpentiel/betterglobekey/issues/new)
- Provide as much context as you can about what you are running into
- Provide the version of **betterglobekey**, platform, OS and its version, depending on what seems relevant

We will then take care of the issue as soon as possible.

<!-- markdownlint-disable -->
<p align="right"><a href="#top">(back to top)</a></p>
<!-- markdownlint-restore -->

## I Want to Contribute

> **N.B.** When contributing to this project, you agree that you have authored all of the content, that you have the
> necessary rights to the content, and that the content you contribute may be provided under the project's license.

### Reporting Bugs

#### Before Submitting a Bug Report

A good bug report should not leave others needing to chase you up for more information. Therefore, we ask you to
investigate carefully, collect information and describe the issue in detail in your report. Please complete the
following steps in advance to help us fix any potential bug as fast as possible.

- Make sure that you are using the latest version
- Determine if your bug is really a bug and not an error on your side, e.g. using incompatible environment
  components/versions—make sure that you have read the
  [documentation](https://github.com/Serpentiel/betterglobekey/wiki). If you are looking for support, you might want to
  check [this section](#i-have-a-question))
- Check if other users have experienced—and potentially already solved—the same issue you are having, check if there is
  not already a bug report existing for your bug or error
  in [GitHub Issues](https://github.com/Serpentiel/betterglobekey/issues?q=label:bug)
- Make sure to search the Internet—including StackOverflow—to see if users outside of the GitHub community have
  discussed the issue
- Check if the issue is reproducible on the latest and/or previous versions of **betterglobekey**
- Collect some relevant information about the bug:
  - Stack trace
  - Platform, OS and its version
  - Version of the tools involved that seems relevant
  - If possible, your input and the output
  - Steps required to reproduce the issue reliably

#### How Do I Submit a Good Bug Report?

> **N.B.** You must never report security related issues, vulnerabilities or bugs including sensitive information to the
> issue tracker, or elsewhere in public. Instead sensitive bugs must be reported according to
> our [security policy](https://github.com/Serpentiel/betterglobekey/blob/main/SECURITY.mdZ).

We use GitHub Issues to track bugs and errors. If you run into an issue with the project:

- Open an [issue](https://github.com/Serpentiel/betterglobekey/issues/new?label=bug)
- Explain the behavior you would expect and the actual behavior
- Please provide as much context as possible and describe the reproduction steps that someone else can follow to
  recreate the issue on their own—for good bug reports you should isolate the problem and create a reduced test case
- Provide all the relevant information that might help to resolve the problem

Once the bug report is filed:

- The project team will check the issue and label and/or relabel it accordingly
- A team member will try to reproduce the issue with your provided steps
  - If there are no reproduction steps or no obvious way to reproduce the issue, the team will ask you for those steps
      and label the issue as `needs repro`—bugs with the `needs repro` label will not be addressed until they are
      reproduced
  - If the team is able to reproduce the issue, it will be labeled as `needs fix`, as well as possibly other labels,
      and the issue will be left to be [implemented by someone](#your-first-code-contribution)

### Suggesting Enhancements

This section guides you through submitting an enhancement suggestion for **betterglobekey**, including completely new
features and minor improvements to existing functionality. Following these guidelines will help maintainers and the
community to understand your suggestion and find related suggestions.

#### Before Submitting an Enhancement

- Make sure that you are using the latest version of **betterglobekey**
- Read the [documentation](https://github.com/Serpentiel/betterglobekey/wiki) carefully and find out if the
  functionality is already covered, maybe by an individual configuration
- Perform a [search](https://github.com/Serpentiel/betterglobekey/issues) to see if the enhancement has already been
  suggested—if it has, add a comment to the existing issue instead of opening a new one
- Find out whether your idea fits with the scope and aims of the project—it is up to you to make a strong case to
  convince the project's developers of the merits of this feature—keep in mind that we want features that will be useful
  to the majority of our users and not just a small subset

#### How Do I Submit a Good Enhancement Suggestion?

Enhancement suggestions are tracked as [GitHub Issues](https://github.com/Serpentiel/betterglobekey/issues).

- Use a clear and descriptive title for the issue to identify the suggestion
- Provide a step-by-step description of the suggested enhancement, in as many details as possible
- Describe the current behavior and explain which behavior you expected to see instead and why—at this point you can
  also tell which alternatives do not work for you
- You may want to include screenshots and animated GIFs which help you demonstrate the steps, or point out the part
  to which the suggestion is related to
- Explain why this enhancement would be useful to most **betterglobekey** users—you may also want to point out the other
  projects that solved it better and which could serve as inspiration

<!-- markdownlint-disable -->
<p align="right"><a href="#top">(back to top)</a></p>
<!-- markdownlint-restore -->

### Your First Code Contribution

<!-- TODO -->

### Improving the Documentation

<!-- TODO -->

<!-- markdownlint-disable -->
<p align="right"><a href="#top">(back to top)</a></p>
<!-- markdownlint-restore -->

## Styleguides

<!-- TODO -->

### Commit Messages

<!-- TODO -->

<!-- markdownlint-disable -->
<p align="right"><a href="#top">(back to top)</a></p>
<!-- markdownlint-restore -->

## Join the Project Team

<!-- TODO -->

<!-- markdownlint-disable -->
<p align="right"><a href="#top">(back to top)</a></p>
<!-- markdownlint-restore -->
