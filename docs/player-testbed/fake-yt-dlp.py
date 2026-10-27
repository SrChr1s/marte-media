#!/usr/bin/env python3
# Fake yt-dlp for the player tests: searches, formats pointing at the local range server, a playlist.
import json, sys, time, urllib.parse
PORT = 18765
args = sys.argv[1:]
if "--version" in args:
    print("2099.01.01"); sys.exit(0)
target = args[-1]
EXP = int(time.time()) + 6 * 3600

def thumb(n, hue):
    svg = (f"<svg xmlns='http://www.w3.org/2000/svg' width='640' height='360'><defs><linearGradient id='g' x1='0' y1='0' x2='0' y2='1'>"
           f"<stop offset='0' stop-color='hsl({hue},55%,22%)'/><stop offset='1' stop-color='hsl({(hue+25)%360},75%,50%)'/></linearGradient></defs>"
           f"<rect width='640' height='360' fill='url(#g)'/><circle cx='{140+n*40}' cy='120' r='46' fill='hsl({(hue+50)%360},90%,82%)'/>"
           f"<path d='M0 280 Q120 200 240 250 T480 240 T640 270 V360 H0Z' fill='hsl({hue},45%,12%)'/></svg>")
    return "data:image/svg+xml;utf8," + urllib.parse.quote(svg, safe="/:=,' ")

VIDEO_TITLES = ["Cómo se formó Valles Marineris", "Olympus Mons: el volcán más alto", "Perseverance: la primera muestra de roca",
                "Un día en Marte, hora por hora", "Por qué Marte es rojo", "Misión InSight: escuchando marsmotos"]
CHANNELS = ["Canal Cosmos", "Astro Lab", "Cuaderno de Órbita", "Canal Cosmos", "Sismos y Planetas", "Astro Lab"]
SONGS = ["Órbita lenta", "Polvo rojo", "Noche en Marte", "Cráter", "Marea de arena", "Satélite"]

def vid(i): return "mrtvid%05d" % i
def sng(i): return "mrtsng%05d" % i

if target.startswith("ytsearch"):
    n = int(target[len("ytsearch"):].split(":")[0])
    entries = [{"_type": "url", "ie_key": "Youtube", "id": vid(i), "title": VIDEO_TITLES[i % 6] + ("" if i < 6 else f" (parte {i//6+1})"),
                "channel": CHANNELS[i % 6], "uploader": CHANNELS[i % 6], "duration": 60 + i * 37, "view_count": 12000 * (i + 1) * (i + 3),
                "thumbnails": [{"url": thumb(i, (i * 47) % 360)}], "url": f"https://www.youtube.com/watch?v={vid(i)}"} for i in range(min(n, 12))]
    print(json.dumps({"_type": "playlist", "entries": entries})); sys.exit(0)

if target.startswith("https://music.youtube.com/search"):
    entries = []
    for i in range(8):
        entries.append({"_type": "url", "ie_key": "Youtube", "id": sng(i), "title": SONGS[i % 6] + ("" if i < 6 else " (versión)"), "url": f"https://music.youtube.com/watch?v={sng(i)}"})
        if i in (1, 4): entries.append({"_type": "url", "ie_key": "YoutubeTab", "id": f"MPREb_album{i}", "title": "Álbum", "url": f"https://music.youtube.com/browse/MPREb_album{i}"})
    print(json.dumps({"_type": "playlist", "entries": entries})); sys.exit(0)

if "playlist" in target and "--flat-playlist" in args:
    entries = [{"_type": "url", "ie_key": "Youtube", "id": sng(i), "title": SONGS[i], "channel": "Artista Demo", "duration": 60, "url": f"https://www.youtube.com/watch?v={sng(i)}"} for i in range(5)]
    print(json.dumps({"_type": "playlist", "id": "PLdemo", "title": "Lista de prueba", "entries": entries})); sys.exit(0)

if "watch?v=" in target:
    vid_id = target.split("v=")[1]
    base = f"http://127.0.0.1:{PORT}"
    hdr = {"User-Agent": "Mozilla/5.0 (X11; Linux x86_64) FakeYtDlp", "Accept-Language": "en-us,en;q=0.5"}
    f = lambda **k: {**k, "url": f"{base}/{k.pop('file')}?expire={EXP}&itag={k['format_id']}", "http_headers": hdr} if False else None
    def fmt(format_id, file, **k):
        return {"format_id": format_id, "url": f"{base}/{file}?expire={EXP}&itag={format_id}", "http_headers": hdr, "protocol": "https", **k}
    formats = [
        fmt("251", "audio.webm", ext="webm", vcodec="none", acodec="opus", abr=128, audio_channels=2, format_note="medium"),
        fmt("251-drc", "audio.webm", ext="webm", vcodec="none", acodec="opus", abr=128, audio_channels=2, format_note="medium, DRC"),
        fmt("247", "video_720.webm", ext="webm", vcodec="vp9", acodec="none", height=720, width=1280, fps=30),
        fmt("243", "video_360.webm", ext="webm", vcodec="vp9", acodec="none", height=360, width=640, fps=30),
        fmt("136", "video_720.webm", ext="mp4", vcodec="avc1.4d401f", acodec="none", height=720, width=1280, fps=30),  # H.264: must not be picked on this webview
    ]
    is_song = vid_id.startswith("mrtsng")
    i = int(vid_id[-5:])
    out = {"id": vid_id, "title": (SONGS[i % 6] if is_song else VIDEO_TITLES[i % 6]), "channel": ("Artista Demo" if is_song else CHANNELS[i % 6]),
           "uploader": "x", "duration": 90, "thumbnail": thumb(i, (i * 47) % 360), "live_status": "not_live", "formats": formats}
    if is_song: out.update({"artist": "Artista Demo", "track": SONGS[i % 6]})
    print(json.dumps(out)); sys.exit(0)

sys.stderr.write("ERROR: fake yt-dlp does not know: " + target + "\n"); sys.exit(1)
