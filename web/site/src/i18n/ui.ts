// src/i18n/ui.ts — Translation dictionaries for the marketing site

export const languages = {
  en: "English",
  es: "Español",
} as const;

export const defaultLang = "en";
export const showDefaultLang = false; // / = English, /es/ = Spanish

export const ui = {
  en: {
    "nav.docs": "Docs",
    "nav.forAgents": "For AI Agents",
    "nav.trust": "Trust",
    "nav.pricing": "Pricing",
    "nav.blog": "Blog",
    "nav.dashboard": "Dashboard",

    "footer.tagline": "Functions deployed. Agents trusted.",
    "footer.product": "Product",
    "footer.documentation": "Documentation",
    "footer.trustLayer": "Trust layer",
    "footer.forAIAgents": "For AI agents",
    "footer.changelog": "Changelog",
    "footer.company": "Company",
    "footer.about": "About",
    "footer.careers": "Careers",
    "footer.contact": "Contact",
    "footer.community": "Community",
    "footer.discord": "Discord",
    "footer.githubDiscussions": "GitHub Discussions",
    "footer.legal": "Legal",
    "footer.privacy": "Privacy",
    "footer.terms": "Terms",
    "footer.sla": "SLA",
    "footer.copyright": "All rights reserved.",

    "language.title": "Language",
    "language.description": "Choose your preferred language",
    "language.detected": "Detected",
    "language.autoDetected": "Auto-detected from your browser",

    // Index page
    "hero.badge": "Now in Invite-Only Beta",
    "hero.titleLine": "The Platform for",
    "hero.titleGradient": "AI Agent Tooling",
    "hero.lead":
      "Publish once, run everywhere, get paid. Verified functions, trust scores, encrypted secrets vault, durable state, and continuous learning — everything agent tooling needs in one platform.",
    "hero.ctaPrimary": "Request Access",
    "hero.ctaSecondary": "View Documentation",
    "hero.stat1Value": "Publish",
    "hero.stat1Label": "once, run everywhere",
    "hero.stat2Value": "Trust",
    "hero.stat2Label": "scores + verification",
    "hero.stat3Value": "Vault",
    "hero.stat3Label": "zero-knowledge encrypted",
    "hero.stat4Value": "Ecosystem",
    "hero.stat4Label": "continuous agent learning",

    "audience.title": "Who benefits most from FunctionFly",
    "audience.intro":
      "From solo founders to enterprise teams, FunctionFly provides the reliability and flexibility you need to focus on building great products.",

    "audience.founder.title": "Indie SaaS Founder",
    "audience.founder.sub": "Deploying your first app",
    "audience.founder.desc":
      "You built an amazing product and now need trustworthy agent tooling without enterprise overhead. FunctionFly gives you verified functions, clear trust signals, and a path to scale safely.",
    "audience.founder.b1": "Publish once, run everywhere",
    "audience.founder.b2": "Trust scores and verification tiers",
    "audience.founder.b3": "Zero-knowledge vault for secrets",
    "audience.founder.b4": "Pay as you grow",

    "audience.startup.title": "Growing Startup",
    "audience.startup.sub": "Scaling agent + API workloads",
    "audience.startup.desc":
      "Your team is shipping fast and agents are calling more tools every week. You need policy-friendly trust metadata, attestations, and execution-backed signals—not a pile of unvetted endpoints.",
    "audience.startup.b1": "Multi-level verification workflows",
    "audience.startup.b2": "Signed artifacts and revocation",
    "audience.startup.b3": "Execution history for trust scoring",
    "audience.startup.b4": "Marketplace discovery with trust filters",

    "audience.enterprise.title": "Enterprise",
    "audience.enterprise.sub": "Governance and compliance",
    "audience.enterprise.desc":
      "Your organization needs auditability, least-privilege access to tools, and consistent trust policies across agents and teams. FunctionFly aligns verification, signing, and vault semantics with how you already think about risk.",
    "audience.enterprise.b1": "Enterprise-grade security posture",
    "audience.enterprise.b2": "Attestations and policy-ready metadata",
    "audience.enterprise.b3": "Multi-team and multi-app patterns",
    "audience.enterprise.b4": "Room for custom SLAs and support",

    "features.title": "Everything your agent tooling platform needs",
    "features.subtitle": "From publishing to production — five integrated capabilities in one platform. "
      + "Verification + trust scoring so agents can act safely.",
    "features.publishing.title": "Verified Tool Publishing",
    "features.publishing.desc":
      "Turn code into trusted functions with verification checks and signed artifacts before they can be used by agents.",
    "features.revocation.title": "Verification Levels & Revocation",
    "features.revocation.desc":
      "Trust is earned. Verify at multiple levels, sign with platform keys, and revoke when issues are found.",
    "features.trace.title": "Execution Trace & Trust Scores",
    "features.trace.desc":
      'Agents get history-backed trust signals so "works" and "safe" are distinguishable and policy-driven.',
    "features.tooling.title": "Agent-ready Tooling",
    "features.tooling.desc":
      "Expose capabilities with clear manifests and schemas—so agents can select tools that match their trust policy.",
    "features.marketplace.title": "Marketplace for Verified Tools",
    "features.marketplace.desc":
      "Creators publish once. Agents discover trusted functions with built-in trust filters and ranked placement.",
    "features.routing.title": "Trust-powered Routing",
    "features.routing.desc":
      "Enforce budgets and trust constraints at runtime—so agent execution is predictable, not reckless.",

    "code.title": "Verify & publish in seconds",
    "code.subtitle": "Turn your code into an agent-ready trusted tool",

    "cta.title": "Where trusted tools meet agent execution",
    "cta.lead":
      "Start free. Verified tooling and auditable execution history help your agents move faster—with fewer surprises.",
    "cta.ctaPrimary": "Request Access",
    "cta.ctaSecondary": "Contact Sales",
  },
  es: {
    "nav.docs": "Docs",
    "nav.forAgents": "Para Agentes IA",
    "nav.trust": "Confianza",
    "nav.pricing": "Precios",
    "nav.blog": "Blog",
    "nav.dashboard": "Panel",

    "footer.tagline": "Funciones desplegadas. Agentes confiados.",
    "footer.product": "Producto",
    "footer.documentation": "Documentación",
    "footer.trustLayer": "Capa de confianza",
    "footer.forAIAgents": "Para agentes IA",
    "footer.changelog": "Historial",
    "footer.company": "Empresa",
    "footer.about": "Acerca de",
    "footer.careers": "Empleo",
    "footer.contact": "Contacto",
    "footer.legal": "Legal",
    "footer.privacy": "Privacidad",
    "footer.terms": "Términos",
    "footer.sla": "SLA",
    "footer.copyright": "Todos los derechos reservados.",

    "language.title": "Idioma",
    "language.description": "Elige tu idioma preferido",
    "language.detected": "Detectado",
    "language.autoDetected": "Detectado automáticamente desde tu navegador",

    // Index page
    "hero.badge": "Disponible ahora",
    "hero.titleLine": "La capa de confianza para",
    "hero.titleGradient": "Agentes IA",
    "hero.lead":
      "Herramientas verificadas, firmadas y auditables con puntuaciones de confianza y una bóveda de conocimiento cero para la ejecución segura de agentes IA.",
    "hero.ctaPrimary": "Solicitar Acceso",
    "hero.ctaSecondary": "Ver Documentación",
    "hero.stat1Value": "Verificado",
    "hero.stat1Label": "niveles de verificación",
    "hero.stat2Value": "Firmado",
    "hero.stat2Label": "atestaciones de funciones",
    "hero.stat3Value": "Privado",
    "hero.stat3Label": "bóveda de conocimiento cero",

    "audience.title": "¿Quién se beneficia más de FunctionFly?",
    "audience.intro":
      "Desde fundadores independientes hasta equipos empresariales, FunctionFly ofrece la fiabilidad y flexibilidad que necesitas para centrarte en crear grandes productos.",

    "audience.founder.title": "Fundador SaaS Independiente",
    "audience.founder.sub": "Desplegando tu primera app",
    "audience.founder.desc":
      "Has construido un producto increíble y ahora necesitas herramientas de agentes confiables sin la complejidad empresarial. FunctionFly te da funciones verificadas, señales claras de confianza y un camino para escalar de forma segura.",
    "audience.founder.b1": "Publica una vez, ejecuta en todas partes",
    "audience.founder.b2":
      "Puntuaciones de confianza y niveles de verificación",
    "audience.founder.b3": "Bóveda de conocimiento cero para secretos",
    "audience.founder.b4": "Paga a medida que creces",

    "audience.startup.title": "Startup en Crecimiento",
    "audience.startup.sub": "Escalando cargas de agentes + API",
    "audience.startup.desc":
      "Tu equipo envía rápido y los agentes llaman a más herramientas cada semana. Necesitas metadatos de confianza compatibles con políticas, atestaciones y señales basadas en ejecución—no un puñado de endpoints sin verificar.",
    "audience.startup.b1": "Flujos de trabajo de verificación multinivel",
    "audience.startup.b2": "Artefactos firmados y revocación",
    "audience.startup.b3":
      "Historial de ejecución para puntuación de confianza",
    "audience.startup.b4":
      "Descubrimiento de marketplace con filtros de confianza",

    "audience.enterprise.title": "Empresa",
    "audience.enterprise.sub": "Gobernanza y cumplimiento",
    "audience.enterprise.desc":
      "Tu organización necesita auditabilidad, acceso de mínimo privilegio a herramientas y políticas de confianza consistentes entre agentes y equipos. FunctionFly alinea verificación, firma y semántica de bóveda con cómo ya piensas sobre el riesgo.",
    "audience.enterprise.b1": "Postura de seguridad de nivel empresarial",
    "audience.enterprise.b2": "Atestaciones y metadatos listos para políticas",
    "audience.enterprise.b3": "Patrones multi-equipo y multi-app",
    "audience.enterprise.b4": "Espacio para SLAs y soporte personalizados",

    "features.title":
      "Todo lo que necesitas para confiar en herramientas de agentes",
    "features.subtitle": "From publishing to production — five integrated capabilities in one platform. "
      + "Verificación + puntuación de confianza para que los agentes actúen de forma segura.",
    "features.publishing.title": "Publicación de Herramientas Verificadas",
    "features.publishing.desc":
      "Convierte código en funciones confiables con verificaciones y artefactos firmados antes de que los agentes puedan usarlas.",
    "features.revocation.title": "Niveles de Verificación y Revocación",
    "features.revocation.desc":
      "La confianza se gana. Verifica en múltiples niveles, firma con claves de plataforma y revoca cuando se encuentren problemas.",
    "features.trace.title":
      "Trazabilidad de Ejecución y Puntuaciones de Confianza",
    "features.trace.desc":
      'Los agentes obtienen señales de confianza respaldadas por historial para que "funciona" y "es seguro" sean distinguibles y governed por políticas.',
    "features.tooling.title": "Herramientas Listas para Agentes",
    "features.tooling.desc":
      "Expón capacidades con manifiestos y esquemas claros—para que los agentes puedan seleccionar herramientas que coincidan con su política de confianza.",
    "features.marketplace.title": "Marketplace de Herramientas Verificadas",
    "features.marketplace.desc":
      "Los creadores publican una vez. Los agentes descubren funciones confiables con filtros de confianza integrados y posicionamiento mejorado.",
    "features.routing.title": "Enrutamiento Impulsado por Confianza",
    "features.routing.desc":
      "Aplica presupuestos y restricciones de confianza en ejecución—para que la ejecución de agentes sea predecible y no reckless.",

    "code.title": "Verifica y publica en segundos",
    "code.subtitle":
      "Convierte tu código en una herramienta confiable lista para agentes",

    "cta.title":
      "Donde las herramientas confiables se encuentran con la ejecución de agentes",
    "cta.lead":
      "Empieza gratis. Herramientas verificadas e historial de ejecución auditable ayudan a tus agentes a moverse más rápido—con menos sorpresas.",
    "cta.ctaPrimary": "Solicitar Acceso",
    "cta.ctaSecondary": "Contactar Ventas",
  },
} as const;
