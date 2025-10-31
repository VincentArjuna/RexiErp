# Starter Template or Existing Project

Based on my analysis of the PRD, this is a greenfield Go-based microservices project. The PRD clearly states:

- **Repository Structure:** Monorepo (explicitly decided in PRD)
- **Service Architecture:** Microservices with domain-driven design
- **Technology Stack:** Go programming language with Gin framework, GORM, PostgreSQL, Redis, and container-based deployment

**Recommendation:** No starter template will be used. This is a custom Go microservices project built from scratch to meet Indonesian MSME requirements, with specific technical constraints around zero-cost licensing and cloud-agnostic deployment.

**Rationale:**
- The project requires specific Indonesian compliance features (e-Faktur, BPJS, PPn 11%) that generic templates don't support
- Zero-cost licensing requirement eliminates most commercial templates
- Cloud-agnostic deployment needs custom infrastructure as code
- Monorepo structure with microservices warrants custom organization
