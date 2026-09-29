export const WsEventType = {
  Progress: "progress",
  VideoCreated: "video_created",
  VideoDeleted: "video_deleted",
} as const;

export type WsEventType = (typeof WsEventType)[keyof typeof WsEventType];
