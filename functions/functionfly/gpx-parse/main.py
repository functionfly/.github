import xml.etree.ElementTree as ET
import re


def strip_ns(tag):
    return re.sub(r'\{[^}]+\}', '', tag)


def find_text(elem, tag):
    for child in elem:
        if strip_ns(child.tag) == tag:
            return child.text or ""
    return ""


def find_elem(elem, tag):
    for child in elem:
        if strip_ns(child.tag) == tag:
            return child
    return None


def find_all_direct(elem, tag):
    return [child for child in elem if strip_ns(child.tag) == tag]


def parse_wpt(wpt_elem, gpx_type="waypoint"):
    """Parse a waypoint/route point/track point element."""
    try:
        lat = float(wpt_elem.get("lat", 0))
        lon = float(wpt_elem.get("lon", 0))
    except (TypeError, ValueError):
        return None

    properties = {"gpx_type": gpx_type}

    name = find_text(wpt_elem, "name")
    if name:
        properties["name"] = name

    desc = find_text(wpt_elem, "desc")
    if desc:
        properties["description"] = desc

    sym = find_text(wpt_elem, "sym")
    if sym:
        properties["symbol"] = sym

    ele = find_text(wpt_elem, "ele")
    time = find_text(wpt_elem, "time")
    if time:
        properties["time"] = time

    coords = [lon, lat]
    if ele:
        try:
            coords.append(float(ele))
        except ValueError:
            pass

    return {
        "type": "Feature",
        "geometry": {"type": "Point", "coordinates": coords},
        "properties": properties
    }


def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}

    gpx = event.get("gpx")
    if gpx is None:
        return {"ok": False, "error": "gpx is required"}

    if not isinstance(gpx, str):
        return {"ok": False, "error": "gpx must be a string"}

    gpx = gpx.strip()
    if not gpx:
        return {"ok": False, "error": "gpx cannot be empty"}

    try:
        root = ET.fromstring(gpx)
        features = []
        metadata = {}
        waypoint_count = 0
        track_count = 0
        route_count = 0

        # Parse metadata
        meta_elem = find_elem(root, "metadata")
        if meta_elem:
            name = find_text(meta_elem, "name")
            if name:
                metadata["name"] = name
            desc = find_text(meta_elem, "desc")
            if desc:
                metadata["description"] = desc
            time = find_text(meta_elem, "time")
            if time:
                metadata["time"] = time

        # Parse waypoints
        for wpt in find_all_direct(root, "wpt"):
            feat = parse_wpt(wpt, "waypoint")
            if feat:
                features.append(feat)
                waypoint_count += 1

        # Parse routes
        for rte in find_all_direct(root, "rte"):
            route_count += 1
            rte_name = find_text(rte, "name")
            coords = []
            for rtept in find_all_direct(rte, "rtept"):
                try:
                    lat = float(rtept.get("lat", 0))
                    lon = float(rtept.get("lon", 0))
                    ele = find_text(rtept, "ele")
                    if ele:
                        try:
                            coords.append([lon, lat, float(ele)])
                        except ValueError:
                            coords.append([lon, lat])
                    else:
                        coords.append([lon, lat])
                except (TypeError, ValueError):
                    pass

            if coords:
                props = {"gpx_type": "route"}
                if rte_name:
                    props["name"] = rte_name
                features.append({
                    "type": "Feature",
                    "geometry": {"type": "LineString", "coordinates": coords},
                    "properties": props
                })

        # Parse tracks
        for trk in find_all_direct(root, "trk"):
            track_count += 1
            trk_name = find_text(trk, "name")
            all_coords = []

            for trkseg in find_all_direct(trk, "trkseg"):
                seg_coords = []
                for trkpt in find_all_direct(trkseg, "trkpt"):
                    try:
                        lat = float(trkpt.get("lat", 0))
                        lon = float(trkpt.get("lon", 0))
                        ele = find_text(trkpt, "ele")
                        if ele:
                            try:
                                seg_coords.append([lon, lat, float(ele)])
                            except ValueError:
                                seg_coords.append([lon, lat])
                        else:
                            seg_coords.append([lon, lat])
                    except (TypeError, ValueError):
                        pass
                if seg_coords:
                    all_coords.append(seg_coords)

            if all_coords:
                props = {"gpx_type": "track"}
                if trk_name:
                    props["name"] = trk_name

                if len(all_coords) == 1:
                    features.append({
                        "type": "Feature",
                        "geometry": {"type": "LineString", "coordinates": all_coords[0]},
                        "properties": props
                    })
                else:
                    features.append({
                        "type": "Feature",
                        "geometry": {"type": "MultiLineString", "coordinates": all_coords},
                        "properties": props
                    })

        return {
            "ok": True,
            "result": {
                "type": "FeatureCollection",
                "features": features,
                "metadata": metadata,
                "waypoint_count": waypoint_count,
                "track_count": track_count,
                "route_count": route_count
            }
        }

    except ET.ParseError as e:
        return {"ok": False, "error": f"invalid GPX XML: {str(e)}"}
    except Exception as e:
        return {"ok": False, "error": f"GPX parsing failed: {str(e)}"}
