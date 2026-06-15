"""AI Content Generator - Generate various types of content."""
import random


CONTENT_TEMPLATES = {
    "blog": {
        "formal": {
            "short": "This article examines {topic} and its significance in contemporary practice. Key findings indicate that {topic} represents a critical area for professionals.",
            "medium": "This comprehensive analysis explores {topic} in depth. The examination reveals several important aspects that {audience} should understand. First, we establish the foundational principles. Then, we explore practical applications. Finally, we discuss future implications.",
            "long": "In today's rapidly evolving landscape, understanding {topic} has become essential for {audience}. This detailed exploration covers multiple dimensions of the subject. We begin by establishing context and defining key terms. Next, we examine current trends and their implications. Practical considerations follow, with real-world examples. We then address common challenges and solutions. Concluding thoughts offer strategic recommendations for moving forward."
        },
        "casual": {
            "short": f"Hey {audience}! Let's talk about {topic}. It's actually pretty interesting once you get it.",
            "medium": f"So you're interested in {topic}? Great choice! Here's the deal - this is something {audience} are getting really excited about. Let me break it down for you in simple terms. First, the basics. Then, the good stuff. And finally, how to actually use it.",
            "long": f"Okay, let's dive deep into {topic}! As {audience}, you're probably wondering what all the fuss is about. Well, buckle up because this is a fun one. We start from square one - what even is {topic}? Then we get into the juicy details. Real talk: there are some things they won't tell you, but I'm sharing everything. By the end, you'll be ready to jump in and start seeing results."
        },
        "technical": {
            "short": f"Technical analysis of {topic}: An examination of core components and their interactions within the system architecture.",
            "medium": f"This technical document provides an in-depth analysis of {topic}. Coverage includes architectural overview, component analysis, integration patterns, and performance considerations. Appropriate for {audience} with intermediate technical background.",
            "long": f"Technical specification document for {topic}. This comprehensive technical analysis covers system architecture, component design, implementation patterns, and operational considerations. Designed for {audience} with advanced technical expertise. Includes detailed specifications, reference implementations, and best practice recommendations."
        }
    },
    "social": {
        "formal": {
            "short": f"Important update regarding {topic}. Key stakeholders should review this information carefully.",
            "medium": f"We are pleased to share important information about {topic}. This announcement contains relevant details for {audience}. Please review and share with your network.",
            "long": f"This official communication addresses {topic} and its implications for {audience}. The following information is provided for your review and consideration. We encourage engagement and dialogue on this important topic."
        },
        "casual": {
            "short": f"Hot take on {topic}! 🔥 What do you think?",
            "medium": f"Okay guys, big news about {topic}! 🌟 Here's everything {audience} need to know. Drop your thoughts below! 👇",
            "long": f"So something exciting is happening with {topic}! 🎉 I had to share this with all of you. {audience}, you're going to want to hear this. Let me break it down..."
        },
        "technical": {
            "short": f"Technical update: {topic}. System notifications for {audience}.",
            "medium": f"Technical advisory on {topic}. This update contains technical information relevant to {audience}. Implementation details included.",
            "long": f"Technical release notes for {topic}. This comprehensive update covers technical changes, system requirements, and implementation guidance for {audience}."
        }
    },
    "email": {
        "formal": {
            "short": f"Regarding {topic}: We wanted to inform {audience} about important developments.",
            "medium": f"Dear {audience}, We are writing to share information about {topic}. This communication contains important details that may affect your operations. Please review the attached materials.",
            "long": f"Dear {audience}, This email serves to provide comprehensive information regarding {topic}. We understand the importance of keeping you informed and would like to outline the key points for your consideration. Should you have any questions, please do not hesitate to reach out."
        },
        "casual": {
            "short": f"Hey {audience}! Quick update on {topic} - check this out!",
            "medium": f"Hi {audience}! 👋 Just wanted to share what's happening with {topic}. Hope this helps! Let me know if you have questions.",
            "long": f"Hey {audience}! Hope you're doing well! We've got some exciting news about {topic} that I think you'll love. Here's the full scoop..."
        },
        "technical": {
            "short": f"Technical notice: {topic}. For {audience} technical teams.",
            "medium": f"Technical communication regarding {topic}. This notice contains technical specifications and implementation guidance for {audience}.",
            "long": f"Technical specification notification for {audience}. This document provides complete technical details regarding {topic} including system requirements, integration protocols, and deployment instructions."
        }
    }
}


def handler(event):
    try:
        topic = event.get("topic", "")
        content_type = event.get("content_type", "blog")
        tone = event.get("tone", "formal")
        length = event.get("length", "medium")

        if not topic:
            return {"ok": False, "error": "topic is required"}
        if content_type not in ["blog", "social", "email"]:
            return {"ok": False, "error": "content_type must be blog, social, or email"}
        if tone not in ["formal", "casual", "technical"]:
            return {"ok": False, "error": "tone must be formal, casual, or technical"}
        if length not in ["short", "medium", "long"]:
            return {"ok": False, "error": "length must be short, medium, or long"}

        audience = event.get("audience", "readers")
        template = CONTENT_TEMPLATES[content_type][tone][length]

        content = template.format(topic=topic, audience=audience)

        title = f"{topic.title()}: " + random.choice([
            "A Comprehensive Overview",
            "Key Insights and Analysis",
            "What You Need to Know",
            "Complete Guide",
            "Essential Information"
        ])

        meta_tags = []
        if content_type == "blog":
            meta_tags = [
                f"{topic} guide",
                f"{topic} tips",
                f"{topic} for {audience}",
                "how to understand " + topic.lower(),
                topic.lower() + " best practices"
            ]

        return {
            "ok": True,
            "title": title,
            "content": content,
            "content_type": content_type,
            "tone": tone,
            "length": length,
            "meta_tags": meta_tags if content_type == "blog" else []
        }
    except Exception as e:
        return {"ok": False, "error": str(e)}
