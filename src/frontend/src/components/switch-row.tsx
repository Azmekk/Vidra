import { type ReactNode, useId } from "react";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { cn } from "@/lib/utils";

export function SwitchRow({
	label,
	description,
	checked,
	onCheckedChange,
	className,
}: {
	label: ReactNode;
	description?: ReactNode;
	checked: boolean;
	onCheckedChange: (checked: boolean) => void;
	className?: string;
}) {
	const id = useId();
	return (
		<div className={cn("flex items-center justify-between gap-4", className)}>
			<Label htmlFor={id} className="block cursor-pointer">
				<span className="block font-semibold">{label}</span>
				{description && (
					<span className="mt-0.5 block font-normal text-muted-foreground text-sm">{description}</span>
				)}
			</Label>
			<Switch id={id} checked={checked} onCheckedChange={onCheckedChange} />
		</div>
	);
}
