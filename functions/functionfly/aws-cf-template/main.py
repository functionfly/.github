import json


def handler(event):
    """Generate an AWS CloudFormation template."""
    try:
        resources = event.get("resources")
        if not resources:
            return {"ok": False, "error": "resources is required"}

        template = {
            "AWSTemplateFormatVersion": "2010-09-09",
            "Resources": resources,
        }

        if event.get("description"):
            template["Description"] = event["description"]
        if event.get("parameters"):
            template["Parameters"] = event["parameters"]
        if event.get("outputs"):
            template["Outputs"] = event["outputs"]

        template_str = json.dumps(template, indent=2)
        return {"ok": True, "result": "CloudFormation template generated", "template": template_str}
    except Exception as e:
        return {"ok": False, "error": str(e)}
