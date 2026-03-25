def _p(h):
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
    return round((r_+m)*255),round((g_+m)*255),round((b_+m)*255)

def handler(event):
    color = event.get("color") if isinstance(event, dict) else None
    amount = float(event.get("amount", 20))
    if not color:
        return {"ok": False, "error": "color is required"}
    try:
        r,g,b=_p(str(color)); h,s,l=_rh(r,g,b)
        s2=min(100,s+amount); nr,ng,nb=_hr(h,s2,l)
        return {"ok": True, "result": f"#{nr:02X}{ng:02X}{nb:02X}", "original_saturation": round(s,2), "new_saturation": round(s2,2)}
    except Exception as e:
        return {"ok": False, "error": str(e)}
