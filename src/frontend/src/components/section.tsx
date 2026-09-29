import type { ReactNode } from "react";

export function Section({
	title,
	description,
	children,
}: {
	title: string;
	description?: string;
	children: ReactNode;
}) {
	return (
		<section className="space-y-5 rounded-[2.5rem] border bg-card p-6">
			<div>
				<h2 className="font-bold text-xl tracking-tight">{title}</h2>
				{description && <p className="mt-0.5 text-muted-foreground text-sm">{description}</p>}
			</div>
			{children}
		</section>
	);
}
