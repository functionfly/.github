---
title: "TRUST PROTOCOL OPEN SOURCE"
---

# FunctionFly Trust Protocol — Open Source Strategy

**Version**: 1.0.0-draft  
**Status**: Draft for Q4 2026 Publication  
**Date**: 2026-03-21  

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Strategic Rationale](#2-strategic-rationale)
3. [Components to Open Source](#3-components-to-open-source)
4. [License Selection](#4-license-selection)
5. [Repository Structure](#5-repository-structure)
6. [Contribution Guidelines](#6-contribution-guidelines)
7. [Community Building](#7-community-building)
8. [Adoption Roadmap](#8-adoption-roadmap)
9. [Governance Model](#9-governance-model)
10. [Risk Mitigation](#10-risk-mitigation)

---

## 1. Executive Summary

### 1.1 Objective

Open-source the FunctionFly Trust Protocol verification and scoring components to establish an industry standard for AI agent function trust. This positions FunctionFly as the authority on trust infrastructure while enabling ecosystem growth.

### 1.2 Target Outcomes

| Outcome | Success Metric | Timeline |
|---------|---------------|----------|
| Protocol adoption | 10+ major platforms using spec | Q4 2026 |
| Community contributions | 50+ external contributors | Q4 2026 |
| SDK availability | Libraries in 5+ languages | Q4 2026 |
| Industry recognition | Referenced in 3+ analyst reports | Q1 2027 |

### 1.3 Key Decision

**Open Source MIT License** for core verification and scoring libraries, with **Apache 2.0** for protocol specification and documentation.

---

## 2. Strategic Rationale

### 2.1 Why Open Source the Trust Protocol?

| Reason | Explanation |
|--------|-------------|
| **Standard Setting** | Open protocols become industry standards; proprietary protocols face adoption resistance |
| **Network Effects** | More platforms adopting trust standard → more verified functions → better trust scores for everyone |
| **Competitive Moat Shift** | Move from "trust protocol" (commoditizable) to "trust data moat" (differentiated) |
| **Talent Acquisition** | Open protocols attract developer mindshare and potential contributors/employees |
| **Enterprise Credibility** | Enterprises prefer open standards over vendor lock-in |

### 2.2 Competitive Landscape Context

From [Moat Competitive Analysis](../plans/MOAT_COMPETITIVE_ANALYSIS.md):

> **Layer 3: Standard Setting** — Trust Protocol, Verification standards, Agent communication

The trust protocol layer is **the most defensible** when it becomes an industry standard. Competitors cannot compete with a freely available standard; they must adopt it or be seen as untrusted by comparison.

### 2.3 What Stays Proprietary

| Component | Reason |
|-----------|--------|
| Trust score calculation service | Execution data moat — accumulated verification history |
| Verification pipeline infrastructure | Proprietary processes and integrations |
| Partner management system | Business logic and billing |
| FunctionFly-specific trust data | Competitive advantage in trust rankings |

---

## 3. Components to Open Source

### 3.1 Open Source Package Map

```
┌─────────────────────────────────────────────────────────────────┐
│                    FunctionFly Platform                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   ┌──────────────────────────────────────────────────────────┐  │
│   │              OPEN SOURCE (MIT/Apache 2.0)                 │  │
│   ├──────────────────────────────────────────────────────────┤  │
│   │  • trust-protocol-spec (protocol definition)             │  │
│   │  • trust-score-calculator (scoring algorithms)           │  │
│   │  • trust-verification-client (verification SDK)          │  │
│   │  • trust-cli (command-line tools)                        │  │
│   │  • trust-examples (sample implementations)               │  │
│   └──────────────────────────────────────────────────────────┘  │
│                                                                  │
│   ┌──────────────────────────────────────────────────────────┐  │
│   │              PROPRIETARY (FunctionFly)                    │  │
│   ├──────────────────────────────────────────────────────────┤  │
│   │  • Trust Score Service (calculation + hosting)            │  │
│   │  • Verification Pipeline (orchestration)                  │  │
│   │  • Partner Management (billing + quotas)                  │  │
│   │  • FunctionFly Trust Data (historical records)           │  │
│   └──────────────────────────────────────────────────────────┘  │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 3.2 Package Specifications

#### Package 1: `trust-protocol-spec`

| Attribute | Value |
|-----------|-------|
| **Description** | Protocol specification, data models, API schemas |
| **License** | Apache 2.0 |
| **Language** | Markdown + OpenAPI 3.0 YAML |
| **Location** | `trust-protocol/spec` |
| **Contents** | Protocol docs, JSON schemas, API specifications |

**Contents**:

- Complete Trust Protocol specification (equivalent to [`docs/TRUST_PROTOCOL_SPEC.md`](docs/TRUST_PROTOCOL_SPEC.md))
- OpenAPI 3.0 specification for Trust API
- JSON Schema definitions for all data models
- Protocol change log and versioning policy

#### Package 2: `trust-score-calculator`

| Attribute | Value |
|-----------|-------|
| **Description** | Trust score calculation algorithms |
| **License** | MIT |
| **Language** | Go (reference), with ports to Python/JS |
| **Location** | `trust-protocol/calculator` |
| **Contents** | Core scoring logic, weight configuration, tier calculation |

**Contents**:

- Reference implementation of [`CalculateTrustScore()`](internal/storage/registry/trust_repository.go:19)
- Default weights from [`DefaultTrustScoreWeights()`](internal/storage/registry/types.go:571)
- Tier classification from [`getTrustTierFromScore()`](internal/api/handlers/registry/trust.go:342)
- Unit tests for all scoring components
- Benchmark suite for performance validation

**Not Included**:

- Server-side data collection
- Database operations
- API hosting

#### Package 3: `trust-verification-client`

| Attribute | Value |
|-----------|-------|
| **Description** | SDK for interacting with verification services |
| **License** | MIT |
| **Language** | Go, Python, TypeScript |
| **Location** | `trust-protocol/sdk` |
| **Contents** | API client, request/response types, error handling |

**Contents**:

- Client library for Trust API endpoints
- Verification request submission
- Batch trust score lookup
- Trust history retrieval
- Partner API key management

**Not Included**:

- Verification pipeline execution (server-side only)
- Partner onboarding UI

#### Package 4: `trust-cli`

| Attribute | Value |
|-----------|-------|
| **Description** | Command-line interface for trust operations |
| **License** | MIT |
| **Language** | Go |
| **Location** | `trust-protocol/cli` |
| **Contents** | CLI tool for developers |

**Commands**:

```
trust score <function_id>          # Get trust score
trust batch <function_ids...>      # Batch lookup
trust history <function_id>        # Get history
trust verify submit <function_id>  # Submit for verification
trust report <function_id>        # Report trust issue
```

#### Package 5: `trust-examples`

| Attribute | Value |
|-----------|-------|
| **Description** | Example implementations and integrations |
| **License** | MIT |
| **Location** | `trust-protocol/examples` |
| **Contents** | Integration examples, reference architectures |

**Examples**:

- LangChain tool integration
- AutoGen agent integration
- CrewAI toolkit integration
- Custom trust-aware load balancer
- Trust dashboard implementation

### 3.3 External Contribution Mapping

| Component | External Contributions Welcome |
|-----------|-------------------------------|
| `trust-protocol-spec` | Yes — via RFC process |
| `trust-score-calculator` | Yes — algorithm improvements |
| `trust-verification-client` | Yes — new language ports |
| `trust-cli` | Yes — new commands |
| `trust-examples` | Yes — new integration examples |

---

## 4. License Selection

### 4.1 License Matrix

| Package | License | Rationale |
|---------|---------|-----------|
| `trust-protocol-spec` | Apache 2.0 | Industry standard for specifications; allows proprietary implementations |
| `trust-score-calculator` | MIT | Maximum adoption; permissive for commercial use |
| `trust-verification-client` | MIT | Maximum adoption; enables commercial SDK products |
| `trust-cli` | MIT | Developer-friendly; common for tools |
| `trust-examples` | MIT | Educational; no restrictions |

### 4.2 Apache 2.0 vs MIT Comparison

| Factor | Apache 2.0 | MIT |
|--------|------------|-----|
| **Commercial use** | ✅ | ✅ |
| **Modification** | ✅ | ✅ |
| **Distribution** | ✅ | ✅ |
| **Patent grant** | ✅ | ❌ |
| **Sublicensing** | ✅ (via same license) | ✅ |
| **Trademark use** | Explicit clause | No explicit clause |
| **Attribution requirement** | YES (NOTICE file) | YES (license text) |
| **Enterprise preference** | High | Medium |

### 4.3 Why Not AGPL?

AGPL was considered but rejected:

- **Adoption barrier**: Many enterprises have policies against AGPL
- **Cloud service concern**: AGPL's "modified version" concept is ambiguous for APIs
- **Competitive differentiation**: Permissive licensing attracts more users

---

## 5. Repository Structure

### 5.1 Proposed Monorepo Structure

```
trust-protocol/
├── README.md
├── LICENSE (Apache 2.0 for spec, MIT for code)
├── NOTICE
├── SPEC.md                    # Main protocol specification
├── GOVERNANCE.md              # Community governance
├── CONTRIBUTING.md            # Contribution guidelines
│
├── spec/
│   ├── README.md
│   ├── openapi/
│   │   └── trust-api.yaml     # OpenAPI 3.0 specification
│   └── json-schema/
│       ├── trust-score.json
│       ├── verification-request.json
│       └── ...
│
├── calculator/
│   ├── go/
│   │   ├── README.md
│   │   ├── go.mod
│   │   ├── calculator.go
│   │   ├── calculator_test.go
│   │   └── weights.go
│   ├── python/
│   │   ├── README.md
│   │   ├── setup.py
│   │   └── trust_score/
│   └── js/
│       ├── README.md
│       ├── package.json
│       └── src/
│
├── sdk/
│   ├── go/
│   │   ├── README.md
│   │   ├── client.go
│   │   └── client_test.go
│   ├── python/
│   │   └── ...
│   └── typescript/
│       └── ...
│
├── cli/
│   ├── README.md
│   ├── main.go
│   ├── cmd/
│   │   ├── score.go
│   │   ├── batch.go
│   │   ├── history.go
│   │   ├── verify.go
│   │   └── report.go
│   └── go.mod
│
└── examples/
    ├── README.md
    ├── langchain/
    │   └── functionfly_tool.py
    ├── autogen/
    │   └── functionfly_agent.py
    ├── crewai/
    │   └── functionfly_tool.py
    └── load-balancer/
        └── trust_aware_lb.go
```

### 5.2 Repository Hosting

| Option | Choice | Rationale |
|--------|--------|-----------|
| **GitHub** | ✅ Primary | Largest AI/developer community |
| **GitLab** | Mirror option | Enterprise preference |
| **Gitea** | Self-host option | For enterprises with on-prem requirements |

**Recommended**: GitHub Organization `trust-protocol` with repository `trust-protocol`

---

## 6. Contribution Guidelines

### 6.1 Code of Conduct

Adopt Contributor Covenant v2.1:

```
trust-protocol/
├── CODE_OF_CONDUCT.md
```

Key points:

- Respectful communication
- Inclusive environment
- Zero tolerance for harassment
- Clear escalation path

### 6.2 Getting Started

```
CONTRIBUTING.md
===============

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request
6. Ensure CI passes
7. Wait for review (response within 48 hours)
```

### 6.3 Contribution Types

| Type | Examples | Process |
|------|----------|---------|
| **Bug Fixes** | Typo corrections, test failures | Direct PR, fast review |
| **Documentation** | Spec improvements, examples | Direct PR, doc review |
| **New Language Ports** | Ruby SDK, Rust SDK | RFC required first |
| **Algorithm Changes** | Scoring formula modifications | RFC + benchmark required |
| **New Examples** | Integration examples | Direct PR, functional review |

### 6.4 RFC Process

For significant changes, contributors MUST submit an RFC:

```
docs/rfcs/
├── 0000-template.md
├── 0001-trust-score-v2.md
├── 0002-verification-level-5.md
└── ...
```

RFC Process:

1. Create RFC document with motivation, detailed design, alternatives
2. Open PR for RFC discussion
3. Community feedback period: 30 days
4. Core team decision within 14 days
5. Implement after RFC approval

### 6.5 Pull Request Requirements

| Type | Tests | Documentation | Benchmarks |
|------|-------|---------------|------------|
| Bug Fix | ✅ Unit tests | ❌ Optional | ❌ Optional |
| Feature | ✅ Unit + Integration | ✅ Required | ✅ Required |
| Algorithm | ✅ Unit + Property-based | ✅ Required | ✅ Required |
| Port | ✅ Unit | ✅ Required | ❌ Optional |

### 6.6 Commit Message Convention

Follow Conventional Commits:

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

Types: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`

Examples:

```
feat(calculator): add weighted average for user ratings
fix(sdk): handle nil response in batch trust score
docs(spec): clarify verification level requirements
```

---

## 7. Community Building

### 7.1 Communication Channels

| Channel | Purpose | Moderation |
|---------|--------|------------|
| GitHub Discussions | Protocol questions, RFC discussions | Core team |
| Discord `#trust-protocol` | Real-time chat, community support | Community leads |
| Twitter/X | Announcements, news | Marketing |
| Newsletter | Monthly digest | Marketing |

### 7.2 Contributor Recognition

| Level | Requirement | Recognition |
|------|-------------|------------|
| **First Contribution** | 1 merged PR | Thank you email, sticker pack |
| **Contributor** | 5 merged PRs | GitHub "Contributor" badge |
| **Reviewer** | 10 PRs reviewed | Reviewer permissions |
| **Maintainer** | Significant ongoing contributions | Commit access, co-author credit |

### 7.3 Advisory Committee

Form an advisory committee with representatives from:

- Major AI agent frameworks (LangChain, AutoGen, CrewAI)
- Cloud providers (AWS, GCP, Azure)
- Enterprise adopters
- Academic researchers

Meet quarterly; provide strategic guidance on protocol direction.

### 7.4 Event Strategy

| Event | Type | Goal |
|-------|------|------|
| **Trust Protocol Summit** (annual) | Conference | Community gathering, spec v1.0 launch |
| **Office Hours** (bi-weekly) | Virtual | Q&A, RFC discussions |
| **Hackathons** (quarterly) | Virtual | New integrations, example contributions |

---

## 8. Adoption Roadmap

### 8.1 Q4 2026 Launch Plan

```
September 2026:
├── September 1-15: Code freeze, documentation finalization
├── September 15-30: Security audit, license review
│
October 2026:
├── October 1: Public beta — early access for partners
├── October 15: Trust Protocol Summit — v1.0 launch
├── October 31: General availability — open source release
│
November-December 2026:
├── Major platform integrations announced
├── Advisory committee first meeting
├── Community contribution milestones
```

### 8.2 Adoption Milestones

| Milestone | Target | Success Indicator |
|-----------|--------|------------------|
| **Beta Partners** | 5 platforms | Active API usage |
| **General Availability** | 20 platforms | API integration in production |
| **LangChain Integration** | Q4 2026 GA | Official LangChain provider |
| **AutoGen Integration** | Q1 2027 | Community provider package |
| **Enterprise Adoption** | 3 enterprise accounts | Paid Trust API usage |

### 8.3 Long-Term Ecosystem Goals

| Year | Goal |
|------|------|
| **2026** | Establish protocol as de facto standard |
| **2027** | 50+ platforms, 500+ contributors |
| **2028** | ISO/IETF standardization consideration |
| **2029** | Global trust network for AI agents |

---

## 9. Governance Model

### 9.1 Governance Layers

```
┌─────────────────────────────────────────────────────────────┐
│                    Advisory Committee                        │
│            (Strategic direction, major decisions)            │
├─────────────────────────────────────────────────────────────┤
│                      Core Team                              │
│        (Day-to-day decisions, PR reviews, releases)         │
├─────────────────────────────────────────────────────────────┤
│                    Maintainers                               │
│        (Package-specific decisions, issue triage)            │
├─────────────────────────────────────────────────────────────┤
│                    Contributors                              │
│              (Code, docs, examples, RFCs)                    │
├─────────────────────────────────────────────────────────────┤
│                     Community                                │
│           (Feedback, discussions, adoption)                   │
└─────────────────────────────────────────────────────────────┘
```

### 9.2 Decision-Making Process

| Decision Type | Who Decides | Process |
|--------------|-------------|---------|
| Bug fix | PR reviewer | Direct approval |
| Documentation change | Maintainer | PR approval |
| New example | Maintainer | PR approval |
| Algorithm change | Core team + RFC | RFC process |
| New package | Core team + RFC | RFC + committee review |
| Spec breaking change | Core team + RFC + committee | Full RFC process |

### 9.3 Release Process

```
Versioning: SemVer (MAJOR.MINOR.PATCH)

MAJOR: Breaking changes to protocol spec
MINOR: New features, backward-compatible
PATCH: Bug fixes, documentation updates

Release Process:
1. Changelog generation from conventional commits
2. Release candidate for 7 days
3. Security review for major/minor
4. Tag and publish
5. Announcement to community
```

---

## 10. Risk Mitigation

### 10.1 Risk Matrix

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Competitors adopt and fork protocol | Medium | Low | Community momentum; first-mover advantage |
| Corporate takeover of project | Low | High | Foundation transfer option; multi-vendor governance |
| Low community adoption | Medium | High | Beta partner commitments; marketing push |
| Fragmentation (forks with incompatible changes) | Medium | Medium | Clear RFC process; "spec compliance" certification |
| Security vulnerabilities in open code | Medium | High | Security audit before release; responsible disclosure |

### 10.2 Mitigation Strategies

#### Risk: Protocol Fragmentation

**Strategy**: "Spec Compliance" Certification

- Create compliance test suite
- Allow "Trust Protocol Compliant" badge for implementations
- Require passing compliance tests for official integration listings

#### Risk: Low Community Adoption

**Strategy**: Incentivized Adoption

- Provide free Trust API access for beta partners
- Co-marketing opportunities for major integrations
- Priority support for early adopters

#### Risk: Security Vulnerabilities

**Strategy**: Responsible Disclosure Program

```
SECURITY.md
===========

Vulnerability Reports: security@functionfly.dev

Scope: trust-protocol repositories
Response: 48 hours acknowledgement
Timeline: 90 days to fix or disclose

Rewards: Bug bounty program (TBD based on severity)
```

#### Risk: Corporate Takeover

**Strategy**: Foundation Option

- Donate protocol to existing foundation (Linux Foundation, Apache)
- Establish separate legal entity if project reaches critical mass
- Ensure governance documents prevent unilateral control

### 10.3 Contingency Plans

| Scenario | Contingency |
|----------|-------------|
| Maintained fork gains more traction | Adopt successful changes; merge back |
| Major enterprise demands proprietary license | Dual-license option with commercial terms |
| Core team burnout/attrition | Document processes; empower maintainers |
| Security incident | Immediate patch; transparent post-mortem |

---

## Appendix A: Checklist for Q4 Launch

### Documentation

- [ ] `SPEC.md` finalized and reviewed
- [ ] `README.md` for each package
- [ ] `CONTRIBUTING.md` complete
- [ ] `CODE_OF_CONDUCT.md` adopted
- [ ] `SECURITY.md` with disclosure policy
- [ ] `GOVERNANCE.md` with decision process

### Code

- [ ] All packages pass CI/CD
- [ ] Security audit completed
- [ ] Benchmarks documented
- [ ] Test coverage > 80%
- [ ] Examples tested end-to-end

### Community

- [ ] Discord server created and moderated
- [ ] Twitter account established
- [ ] Newsletter signup configured
- [ ] Advisory committee members confirmed
- [ ] First blog post drafted

### Launch

- [ ] Press release drafted
- [ ] Partner announcements coordinated
- [ ] Launch event scheduled
- [ ] Analytics tracking configured
- [ ] Support channels staffed

---

## Appendix B: Reference Documents

| Document | Location |
|----------|----------|
| Trust Protocol Specification | [`docs/TRUST_PROTOCOL_SPEC.md`](docs/TRUST_PROTOCOL_SPEC.md) |
| Moat Competitive Analysis | [`plans/MOAT_COMPETITIVE_ANALYSIS.md`](../plans/MOAT_COMPETITIVE_ANALYSIS.md) |
| Moat Action Items | [`plans/MOAT_ANALYSIS_TODO.md`](../plans/MOAT_ANALYSIS_TODO.md) |
| Trust API Implementation | [`internal/api/handlers/trustapi/`](internal/api/handlers/trustapi/) |
| Trust Score Calculation | [`internal/storage/registry/trust_repository.go`](internal/storage/registry/trust_repository.go) |

---

**Document Version**: 1.0.0-draft  
**Last Updated**: 2026-03-21  
**Next Review**: Q2 2026  
**Owner**: FunctionFly Platform Team  
**Contact**: <trust-protocol@functionfly.dev>
