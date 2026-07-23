# Capability Specification Template

> This document defines the standard template for all business capabilities in the Investment AI platform.

---

# 1. Overview

## Capability Name

Describe the official name of the capability.

## Purpose

Explain why this capability exists and the business problem it solves.

## Business Value

Describe the value delivered to end users and the overall platform.

---

# 2. Scope

## In Scope

List the responsibilities that belong to this capability.

## Out of Scope

List the responsibilities that explicitly do not belong to this capability.

---

# 3. Responsibilities

Describe the primary business responsibilities.

For each responsibility:

* Objective
* Expected Outcome

---

# 4. Business Inputs

Describe the business information required by this capability.

For each input:

* Name
* Description
* Source
* Required / Optional

---

# 5. Business Outputs

Describe the business objects produced by this capability.

For each output:

* Name
* Description
* Consumers

Outputs should represent business concepts rather than implementation details.

---

# 6. Domain Ownership

Define the concepts owned by this capability.

## Owns

Business entities, concepts and rules maintained by this capability.

## References

Business concepts owned by other capabilities but referenced here.

---

# 7. Core Business Concepts

Define the ubiquitous language for this capability.

Each concept should include:

* Name
* Definition
* Business Meaning

---

# 8. Business Rules

Describe business rules in a technology-independent way.

Example format:

Rule ID

Description

Condition

Expected Outcome

Priority

---

# 9. Domain Events

Describe business events published by this capability.

For each event:

* Event Name
* Trigger
* Description
* Consumers

---

# 10. Dependencies

Describe the business capabilities required by this capability.

For each dependency:

* Capability
* Purpose
* Required Information

Avoid technical dependencies such as databases, APIs or messaging systems.

---

# 11. Consumers

Identify the capabilities that consume the outputs of this capability.

Examples:

* Reporting
* Portfolio Intelligence
* Risk Intelligence
* Research

---

# 12. Business KPIs

Define measurable indicators of business success.

Examples:

* Insight Accuracy
* Research Time Reduction
* Coverage
* Explanation Quality
* User Satisfaction

---

# 13. Non-Functional Requirements

Describe quality attributes required by the business.

Examples:

* Freshness
* Explainability
* Traceability
* Reliability
* Availability
* Auditability

---

# 14. Risks and Constraints

Describe business risks.

Examples:

* Missing Data
* Delayed Data
* Low Confidence
* Conflicting Evidence

---

# 15. Future Enhancements

Describe possible future evolution.

Examples:

* AI Reasoning
* Forecasting
* Reflection Learning
* Cross Market Analysis

---

# 16. Related Workflows

List workflows that use this capability.

Examples:

* Daily Market Analysis
* Investment Research
* Portfolio Review
* Report Generation

---

# 17. Related Architecture Decisions

Reference the Architecture Decision Records (ADRs) related to this capability.

---

# 18. Open Questions

Capture unresolved business questions that require further discussion.

Examples:

* Should confidence scores be rule-based or AI-generated?
* How should conflicting evidence be resolved?
* What level of explanation is sufficient for end users?

---

# Document History

| Version | Date | Author            | Description     |
| ------- | ---- | ----------------- | --------------- |
| 1.0     | TBD  | Architecture Team | Initial version |
