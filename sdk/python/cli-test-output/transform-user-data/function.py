@flypy.function(
    name="transform-user-data",
    description="Transform user data for analytics",
    deterministic=True,
    enable_performance_monitoring=True
)
def transform_user_data(user_data: Dict[str, Any]) -> Dict[str, Any]:
    """Transform raw user data into analytics-ready format."""
    # Extract and normalize user information
    user = {
        "user_id": user_data["id"],
        "full_name": f"{user_data['first_name']} {user_data['last_name']}",
        "email_domain": user_data["email"].split("@")[1],
        "age_group": "18-24" if user_data["age"] < 25 else "25+",
        "account_status": "active" if user_data["is_active"] else "inactive"
    }

    # Process activity data
    activities = user_data.get("activities", [])
    activity_summary = {
        "total_activities": len(activities),
        "last_activity": max(act["timestamp"] for act in activities) if activities else None,
        "activity_types": list(set(act["type"] for act in activities))
    }

    # Calculate engagement score (simplified)
    engagement_score = min(100, len(activities) * 10 + (user_data.get("login_count", 0) * 5))

    return {
        "user": user,
        "activity_summary": activity_summary,
        "engagement_score": engagement_score,
        "processed_at": "2024-01-01T00:00:00Z"  # Would be dynamic in real implementation
    }
