def handler(event):
    """Generate Terraform HCL configuration templates."""
    try:
        provider = event.get("provider")
        resource_type = event.get("resource_type")
        resource_name = event.get("resource_name")
        if not all([provider, resource_type, resource_name]):
            return {"ok": False, "error": "provider, resource_type, and resource_name are required"}

        config = event.get("config", {})
        variables = event.get("variables", {})

        provider_sources = {
            "aws": "hashicorp/aws",
            "gcp": "hashicorp/google",
            "azure": "hashicorp/azurerm",
            "kubernetes": "hashicorp/kubernetes",
            "helm": "hashicorp/helm",
        }

        lines = []

        # Terraform block
        source = provider_sources.get(provider, f"hashicorp/{provider}")
        lines.append("terraform {")
        lines.append("  required_providers {")
        lines.append(f"    {provider} = {{")
        lines.append(f'      source = "{source}"')
        lines.append("    }")
        lines.append("  }")
        lines.append("}")
        lines.append("")

        # Variables
        for var_name, var_config in variables.items():
            lines.append(f'variable "{var_name}" {{')
            if isinstance(var_config, dict):
                for k, v in var_config.items():
                    if isinstance(v, str):
                        lines.append(f'  {k} = "{v}"')
                    else:
                        lines.append(f"  {k} = {v}")
            lines.append("}")
            lines.append("")

        # Resource block
        lines.append(f'resource "{resource_type}" "{resource_name}" {{')
        for k, v in config.items():
            if isinstance(v, str):
                lines.append(f'  {k} = "{v}"')
            elif isinstance(v, bool):
                lines.append(f"  {k} = {str(v).lower()}")
            elif isinstance(v, (int, float)):
                lines.append(f"  {k} = {v}")
            elif isinstance(v, dict):
                lines.append(f"  {k} {{")
                for sk, sv in v.items():
                    if isinstance(sv, str):
                        lines.append(f'    {sk} = "{sv}"')
                    else:
                        lines.append(f"    {sk} = {sv}")
                lines.append("  }")
            else:
                lines.append(f'  {k} = "{v}"')
        lines.append("}")

        hcl = "\n".join(lines) + "\n"
        return {"ok": True, "result": "Terraform template generated", "hcl": hcl}
    except Exception as e:
        return {"ok": False, "error": str(e)}
