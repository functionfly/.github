@flypy.input_schema({
    "type": "object",
    "properties": {
        "name": {"type": "string"},
        "age": {"type": "integer", "minimum": 0},
        "email": {"type": "string", "format": "email"}
    },
    "required": ["name", "age"]
})
@flypy.output_schema({
    "type": "object",
    "properties": {
        "user_id": {"type": "string"},
        "profile_complete": {"type": "boolean"},
        "age_group": {"type": "string"},
        "validation_status": {"type": "string"}
    }
})
@flypy.function(
    name="validate-user-profile",
    description="Validate and process user profile data",
    deterministic=True
)
def validate_user_profile(user_data: Dict[str, Any]) -> Dict[str, Any]:
    """Validate user profile and return processed information."""

    # Generate a simple user ID
    user_id = f"user_{hash(user_data['name'] + str(user_data['age'])) % 10000}"

    # Determine if profile is complete
    required_fields = ["name", "age", "email"]
    profile_complete = all(field in user_data for field in required_fields)

    # Age group classification
    age = user_data["age"]
    if age < 18:
        age_group = "minor"
    elif age < 25:
        age_group = "young_adult"
    elif age < 65:
        age_group = "adult"
    else:
        age_group = "senior"

    # Validation status
    validation_status = "valid" if profile_complete else "incomplete"

    return {
        "user_id": user_id,
        "profile_complete": profile_complete,
        "age_group": age_group,
        "validation_status": validation_status
    }
