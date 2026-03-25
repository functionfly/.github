SHORTCODES = {
    ":smile:":"😄",":laughing:":"😆",":blush:":"😊",":smiley:":"😃",":relaxed:":"☺️",":smirk:":"😏",":heart_eyes:":"😍",":kissing_heart:":"😘",":wink:":"😉",":stuck_out_tongue:":"😛",":grinning:":"😀",":thumbsup:":"👍",":thumbsdown:":"👎",":clap:":"👏",":wave:":"👋",":pray:":"🙏",":fire:":"🔥",":100:":"💯",":tada:":"🎉",":heart:":"❤️",":broken_heart:":"💔",":star:":"⭐",":star2:":"🌟",":sparkles:":"✨",":zap:":"⚡",":sunny:":"☀️",":cloud:":"☁️",":snowflake:":"❄️",":umbrella:":"☔",":earth_americas:":"🌎",":earth_africa:":"🌍",":earth_asia:":"🌏",":rocket:":"🚀",":car:":"🚗",":house:":"🏠",":coffee:":"☕",":pizza:":"🍕",":hamburger:":"🍔",":beer:":"🍺",":wine_glass:":"🍷",":trophy:":"🏆",":soccer:":"⚽",":basketball:":"🏀",":football:":"🏈",":music:":"🎵",":headphones:":"🎧",":book:":"📚",":phone:":"📱",":computer:":"💻",":lock:":"🔒",":key:":"🔑",":eyes:":"👀",":white_check_mark:":"✅",":x:":"❌",":warning:":"⚠️",":info:":"ℹ️",":question:":"❓",":exclamation:":"❗",":arrow_right:":"➡️",":back:":"🔙",":cat:":"🐱",":dog:":"🐶",":penguin:":"🐧",":panda_face:":"🐼",":monkey:":"🐒",":octocat:":"🐙",":cry:":"😢",":sob:":"😭",":angry:":"😠",":rage:":"😡",":sweat_smile:":"😅",":joy:":"😂",":rofl:":"🤣",":thinking:":"🤔",":shrug:":"🤷",":facepalm:":"🤦",":ok_hand:":"👌",":point_up:":"☝️",":v:":"✌️",":crossed_fingers:":"🤞",":muscle:":"💪",":running:":"🏃",":sleeping:":"😴",":sick:":"🤒",":poop:":"💩",":ghost:":"👻",":alien:":"👽",":robot:":"🤖",":rainbow:":"🌈",":four_leaf_clover:":"🍀",":christmas_tree:":"🎄",":birthday:":"🎂",":gift:":"🎁",
}

REVERSE = {v: k for k, v in SHORTCODES.items()}

import re
EMOJI_RE = re.compile("["
    "\U0001F600-\U0001F64F"
    "\U0001F300-\U0001F5FF"
    "\U0001F680-\U0001F6FF"
    "\U0001F900-\U0001F9FF"
    "\U0001FA70-\U0001FAFF"
    "\U00002702-\U000027B0"
    "]+", flags=re.UNICODE)


def handler(event):
    text = event.get("text") if isinstance(event, dict) else None
    direction = event.get("direction", "to_emoji")
    if not text:
        return {"ok": False, "error": "text is required"}
    try:
        t = str(text)
        if direction == "to_emoji":
            result = re.sub(r':[a-z0-9_+\-]+:', lambda m: SHORTCODES.get(m.group(0), m.group(0)), t)
        else:
            result = EMOJI_RE.sub(lambda m: REVERSE.get(m.group(0), m.group(0)), t)
        return {"ok": True, "result": result, "converted": result}
    except Exception as e:
        return {"ok": False, "error": str(e)}
