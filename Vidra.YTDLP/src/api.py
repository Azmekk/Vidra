from fastapi import APIRouter, Query, HTTPException, WebSocket, WebSocketDisconnect
import uuid
import threading
import asyncio
from typing import List, Optional
from .schemas import VideoInfo, DownloadStatus, ThumbnailResponse, SizeResponse, DownloadIdResponse, FormatsItem, FormatsResponse
from pydantic import BaseModel
from .utils import get_info, download_worker, download_progresses

router = APIRouter()

YDL_OPTS = {
    "quiet": True,
    "no_warnings": True,
    "skip_download": True,
    "format": "best"
}

@router.get("/metadata", response_model=VideoInfo)
def get_metadata(url: str = Query(..., description="URL of the video")):
    info = get_info(url, YDL_OPTS)
    if info is None:
        raise HTTPException(status_code=404, detail="Video info not found")
    return {
        "id": info.get("id"),
        "title": info.get("title"),
        "duration": info.get("duration"),
        "uploader": info.get("uploader"),
        "view_count": info.get("view_count"),
        "like_count": info.get("like_count"),
        "thumbnail": info.get("thumbnail"),
        "filesize": info.get("filesize") or info.get("filesize_approx"),
        "url": info.get("webpage_url"),
    }

@router.get("/thumbnail", response_model=ThumbnailResponse)
def get_thumbnail(url: str):
    info = get_info(url, YDL_OPTS)
    if info is None or "thumbnail" not in info:
        raise HTTPException(status_code=404, detail="Thumbnail not found")
    return {"thumbnail_url": info.get("thumbnail")}

@router.get("/size", response_model=SizeResponse)
def get_size(url: str):
    info = get_info(url, YDL_OPTS)
    if info is None:
        raise HTTPException(status_code=404, detail="Video info not found")
    size = info.get("filesize") or info.get("filesize_approx")
    return {"filesize_bytes": size}

@router.get("/download", response_model=DownloadIdResponse)
def download_video(
    url: str,
    name: str = Query(None, description="Optional name for the downloaded file"),
    include_thumbnail: bool = Query(False, description="If true, also download the thumbnail as a PNG")
):
    guid = str(uuid.uuid4())
    download_progresses[guid] = {'status': 'queued'}
    thread = threading.Thread(target=download_worker, args=(url, guid, name, include_thumbnail), daemon=True)
    thread.start()
    return {"download_id": guid}

@router.get("/download_status", response_model=DownloadStatus)
def download_status(download_id: str):
    status = download_progresses.get(download_id)
    if not status:
        raise HTTPException(status_code=404, detail="Download ID not found")
    return status

@router.websocket("/ws/download_status")
async def websocket_download_status(websocket: WebSocket):
    await websocket.accept()
    last_snapshot = None
    try:
        while True:
            snapshot = download_progresses.copy()
            if snapshot != last_snapshot:
                await websocket.send_json(snapshot)
                last_snapshot = snapshot
            await asyncio.sleep(0.25)
    except WebSocketDisconnect:
        pass

@router.get("/formats", response_model=FormatsResponse)
def get_formats(url: str = Query(..., description="URL of the video")):
    info = get_info(url, YDL_OPTS)
    if info is None or "formats" not in info:
        raise HTTPException(status_code=404, detail="Formats not found")
    formats = [
        {"format_id": f["format_id"], "ext": f.get("ext"), "format_note": f.get("format_note"), "filesize": f.get("filesize"), "resolution": f.get("resolution"), "fps": f.get("fps"), "vcodec": f.get("vcodec"), "acodec": f.get("acodec")} 
        for f in info["formats"]
    ]
    return {"formats": formats}
