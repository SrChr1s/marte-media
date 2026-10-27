# Third-party software

The installers include these programs, run as separate processes (not linked into the app):

| Program | License | Source |
|---|---|---|
| yt-dlp | The Unlicense (public domain) | https://github.com/yt-dlp/yt-dlp |
| FFmpeg (LGPL build) | LGPL v2.1+ | https://ffmpeg.org/ · builds from https://github.com/BtbN/FFmpeg-Builds |
| Deno | MIT | https://github.com/denoland/deno (JavaScript runtime yt-dlp uses to solve YouTube's challenges) |

The FFmpeg license text ships next to the binary (`FFMPEG-LICENSE.txt`). The app also downloads newer
yt-dlp releases into the user's config directory when the user clicks "Actualizar yt-dlp".
