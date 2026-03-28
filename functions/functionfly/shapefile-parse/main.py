import base64
import struct


# Shapefile shape types
SHAPE_TYPES = {
    0: "Null",
    1: "Point",
    3: "Polyline",
    5: "Polygon",
    8: "MultiPoint",
    11: "PointZ",
    13: "PolylineZ",
    15: "PolygonZ",
    18: "MultiPointZ",
    21: "PointM",
    23: "PolylineM",
    25: "PolygonM",
    28: "MultiPointM",
    31: "MultiPatch"
}


def parse_shp_header(data):
    """Parse shapefile header (100 bytes)."""
    if len(data) < 100:
        return None, "shapefile too short (less than 100 bytes)"

    # File code (big-endian)
    file_code = struct.unpack('>i', data[0:4])[0]
    if file_code != 9994:
        return None, f"invalid shapefile file code: {file_code} (expected 9994)"

    # File length (big-endian, in 16-bit words)
    file_length = struct.unpack('>i', data[24:28])[0] * 2

    # Version (little-endian)
    version = struct.unpack('<i', data[28:32])[0]

    # Shape type (little-endian)
    shape_type = struct.unpack('<i', data[32:36])[0]

    # Bounding box
    xmin = struct.unpack('<d', data[36:44])[0]
    ymin = struct.unpack('<d', data[44:52])[0]
    xmax = struct.unpack('<d', data[52:60])[0]
    ymax = struct.unpack('<d', data[60:68])[0]

    return {
        "file_code": file_code,
        "file_length": file_length,
        "version": version,
        "shape_type": shape_type,
        "shape_type_name": SHAPE_TYPES.get(shape_type, "Unknown"),
        "bbox": [xmin, ymin, xmax, ymax]
    }, None


def parse_shp_records(data, header):
    """Parse shapefile records."""
    features = []
    offset = 100  # Start after header
    shape_type = header["shape_type"]

    while offset < len(data) - 8:
        try:
            # Record header
            record_num = struct.unpack('>i', data[offset:offset+4])[0]
            content_length = struct.unpack('>i', data[offset+4:offset+8])[0] * 2
            offset += 8

            if offset + content_length > len(data):
                break

            record_data = data[offset:offset+content_length]
            offset += content_length

            if len(record_data) < 4:
                continue

            rec_shape_type = struct.unpack('<i', record_data[0:4])[0]

            if rec_shape_type == 0:
                # Null shape
                features.append({
                    "type": "Feature",
                    "geometry": None,
                    "properties": {"record_num": record_num}
                })

            elif rec_shape_type == 1:
                # Point
                if len(record_data) >= 20:
                    x = struct.unpack('<d', record_data[4:12])[0]
                    y = struct.unpack('<d', record_data[12:20])[0]
                    features.append({
                        "type": "Feature",
                        "geometry": {"type": "Point", "coordinates": [x, y]},
                        "properties": {"record_num": record_num}
                    })

            elif rec_shape_type in (3, 5):
                # Polyline or Polygon
                if len(record_data) >= 44:
                    num_parts = struct.unpack('<i', record_data[36:40])[0]
                    num_points = struct.unpack('<i', record_data[40:44])[0]

                    parts_offset = 44
                    parts = []
                    for i in range(num_parts):
                        part_start = struct.unpack('<i', record_data[parts_offset:parts_offset+4])[0]
                        parts.append(part_start)
                        parts_offset += 4

                    points_offset = parts_offset
                    all_points = []
                    for i in range(num_points):
                        if points_offset + 16 <= len(record_data):
                            x = struct.unpack('<d', record_data[points_offset:points_offset+8])[0]
                            y = struct.unpack('<d', record_data[points_offset+8:points_offset+16])[0]
                            all_points.append([x, y])
                            points_offset += 16

                    rings = []
                    for i, part_start in enumerate(parts):
                        part_end = parts[i+1] if i+1 < len(parts) else len(all_points)
                        rings.append(all_points[part_start:part_end])

                    if rec_shape_type == 3:
                        geom_type = "MultiLineString" if len(rings) > 1 else "LineString"
                        coords = rings if len(rings) > 1 else rings[0]
                    else:
                        geom_type = "Polygon"
                        coords = rings

                    features.append({
                        "type": "Feature",
                        "geometry": {"type": geom_type, "coordinates": coords},
                        "properties": {"record_num": record_num}
                    })

        except (struct.error, IndexError):
            break

    return features


def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}

    data_b64 = event.get("data")
    url = event.get("url")
    fmt = event.get("format", "url")

    if data_b64 is None and url is None:
        return {"ok": False, "error": "either data or url is required"}

    try:
        if data_b64 and fmt == "base64":
            # Parse actual shapefile data
            try:
                shp_data = base64.b64decode(data_b64)
            except Exception:
                return {"ok": False, "error": "invalid base64 data"}

            header, err = parse_shp_header(shp_data)
            if err:
                return {"ok": False, "error": err}

            features = parse_shp_records(shp_data, header)

            return {
                "ok": True,
                "result": {
                    "type": "FeatureCollection",
                    "features": features,
                    "feature_count": len(features),
                    "geometry_type": header["shape_type_name"],
                    "bbox": header["bbox"],
                    "version": header["version"]
                }
            }

        else:
            # Mock response for URL-based requests
            mock_features = [
                {
                    "type": "Feature",
                    "geometry": {"type": "Point", "coordinates": [-74.006, 40.7128]},
                    "properties": {"id": 1, "name": "Sample Point 1"}
                },
                {
                    "type": "Feature",
                    "geometry": {"type": "Point", "coordinates": [-87.6298, 41.8781]},
                    "properties": {"id": 2, "name": "Sample Point 2"}
                },
                {
                    "type": "Feature",
                    "geometry": {"type": "Polygon", "coordinates": [
                        [[-74.1, 40.6], [-73.9, 40.6], [-73.9, 40.8], [-74.1, 40.8], [-74.1, 40.6]]
                    ]},
                    "properties": {"id": 3, "name": "Sample Polygon"}
                }
            ]

            return {
                "ok": True,
                "result": {
                    "type": "FeatureCollection",
                    "features": mock_features,
                    "feature_count": len(mock_features),
                    "geometry_type": "Mixed",
                    "source_url": url,
                    "note": "Mock data - provide base64-encoded .shp file data for actual parsing"
                }
            }

    except Exception as e:
        return {"ok": False, "error": f"shapefile parsing failed: {str(e)}"}
