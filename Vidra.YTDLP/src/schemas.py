from pydantic import BaseModel
from typing import List, Optional

class VideoInfo(BaseModel):
    id: str
    title: str
    duration: float
    uploader: str
    view_count: Optional[int] = None
    like_count: Optional[int] = None
    thumbnail: str
    filesize: Optional[int] = None
    url: str

class DownloadStatus(BaseModel):
    status: str
    percent: Optional[float] = None
    filename: Optional[str] = None
    downloaded_bytes: Optional[int] = None
    total_bytes: Optional[int] = None
    error: Optional[str] = None
    thumbnail: Optional[str] = None
    thumbnail_error: Optional[str] = None

class ThumbnailResponse(BaseModel):
    thumbnail_url: str

class SizeResponse(BaseModel):
    filesize_bytes: Optional[int]

class DownloadIdResponse(BaseModel):
    download_id: str

class Format(BaseModel):
    format_id: str
    ext: Optional[str]
    format_note: Optional[str]
    filesize: Optional[int]
    resolution: Optional[str]
    fps: Optional[float]
    vcodec: Optional[str]
    acodec: Optional[str]

class FormatsResponse(BaseModel):
    formats: List[Format]

class CombinedVideoInfo(BaseModel):
    id: str
    title: str
    duration: float
    uploader: str
    view_count: Optional[int] = None
    like_count: Optional[int] = None
    thumbnail: str
    filesize: Optional[int] = None
    url: str
    formats: List[Format]
