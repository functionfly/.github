def _parse(h):
    h=h.strip().lstrip("#")
    if len(h)==3: h="".join(c*2 for c in h)
    return int(h[0:2],16),int(h[2:4],16),int(h[4:6],16)
def _rh(r,g,b):
    r_,g_,b_=r/255,g/255,b/255; mx,mn=max(r_,g_,b_),min(r_,g_,b_); d=mx-mn
    l=(mx+mn)/2; s=0 if d==0 else d/(1-abs(2*l-1))
    if d==0: h=0
    elif mx==r_: h=60*(((g_-b_)/d)%6)
    elif mx==g_: h=60*((b_-r_)/d+2)
    else: h=60*((r_-g_)/d+4)
    return h,s*100,l*100
def _hr(h,s,l):
    s_,l_=s/100,l/100; c=(1-abs(2*l_-1))*s_; x=c*(1-abs((h/60)%2-1)); m=l_-c/2
    if 0<=h<60: r_,g_,b_=c,x,0
    elif 60<=h<120: r_,g_,b_=x,c,0
    elif 120<=h<180: r_,g_,b_=0,c,x
    elif 180<=h<240: r_,g_,b_=0,x,c
    elif 240<=h<300: r_,g_,b_=x,0,c
    else: r_,g_,b_=c,0,x
    return f"#{round((r_+m)*255):02X}{round((g_+m)*255):02X}{round((b_+m)*255):02X}"

def handler(event):
    color = event.get("color") if isinstance(event, dict) else None
    scheme = event.get("scheme", "complementary")
    if not color:
        return {"ok": False, "error": "color is required"}
    SCHEMES = ["complementary", "analogous", "triadic", "tetradic", "split-complementary", "monochromatic"]
    if scheme not in SCHEMES:
        return {"ok": False, "error": f"scheme must be one of: {', '.join(SCHEMES)}"}
    try:
        r,g,b=_parse(str(color)); h,s,l=_rh(r,g,b)
        base = f"#{r:02X}{g:02X}{b:02X}"
        if scheme == "complementary":
            palette = [base, _hr((h+180)%360,s,l)]
        elif scheme == "analogous":
            palette = [_hr((h-30)%360,s,l), base, _hr((h+30)%360,s,l)]
        elif scheme == "triadic":
            palette = [base, _hr((h+120)%360,s,l), _hr((h+240)%360,s,l)]
        elif scheme == "tetradic":
            palette = [base, _hr((h+90)%360,s,l), _hr((h+180)%360,s,l), _hr((h+270)%360,s,l)]
        elif scheme == "split-complementary":
            palette = [base, _hr((h+150)%360,s,l), _hr((h+210)%360,s,l)]
        elif scheme == "monochromatic":
            steps = [0,20,40,60,80]
            palette = [_hr(h,s,lt) for lt in steps]
        return {"ok": True, "result": palette, "scheme": scheme, "base": base}
    except Exception as e:
        return {"ok": False, "error": str(e)}
