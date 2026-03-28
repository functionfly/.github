VALID_TYPES = {
    "Point", "MultiPoint", "LineString", "MultiLineString",
    "Polygon", "MultiPolygon", "GeometryCollection",
    "Feature", "FeatureCollection"
}

GEOMETRY_TYPES = {
    "Point", "MultiPoint", "LineString", "MultiLineString",
    "Polygon", "MultiPolygon", "GeometryCollection"
}


def validate_coordinate(coord, path=""):
    errors = []
    warnings = []
    if not isinstance(coord, (list, tuple)) or len(coord) < 2:
        errors.append(f"{path}: coordinate must be [lng, lat] array")
        return errors, warnings
    try:
        lng = float(coord[0])
        lat = float(coord[1])
    except (TypeError, ValueError):
        errors.append(f"{path}: coordinate values must be numbers")
        return errors, warnings
    if not (-180 <= lng <= 180):
        errors.append(f"{path}: longitude {lng} out of range [-180, 180]")
    if not (-90 <= lat <= 90):
        errors.append(f"{path}: latitude {lat} out of range [-90, 90]")
    if len(coord) > 3:
        warnings.append(f"{path}: coordinate has more than 3 values (extra values ignored)")
    return errors, warnings


def validate_geometry(geom, path="geometry"):
    errors = []
    warnings = []

    if not isinstance(geom, dict):
        errors.append(f"{path}: must be an object")
        return errors, warnings

    gtype = geom.get("type")
    if gtype not in GEOMETRY_TYPES:
        errors.append(f"{path}.type: must be a valid geometry type, got '{gtype}'")
        return errors, warnings

    if gtype == "GeometryCollection":
        geometries = geom.get("geometries")
        if not isinstance(geometries, list):
            errors.append(f"{path}.geometries: must be an array")
        else:
            for i, g in enumerate(geometries):
                e, w = validate_geometry(g, f"{path}.geometries[{i}]")
                errors.extend(e)
                warnings.extend(w)
        return errors, warnings

    coords = geom.get("coordinates")
    if coords is None:
        errors.append(f"{path}.coordinates: required")
        return errors, warnings

    if gtype == "Point":
        e, w = validate_coordinate(coords, f"{path}.coordinates")
        errors.extend(e)
        warnings.extend(w)

    elif gtype == "MultiPoint":
        if not isinstance(coords, list):
            errors.append(f"{path}.coordinates: must be array of positions")
        else:
            for i, c in enumerate(coords):
                e, w = validate_coordinate(c, f"{path}.coordinates[{i}]")
                errors.extend(e)
                warnings.extend(w)

    elif gtype == "LineString":
        if not isinstance(coords, list) or len(coords) < 2:
            errors.append(f"{path}.coordinates: LineString must have at least 2 positions")
        else:
            for i, c in enumerate(coords):
                e, w = validate_coordinate(c, f"{path}.coordinates[{i}]")
                errors.extend(e)
                warnings.extend(w)

    elif gtype == "MultiLineString":
        if not isinstance(coords, list):
            errors.append(f"{path}.coordinates: must be array of LineString coordinate arrays")
        else:
            for i, line in enumerate(coords):
                if not isinstance(line, list) or len(line) < 2:
                    errors.append(f"{path}.coordinates[{i}]: LineString must have at least 2 positions")
                else:
                    for j, c in enumerate(line):
                        e, w = validate_coordinate(c, f"{path}.coordinates[{i}][{j}]")
                        errors.extend(e)
                        warnings.extend(w)

    elif gtype == "Polygon":
        if not isinstance(coords, list) or len(coords) < 1:
            errors.append(f"{path}.coordinates: Polygon must have at least one ring")
        else:
            for i, ring in enumerate(coords):
                if not isinstance(ring, list) or len(ring) < 4:
                    errors.append(f"{path}.coordinates[{i}]: ring must have at least 4 positions")
                else:
                    for j, c in enumerate(ring):
                        e, w = validate_coordinate(c, f"{path}.coordinates[{i}][{j}]")
                        errors.extend(e)
                        warnings.extend(w)
                    # Check if ring is closed
                    if ring[0] != ring[-1]:
                        warnings.append(f"{path}.coordinates[{i}]: ring is not closed (first and last positions should be equal)")

    elif gtype == "MultiPolygon":
        if not isinstance(coords, list):
            errors.append(f"{path}.coordinates: must be array of Polygon coordinate arrays")
        else:
            for i, poly in enumerate(coords):
                if not isinstance(poly, list) or len(poly) < 1:
                    errors.append(f"{path}.coordinates[{i}]: Polygon must have at least one ring")
                else:
                    for j, ring in enumerate(poly):
                        if not isinstance(ring, list) or len(ring) < 4:
                            errors.append(f"{path}.coordinates[{i}][{j}]: ring must have at least 4 positions")
                        else:
                            for k, c in enumerate(ring):
                                e, w = validate_coordinate(c, f"{path}.coordinates[{i}][{j}][{k}]")
                                errors.extend(e)
                                warnings.extend(w)

    return errors, warnings


def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}

    geojson = event.get("geojson")
    if geojson is None:
        return {"ok": False, "error": "geojson is required"}

    if not isinstance(geojson, dict):
        return {"ok": False, "error": "geojson must be an object"}

    errors = []
    warnings = []

    gtype = geojson.get("type")
    if not gtype:
        errors.append("type: required")
    elif gtype not in VALID_TYPES:
        errors.append(f"type: '{gtype}' is not a valid GeoJSON type")

    if gtype == "Feature":
        geometry = geojson.get("geometry")
        if geometry is None:
            warnings.append("geometry: null geometry is allowed but unusual")
        elif isinstance(geometry, dict):
            e, w = validate_geometry(geometry)
            errors.extend(e)
            warnings.extend(w)
        else:
            errors.append("geometry: must be a geometry object or null")

        if "properties" not in geojson:
            warnings.append("properties: missing (null is allowed)")

    elif gtype == "FeatureCollection":
        features = geojson.get("features")
        if not isinstance(features, list):
            errors.append("features: must be an array")
        else:
            for i, feat in enumerate(features):
                if not isinstance(feat, dict) or feat.get("type") != "Feature":
                    errors.append(f"features[{i}]: must be a Feature object")
                else:
                    geometry = feat.get("geometry")
                    if geometry and isinstance(geometry, dict):
                        e, w = validate_geometry(geometry, f"features[{i}].geometry")
                        errors.extend(e)
                        warnings.extend(w)

    elif gtype in GEOMETRY_TYPES:
        e, w = validate_geometry(geojson)
        errors.extend(e)
        warnings.extend(w)

    valid = len(errors) == 0

    return {
        "ok": True,
        "result": {
            "valid": valid,
            "type": gtype,
            "errors": errors,
            "warnings": warnings
        }
    }
