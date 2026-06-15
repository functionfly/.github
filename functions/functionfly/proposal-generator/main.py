"""Business Proposal Generator - Generate business proposals."""
from datetime import datetime
from typing import Any


def handler(event: dict) -> dict:
    """Generate a business proposal."""
    try:
        client_name = event.get("client_name")
        proposed_solution = event.get("proposed_solution")
        timeline_weeks = event.get("timeline_weeks")
        budget = event.get("budget")
        team_members = event.get("team_members", [])

        if not client_name:
            return {"ok": False, "error": "client_name is required"}
        if not proposed_solution:
            return {"ok": False, "error": "proposed_solution is required"}
        if not timeline_weeks:
            return {"ok": False, "error": "timeline_weeks is required"}
        if not budget:
            return {"ok": False, "error": "budget is required"}

        try:
            timeline_weeks = int(timeline_weeks)
            if timeline_weeks < 1:
                return {"ok": False, "error": "timeline_weeks must be a positive integer"}
        except (ValueError, TypeError):
            return {"ok": False, "error": "timeline_weeks must be a valid integer"}

        today = datetime.now()
        proposal_id = f"PROP-{today.strftime('%Y%m%d')}-{hash(client_name) % 10000:04d}"
        valid_until = (today.replace(day=1) + datetime.timedelta(days=32)).replace(day=1)

        sections = [
            "1. Executive Summary",
            "2. Understanding Your Needs",
            "3. Proposed Solution",
            "4. Approach and Methodology",
            "5. Timeline & Milestones",
            "6. Team Introduction",
            "7. Pricing Breakdown",
            "8. Terms and Conditions",
            "9. Next Steps"
        ]

        phases = []
        remaining_weeks = timeline_weeks
        phase_num = 1

        while remaining_weeks > 0:
            phase_weeks = min(4, remaining_weeks)
            phases.append({
                "phase": phase_num,
                "title": ["Discovery & Planning", "Development & Implementation", "Testing & Refinement", "Deployment & Launch"][min(phase_num - 1, 3)],
                "weeks": phase_weeks,
                "deliverables": [
                    f"Phase {phase_num} deliverable {i+1}" for i in range(min(3, phase_weeks))
                ]
            })
            remaining_weeks -= phase_weeks
            phase_num += 1
            if phase_num > 4:
                break

        budget_num = float(budget.replace("$", "").replace(",", "")) if isinstance(budget, str) else float(budget)
        budget_formatted = f"${budget_num:,.2f}"

        pricing_breakdown = {
            "subtotal": budget_formatted,
            "contingency": f"${budget_num * 0.1:,.2f}",
            "total": f"${budget_num * 1.1:,.2f}"
        }

        proposal_text = f"""
{'='*70}
                    BUSINESS PROPOSAL
{'='*70}

Proposal ID: {proposal_id}
Date: {today.strftime('%B %d, %Y')}
Valid Until: {valid_until.strftime('%B %d, %Y')}

Prepared For: {client_name}
Prepared By: [Your Company Name]

{'='*70}
{sections[0]}
{'='*70}

This proposal outlines our approach to delivering {proposed_solution} for {client_name}.
Our team is excited about the opportunity to partner with you on this project.

Key Benefits:
   • Expert implementation of the proposed solution
   • Clear communication and regular progress updates
   • On-time delivery within the agreed timeline
   • Post-project support and documentation

Estimated Investment: {budget_formatted}
Project Duration: {timeline_weeks} weeks

{'='*70}
{sections[1]}
{'='*70}

Based on our preliminary discussions, we understand that {client_name} requires:
   • [Understanding point 1]
   • [Understanding point 2]
   • [Understanding point 3]

Our proposed solution directly addresses these requirements while providing
additional value through our proven methodology and experienced team.

{'='*70}
{sections[2]}
{'='*70}

{proposed_solution}

Our solution includes:
   • Comprehensive planning and requirements gathering
   • Professional implementation by certified experts
   • Rigorous testing and quality assurance
   • Complete documentation and knowledge transfer
   • Post-launch support period

{'='*70}
{sections[3]}
{'='*70}

Our methodology follows industry best practices:

Phase 1: Discovery & Planning
   • Requirements analysis and documentation
   • Technical architecture design
   • Project timeline finalization

Phase 2: Development & Implementation
   • Core development activities
   • Regular client check-ins and demos
   • Agile iterations based on feedback

Phase 3: Testing & Refinement
   • Comprehensive testing
   • Performance optimization
   • User acceptance testing

Phase 4: Deployment & Launch
   • Production deployment
   • Team training
   • Handover and documentation

{'='*70}
{sections[4]}
{'='*70}

Project Timeline: {timeline_weeks} weeks

"""

        for phase in phases:
            proposal_text += f"Phase {phase['phase']}: {phase['title']} ({phase['weeks']} weeks)\n"
            for deliverable in phase['deliverables']:
                proposal_text += f"   • {deliverable}\n"
            proposal_text += "\n"

        team_section = ""
        if team_members:
            team_section = f"""
{'='*70}
{sections[5]}
{'='*70}

Our dedicated team for this project includes:

"""
            for member in team_members:
                if isinstance(member, dict):
                    team_section += f"• {member.get('name', 'Team Member')} - {member.get('role', 'Role')}\n"
                else:
                    team_section += f"• {member}\n"

            proposal_text += team_section

        proposal_text += f"""
{'='*70}
{sections[6]}
{'='*70}

Pricing Breakdown:

   Subtotal:              {pricing_breakdown['subtotal']}
   Contingency (10%):      {pricing_breakdown['contingency']}
   ─────────────────────────────
   Total Investment:       {pricing_breakdown['total']}

Note: Final pricing may vary based on scope changes mutually agreed upon in writing.

{'='*70}
{sections[7]}
{'='*70}

Terms and Conditions:
   • 50% payment upon project initiation
   • 50% payment upon project completion
   • All prices in USD unless otherwise specified
   • Valid for 30 days from the proposal date
   • Either party may terminate with 14 days written notice

{'='*70}
{sections[8]}
{'='*70}

We look forward to working with {client_name} on this exciting project!

Next Steps:
   1. Review this proposal
   2. Schedule a call to discuss any questions
   3. Provide approval to proceed
   4. We'll kick off the project within 5 business days

Contact Information:
[Your Company Name]
Email: [your.email@company.com]
Phone: [your phone number]

{'='*70}
"""

        next_steps = [
            "Review proposal details",
            "Schedule follow-up call to address questions",
            "Provide written approval to proceed",
            "Facilitate kickoff meeting with stakeholders",
            "Begin discovery phase"
        ]

        return {
            "ok": True,
            "proposal_id": proposal_id,
            "client_name": client_name,
            "proposal_text": proposal_text.strip(),
            "sections": sections,
            "pricing_breakdown": pricing_breakdown,
            "next_steps": next_steps,
            "timeline_weeks": timeline_weeks,
            "valid_until": valid_until.strftime('%Y-%m-%d'),
            "generated_at": datetime.now().isoformat()
        }

    except Exception as e:
        return {"ok": False, "error": f"Failed to generate proposal: {str(e)}"}


if __name__ == "__main__":
    result = handler({
        "client_name": "Acme Corp",
        "proposed_solution": "Implementation of a comprehensive cloud migration strategy to modernize your infrastructure and improve scalability.",
        "timeline_weeks": 12,
        "budget": "$45000",
        "team_members": [
            {"name": "John Smith", "role": "Project Manager"},
            {"name": "Jane Doe", "role": "Cloud Architect"},
            {"name": "Mike Johnson", "role": "Senior Developer"}
        ]
    })
    print(result)
