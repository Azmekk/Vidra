import { useQueryClient } from "@tanstack/react-query";
import { useEffect } from "react";
import { z } from "zod";
import { getListBackupTargetsQueryKey } from "@/api/gen/backups/backups";
import type { HandlersBackupOverviewResponse } from "@/api/gen/model";
import { CreateBackupTargetResponse, GetVideoResponse, ListProgressResponseItem } from "@/api/gen/zod";
import { removeVideo, setProgress, upsertVideo } from "@/lib/videos";

const event = z.discriminatedUnion("type", [
	z.object({ type: z.enum(["video_created", "video_updated"]), payload: GetVideoResponse }),
	z.object({ type: z.literal("video_deleted"), payload: z.object({ id: z.string() }) }),
	z.object({ type: z.literal("file_progress"), payload: ListProgressResponseItem }),
	z.object({ type: z.literal("backup_status"), payload: CreateBackupTargetResponse }),
]);

export function useLiveUpdates() {
	const qc = useQueryClient();

	useEffect(() => {
		let socket: WebSocket | undefined;
		let retry: ReturnType<typeof setTimeout> | undefined;
		let attempts = 0;
		let closed = false;

		const connect = () => {
			const proto = location.protocol === "https:" ? "wss:" : "ws:";
			socket = new WebSocket(`${proto}//${location.host}/api/ws`);
			socket.onopen = () => {
				if (attempts > 0) qc.invalidateQueries();
				attempts = 0;
			};
			socket.onmessage = (msg) => {
				const parsed = event.safeParse(JSON.parse(msg.data));
				if (!parsed.success) return;
				const e = parsed.data;
				switch (e.type) {
					case "video_created":
					case "video_updated":
						upsertVideo(qc, e.payload);
						break;
					case "video_deleted":
						removeVideo(qc, e.payload.id);
						break;
					case "file_progress":
						setProgress(qc, e.payload);
						break;
					case "backup_status":
						qc.setQueryData<HandlersBackupOverviewResponse>(getListBackupTargetsQueryKey(), (data) =>
							data
								? { ...data, targets: data.targets.map((t) => (t.id === e.payload.id ? e.payload : t)) }
								: data,
						);
						break;
				}
			};
			socket.onclose = () => {
				if (closed) return;
				attempts++;
				retry = setTimeout(connect, Math.min(1000 * 2 ** attempts, 15_000));
			};
		};

		const onVisible = () => {
			if (document.visibilityState === "visible" && socket?.readyState === WebSocket.CLOSED) {
				clearTimeout(retry);
				connect();
			}
		};

		connect();
		document.addEventListener("visibilitychange", onVisible);
		return () => {
			closed = true;
			clearTimeout(retry);
			document.removeEventListener("visibilitychange", onVisible);
			socket?.close();
		};
	}, [qc]);
}
