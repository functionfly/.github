"""
FOM Synthetic Data Generator

Generates synthetic FOM training data using templates and variations
to create Goal→Workflow→Outcome records for training data augmentation.
"""

import json
import random
from dataclasses import dataclass, field
from typing import Any


@dataclass
class GoalTemplate:
    type: str
    category: str
    template: str
    variables: list[str]
    typical_workflow: list[str]
    variations: int


GOAL_TEMPLATES = [
    GoalTemplate(
        type="deploy",
        category="web",
        template="Launch a {business} {business_type} website",
        variables=["business", "business_type"],
        typical_workflow=["create_project", "generate_code", "configure_seo", "deploy"],
        variations=50,
    ),
    GoalTemplate(
        type="deploy",
        category="web",
        template="Deploy a {tech_stack} application to {cloud_provider}",
        variables=["tech_stack", "cloud_provider"],
        typical_workflow=["setup_project", "configure_infra", "deploy"],
        variations=40,
    ),
    GoalTemplate(
        type="build",
        category="api",
        template="Build a {api_type} API for {use_case}",
        variables=["api_type", "use_case"],
        typical_workflow=["design_schema", "implement_endpoints", "add_auth", "write_tests", "deploy"],
        variations=40,
    ),
    GoalTemplate(
        type="build",
        category="api",
        template="Create a REST API for {resource_name} management",
        variables=["resource_name"],
        typical_workflow=["define_models", "create_routes", "implement_crud", "add_validation", "deploy"],
        variations=30,
    ),
    GoalTemplate(
        type="configure",
        category="integration",
        template="Set up {integration_name} integration with {platform}",
        variables=["integration_name", "platform"],
        typical_workflow=["install_integration", "configure_credentials", "test_connection", "enable_webhooks"],
        variations=35,
    ),
    GoalTemplate(
        type="configure",
        category="infra",
        template="Configure {service_name} monitoring and alerting",
        variables=["service_name"],
        typical_workflow=["setup_monitoring", "configure_alerts", "add_dashboards", "test_notifications"],
        variations=25,
    ),
]

VARIABLE_VALUES = {
    "business": [
        "plumbing", "restaurant", "consulting", "ecommerce", "healthcare",
        "real estate", "automotive", "education", "fitness", "beauty",
    ],
    "business_type": [
        "local service", "SaaS", "marketplace", "content", "membership",
    ],
    "tech_stack": [
        "React", "Vue", "Angular", "Next.js", "Nuxt", "Node.js", "Django", "FastAPI",
    ],
    "cloud_provider": [
        "Vercel", "Fly.io", "Railway", "AWS", "GCP", "Cloudflare",
    ],
    "api_type": [
        "REST", "GraphQL", "gRPC", "WebSocket", "tRPC",
    ],
    "use_case": [
        "user authentication", "payment processing", "data analytics", "notification system",
    ],
    "resource_name": [
        "users", "products", "orders", "invoices", "subscriptions",
    ],
    "integration_name": [
        "Stripe", "SendGrid", "Twilio", "Slack", "Discord", "Analytics",
    ],
    "platform": [
        "my application", "the dashboard", "our platform", "the API",
    ],
    "service_name": [
        "API gateway", "database", "cache layer", "queue system", "storage service",
    ],
}

OPTIONAL_STEPS = [
    "setup_monitoring",
    "add_logging",
    "configure_caching",
    "enable_autoscaling",
    "add_backup",
    "setup_ci_cd",
    "add_rate_limiting",
]


class SyntheticGenerator:
    def __init__(self, physics_engine=None):
        self.physics = physics_engine
        self.templates = GOAL_TEMPLATES

    def generate_batch(self, count: int) -> list[dict]:
        records = []
        for _ in range(count):
            template = random.choice(self.templates)
            record = self._generate_from_template(template)
            records.append(record)
        return records

    def _generate_from_template(self, template: GoalTemplate) -> dict:
        variables = {
            v: random.choice(VARIABLE_VALUES.get(v, ["default"]))
            for v in template.variables
        }
        goal_text = template.template.format(**variables)

        workflow = self._mutate_workflow(template.typical_workflow)

        prediction = None
        if self.physics:
            prediction = self.physics.predict_workflow_outcome(workflow)

        is_success = True
        if prediction:
            is_success = random.random() < prediction.get("success_probability", 0.9)

        outcome_score = self._calculate_outcome_score(prediction, is_success)

        return {
            "goal_text": goal_text,
            "goal_type": template.type,
            "goal_category": template.category,
            "workflow_json": workflow,
            "outcome_success": is_success,
            "outcome_score": outcome_score,
            "total_cost": prediction.get("estimated_cost", 0.05) if prediction else 0.05,
            "total_time_ms": prediction.get("estimated_time_ms", 5000) if prediction else 5000,
            "is_synthetic": True,
            "generation_method": "synthetic_v1",
            "data_source": "synthetic",
            "confidence_level": "medium",
            "split": self._assign_split(),
        }

    def _mutate_workflow(self, workflow: list[str]) -> list[str]:
        if not workflow:
            return workflow

        mutation = random.randint(0, 3)
        result = workflow.copy()

        if mutation == 1 and len(result) > 2:
            result = result[:-1]
        elif mutation == 2 and len(OPTIONAL_STEPS) > 0:
            result = result + [random.choice(OPTIONAL_STEPS)]
        elif mutation == 3 and len(result) > 2:
            a, b = random.sample(range(len(result)), 2)
            result[a], result[b] = result[b], result[a]

        return result

    def _calculate_outcome_score(self, prediction: dict = None, is_success: bool = True) -> int:
        if not is_success:
            return random.randint(20, 40)

        base_score = random.randint(70, 95)

        if prediction:
            cost = prediction.get("estimated_cost", 0)
            time_ms = prediction.get("estimated_time_ms", 5000)

            if cost < 0.10:
                base_score += 5
            elif cost > 0.50:
                base_score -= 10

            if time_ms < 5000:
                base_score += 5
            elif time_ms > 15000:
                base_score -= 10

        return min(100, max(0, base_score))

    def _assign_split(self) -> str:
        r = random.random()
        if r < 0.8:
            return "train"
        elif r < 0.9:
            return "val"
        else:
            return "test"

    def generate_for_goal_type(self, goal_type: str, count: int) -> list[dict]:
        matching_templates = [t for t in self.templates if t.type == goal_type]
        if not matching_templates:
            matching_templates = self.templates

        records = []
        for _ in range(count):
            template = random.choice(matching_templates)
            record = self._generate_from_template(template)
            records.append(record)
        return records


class WorkflowSimulator:
    def __init__(self, physics_engine):
        self.physics = physics_engine

    def simulate_goal(
        self,
        goal_text: str,
        goal_type: str,
        num_workflows: int = 20,
        iterations: int = 100,
    ) -> list[dict]:
        candidates = self._generate_candidates(goal_type, num_workflows)

        results = []
        for workflow in candidates:
            sim_result = self.physics.simulate_workflow(workflow, iterations=iterations)

            results.append({
                "workflow": workflow,
                "sim_success_rate": sim_result.get("success_rate", 0),
                "sim_avg_cost": sim_result.get("avg_cost", 0),
                "sim_avg_time": sim_result.get("avg_time_ms", 0),
                "sim_p95_cost": sim_result.get("p95_cost", 0),
                "composite_score": self._calculate_composite_score(sim_result),
            })

        ranked = self._rank_workflows(results)
        return ranked

    def _generate_candidates(self, goal_type: str, count: int) -> list[list[str]]:
        base_workflows = {
            "deploy": [
                ["create_project", "generate_code", "configure_seo", "deploy"],
                ["setup_project", "configure_infra", "deploy"],
                ["create_project", "generate_code", "test", "deploy"],
            ],
            "build": [
                ["design_schema", "implement_endpoints", "add_auth", "write_tests", "deploy"],
                ["define_models", "create_routes", "implement_crud", "add_validation", "deploy"],
                ["design_schema", "implement_endpoints", "deploy"],
            ],
            "configure": [
                ["install_integration", "configure_credentials", "test_connection", "enable_webhooks"],
                ["setup_monitoring", "configure_alerts", "add_dashboards", "test_notifications"],
            ],
        }

        workflows = base_workflows.get(goal_type, base_workflows["deploy"]).copy()

        while len(workflows) < count:
            base = random.choice(base_workflows.get(goal_type, base_workflows["deploy"]))
            mutated = self._mutate_workflow(base)
            if mutated not in workflows:
                workflows.append(mutated)

        return workflows[:count]

    def _mutate_workflow(self, workflow: list[str]) -> list[str]:
        if not workflow:
            return workflow

        mutation = random.randint(0, 3)
        result = workflow.copy()

        if mutation == 1 and len(result) > 2:
            result = result[:-1]
        elif mutation == 2 and len(OPTIONAL_STEPS) > 0:
            result = result + [random.choice(OPTIONAL_STEPS)]
        elif mutation == 3 and len(result) > 2:
            a, b = random.sample(range(len(result)), 2)
            result[a], result[b] = result[b], result[a]

        return result

    def _calculate_composite_score(self, sim_result: dict) -> float:
        reliability = sim_result.get("sim_success_rate", 0)
        efficiency = 1.0 - min(1.0, sim_result.get("sim_avg_cost", 0) / 1.0)
        speed = 1.0 - min(1.0, sim_result.get("sim_avg_time", 0) / 60000.0)

        return reliability * 0.4 + efficiency * 0.3 + speed * 0.3

    def _rank_workflows(self, results: list[dict]) -> list[dict]:
        return sorted(results, key=lambda x: x["composite_score"], reverse=True)