import math


CITY_DATABASE = {
    "new york": {"lat": 40.7128, "lng": -74.0060, "formatted": "New York, NY, USA", "type": "city"},
    "new york, ny": {"lat": 40.7128, "lng": -74.0060, "formatted": "New York, NY, USA", "type": "city"},
    "los angeles": {"lat": 34.0522, "lng": -118.2437, "formatted": "Los Angeles, CA, USA", "type": "city"},
    "los angeles, ca": {"lat": 34.0522, "lng": -118.2437, "formatted": "Los Angeles, CA, USA", "type": "city"},
    "chicago": {"lat": 41.8781, "lng": -87.6298, "formatted": "Chicago, IL, USA", "type": "city"},
    "chicago, il": {"lat": 41.8781, "lng": -87.6298, "formatted": "Chicago, IL, USA", "type": "city"},
    "houston": {"lat": 29.7604, "lng": -95.3698, "formatted": "Houston, TX, USA", "type": "city"},
    "phoenix": {"lat": 33.4484, "lng": -112.0740, "formatted": "Phoenix, AZ, USA", "type": "city"},
    "philadelphia": {"lat": 39.9526, "lng": -75.1652, "formatted": "Philadelphia, PA, USA", "type": "city"},
    "san antonio": {"lat": 29.4241, "lng": -98.4936, "formatted": "San Antonio, TX, USA", "type": "city"},
    "san diego": {"lat": 32.7157, "lng": -117.1611, "formatted": "San Diego, CA, USA", "type": "city"},
    "dallas": {"lat": 32.7767, "lng": -96.7970, "formatted": "Dallas, TX, USA", "type": "city"},
    "san jose": {"lat": 37.3382, "lng": -121.8863, "formatted": "San Jose, CA, USA", "type": "city"},
    "austin": {"lat": 30.2672, "lng": -97.7431, "formatted": "Austin, TX, USA", "type": "city"},
    "seattle": {"lat": 47.6062, "lng": -122.3321, "formatted": "Seattle, WA, USA", "type": "city"},
    "denver": {"lat": 39.7392, "lng": -104.9903, "formatted": "Denver, CO, USA", "type": "city"},
    "boston": {"lat": 42.3601, "lng": -71.0589, "formatted": "Boston, MA, USA", "type": "city"},
    "miami": {"lat": 25.7617, "lng": -80.1918, "formatted": "Miami, FL, USA", "type": "city"},
    "atlanta": {"lat": 33.7490, "lng": -84.3880, "formatted": "Atlanta, GA, USA", "type": "city"},
    "london": {"lat": 51.5074, "lng": -0.1278, "formatted": "London, UK", "type": "city"},
    "paris": {"lat": 48.8566, "lng": 2.3522, "formatted": "Paris, France", "type": "city"},
    "tokyo": {"lat": 35.6762, "lng": 139.6503, "formatted": "Tokyo, Japan", "type": "city"},
    "berlin": {"lat": 52.5200, "lng": 13.4050, "formatted": "Berlin, Germany", "type": "city"},
    "sydney": {"lat": -33.8688, "lng": 151.2093, "formatted": "Sydney, NSW, Australia", "type": "city"},
    "toronto": {"lat": 43.6532, "lng": -79.3832, "formatted": "Toronto, ON, Canada", "type": "city"},
    "singapore": {"lat": 1.3521, "lng": 103.8198, "formatted": "Singapore", "type": "city"},
    "dubai": {"lat": 25.2048, "lng": 55.2708, "formatted": "Dubai, UAE", "type": "city"},
    "san francisco": {"lat": 37.7749, "lng": -122.4194, "formatted": "San Francisco, CA, USA", "type": "city"},
    "san francisco, ca": {"lat": 37.7749, "lng": -122.4194, "formatted": "San Francisco, CA, USA", "type": "city"},
    "washington": {"lat": 38.9072, "lng": -77.0369, "formatted": "Washington, DC, USA", "type": "city"},
    "washington, dc": {"lat": 38.9072, "lng": -77.0369, "formatted": "Washington, DC, USA", "type": "city"},
    "portland": {"lat": 45.5051, "lng": -122.6750, "formatted": "Portland, OR, USA", "type": "city"},
    "las vegas": {"lat": 36.1699, "lng": -115.1398, "formatted": "Las Vegas, NV, USA", "type": "city"},
    "minneapolis": {"lat": 44.9778, "lng": -93.2650, "formatted": "Minneapolis, MN, USA", "type": "city"},
    "nashville": {"lat": 36.1627, "lng": -86.7816, "formatted": "Nashville, TN, USA", "type": "city"},
    "new orleans": {"lat": 29.9511, "lng": -90.0715, "formatted": "New Orleans, LA, USA", "type": "city"},
    "salt lake city": {"lat": 40.7608, "lng": -111.8910, "formatted": "Salt Lake City, UT, USA", "type": "city"},
    "kansas city": {"lat": 39.0997, "lng": -94.5786, "formatted": "Kansas City, MO, USA", "type": "city"},
    "columbus": {"lat": 39.9612, "lng": -82.9988, "formatted": "Columbus, OH, USA", "type": "city"},
    "charlotte": {"lat": 35.2271, "lng": -80.8431, "formatted": "Charlotte, NC, USA", "type": "city"},
    "indianapolis": {"lat": 39.7684, "lng": -86.1581, "formatted": "Indianapolis, IN, USA", "type": "city"},
    "memphis": {"lat": 35.1495, "lng": -90.0490, "formatted": "Memphis, TN, USA", "type": "city"},
    "baltimore": {"lat": 39.2904, "lng": -76.6122, "formatted": "Baltimore, MD, USA", "type": "city"},
    "milwaukee": {"lat": 43.0389, "lng": -87.9065, "formatted": "Milwaukee, WI, USA", "type": "city"},
    "albuquerque": {"lat": 35.0844, "lng": -106.6504, "formatted": "Albuquerque, NM, USA", "type": "city"},
    "tucson": {"lat": 32.2226, "lng": -110.9747, "formatted": "Tucson, AZ, USA", "type": "city"},
    "fresno": {"lat": 36.7378, "lng": -119.7871, "formatted": "Fresno, CA, USA", "type": "city"},
    "sacramento": {"lat": 38.5816, "lng": -121.4944, "formatted": "Sacramento, CA, USA", "type": "city"},
    "mesa": {"lat": 33.4152, "lng": -111.8315, "formatted": "Mesa, AZ, USA", "type": "city"},
    "omaha": {"lat": 41.2565, "lng": -95.9345, "formatted": "Omaha, NE, USA", "type": "city"},
    "raleigh": {"lat": 35.7796, "lng": -78.6382, "formatted": "Raleigh, NC, USA", "type": "city"},
    "colorado springs": {"lat": 38.8339, "lng": -104.8214, "formatted": "Colorado Springs, CO, USA", "type": "city"},
    "long beach": {"lat": 33.7701, "lng": -118.1937, "formatted": "Long Beach, CA, USA", "type": "city"},
    "virginia beach": {"lat": 36.8529, "lng": -75.9780, "formatted": "Virginia Beach, VA, USA", "type": "city"},
    "oakland": {"lat": 37.8044, "lng": -122.2712, "formatted": "Oakland, CA, USA", "type": "city"},
    "tampa": {"lat": 27.9506, "lng": -82.4572, "formatted": "Tampa, FL, USA", "type": "city"},
    "tulsa": {"lat": 36.1540, "lng": -95.9928, "formatted": "Tulsa, OK, USA", "type": "city"},
    "arlington": {"lat": 32.7357, "lng": -97.1081, "formatted": "Arlington, TX, USA", "type": "city"},
    "new orleans, la": {"lat": 29.9511, "lng": -90.0715, "formatted": "New Orleans, LA, USA", "type": "city"},
    "wichita": {"lat": 37.6872, "lng": -97.3301, "formatted": "Wichita, KS, USA", "type": "city"},
    "bakersfield": {"lat": 35.3733, "lng": -119.0187, "formatted": "Bakersfield, CA, USA", "type": "city"},
    "aurora": {"lat": 39.7294, "lng": -104.8319, "formatted": "Aurora, CO, USA", "type": "city"},
    "anaheim": {"lat": 33.8366, "lng": -117.9143, "formatted": "Anaheim, CA, USA", "type": "city"},
    "santa ana": {"lat": 33.7455, "lng": -117.8677, "formatted": "Santa Ana, CA, USA", "type": "city"},
    "corpus christi": {"lat": 27.8006, "lng": -97.3964, "formatted": "Corpus Christi, TX, USA", "type": "city"},
    "riverside": {"lat": 33.9806, "lng": -117.3755, "formatted": "Riverside, CA, USA", "type": "city"},
    "lexington": {"lat": 38.0406, "lng": -84.5037, "formatted": "Lexington, KY, USA", "type": "city"},
    "st. louis": {"lat": 38.6270, "lng": -90.1994, "formatted": "St. Louis, MO, USA", "type": "city"},
    "pittsburgh": {"lat": 40.4406, "lng": -79.9959, "formatted": "Pittsburgh, PA, USA", "type": "city"},
    "anchorage": {"lat": 61.2181, "lng": -149.9003, "formatted": "Anchorage, AK, USA", "type": "city"},
    "stockton": {"lat": 37.9577, "lng": -121.2908, "formatted": "Stockton, CA, USA", "type": "city"},
    "cincinnati": {"lat": 39.1031, "lng": -84.5120, "formatted": "Cincinnati, OH, USA", "type": "city"},
    "st. paul": {"lat": 44.9537, "lng": -93.0900, "formatted": "St. Paul, MN, USA", "type": "city"},
    "greensboro": {"lat": 36.0726, "lng": -79.7920, "formatted": "Greensboro, NC, USA", "type": "city"},
    "toledo": {"lat": 41.6639, "lng": -83.5552, "formatted": "Toledo, OH, USA", "type": "city"},
    "newark": {"lat": 40.7357, "lng": -74.1724, "formatted": "Newark, NJ, USA", "type": "city"},
    "plano": {"lat": 33.0198, "lng": -96.6989, "formatted": "Plano, TX, USA", "type": "city"},
    "henderson": {"lat": 36.0395, "lng": -114.9817, "formatted": "Henderson, NV, USA", "type": "city"},
    "orlando": {"lat": 28.5383, "lng": -81.3792, "formatted": "Orlando, FL, USA", "type": "city"},
    "chandler": {"lat": 33.3062, "lng": -111.8413, "formatted": "Chandler, AZ, USA", "type": "city"},
    "laredo": {"lat": 27.5306, "lng": -99.4803, "formatted": "Laredo, TX, USA", "type": "city"},
    "madison": {"lat": 43.0731, "lng": -89.4012, "formatted": "Madison, WI, USA", "type": "city"},
    "durham": {"lat": 35.9940, "lng": -78.8986, "formatted": "Durham, NC, USA", "type": "city"},
    "lubbock": {"lat": 33.5779, "lng": -101.8552, "formatted": "Lubbock, TX, USA", "type": "city"},
    "winston-salem": {"lat": 36.0999, "lng": -80.2442, "formatted": "Winston-Salem, NC, USA", "type": "city"},
    "garland": {"lat": 32.9126, "lng": -96.6389, "formatted": "Garland, TX, USA", "type": "city"},
    "glendale": {"lat": 33.5387, "lng": -112.1860, "formatted": "Glendale, AZ, USA", "type": "city"},
    "hialeah": {"lat": 25.8576, "lng": -80.2781, "formatted": "Hialeah, FL, USA", "type": "city"},
    "reno": {"lat": 39.5296, "lng": -119.8138, "formatted": "Reno, NV, USA", "type": "city"},
    "baton rouge": {"lat": 30.4515, "lng": -91.1871, "formatted": "Baton Rouge, LA, USA", "type": "city"},
    "irvine": {"lat": 33.6846, "lng": -117.8265, "formatted": "Irvine, CA, USA", "type": "city"},
    "chesapeake": {"lat": 36.7682, "lng": -76.2875, "formatted": "Chesapeake, VA, USA", "type": "city"},
    "scottsdale": {"lat": 33.4942, "lng": -111.9261, "formatted": "Scottsdale, AZ, USA", "type": "city"},
    "north las vegas": {"lat": 36.1989, "lng": -115.1175, "formatted": "North Las Vegas, NV, USA", "type": "city"},
    "fremont": {"lat": 37.5485, "lng": -121.9886, "formatted": "Fremont, CA, USA", "type": "city"},
    "gilbert": {"lat": 33.3528, "lng": -111.7890, "formatted": "Gilbert, AZ, USA", "type": "city"},
    "san bernardino": {"lat": 34.1083, "lng": -117.2898, "formatted": "San Bernardino, CA, USA", "type": "city"},
    "birmingham": {"lat": 33.5186, "lng": -86.8104, "formatted": "Birmingham, AL, USA", "type": "city"},
    "rochester": {"lat": 43.1566, "lng": -77.6088, "formatted": "Rochester, NY, USA", "type": "city"},
    "richmond": {"lat": 37.5407, "lng": -77.4360, "formatted": "Richmond, VA, USA", "type": "city"},
    "spokane": {"lat": 47.6588, "lng": -117.4260, "formatted": "Spokane, WA, USA", "type": "city"},
    "des moines": {"lat": 41.5868, "lng": -93.6250, "formatted": "Des Moines, IA, USA", "type": "city"},
    "montgomery": {"lat": 32.3668, "lng": -86.3000, "formatted": "Montgomery, AL, USA", "type": "city"},
    "modesto": {"lat": 37.6391, "lng": -120.9969, "formatted": "Modesto, CA, USA", "type": "city"},
    "fayetteville": {"lat": 36.0626, "lng": -94.1574, "formatted": "Fayetteville, AR, USA", "type": "city"},
    "tacoma": {"lat": 47.2529, "lng": -122.4443, "formatted": "Tacoma, WA, USA", "type": "city"},
    "akron": {"lat": 41.0814, "lng": -81.5190, "formatted": "Akron, OH, USA", "type": "city"},
    "yonkers": {"lat": 40.9312, "lng": -73.8988, "formatted": "Yonkers, NY, USA", "type": "city"},
    "little rock": {"lat": 34.7465, "lng": -92.2896, "formatted": "Little Rock, AR, USA", "type": "city"},
    "aurora, il": {"lat": 41.7606, "lng": -88.3201, "formatted": "Aurora, IL, USA", "type": "city"},
    "oxnard": {"lat": 34.1975, "lng": -119.1771, "formatted": "Oxnard, CA, USA", "type": "city"},
    "fontana": {"lat": 34.0922, "lng": -117.4350, "formatted": "Fontana, CA, USA", "type": "city"},
    "moreno valley": {"lat": 33.9425, "lng": -117.2297, "formatted": "Moreno Valley, CA, USA", "type": "city"},
    "glendale, ca": {"lat": 34.1425, "lng": -118.2551, "formatted": "Glendale, CA, USA", "type": "city"},
    "huntington beach": {"lat": 33.6595, "lng": -117.9988, "formatted": "Huntington Beach, CA, USA", "type": "city"},
    "salt lake city, ut": {"lat": 40.7608, "lng": -111.8910, "formatted": "Salt Lake City, UT, USA", "type": "city"},
    "grand rapids": {"lat": 42.9634, "lng": -85.6681, "formatted": "Grand Rapids, MI, USA", "type": "city"},
    "tallahassee": {"lat": 30.4518, "lng": -84.2807, "formatted": "Tallahassee, FL, USA", "type": "city"},
    "huntsville": {"lat": 34.7304, "lng": -86.5861, "formatted": "Huntsville, AL, USA", "type": "city"},
    "worcester": {"lat": 42.2626, "lng": -71.8023, "formatted": "Worcester, MA, USA", "type": "city"},
    "knoxville": {"lat": 35.9606, "lng": -83.9207, "formatted": "Knoxville, TN, USA", "type": "city"},
    "newport news": {"lat": 37.0871, "lng": -76.4730, "formatted": "Newport News, VA, USA", "type": "city"},
    "brownsville": {"lat": 25.9017, "lng": -97.4975, "formatted": "Brownsville, TX, USA", "type": "city"},
    "santa clarita": {"lat": 34.3917, "lng": -118.5426, "formatted": "Santa Clarita, CA, USA", "type": "city"},
    "providence": {"lat": 41.8240, "lng": -71.4128, "formatted": "Providence, RI, USA", "type": "city"},
    "garden grove": {"lat": 33.7739, "lng": -117.9414, "formatted": "Garden Grove, CA, USA", "type": "city"},
    "oceanside": {"lat": 33.1959, "lng": -117.3795, "formatted": "Oceanside, CA, USA", "type": "city"},
    "chattanooga": {"lat": 35.0456, "lng": -85.3097, "formatted": "Chattanooga, TN, USA", "type": "city"},
    "fort lauderdale": {"lat": 26.1224, "lng": -80.1373, "formatted": "Fort Lauderdale, FL, USA", "type": "city"},
    "rancho cucamonga": {"lat": 34.1064, "lng": -117.5931, "formatted": "Rancho Cucamonga, CA, USA", "type": "city"},
    "santa rosa": {"lat": 38.4404, "lng": -122.7141, "formatted": "Santa Rosa, CA, USA", "type": "city"},
    "port arthur": {"lat": 29.8849, "lng": -93.9399, "formatted": "Port Arthur, TX, USA", "type": "city"},
    "tempe": {"lat": 33.4255, "lng": -111.9400, "formatted": "Tempe, AZ, USA", "type": "city"},
    "cape coral": {"lat": 26.5629, "lng": -81.9495, "formatted": "Cape Coral, FL, USA", "type": "city"},
    "oxnard, ca": {"lat": 34.1975, "lng": -119.1771, "formatted": "Oxnard, CA, USA", "type": "city"},
    "sioux falls": {"lat": 43.5446, "lng": -96.7311, "formatted": "Sioux Falls, SD, USA", "type": "city"},
    "ontario": {"lat": 34.0633, "lng": -117.6509, "formatted": "Ontario, CA, USA", "type": "city"},
    "vancouver": {"lat": 49.2827, "lng": -123.1207, "formatted": "Vancouver, BC, Canada", "type": "city"},
    "montreal": {"lat": 45.5017, "lng": -73.5673, "formatted": "Montreal, QC, Canada", "type": "city"},
    "calgary": {"lat": 51.0447, "lng": -114.0719, "formatted": "Calgary, AB, Canada", "type": "city"},
    "edmonton": {"lat": 53.5461, "lng": -113.4938, "formatted": "Edmonton, AB, Canada", "type": "city"},
    "ottawa": {"lat": 45.4215, "lng": -75.6972, "formatted": "Ottawa, ON, Canada", "type": "city"},
    "mexico city": {"lat": 19.4326, "lng": -99.1332, "formatted": "Mexico City, Mexico", "type": "city"},
    "guadalajara": {"lat": 20.6597, "lng": -103.3496, "formatted": "Guadalajara, Mexico", "type": "city"},
    "monterrey": {"lat": 25.6866, "lng": -100.3161, "formatted": "Monterrey, Mexico", "type": "city"},
    "sao paulo": {"lat": -23.5505, "lng": -46.6333, "formatted": "São Paulo, Brazil", "type": "city"},
    "rio de janeiro": {"lat": -22.9068, "lng": -43.1729, "formatted": "Rio de Janeiro, Brazil", "type": "city"},
    "buenos aires": {"lat": -34.6037, "lng": -58.3816, "formatted": "Buenos Aires, Argentina", "type": "city"},
    "lima": {"lat": -12.0464, "lng": -77.0428, "formatted": "Lima, Peru", "type": "city"},
    "bogota": {"lat": 4.7110, "lng": -74.0721, "formatted": "Bogotá, Colombia", "type": "city"},
    "santiago": {"lat": -33.4489, "lng": -70.6693, "formatted": "Santiago, Chile", "type": "city"},
    "madrid": {"lat": 40.4168, "lng": -3.7038, "formatted": "Madrid, Spain", "type": "city"},
    "barcelona": {"lat": 41.3851, "lng": 2.1734, "formatted": "Barcelona, Spain", "type": "city"},
    "rome": {"lat": 41.9028, "lng": 12.4964, "formatted": "Rome, Italy", "type": "city"},
    "milan": {"lat": 45.4654, "lng": 9.1859, "formatted": "Milan, Italy", "type": "city"},
    "amsterdam": {"lat": 52.3676, "lng": 4.9041, "formatted": "Amsterdam, Netherlands", "type": "city"},
    "brussels": {"lat": 50.8503, "lng": 4.3517, "formatted": "Brussels, Belgium", "type": "city"},
    "vienna": {"lat": 48.2082, "lng": 16.3738, "formatted": "Vienna, Austria", "type": "city"},
    "zurich": {"lat": 47.3769, "lng": 8.5417, "formatted": "Zurich, Switzerland", "type": "city"},
    "stockholm": {"lat": 59.3293, "lng": 18.0686, "formatted": "Stockholm, Sweden", "type": "city"},
    "oslo": {"lat": 59.9139, "lng": 10.7522, "formatted": "Oslo, Norway", "type": "city"},
    "copenhagen": {"lat": 55.6761, "lng": 12.5683, "formatted": "Copenhagen, Denmark", "type": "city"},
    "helsinki": {"lat": 60.1699, "lng": 24.9384, "formatted": "Helsinki, Finland", "type": "city"},
    "warsaw": {"lat": 52.2297, "lng": 21.0122, "formatted": "Warsaw, Poland", "type": "city"},
    "prague": {"lat": 50.0755, "lng": 14.4378, "formatted": "Prague, Czech Republic", "type": "city"},
    "budapest": {"lat": 47.4979, "lng": 19.0402, "formatted": "Budapest, Hungary", "type": "city"},
    "bucharest": {"lat": 44.4268, "lng": 26.1025, "formatted": "Bucharest, Romania", "type": "city"},
    "athens": {"lat": 37.9838, "lng": 23.7275, "formatted": "Athens, Greece", "type": "city"},
    "istanbul": {"lat": 41.0082, "lng": 28.9784, "formatted": "Istanbul, Turkey", "type": "city"},
    "moscow": {"lat": 55.7558, "lng": 37.6173, "formatted": "Moscow, Russia", "type": "city"},
    "st. petersburg": {"lat": 59.9311, "lng": 30.3609, "formatted": "St. Petersburg, Russia", "type": "city"},
    "beijing": {"lat": 39.9042, "lng": 116.4074, "formatted": "Beijing, China", "type": "city"},
    "shanghai": {"lat": 31.2304, "lng": 121.4737, "formatted": "Shanghai, China", "type": "city"},
    "guangzhou": {"lat": 23.1291, "lng": 113.2644, "formatted": "Guangzhou, China", "type": "city"},
    "shenzhen": {"lat": 22.5431, "lng": 114.0579, "formatted": "Shenzhen, China", "type": "city"},
    "hong kong": {"lat": 22.3193, "lng": 114.1694, "formatted": "Hong Kong", "type": "city"},
    "seoul": {"lat": 37.5665, "lng": 126.9780, "formatted": "Seoul, South Korea", "type": "city"},
    "osaka": {"lat": 34.6937, "lng": 135.5023, "formatted": "Osaka, Japan", "type": "city"},
    "taipei": {"lat": 25.0330, "lng": 121.5654, "formatted": "Taipei, Taiwan", "type": "city"},
    "bangkok": {"lat": 13.7563, "lng": 100.5018, "formatted": "Bangkok, Thailand", "type": "city"},
    "jakarta": {"lat": -6.2088, "lng": 106.8456, "formatted": "Jakarta, Indonesia", "type": "city"},
    "kuala lumpur": {"lat": 3.1390, "lng": 101.6869, "formatted": "Kuala Lumpur, Malaysia", "type": "city"},
    "manila": {"lat": 14.5995, "lng": 120.9842, "formatted": "Manila, Philippines", "type": "city"},
    "mumbai": {"lat": 19.0760, "lng": 72.8777, "formatted": "Mumbai, India", "type": "city"},
    "delhi": {"lat": 28.7041, "lng": 77.1025, "formatted": "Delhi, India", "type": "city"},
    "new delhi": {"lat": 28.6139, "lng": 77.2090, "formatted": "New Delhi, India", "type": "city"},
    "bangalore": {"lat": 12.9716, "lng": 77.5946, "formatted": "Bangalore, India", "type": "city"},
    "hyderabad": {"lat": 17.3850, "lng": 78.4867, "formatted": "Hyderabad, India", "type": "city"},
    "chennai": {"lat": 13.0827, "lng": 80.2707, "formatted": "Chennai, India", "type": "city"},
    "kolkata": {"lat": 22.5726, "lng": 88.3639, "formatted": "Kolkata, India", "type": "city"},
    "karachi": {"lat": 24.8607, "lng": 67.0011, "formatted": "Karachi, Pakistan", "type": "city"},
    "lahore": {"lat": 31.5204, "lng": 74.3587, "formatted": "Lahore, Pakistan", "type": "city"},
    "dhaka": {"lat": 23.8103, "lng": 90.4125, "formatted": "Dhaka, Bangladesh", "type": "city"},
    "tehran": {"lat": 35.6892, "lng": 51.3890, "formatted": "Tehran, Iran", "type": "city"},
    "riyadh": {"lat": 24.7136, "lng": 46.6753, "formatted": "Riyadh, Saudi Arabia", "type": "city"},
    "cairo": {"lat": 30.0444, "lng": 31.2357, "formatted": "Cairo, Egypt", "type": "city"},
    "lagos": {"lat": 6.5244, "lng": 3.3792, "formatted": "Lagos, Nigeria", "type": "city"},
    "nairobi": {"lat": -1.2921, "lng": 36.8219, "formatted": "Nairobi, Kenya", "type": "city"},
    "johannesburg": {"lat": -26.2041, "lng": 28.0473, "formatted": "Johannesburg, South Africa", "type": "city"},
    "cape town": {"lat": -33.9249, "lng": 18.4241, "formatted": "Cape Town, South Africa", "type": "city"},
    "melbourne": {"lat": -37.8136, "lng": 144.9631, "formatted": "Melbourne, VIC, Australia", "type": "city"},
    "brisbane": {"lat": -27.4698, "lng": 153.0251, "formatted": "Brisbane, QLD, Australia", "type": "city"},
    "perth": {"lat": -31.9505, "lng": 115.8605, "formatted": "Perth, WA, Australia", "type": "city"},
    "auckland": {"lat": -36.8485, "lng": 174.7633, "formatted": "Auckland, New Zealand", "type": "city"},
    "wellington": {"lat": -41.2865, "lng": 174.7762, "formatted": "Wellington, New Zealand", "type": "city"},
}


def handler(event):
    if not isinstance(event, dict):
        return {"ok": False, "error": "event must be an object"}

    address = event.get("address")
    if not address:
        return {"ok": False, "error": "address is required"}

    if not isinstance(address, str):
        return {"ok": False, "error": "address must be a string"}

    address = address.strip()
    if not address:
        return {"ok": False, "error": "address cannot be empty"}

    try:
        # Try exact match first (case-insensitive)
        key = address.lower()
        if key in CITY_DATABASE:
            city = CITY_DATABASE[key]
            return {
                "ok": True,
                "result": {
                    "address": address,
                    "lat": city["lat"],
                    "lng": city["lng"],
                    "formatted_address": city["formatted"],
                    "confidence": 0.95,
                    "type": city["type"]
                }
            }

        # Try partial match
        for db_key, city in CITY_DATABASE.items():
            if db_key in key or key in db_key:
                return {
                    "ok": True,
                    "result": {
                        "address": address,
                        "lat": city["lat"],
                        "lng": city["lng"],
                        "formatted_address": city["formatted"],
                        "confidence": 0.75,
                        "type": city["type"]
                    }
                }

        # Generate a deterministic mock location based on address hash
        h = 0
        for c in address:
            h = (h * 31 + ord(c)) & 0xFFFFFFFF

        lat = ((h % 18000) - 9000) / 100.0  # -90 to 90
        lng = (((h >> 8) % 36000) - 18000) / 100.0  # -180 to 180

        return {
            "ok": True,
            "result": {
                "address": address,
                "lat": round(lat, 6),
                "lng": round(lng, 6),
                "formatted_address": address,
                "confidence": 0.3,
                "type": "unknown"
            }
        }

    except Exception as e:
        return {"ok": False, "error": f"geocoding failed: {str(e)}"}
