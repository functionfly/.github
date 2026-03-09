from urllib.parse import urlparse, urlunparse
import posixpath


def handler(event):
    path = event.get("path")
    base_url = event.get("base_url")

    if not path:
        return {"ok": False, "error": "path is required"}

    if not isinstance(path, str):
        return {"ok": False, "error": "path must be a string"}

    try:
        # Handle URL paths vs standalone paths
        if base_url:
            # Extract path from URL
            parsed_url = urlparse(base_url)
            base_path = parsed_url.path
            # Combine base path with relative path
            if path.startswith('/'):
                normalized_path = posixpath.normpath(path)
            else:
                combined_path = posixpath.join(base_path, path)
                normalized_path = posixpath.normpath(combined_path)
        else:
            # Normalize standalone path
            normalized_path = posixpath.normpath(path)

        # Ensure path starts with / for web paths
        if not normalized_path.startswith('/') and not path.startswith('./') and not path.startswith('../'):
            normalized_path = '/' + normalized_path

        # Remove redundant slashes but keep leading slash
        parts = normalized_path.split('/')
        clean_parts = []
        for part in parts:
            if part not in ('', '.'):
                clean_parts.append(part)

        # Handle .. navigation
        final_parts = []
        for part in clean_parts:
            if part == '..':
                if final_parts:
                    final_parts.pop()
            else:
                final_parts.append(part)

        # Reconstruct path
        if normalized_path.startswith('/'):
            result_path = '/' + '/'.join(final_parts)
        else:
            result_path = '/'.join(final_parts)

        # Handle edge cases
        if not result_path:
            result_path = '/'

        result = {
            "original_path": path,
            "normalized_path": result_path,
            "was_modified": result_path != path,
            "is_absolute": result_path.startswith('/'),
            "segments": [s for s in result_path.split('/') if s],
            "depth": len([s for s in result_path.split('/') if s]),
            "has_trailing_slash": result_path.endswith('/') and result_path != '/',
            "is_root": result_path == '/'
        }

        if base_url:
            result["base_url"] = base_url
            # Reconstruct full URL if base was provided
            parsed_base = urlparse(base_url)
            full_url = urlunparse((
                parsed_base.scheme,
                parsed_base.netloc,
                result_path,
                parsed_base.params,
                parsed_base.query,
                parsed_base.fragment
            ))
            result["full_url"] = full_url

        return {
            "ok": True,
            "result": result
        }

    except Exception as e:
        return {"ok": False, "error": f"failed to normalize path: {str(e)}"}