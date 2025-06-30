from fastapi import APIRouter, Query, HTTPException, WebSocket, WebSocketDisconnect, Body
import uuid
import threading
import asyncio
from typing import List, Optional
from .schemas import VideoInfo, DownloadStatus, ThumbnailResponse, SizeResponse, DownloadIdResponse, Format, FormatsResponse, CombinedVideoInfo
from pydantic import BaseModel
from .utils import get_info, download_worker, download_progresses
import logging

router = APIRouter()

YDL_OPTS = {
    "quiet": True,
    "no_warnings": True,
    "skip_download": True,
    "format": "best"
}

@router.post("/metadata", response_model=VideoInfo)
def get_metadata(url: str = Body(..., embed=True, description="URL of the video")):
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

@router.post("/thumbnail", response_model=ThumbnailResponse)
def get_thumbnail(url: str = Body(..., embed=True, description="URL of the video")):
    info = get_info(url, YDL_OPTS)
    if info is None or "thumbnail" not in info:
        raise HTTPException(status_code=404, detail="Thumbnail not found")
    return {"thumbnail_url": info.get("thumbnail")}

@router.post("/size", response_model=SizeResponse)
def get_size(url: str = Body(..., embed=True, description="URL of the video")):
    info = get_info(url, YDL_OPTS)
    if info is None:
        raise HTTPException(status_code=404, detail="Video info not found")
    size = info.get("filesize") or info.get("filesize_approx")
    return {"filesize_bytes": size}

@router.post("/download", response_model=DownloadIdResponse)
def download_video(
    url: str = Body(..., embed=True, description="URL of the video to download"),
    name: str = Body(None, embed=True, description="Optional name for the downloaded file"),
    include_thumbnail: bool = Body(False, embed=True, description="If true, also download the thumbnail as a PNG"),
    format_id: Optional[str] = Body(None, embed=True, description="Optional format ID to download specific format"),
    audio_format_id: Optional[str] = Body(None, embed=True, description="Optional audio format ID to download specific audio format"),
    video_format_id: Optional[str] = Body(None, embed=True, description="Optional video format ID to download specific video format")
):
    guid = str(uuid.uuid4())
    download_progresses[guid] = {'status': 'queued'}
    thread = threading.Thread(target=download_worker, args=(url, guid, name, include_thumbnail, audio_format_id, video_format_id), daemon=True)
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

@router.post("/formats", response_model=FormatsResponse)
def get_formats(url: str = Body(..., embed=True, description="URL of the video")):
    info = get_info(url, YDL_OPTS)
    if info is None or "formats" not in info:
        raise HTTPException(status_code=404, detail="Formats not found")
    formats = [
        {
            "format_id": f["format_id"],
            "ext": f.get("ext"),
            "format_note": f.get("format_note"),
            "filesize": f.get("filesize"),
            "resolution": f.get("resolution"),
            "fps": f.get("fps"),
            "vcodec": f.get("vcodec"),
            "acodec": f.get("acodec")
        }
        for f in info["formats"]
    ]
    return {"formats": formats}

@router.post("/combined_video_info", response_model=CombinedVideoInfo)
def get_combined_video_info(url: str = Body(..., embed=True, description="URL of the video")):
    info = get_info(url, YDL_OPTS)
    if info is None:
        raise HTTPException(status_code=404, detail="Video info not found")
    formats = []
    if "formats" in info:
        formats = [
            {
                "format_id": f["format_id"],
                "ext": f.get("ext"),
                "format_note": f.get("format_note"),
                "filesize": f.get("filesize"),
                "resolution": f.get("resolution"),
                "fps": f.get("fps"),
                "vcodec": f.get("vcodec"),
                "acodec": f.get("acodec")
            }
            for f in info["formats"]
        ]
        logging.info(f"Combined video info for URL {url}: {info}")
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
        "formats": formats
    }

