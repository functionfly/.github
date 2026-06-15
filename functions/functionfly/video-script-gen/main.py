"""Video Script Generator - Generate video scripts with timing and scenes."""
from datetime import datetime
from typing import Any


VIDEO_TYPES = {
    "tutorial": {
        "structure": ["intro", "overview", "steps", "tips", "outro"],
        "pace": "steady and clear",
        "tone_options": ["professional", "friendly", "authoritative"]
    },
    "explainer": {
        "structure": ["hook", "problem", "solution", "benefits", "cta"],
        "pace": "engaging and dynamic",
        "tone_options": ["professional", "funny", "authoritative"]
    },
    "commercial": {
        "structure": ["hook", "value_prop", "features", "social_proof", "cta"],
        "pace": "energetic",
        "tone_options": ["professional", "funny"]
    }
}


def handler(event: dict) -> dict:
    """Generate a video script."""
    try:
        topic = event.get("topic")
        video_type = event.get("video_type")
        duration_seconds = event.get("duration_seconds")
        tone = event.get("tone", "professional")

        if not topic:
            return {"ok": False, "error": "topic is required"}
        if not video_type:
            return {"ok": False, "error": "video_type is required (tutorial/explainer/commercial)"}
        if not duration_seconds:
            return {"ok": False, "error": "duration_seconds is required"}

        if video_type not in ["tutorial", "explainer", "commercial"]:
            return {"ok": False, "error": "video_type must be one of: tutorial, explainer, commercial"}

        if not isinstance(duration_seconds, int) or duration_seconds < 15 or duration_seconds > 3600:
            return {"ok": False, "error": "duration_seconds must be an integer between 15 and 3600"}

        if tone not in ["professional", "friendly", "authoritative", "funny"]:
            return {"ok": False, "error": "tone must be one of: professional, friendly, authoritative, funny"}

        config = VIDEO_TYPES[video_type]
        if tone not in config["tone_options"]:
            tone = config["tone_options"][0]

        duration_minutes = duration_seconds // 60
        intro_duration = min(15, duration_seconds // 10)
        outro_duration = min(10, duration_seconds // 12)
        main_content_duration = duration_seconds - intro_duration - outro_duration

        scenes = []

        if "intro" in config["structure"]:
            scenes.append({
                "scene_number": len(scenes) + 1,
                "title": "Introduction",
                "description": f"Host welcomes viewers and introduces today's topic: {topic}",
                "duration_seconds": intro_duration,
                "dialogue": f"[{'HOST' if video_type != 'commercial' else 'VOICEOVER'}] Welcome to today's video! We're diving into {topic}. By the end, you'll understand the key concepts and how to apply them.",
                "b_roll": f"B-roll of {topic} context, title card animation"
            })

        if "overview" in config["structure"]:
            scenes.append({
                "scene_number": len(scenes) + 1,
                "title": "Overview",
                "description": "Brief overview of what will be covered",
                "duration_seconds": min(20, main_content_duration // 4),
                "dialogue": f"[HOST] Here's what we'll cover today: First, we'll look at the basics. Then we'll dive into practical applications. And finally, I'll share some tips to get started.",
                "b_roll": "Animated bullet points or graphics"
            })

        if "hook" in config["structure"]:
            scenes.append({
                "scene_number": len(scenes) + 1,
                "title": "Hook",
                "description": "Attention-grabbing opening to engage viewers",
                "duration_seconds": intro_duration,
                "dialogue": f"[{'HOST' if tone != 'funny' else 'HOST (with humor)'}] What if I told you everything you know about {topic} is incomplete? Let's fix that.",
                "b_roll": f"Engaging visuals related to {topic}, dramatic graphics"
            })

        if "problem" in config["structure"]:
            scenes.append({
                "scene_number": len(scenes) + 1,
                "title": "The Problem",
                "description": "Identify the problem or knowledge gap",
                "duration_seconds": min(30, main_content_duration // 4),
                "dialogue": f"[HOST] Most people struggle with {topic} because they don't understand the fundamentals. Today, we're going to change that.",
                "b_roll": "Relatable problem scenarios, interview clips (if applicable)"
            })

        if "value_prop" in config["structure"]:
            scenes.append({
                "scene_number": len(scenes) + 1,
                "title": "Value Proposition",
                "description": "Highlight key benefits",
                "duration_seconds": min(20, main_content_duration // 4),
                "dialogue": f"[VOICEOVER] With {topic}, you can achieve better results in less time. It's the solution you've been looking for.",
                "b_roll": "Product/demo visuals, benefits animation"
            })

        if "steps" in config["structure"] or "solution" in config["structure"]:
            steps_title = "Step-by-Step Guide" if "steps" in config["structure"] else "The Solution"
            num_steps = min(5, max(3, main_content_duration // 40))

            for step in range(1, num_steps + 1):
                step_duration = main_content_duration // (num_steps + 2)
                scenes.append({
                    "scene_number": len(scenes) + 1,
                    "title": f"{steps_title}: Step {step}",
                    "description": f"Detailed step {step} of the process",
                    "duration_seconds": step_duration,
                    "dialogue": f"[HOST] Step {step}: [Clear explanation of this part of the process]. Let me show you exactly how it's done.",
                    "b_roll": f"Screen recording or demonstration of step {step}"
                })

        if "features" in config["structure"]:
            scenes.append({
                "scene_number": len(scenes) + 1,
                "title": "Key Features",
                "description": "Highlight main features or benefits",
                "duration_seconds": min(30, main_content_duration // 4),
                "dialogue": "[VOICEOVER] Here are the key features that set this apart: [Feature 1], [Feature 2], and [Feature 3].",
                "b_roll": "Product showcase, feature highlights"
            })

        if "benefits" in config["structure"]:
            scenes.append({
                "scene_number": len(scenes) + 1,
                "title": "Benefits",
                "description": "Summary of key benefits",
                "duration_seconds": min(20, main_content_duration // 4),
                "dialogue": "[HOST] The benefits are clear: Save time, get better results, and simplify your workflow.",
                "b_roll": "Customer testimonials or success metrics"
            })

        if "tips" in config["structure"]:
            scenes.append({
                "scene_number": len(scenes) + 1,
                "title": "Pro Tips",
                "description": "Helpful tips for viewers",
                "duration_seconds": min(25, main_content_duration // 4),
                "dialogue": "[HOST] Before we wrap up, here are my top tips: First, start small. Second, practice consistently. And third, don't be afraid to experiment.",
                "b_roll": "Tips displayed as text overlays or icons"
            })

        if "social_proof" in config["structure"]:
            scenes.append({
                "scene_number": len(scenes) + 1,
                "title": "Social Proof",
                "description": "Customer testimonials or success stories",
                "duration_seconds": min(20, main_content_duration // 5),
                "dialogue": "[VOICEOVER] Don't just take my word for it. Here's what our customers are saying: [Testimonial 1]...",
                "b_roll": "Customer photos, review snippets, logos"
            })

        if "cta" in config["structure"]:
            scenes.append({
                "scene_number": len(scenes) + 1,
                "title": "Call to Action",
                "description": "Final call to action",
                "duration_seconds": outro_duration,
                "dialogue": "[HOST] If you found this helpful, smash that like button and subscribe for more content. And drop a comment below with your thoughts!",
                "b_roll": "Subscribe button animation, channel trailer preview"
            })

        if "outro" in config["structure"]:
            scenes.append({
                "scene_number": len(scenes) + 1,
                "title": "Outro",
                "description": "Closing remarks and sign-off",
                "duration_seconds": outro_duration,
                "dialogue": f"[HOST] Thanks for watching this video on {topic}. Remember to like, subscribe, and hit the notification bell. I'll see you in the next one!",
                "b_roll": "End screen, suggested videos, channel logo"
            })

        script_lines = []
        for scene in scenes:
            script_lines.append(f"\n{'='*50}")
            script_lines.append(f"SCENE {scene['scene_number']}: {scene['title']}")
            script_lines.append(f"Duration: {scene['duration_seconds']}s")
            script_lines.append(f"{'='*50}")
            script_lines.append(f"VISUAL: {scene['description']}")
            script_lines.append(f"\nDIALOGUE:\n{scene['dialogue']}")
            script_lines.append(f"\nB-ROLL: {scene['b_roll']}")

        script = "\n".join(script_lines)

        b_roll_suggestions = [
            f"Screen recordings of {topic} in action",
            f"Relevant stock footage for context",
            f"Graphics and animations explaining key points",
            f"Interview clips with subject matter experts",
            f"Before/after comparisons",
            f"Data visualizations and charts"
        ]

        return {
            "ok": True,
            "topic": topic,
            "video_type": video_type,
            "duration_seconds": duration_seconds,
            "tone": tone,
            "script": script.strip(),
            "scenes": scenes,
            "b_roll_suggestions": b_roll_suggestions,
            "scene_count": len(scenes),
            "generated_at": datetime.now().isoformat()
        }

    except Exception as e:
        return {"ok": False, "error": f"Failed to generate video script: {str(e)}"}


if __name__ == "__main__":
    result = handler({
        "topic": "How to Build a Mobile App",
        "video_type": "tutorial",
        "duration_seconds": 300,
        "tone": "professional"
    })
    print(result)
