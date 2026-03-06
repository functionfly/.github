# 🚀 FunctionFly Production Deployment

> **High-level overview for DevOps, SRE, and Platform Engineers**

FunctionFly's production deployment runs on a cost-optimized, two-server bare metal architecture with managed PostgreSQL. This setup delivers enterprise-grade serverless function execution starting at **$65/month**.

---

## 🏗️ Architecture Summary

```
┌─────────────────────────────────────────────────────────────────────┐
│                    FunctionFly Production Stack                     │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌─────────────────────┐        ┌──────────────────────────────┐   │
│  │  🖥️ Server 1         │◄──────►│  🖥️ Server 2                  │   │
│  │  App Stack ($20/mo) │        │  Runtime Stack ($20/mo)      │   │
│  ├─────────────────────┤        ├──────────────────────────────┤   │
│  │ • Orchestrator API  │        │ • runtime-local              │   │
│  │ • PostgreSQL        │        │ • runtime-nodejs             │   │
│  │ • Redis Cache       │        │ • runtime-python             │   │
│  │ • Caddy Reverse     │        │ • ClamAV Scanner             │   │
│  │ • Frontend (Static) │        │ • YARA Security              │   │
│  └─────────────────────┘        └──────────────────────────────┘   │
│           │                                    │                    │
│           └──────────────┬─────────────────────┘                    │
│                          │                                          │
│           ┌──────────────▼──────────────┐                          │
│           │  🗄️ Neon PostgreSQL          │                          │
│           │  Managed Database ($25/mo)   │                          │
│           │  • Auto backups • Branching  │                          │
│           └─────────────────────────────┘                          │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

**Two-server design** separates concerns for security and scalability:
- **Server 1** handles API requests, caching, and web frontend
- **Server 2** isolates function execution and security scanning

---

## 💰 Infrastructure Costs

| Component | Provider | Specs | Monthly Cost |
|-----------|----------|-------|--------------|
| **Server 1 (App Stack)** | OVHcloud | KS-5: 64GB RAM, 1TB SSD | $20 |
| **Server 2 (Runtime)** | OVHcloud | KS-5: 64GB RAM, 1TB SSD | $20 |
| **Database** | Neon | Managed PostgreSQL | $25 |
| **DNS/CDN** | Cloudflare | Free Plan | $0 |
| **🎯 Total Baseline** | - | - | **$65/month** |

> 💡 **Scaling Note**: Costs scale linearly with additional runtime servers. See [COST_OPTIMIZED_DEPLOYMENT.md](./COST_OPTIMIZED_DEPLOYMENT.md) for cost-saving strategies.

---

## 📚 Quick Links

| Document | Purpose | Time |
|----------|---------|------|
| [📖 PRODUCTION_DEPLOYMENT.md](./PRODUCTION_DEPLOYMENT.md) | Complete step-by-step deployment guide | ~2 hours |
| [⚡ QUICK_START.md](./QUICK_START.md) | Get running in 5 minutes | ~5 min |
| [🔧 PERFORMANCE_TUNING_GUIDE.md](./PERFORMANCE_TUNING_GUIDE.md) | Optimize for high throughput | - |
| [🚨 DISASTER_RECOVERY_RUNBOOK.md](./DISASTER_RECOVERY_RUNBOOK.md) | Recovery procedures & incident response | - |
| [💸 COST_OPTIMIZED_DEPLOYMENT.md](./COST_OPTIMIZED_DEPLOYMENT.md) | Reduce costs without sacrificing reliability | - |

---

## ✅ Prerequisites

Before deploying, ensure you have:

- [ ] **Domain name** registered and ready for configuration
- [ ] **OVHcloud account** with bare metal server access
- [ ] **Cloudflare account** for DNS and SSL management
- [ ] **Neon account** for managed PostgreSQL
- [ ] **Basic knowledge** of Docker, Linux, and SSH

---

## 🆘 Support

- **GitHub Issues**: [github.com/functionfly/functionfly/issues](https://github.com/functionfly/functionfly/issues)
- **Documentation**: [Full docs index](../README.md#documentation)

---

<p align="center">
  <sub>FunctionFly Production Deployment • Last updated: 2026</sub>
</p>
