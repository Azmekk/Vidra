import type { ReactNode } from "react";
import {
	Drawer,
	DrawerContent,
	DrawerDescription,
	DrawerFooter,
	DrawerHeader,
	DrawerTitle,
} from "@/components/ui/drawer";

export function Sheet({
	open,
	onOpenChange,
	title,
	description,
	footer,
	children,
}: {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	title: ReactNode;
	description?: ReactNode;
	footer?: ReactNode;
	children: ReactNode;
}) {
	return (
		<Drawer open={open} onOpenChange={onOpenChange} repositionInputs={false}>
			<DrawerContent className="mx-auto max-h-[92dvh]! max-w-2xl rounded-t-[2.5rem]!">
				<DrawerHeader className="px-6 text-left">
					<DrawerTitle className="font-extrabold text-2xl tracking-tight">{title}</DrawerTitle>
					{description && <DrawerDescription className="font-medium">{description}</DrawerDescription>}
				</DrawerHeader>
				<div className="flex-1 overflow-y-auto overscroll-contain px-6 pb-4">{children}</div>
				{footer && (
					<DrawerFooter className="border-t px-6 pb-[max(1rem,env(safe-area-inset-bottom))]">
						{footer}
					</DrawerFooter>
				)}
				{!footer && <div className="pb-[env(safe-area-inset-bottom)]" />}
			</DrawerContent>
		</Drawer>
	);
}
