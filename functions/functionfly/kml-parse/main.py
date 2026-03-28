import xml.etree.ElementTree as ET
import re


KML_NS = "http://www.opengis.net/kml/2.2"
KML_NS_ALT = "http://earth.google.com/kml/2.2"
KML_NS_ALT2 = "http://earth.google.com/kml/2.1"


def strip_ns(tag):
    """Remove namespace from tag."""
    return re.sub(r'\{[^}]+\}', '', tag)


def parse_coordinates(coord_text):
    """Parse KML coordinates string to list of [lng, lat, alt?]."""
    coords = []
    for part in coord_text.strip().split():
        vals = part.split(',')
        if len(vals) >= 2:
            try:
                lng = float(vals[0])
                lat = float(vals[1])
                if len(vals) >= 3:
                    alt = float(vals[2])
                    coords.append([lng, lat, alt])
                else:
                    coords.append([lng, lat])
            except ValueError:
                pass
    return coords


def find_text(elem, tag):
    """Find text of a child element by tag name (ignoring namespace)."""
    for child in elem:
        if strip_ns(child.tag) == tag:
            return child.text or ""
    return ""


def find_elem(elem, tag):
    """Find first child element by tag name (ignoring namespace)."""
    for child in elem:
        if strip_ns(child.tag) == tag:
            return child
    return None


def find_all(elem, tag):
    """Find all child elements by tag name (ignoring namespace)."""
    return [child for child in elem if strip_ns(child.tag) == tag]


def parse_geometry(placemark):
    """Extract geometry from a Placemark element."""
    for child in placemark:
        tag = strip_ns(child.tag)

        if tag == "Point":
            coord_elem = find_elem(child, "coordinates")
            if coord_elem is not None and coord_elem.text:
                coords = parse_coordinates(coord_elem.text)
                if coords:
                    return {"type": "Point", "coordinates": coords[0]}

        elif tag == "LineString":
            coord_elem = find_elem(child, "coordinates")
            if coord_elem is not None and coord_elem.text:
                coords = parse_coordinates(coord_elem.text)
                if coords:
                    return {"type": "LineString", "coordinates": coords}

        elif tag == "LinearRing":
            coord_elem = find_elem(child, "coordinates")
            if coord_elem is not None and coord_elem.text:
                coords = parse_coordinates(coord_elem.text)
                if coords:
                    return {"type": "LineString", "coordinates": coords}

        elif tag == "Polygon":
            outer = find_elem(child, "outerBoundaryIs")
            rings = []
            if outer:
                lr = find_elem(outer, "LinearRing")
                if lr:
                    coord_elem = find_elem(lr, "coordinates")
                    if coord_elem is not None and coord_elem.text:
                        rings.append(parse_coordinates(coord_elem.text))

            inner_boundaries = find_all(child, "innerBoundaryIs")
            for inner in inner_boundaries:
                lr = find_elem(inner, "LinearRing")
                if lr:
                    coord_elem = find_elem(lr, "coordinates")
                    if coord_elem is not None and coord_elem.text:
                        rings.append(parse_coordinates(coord_elem.text))

            if rings:
                return {"type": "Polygon", "coordinates": rings}

        elif tag == "MultiGeometry":
            geometries = []
            for sub in child:
                sub_geom = parse_geometry_from_elem(sub)
                if sub_geom:
                    geometries.append(sub_geom)
            if geometries:
                return {"type": "GeometryCollection", "geometries": geometries}

    return None


def parse_geometry_from_elem(elem):
    """Parse geometry from a geometry element directly."""
    tag = strip_ns(elem.tag)

    if tag == "Point":
        coord_elem = find_elem(elem, "coordinates")
        if coord_elem is not None and coord_elem.text:
            coords = parse_coordinates(coord_elem.text)
            if coords:
                return {"type": "Point", "coordinates": coords[0]}

    elif tag == "LineString":
        coord_elem = find_elem(elem, "coordinates")
        if coord_elem is not None and coord_elem.text:
            coords = parse_coordinates(coord_elem.text)
            if coords:
                return {"type": "LineString", "coordinates": coords}

    elif tag == "Polygon":
        outer = find_elem(elem, "outerBoundaryIs")
        rings = []
        if outer:
            lr = find_elem(outer, "LinearRing")
            if lr:
                coord_elem = find_elem(lr, "coordinates")
                if coord_elem is not None and coord_elem.text:
                    rings.append(parse_coordinates(coord_elem.text))
        if rings:
            return {"type": "Polygon", "coordinates": rings}

    return None


def parse_placemarks(root):
    """Recursively find all Placemarks in the KML tree."""
    features = []

    def recurse(elem):
        tag = strip_ns(elem.tag)
        if tag == "Placemark":
            name = find_text(elem, "name")
            description = find_text(elem, "description")
            geometry = parse_geometry(elem)

            properties = {}
            if name:
                properties["name"] = name
            if description:
                properties["description"] = description

            # Extended data
            ext_data = find_elem(elem, "ExtendedData")
            if ext_data:
                for data in find_all(ext_data, "Data"):
                    key = data.get("name", "")
                    val_elem = find_elem(data, "value")
                    if key and val_elem is not None:
                        properties[key] = val_elem.text or ""

            feature = {
                "type": "Feature",
                "geometry": geometry,
                "properties": properties
            }
            features.append(feature)
        else:
            for child in elem:
                recurse(child)

    recurse(root)
    return features


def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}

    kml = event.get("kml")
    if kml is None:
        return {"ok": False, "error": "kml is required"}

    if not isinstance(kml, str):
        return {"ok": False, "error": "kml must be a string"}

    kml = kml.strip()
    if not kml:
        return {"ok": False, "error": "kml cannot be empty"}

    try:
        root = ET.fromstring(kml)
        features = parse_placemarks(root)

        return {
            "ok": True,
            "result": {
                "type": "FeatureCollection",
                "features": features,
                "feature_count": len(features)
            }
        }

    except ET.ParseError as e:
        return {"ok": False, "error": f"invalid KML XML: {str(e)}"}
    except Exception as e:
        return {"ok": False, "error": f"KML parsing failed: {str(e)}"}
