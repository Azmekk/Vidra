import os
import threading
from typing import Dict, Optional
from yt_dlp import YoutubeDL
from .schemas import DownloadStatus

download_progresses: Dict[str, dict] = {}

YDL_OUTTMPL = "downloads/%(title)s.%(ext)s"

def get_info(url: str, ydl_opts):
    with YoutubeDL(ydl_opts) as ydl:
        info = ydl.extract_info(url, download=False)
        return info

def download_worker(url: str, guid: str, name: str = "", include_thumbnail: bool = False, audio_format_id: Optional[str] = None, video_format_id: Optional[str] = None):
    from .utils import download_progresses  # for circular import safety if needed
    def progress_hook(d):
        if d['status'] == 'downloading':
            percent = d.get('downloaded_bytes', 0) / max(d.get('total_bytes', 1), 1) * 100 if d.get('total_bytes') else None
            download_progresses[guid] = DownloadStatus(
                status='downloading',
                percent=percent,
                filename=d.get('filename'),
                downloaded_bytes=d.get('downloaded_bytes'),
                total_bytes=d.get('total_bytes')
            ).model_dump()
        elif d['status'] == 'finished':
            download_progresses[guid] = DownloadStatus(
                status='finished',
                filename=d.get('filename')
            ).model_dump()
            # Remove entry after short delay
            threading.Timer(60, lambda: download_progresses.pop(guid, None)).start()
        elif d['status'] == 'error':
            download_progresses[guid] = DownloadStatus(
                status='error',
                error=d.get('error')
            ).model_dump()
            # Remove entry after short delay
            threading.Timer(60, lambda: download_progresses.pop(guid, None)).start()

    ydl_opts = {
        "outtmpl": YDL_OUTTMPL if not name else f"downloads/{name}.%(ext)s",
        "quiet": True,
        "progress_hooks": [progress_hook],
    }
    # Always use one of the following for the format string:
    # video_format_id+bestaudio, best+audio_format_id, or video_format_id+audio_format_id
    if video_format_id and audio_format_id:
        ydl_opts["format"] = f"{video_format_id}+{audio_format_id}"
    elif video_format_id:
        ydl_opts["format"] = f"{video_format_id}+bestaudio"
    elif audio_format_id:
        ydl_opts["format"] = f"bestvideo+{audio_format_id}"
    else:
        ydl_opts["format"] = "bestvideo+bestaudio"
    if include_thumbnail:
        ydl_opts["writethumbnail"] = True
        ydl_opts["convert-thumbnails"] = "png"
    os.makedirs("downloads", exist_ok=True)
    try:
        with YoutubeDL(ydl_opts) as ydl:
            info = ydl.extract_info(url, download=True)
            if include_thumbnail and info and ("thumbnails" in info):
                # Find the PNG thumbnail file
                base_name = name if name else info.get("title", guid)
                thumb_path = os.path.join("downloads", f"{base_name}.png")
                if os.path.exists(thumb_path):
                    # Update using DownloadStatus to add thumbnail
                    current = download_progresses.get(guid, {})
                    current['thumbnail'] = thumb_path
                    download_progresses[guid] = DownloadStatus(**current).dict()
    except Exception as e:
        download_progresses[guid] = DownloadStatus(status='error', error=str(e)).dict()
        threading.Timer(60, lambda: download_progresses.pop(guid, None)).start()
