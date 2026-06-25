#!/bin/bash
# Script to download and generate a comprehensive world cities seed CSV
# Data source: GeoNames (https://www.geonames.org/) - CC BY 4.0 license
#
# Usage: ./scripts/generate-world-cities-seed.sh
#
# This script downloads cities with population > 15000 from GeoNames
# and converts them to the format expected by SeedFromCSV.
#
# Output: data/cities_seed.csv

set -e

DATA_DIR="data"
OUTPUT_FILE="${DATA_DIR}/cities_seed.csv"
TMP_DIR=$(mktemp -d)
ARCHIVE_URL="https://download.geonames.org/export/dump/cities15000.zip"
ARCHIVE_FILE="${TMP_DIR}/cities15000.zip"

echo "Downloading GeoNames cities dataset..."
curl -L -o "$ARCHIVE_FILE" "$ARCHIVE_URL"

echo "Extracting..."
unzip -o "$ARCHIVE_FILE" -d "$TMP_DIR"

# The main file is cities15000.txt
CITIES_FILE="${TMP_DIR}/cities15000.txt"

if [ ! -f "$CITIES_FILE" ]; then
    echo "Error: cities15000.txt not found after extraction"
    exit 1
fi

# Create output directory if needed
mkdir -p "$DATA_DIR"

# US state code to name mapping (abbreviated - extend as needed)
declare -A US_STATES=(
    ["AL"]="Alabama" ["AK"]="Alaska" ["AZ"]="Arizona" ["AR"]="Arkansas"
    ["CA"]="California" ["CO"]="Colorado" ["CT"]="Connecticut" ["DE"]="Delaware"
    ["FL"]="Florida" ["GA"]="Georgia" ["HI"]="Hawaii" ["ID"]="Idaho"
    ["IL"]="Illinois" ["IN"]="Indiana" ["IA"]="Iowa" ["KS"]="Kansas"
    ["KY"]="Kentucky" ["LA"]="Louisiana" ["ME"]="Maine" ["MD"]="Maryland"
    ["MA"]="Massachusetts" ["MI"]="Michigan" ["MN"]="Minnesota" ["MS"]="Mississippi"
    ["MO"]="Missouri" ["MT"]="Montana" ["NE"]="Nebraska" ["NV"]="Nevada"
    ["NH"]="New Hampshire" ["NJ"]="New Jersey" ["NM"]="New Mexico" ["NY"]="New York"
    ["NC"]="North Carolina" ["ND"]="North Dakota" ["OH"]="Ohio" ["OK"]="Oklahoma"
    ["OR"]="Oregon" ["PA"]="Pennsylvania" ["RI"]="Rhode Island" ["SC"]="South Carolina"
    ["SD"]="South Dakota" ["TN"]="Tennessee" ["TX"]="Texas" ["UT"]="Utah"
    ["VT"]="Vermont" ["VA"]="Virginia" ["WA"]="Washington" ["WV"]="West Virginia"
    ["WI"]="Wisconsin" ["WY"]="Wyoming" ["DC"]="District of Columbia"
)

# Country code to country name mapping
declare -A COUNTRY_NAMES=(
    ["US"]="United States" ["GB"]="United Kingdom" ["CA"]="Canada"
    ["AU"]="Australia" ["DE"]="Germany" ["FR"]="France" ["JP"]="Japan"
    ["IN"]="India" ["BR"]="Brazil" ["MX"]="Mexico" ["SG"]="Singapore"
    ["IL"]="Israel" ["AE"]="United Arab Emirates" ["KR"]="South Korea"
    ["CH"]="Switzerland" ["NL"]="Netherlands" ["SE"]="Sweden" ["NZ"]="New Zealand"
    ["ZA"]="South Africa" ["AR"]="Argentina" ["ES"]="Spain" ["PT"]="Portugal"
)

# Write header
echo "slug,name,country_code,population,latitude,longitude,state_code,state_name,country_name,metro_slug" > "$OUTPUT_FILE"

# Process the TSV file
# Fields: geonameid,name,asciiname,alternatenames,latitude,longitude,feature_class,feature_code,country_code,cc2,admin1_code,admin2_code,admin3_code,admin4_code,population,elevation,dem,timezone,modification_date
awk -F'\t' '
BEGIN { OFS="," }
{
    # Skip if population is 0 or empty
    if ($15 == "" || $15 == "0") next

    # Get fields
    name = $2
    lat = $5
    lon = $6
    country = $10
    state_code = $11
    population = $15

    # Skip if name is empty
    if (name == "") next

    # Generate slug
    slug = tolower(name)
    gsub(/[^a-z0-9]/, "-", slug)
    gsub(/-+/, "-", slug)
    gsub(/^-|-$/, "", slug)

    # Add country code to slug
    slug = slug "-" tolower(country)

    # For US cities, add state to slug
    if (country == "US" && state_code != "") {
        slug = slug "-" tolower(state_code)
    }

    # Get country name
    country_name = ""
    if (country == "US") country_name = "United States"
    else if (country == "GB") country_name = "United Kingdom"
    else if (country == "CA") country_name = "Canada"
    else if (country == "AU") country_name = "Australia"
    else if (country == "DE") country_name = "Germany"
    else if (country == "FR") country_name = "France"
    else if (country == "JP") country_name = "Japan"
    else if (country == "IN") country_name = "India"
    else if (country == "BR") country_name = "Brazil"
    else if (country == "MX") country_name = "Mexico"
    else if (country == "SG") country_name = "Singapore"
    else if (country == "IL") country_name = "Israel"
    else if (country == "AE") country_name = "United Arab Emirates"
    else if (country == "KR") country_name = "South Korea"
    else if (country == "CH") country_name = "Switzerland"
    else if (country == "NL") country_name = "Netherlands"
    else if (country == "SE") country_name = "Sweden"
    else if (country == "NZ") country_name = "New Zealand"
    else if (country == "ZA") country_name = "South Africa"
    else if (country == "AR") country_name = "Argentina"
    else if (country == "ES") country_name = "Spain"
    else if (country == "PT") country_name = "Portugal"
    else country_name = country  # Fallback to country code

    # State name (for US)
    state_name = ""
    if (country == "US" && state_code != "") {
        state_name = state_code  # Would need lookup table for full names
    }

    # Metro slug (for now, same as city slug)
    metro_slug = slug

    # Escape fields for CSV
    gsub(/"/, "\"\"", name)
    gsub(/"/, "\"\"", country_name)

    # Print CSV row
    print "\"" slug "\",\"" name "\"," country "," population "," lat "," lon ",\"" state_code "\",\"" state_name "\",\"" country_name "\",\"" metro_slug "\""
}' "$CITIES_FILE" >> "$OUTPUT_FILE"

# Count lines
LINE_COUNT=$(wc -l < "$OUTPUT_FILE")
echo "Generated ${LINE_COUNT} lines (including header)"

# Cleanup
rm -rf "$TMP_DIR"

echo "Output written to: $OUTPUT_FILE"
echo ""
echo "To load into database, restart the orchestrator API or run:"
echo "  SELECT cityranking_load_seed('/path/to/${OUTPUT_FILE}')"
