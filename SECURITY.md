# Security Policy

## Reporting a vulnerability

Report suspected vulnerabilities through GitHub's private vulnerability
reporting: open the repository's **Security** tab, choose **Advisories**, then
**Report a vulnerability**. If private reporting is unavailable, contact the
repository owner directly through a private channel.

**Do not open a public issue, pull request or discussion for a suspected
security problem.** This includes any report that could expose protected health
information (PHI) or personally identifiable information (PII): names, email
addresses, phone numbers, dates of birth, appointment details, credentials,
tokens or database contents. Redact all patient data from screenshots, logs and
reproduction steps. A PHI leak must be reported privately so it can be contained
before it is disclosed.

Please include:

- affected endpoint, version or commit,
- reproduction steps and expected versus actual behaviour,
- impact assessment (data exposure, privilege escalation, availability),
- any suggested remediation.

## What to expect

| Stage | Target |
|---|---|
| Acknowledgement of the report | 2 business days |
| Initial triage and severity assessment | 5 business days |
| Fix or mitigation plan for critical/high findings | 10 business days |

We will keep you informed of progress and credit you in the advisory unless you
ask to remain anonymous.

## Supported versions

| Version | Supported |
|---|---|
| `main` (latest deployment) | Yes |
| Older commits and release tags | No |

Security fixes land on `main` only; deploy the latest image to receive them.

## Scope

In scope: the Go API in `backend/`, its authentication and authorization model,
database migrations, and the CI/CD workflow in `.github/workflows/`.

Out of scope: the static frontend demo data, third-party services, and findings
that require physical access to a developer machine.

## Handling secrets

Never commit `.env` files, JWT secrets, database URLs, webhook URLs or any
patient data. Rotate a secret immediately if it is exposed, then report the
exposure privately.
