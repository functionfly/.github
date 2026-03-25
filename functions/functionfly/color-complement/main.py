def _parse(h):
    h = h.strip().lstrip("#")
    if len(h) == 3: h = "".join(c*2 for c in h)
    return int(h[0:2],16), int(h[2:4],16), int(h[4:6],16)

def _rgb_to_hsl(r,g,b):
    r_,g_,b_ = r/255, g/255, b/255
    cmax,cmin = max(r_,g_,b_), min(r_,g_,b_)
    delta = cmax - cmin
    l = (cmax+cmin)/2
    s = 0 if delta==0 else delta/(1-abs(2*l-1))
    if delta==0: h=0
    elif cmax==r_: h=60*(((g_-b_)/delta)%6)
    elif cmax==g_: h=60*((b_-r_)/delta+2)
    else: h=60*((r_-g_)/delta+4)
    return h, s*100, l*100

def _hsl_to_rgb(h,s,l):
    s_,l_ = s/100, l/100
    c=(1-abs(2*l_-1))*s_; x=c*(1-abs((h/60)%2-1)); m=l_-c/2
    if 0<=h<60: r_,g_,b_=c,x,0
    elif 60<=h<120: r_,g_,b_=x,c,0
    elif 120<=h<180: r_,g_,b_=0,c,x
    elif 180<=h<240: r_,g_,b_=0,x,c
    elif 240<=h<300: r_,g_,b_=x,0,c
    else: r_,g_,b_=c,0,x
    return round((r_+m)*255), round((g_+m)*255), round((b_+m)*255)

def handler(event):
    color = event.get("color") if isinstance(event, dict) else None
    if not color:
        return {"ok": False, "error": "color is required"}
    try:
        r,g,b = _parse(str(color))
        h,s,l = _rgb_to_hsl(r,g,b)
        cr,cg,cb = _hsl_to_rgb((h+180)%360, s, l)
        return {"ok": True, "result": f"#{cr:02X}{cg:02X}{cb:02X}", "r": cr, "g": cg, "b": cb}
    except Exception as e:
        return {"ok": False, "error": str(e)}
