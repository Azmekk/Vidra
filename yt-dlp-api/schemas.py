from pydantic import BaseModel
from typing import Optional

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

class DownloadStatus(BaseModel):
    status: str
    percent: Optional[float] = None
    filename: Optional[str] = None
    downloaded_bytes: Optional[int] = None
    total_bytes: Optional[int] = None
    error: Optional[str] = None
    thumbnail: Optional[str] = None
    thumbnail_error: Optional[str] = None
