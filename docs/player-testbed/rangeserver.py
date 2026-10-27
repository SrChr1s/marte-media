#!/usr/bin/env python3
# Tiny static server with byte-range support (what googlevideo does), for the player tests.
import os, re, sys
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
ROOT = os.path.dirname(os.path.abspath(__file__))
class H(BaseHTTPRequestHandler):
    def log_message(self, *a): pass
    def do_GET(self):
        name = self.path.split('?')[0].lstrip('/')
        path = os.path.join(ROOT, os.path.basename(name))
        if not os.path.isfile(path): self.send_error(404); return
        size = os.path.getsize(path)
        rng = self.headers.get('Range')
        start, end = 0, size - 1
        code = 200
        if rng:
            m = re.match(r'bytes=(\d*)-(\d*)', rng)
            if m:
                if m.group(1): start = int(m.group(1))
                if m.group(2): end = min(int(m.group(2)), size - 1)
                if not m.group(1) and m.group(2): start = max(0, size - int(m.group(2))); end = size - 1
                if start >= size: self.send_response(416); self.send_header('Content-Range', f'bytes */{size}'); self.end_headers(); return
                code = 206
        self.send_response(code)
        self.send_header('Content-Type', 'video/webm' if 'video' in name else 'audio/webm')
        self.send_header('Accept-Ranges', 'bytes')
        self.send_header('Content-Length', str(end - start + 1))
        if code == 206: self.send_header('Content-Range', f'bytes {start}-{end}/{size}')
        self.end_headers()
        if self.command == 'HEAD': return
        with open(path, 'rb') as f:
            f.seek(start); left = end - start + 1
            while left > 0:
                b = f.read(min(65536, left))
                if not b: break
                try: self.wfile.write(b)
                except Exception: return
                left -= len(b)
ThreadingHTTPServer(('127.0.0.1', int(sys.argv[1])), H).serve_forever()
