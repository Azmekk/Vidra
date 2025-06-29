from fastapi import FastAPI, Query, HTTPException
from pydantic import BaseModel
from yt_dlp import YoutubeDL
import os
import json
import requests

app = FastAPI()

YDL_OPTS = {
    "quiet": True,
    "no_warnings": True,
    "skip_download": True,
    "format": "best"
}


class VideoInfo(BaseModel):
    id: str
    title: str
    duration: float
    uploader: str
    view_count: int
    like_count: int | None
    thumbnail: str
    filesize: int | None
    url: str


def get_info(url: str):
    with YoutubeDL(YDL_OPTS) as ydl:
        info = ydl.extract_info(url, download=False)
        return info


@app.get("/metadata", response_model=VideoInfo)
def get_metadata(url: str = Query(..., description="URL of the video")):
    info = get_info(url)
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


@app.get("/thumbnail")
def get_thumbnail(url: str):
    info = get_info(url)

    if info is None or "thumbnail" not in info:
        raise HTTPException(status_code=404, detail="Thumbnail not found")
    
    return {"thumbnail_url": info.get("thumbnail")}


@app.get("/size")
def get_size(url: str):
    info = get_info(url)
    if info is None:
        raise HTTPException(status_code=404, detail="Video info not found")

    size = info.get("filesize") or info.get("filesize_approx")
    return {"filesize_bytes": size}


@app.get("/download")
def download_video(url: str):
    outtmpl = "downloads/%(title)s.%(ext)s"
    ydl_opts = {
        "outtmpl": outtmpl,
        "quiet": True,
        "format": "best",
    }

    os.makedirs("downloads", exist_ok=True)

    with YoutubeDL(ydl_opts) as ydl:
        info = ydl.extract_info(url, download=True)
        filename = ydl.prepare_filename(info)
    return {"downloaded_file": filename}
