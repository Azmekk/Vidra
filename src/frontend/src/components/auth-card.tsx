import type { ReactNode } from "react";
import { Logo } from "@/components/logo";

export function AuthCard({
	title,
	description,
	children,
}: {
	title: string;
	description: string;
	children: ReactNode;
}) {
	return (
		<div className="flex min-h-dvh items-center justify-center px-4 pt-[env(safe-area-inset-top)] pb-[env(safe-area-inset-bottom)]">
			<div className="w-full max-w-sm rounded-[2.5rem] border bg-card p-8 shadow-primary/5 shadow-xl">
				<div className="mb-8 flex flex-col items-center text-center">
					<Logo className="mb-5 size-14" />
					<h1 className="font-extrabold text-3xl tracking-tight">{title}</h1>
					<p className="mt-2 font-medium text-muted-foreground">{description}</p>
				</div>
				{children}
			</div>
		</div>
	);
}
