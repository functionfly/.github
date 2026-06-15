"""Competitor Analyzer - Analyze competitors in a given industry."""
import random


COMPETITOR_DATA = {
    "technology": [
        {"name": "TechCorp Solutions", "strength": "Strong brand recognition and R&D capabilities", "weakness": "High prices, complex products", "opportunity": "Enterprise market expansion"},
        {"name": "InnovateTech", "strength": "Agile development, modern tech stack", "weakness": "Limited market reach", "opportunity": "SMB market penetration"},
        {"name": "GlobalTech Inc", "strength": "Global presence, diverse portfolio", "weakness": "Slow decision making", "opportunity": "Emerging markets entry"},
        {"name": "StartupHub", "strength": "Innovation-driven culture, talented team", "weakness": "Limited resources", "opportunity": "Partnership opportunities"},
        {"name": "MegaSystems", "strength": "Economies of scale, established distribution", "weakness": "Legacy system包袱", "opportunity": "Digital transformation services"}
    ],
    "business": [
        {"name": "Enterprise Plus", "strength": "Comprehensive service offerings", "weakness": "Generic approach", "opportunity": "Niche market specialization"},
        {"name": "BusinessPro", "strength": "Strong client relationships", "weakness": "Limited scalability", "opportunity": "Technology-enabled growth"},
        {"name": "Strategic Solutions", "strength": "Expert consultants, deep expertise", "weakness": "High overhead costs", "opportunity": "Training and education services"},
        {"name": "Growth Partners", "strength": "Results-oriented approach", "weakness": "Over-reliance on key clients", "opportunity": "New service lines"},
        {"name": "Success Corp", "strength": " Proven track record", "weakness": "Traditional methodologies", "opportunity": "Innovation consulting"}
    ],
    "marketing": [
        {"name": "Digital Dynamics", "strength": "Full-service digital marketing", "weakness": "High client turnover", "opportunity": "AI-powered personalization"},
        {"name": "Creative Agency Co", "strength": "Award-winning creative team", "weakness": "Premium pricing", "opportunity": "Mid-market expansion"},
        {"name": "Social Buzz", "strength": "Social media expertise", "weakness": "Limited traditional skills", "opportunity": "Integrated campaigns"},
        {"name": "Brand Builders", "strength": "Strong brand strategy", "weakness": "Slow turnaround", "opportunity": "Rapid deployment services"},
        {"name": "Media Masters", "strength": "Media buying power", "weakness": "Dependence on ad spend", "opportunity": "Content marketing pivot"}
    ],
    "healthcare": [
        {"name": "MediCare Systems", "strength": "Established hospital network", "weakness": "Bureaucratic structure", "opportunity": "Telemedicine expansion"},
        {"name": "HealthTech Pro", "strength": "Innovative technology solutions", "weakness": "Limited physical presence", "opportunity": "Home healthcare"},
        {"name": "Wellness Corp", "strength": "Preventive care focus", "weakness": "Insurance limitations", "opportunity": "Corporate wellness programs"},
        {"name": "Care Network", "strength": "Comprehensive care coordination", "weakness": "Fragmented services", "opportunity": "Integrated care platforms"},
        {"name": "Life Sciences Inc", "strength": "Research capabilities", "weakness": "Regulatory challenges", "opportunity": "Personalized medicine"}
    ],
    "finance": [
        {"name": "FinServe Group", "strength": "Full-service financial services", "weakness": "Complex organizational structure", "opportunity": "Digital banking expansion"},
        {"name": "Wealth Advisors Pro", "strength": "High-net-worth client base", "weakness": "Exclusive focus on elites", "opportunity": "Mass affluent market"},
        {"name": "Investment Partners", "strength": "Strong investment performance", "weakness": "Market volatility exposure", "opportunity": "Alternative investments"},
        {"name": "Crypto Finance", "strength": "Cryptocurrency expertise", "weakness": "Regulatory uncertainty", "opportunity": "Blockchain services"},
        {"name": "Community Bank Plus", "strength": "Local relationships, trust", "weakness": "Limited product range", "opportunity": "Digital transformation"}
    ]
}

DEFAULT_COMPETITORS = [
    {"name": "Industry Leader Inc", "strength": "Market dominance, brand recognition", "weakness": "Complacency, slow innovation", "opportunity": "Emerging technologies"},
    {"name": "Challenger Co", "strength": "Agile, customer-focused", "weakness": "Limited resources", "opportunity": "Market disruption"},
    {"name": "Niche Player", "strength": "Specialized expertise", "weakness": "Limited market scope", "opportunity": "Market expansion"},
    {"name": "Regional Services", "strength": "Local presence, relationships", "weakness": "Geographic limitations", "opportunity": "National scaling"},
    {"name": "Innovation Labs", "strength": "Cutting-edge solutions", "weakness": "Unproven business model", "opportunity": "Strategic partnerships"}
]


def handler(event):
    try:
        business_name = event.get("business_name", "")
        industry = event.get("industry", "general")
        location = event.get("location", "")

        if not business_name:
            return {"ok": False, "error": "business_name is required"}

        industry_lower = industry.lower()
        competitors = COMPETITOR_DATA.get(industry_lower, DEFAULT_COMPETITORS)

        competitors = competitors[:5]

        market_positions = [
            "You are positioned as a challenger in the market with room to grow.",
            "Your main opportunity lies in differentiating from established players.",
            "Consider focusing on underserved market segments.",
            "There is room for disruption through innovation and customer focus.",
            "Your competitive advantage should focus on unique value proposition."
        ]
        market_position = random.choice(market_positions)

        swot_strengths = ["Innovation", "Customer focus", "Agility", "Expertise", "Quality"]
        swot_weaknesses = ["Brand awareness", "Resources", "Scale", "Market presence", "Track record"]
        swot_opportunities = ["Market expansion", "Partnerships", "Technology", "New customers", "Innovation"]
        swot_threats = ["Competition", "Market changes", "Economic conditions", "Regulatory", "Technology disruption"]

        swot_analysis = {
            "strengths": random.sample(swot_strengths, 3),
            "weaknesses": random.sample(swot_weaknesses, 3),
            "opportunities": random.sample(swot_opportunities, 3),
            "threats": random.sample(swot_threats, 3)
        }

        return {
            "ok": True,
            "competitors": competitors,
            "market_position": market_position,
            "swot_analysis": swot_analysis,
            "business_name": business_name,
            "industry": industry,
            "location": location
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
