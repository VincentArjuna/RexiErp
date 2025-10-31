# Project Brief: RexiERP

<!-- Powered by BMAD™ Core -->

## Executive Summary

**RexiERP** is a cloud-agnostic zero-cost microservices ERP backend for Indonesian micro, small, and medium enterprises (MSMEs), addressing the critical gap where 64 million MSMEs struggle with high-cost ERP solutions like SAP and Oracle while local alternatives are often cloud-locked or expensive. The platform provides comprehensive accounting, inventory, HR & payroll, tax compliance, and reporting capabilities as a self-hosted, API-first solution built entirely on zero-cost open-source infrastructure.

**Key Value Proposition:** Free software with no licensing fees, deployment flexibility across any cloud provider (AWS, Azure, GCP) or on-premises, and comprehensive functionality tailored to Indonesian business needs including PPn 11% tax compliance and Mekari integration.

## Problem Statement

### Current State and Pain Points
- **Prohibitive Costs**: Major ERP solutions (SAP, Oracle) cost $1,000-2,500 per user monthly - completely inaccessible for Indonesian MSMEs
- **Cloud Lock-in**: Local solutions like Mekari, JojoWay require specific cloud providers, limiting flexibility
- **Integration Complexity**: Existing systems lack API-first design, making integration with existing tools difficult
- **Regulatory Compliance Gap**: Many solutions don't properly handle Indonesian tax requirements (PPn 11%, PPh 21, 23, 26)
- **Feature Bloat**: Enterprise ERPs include unnecessary complexity for small business needs
- **Technical Debt**: Legacy systems are difficult to maintain and scale

### Impact and Urgency
- **Market Size**: 64 million Indonesian MSMEs representing 97% of domestic businesses
- **Economic Impact**: MSMEs contribute 60.7% to Indonesia's GDP
- **Competitive Disadvantage**: Businesses without proper ERP systems struggle with efficiency and compliance
- **Digital Transformation Gap**: Post-COVID acceleration makes modern ERP systems essential

## Proposed Solution

**RexiERP** is a comprehensive microservices-based ERP backend that delivers enterprise-grade functionality at zero cost through strategic use of open-source technologies and cloud-agnostic architecture.

### Core Concept
- **Zero-Cost Stack**: Entire solution built on free, open-source infrastructure
- **Microservices Architecture**: Scalable, maintainable services with clear domain boundaries
- **API-First Design**: RESTful APIs with comprehensive documentation for easy integration
- **Cloud Agnostic**: Deploy anywhere - AWS, Azure, GCP, or on-premises
- **Indonesian Compliance**: Built-in support for local tax regulations and business practices

### Key Differentiators
- **Cost Structure**: No licensing fees vs competitors' $1,000-2,500/user/month
- **Deployment Freedom**: Multi-cloud capable vs single-cloud competitors
- **Integration Ready**: Modern API architecture vs legacy systems
- **Local Expertise**: Indonesian compliance built-in vs generic international solutions

## Target Users

### Primary User Segment: Indonesian MSMEs (50-250 employees)
**Profile:**
- **Revenue Range**: Rp 2.5B - 50B annually
- **Industries:** Manufacturing, Retail, Services, Distribution
- **Current Pain Points:** Excel-based operations, fragmented systems, compliance concerns
- **Technical Capability:** Basic IT infrastructure, may have dedicated IT staff
- **Workflow Needs:** Invoicing, inventory tracking, payroll, tax reporting, financial statements

**Specific Needs:**
- PPn 11% tax invoice generation and reporting
- Multi-warehouse inventory management
- Employee payroll with BPJS calculations
- Financial reporting for tax purposes
- Integration with existing accounting systems

### Secondary User Segment: Indonesian System Integrators
**Profile:**
- **Business Type:** IT consulting firms, software development agencies
- **Current Services:** ERP implementation, custom software development
- **Technical Capability:** Advanced development and DevOps skills
- **Opportunity:** White-label solution for their clients

## Goals & Success Metrics

### Business Objectives
- **Market Penetration**: Achieve 1,000 active MSME deployments within 24 months
- **Cost Leadership**: Maintain zero software licensing costs while delivering premium functionality
- **Ecosystem Growth**: Establish 50+ system integrator partners within 18 months
- **User Engagement**: Achieve 80% monthly active user rate across deployed instances
- **Technical Excellence**: Maintain 99.9% uptime across all microservices

### User Success Metrics
- **Implementation Time**: Complete ERP deployment in < 2 weeks vs industry average 3-6 months
- **User Adoption**: 90% of target employees actively using core features within 30 days
- **Process Efficiency**: 40% reduction in manual accounting/inventory processes
- **Compliance Accuracy**: 100% accurate tax report generation
- **Cost Savings**: Average 70% reduction in ERP TCO vs commercial alternatives

### Key Performance Indicators (KPIs)
- **Deployment Velocity**: 50+ new deployments per month by month 12
- **Integration Success**: 95% success rate for API integrations with partner systems
- **Support Efficiency**: <24 hour response time for critical issues
- **Community Growth**: 1,000+ active contributors in open-source community
- **Documentation Quality**: 90%+ user satisfaction with API documentation

## MVP Scope

### Core Features (Must Have)
- **User Management**: Role-based access control (Super Admin, Finance Admin, Sales Admin, Inventory Admin, Employee)
- **Chart of Accounts**: Full Indonesian accounting standards (SAK) compliance
- **Journal Entry**: Manual journal entries with approval workflows
- **Bank Integration**: Transaction sync with major Indonesian banks (BCA, Mandiri, BNI, BRI)
- **Cash Book**: Real-time cash position tracking
- **AR Management**: Invoice generation, payment tracking, aging reports
- **AP Management**: Bill processing, payment scheduling, vendor management
- **General Ledger**: Trial balance, income statements, balance sheets, cash flow statements
- **Tax Compliance**: PPn 11% invoicing and reporting, PPh calculations
- **Product Management**: Multi-category product catalog with pricing variants
- **Stock Management**: Real-time inventory tracking, low stock alerts
- **Purchase Orders**: PO creation, approval workflows, receipt processing
- **Basic Reporting**: Financial statements, inventory reports, sales reports
- **API Access**: Complete RESTful API with Swagger documentation

### Out of Scope for MVP
- **Advanced Reporting**: Custom report builder, advanced analytics dashboards
- **Multi-tenant Architecture**: Single-tenant deployment only in MVP
- **Mobile Applications**: Web-based only initially
- **Advanced Workflows**: Complex approval chains beyond 2 levels
- **BI Integration**: Power BI, Tableau connectors
- **Advanced Inventory**: Serial tracking, batch management
- **Fixed Assets Management**: Asset depreciation, maintenance scheduling
- **HR Management**: Leave requests, performance reviews, recruitment
- **Payroll System**: Employee salary calculations, BPJS integrations

### MVP Success Criteria
**MVP succeeds when:**
- 10 pilot MSMEs successfully deploy and use the system for 3+ months
- Core financial workflows (invoicing, payment, reporting) function without errors
- API documentation enables smooth integrations with external systems
- Indonesian tax compliance features generate accurate reports
- System handles 1,000+ concurrent users across 10 deployments
- Average response time <500ms for critical operations

## Post-MVP Vision

### Phase 2 Features
- **Advanced Inventory**: Serial number tracking, batch/lot management, stock adjustments
- **Fixed Assets**: Asset registration, depreciation calculations, disposal tracking
- **HR Management**: Employee profiles, leave management, attendance tracking
- **Payroll System**: Salary calculations, BPJS integrations, payslip generation
- **Advanced Reporting**: Custom report builder, scheduled reports, export formats
- **Multi-tenant Architecture**: Shared infrastructure for cost efficiency
- **Enhanced APIs**: GraphQL support, webhook subscriptions, rate limiting

### Long-term Vision
**12-24 month roadmap:**
- **Mobile Applications**: Native iOS/Android apps for field operations
- **AI-powered Insights**: Cash flow predictions, inventory optimization, anomaly detection
- **Marketplace Integration**: E-commerce platform connectors (Tokopedia, Shopee)
- **Advanced Analytics**: Business intelligence dashboards, predictive analytics
- **Industry Verticals**: Specialized modules for manufacturing, retail, services
- **International Expansion**: Southeast Asian market adaptation (Malaysia, Singapore, Thailand)

### Expansion Opportunities
- **Managed Service**: RexiERP Cloud - fully managed hosting option
- **Certification Program**: Official RexiERP implementation partner certification
- **Plugin Ecosystem**: Third-party app marketplace for extensions
- **Education Platform**: ERP training and certification for Indonesian businesses
- **Compliance Services**: Automated tax filing, audit preparation tools

## Technical Considerations

### Platform Requirements
- **Target Platforms:** Any cloud provider (AWS, Azure, GCP) or on-premises deployment
- **Container Orchestration:** Docker, Docker Swarm, Kubernetes support
- **Load Balancing:** Nginx, Traefik, or cloud provider load balancers
- **Performance Requirements:** <500ms response time for 95% of requests
- **Scalability:** Horizontal scaling for all services
- **Monitoring:** Prometheus, Grafana for observability

### Technology Preferences
- **Backend:** Go (Golang) for microservices performance and concurrency
- **API Gateway:** Go-zero framework for high-performance routing
- **Database:** PostgreSQL for transactional data, Redis for caching
- **Message Queue:** Redis Pub/Sub for real-time communication
- **Authentication:** JWT with refresh token mechanism
- **Documentation:** Swagger/OpenAPI 3.0 for API docs
- **Testing:** Go testing framework with 80%+ coverage requirement
- **CI/CD:** GitHub Actions for automated testing and deployment

### Architecture Considerations
- **Repository Structure:** Monorepo with clear service boundaries
- **Service Architecture:** Microservices with domain-driven design
- **Integration Requirements:** RESTful APIs, event-driven communication
- **Security/Compliance:** JWT authentication, RBAC, encrypted data storage
- **Deployment Strategy:** Docker containers with optional Kubernetes
- **Monitoring Strategy:** Structured logging, metrics collection, health checks
- **Data Migration:** Tools for importing from Excel, existing systems

## Constraints & Assumptions

### Constraints
- **Budget:** Zero software licensing costs - all components must be open-source
- **Timeline:** MVP delivery in 6 months, full platform in 18 months
- **Resources:** Small core team (3-5 developers) supported by open-source community
- **Technical:** Must run on commodity hardware, no specialized requirements
- **Compliance:** Must meet Indonesian accounting and tax regulations
- **Language:** Primary interface in Bahasa Indonesia, English secondary

### Key Assumptions
- Target MSMEs have basic IT infrastructure and technical capability
- Cloud-agnostic deployment is a key decision factor for Indonesian businesses
- Open-source approach can attract sufficient community contribution
- API-first design enables integration with existing business tools
- Indonesian compliance requirements can be met through configuration
- Performance requirements can be met with zero-cost infrastructure stack

## Risks & Open Questions

### Key Risks
- **Community Adoption:** Risk of insufficient open-source community engagement
- **Compliance Complexity:** Indonesian tax regulations may change more frequently than expected
- **Support Scaling:** Challenge of providing quality support at zero cost
- **Technical Debt:** Rapid development may accumulate maintenance challenges
- **Market Education:** Need to educate MSMEs about self-hosted vs SaaS models
- **Security Vulnerabilities:** Open-source components may introduce security risks

### Open Questions
- **Monetization Strategy:** How to sustain development without licensing fees?
- **Support Model:** What level of support can be provided to zero-cost customers?
- **Feature Prioritization:** How to balance Indonesian-specific vs global feature needs?
- **Community Governance:** How to manage open-source contributions effectively?
- **Partnership Strategy:** Which system integrators should be priority targets?

### Areas Needing Further Research
- **Competitive Analysis:** Deep dive into local competitor pricing and features
- **User Interviews:** Direct feedback from 20+ target MSMEs on feature priorities
- **Technical Feasibility:** Performance testing of proposed technology stack
- **Compliance Validation:** Review with Indonesian accounting/tax experts
- **Community Interest:** Gauging open-source developer interest in the project

## Appendices

### A. Research Summary
**Market Analysis:**
- Indonesian MSME market: 64 million businesses, 60.7% GDP contribution
- Competitive landscape dominated by high-cost international solutions
- Gap exists for zero-cost, cloud-agnostic alternatives
- Strong demand for Indonesian compliance features

**Technical Validation:**
- Go language ecosystem mature for microservices development
- Zero-cost stack proven scalable (Docker, PostgreSQL, Redis)
- Cloud deployment patterns well-established
- API-first architecture gaining market acceptance

### B. Stakeholder Input
**Initial Feedback from Brainstorming Session:**
- Strong validation of zero-cost approach
- Emphasis on Indonesian compliance requirements
- Request for API-first design for integration flexibility
- Concerns about support model for free software
- Interest in community-driven development approach

### C. References
- **Source Documents:** docs/brainstorming-session-results.md
- **Market Data:** Indonesian Ministry of Cooperatives and SMEs statistics
- **Technical Standards:** Indonesian Accounting Standards (SAK)
- **Tax Regulations:** Indonesian Directorate General of Taxes guidelines

## Next Steps

### Immediate Actions
1. **Technical Validation:** Set up proof-of-concept with core microservices stack
2. **User Research:** Conduct interviews with 20 target MSMEs for feature validation
3. **Community Building:** Launch GitHub repository and initial documentation
4. **Legal Review:** Confirm compliance requirements with Indonesian tax advisors
5. **Team Planning:** Identify core development team and advisory board members
6. **Infrastructure Setup:** Establish development, testing, and CI/CD environments

### PM Handoff
This Project Brief provides the full context for RexiERP. Please start in 'PRD Generation Mode', review the brief thoroughly to work with the user to create the PRD section by section as the template indicates, asking for any necessary clarification or suggesting improvements.

---

*Project Brief generated on October 29, 2025*
*Based on comprehensive brainstorming session results*
*Next: Product Requirements Document (PRD) Development*