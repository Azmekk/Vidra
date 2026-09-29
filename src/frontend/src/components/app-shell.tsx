import { Link, Outlet, useNavigate, useRouteContext } from "@tanstack/react-router";
import { CircleAlert, Download, LogOut, Monitor, Moon, Settings2, Sun, UserRound, Video } from "lucide-react";
import { useLogout } from "@/api/gen/auth/auth";
import { Logo } from "@/components/logo";
import { Button } from "@/components/ui/button";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuLabel,
	DropdownMenuRadioGroup,
	DropdownMenuRadioItem,
	DropdownMenuSeparator,
	DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { clearCache } from "@/lib/query";
import { type Theme, useTheme } from "@/lib/theme";
import { useLiveUpdates } from "@/lib/ws";

const links = [
	{ to: "/", label: "Library", icon: Video },
	{ to: "/download", label: "Download", icon: Download },
	{ to: "/errors", label: "Errors", icon: CircleAlert },
	{ to: "/settings", label: "Settings", icon: Settings2 },
] as const;

export function AppShell() {
	useLiveUpdates();
	return (
		<div className="flex min-h-dvh flex-col">
			<header className="sticky top-0 z-40 border-b bg-background/80 pt-[env(safe-area-inset-top)] backdrop-blur-lg">
				<div className="mx-auto flex h-16 max-w-2xl items-center gap-6 px-4">
					<Link to="/" className="flex items-center gap-2.5">
						<Logo className="size-8" />
						<span className="font-bold text-xl tracking-tight">Vidra</span>
					</Link>
					<nav className="hidden flex-1 items-center gap-5 font-medium text-sm md:flex">
						{links.map((l) => (
							<Link
								key={l.to}
								to={l.to}
								className="text-foreground/60 transition-colors hover:text-foreground/80"
								activeProps={{ className: "text-foreground!" }}
								activeOptions={{ exact: true, includeSearch: false }}
							>
								{l.label}
							</Link>
						))}
					</nav>
					<div className="ml-auto">
						<UserMenu />
					</div>
				</div>
			</header>

			<main className="mx-auto w-full max-w-2xl flex-1 px-4 pt-8 pb-[calc(6rem+env(safe-area-inset-bottom))] md:pb-12">
				<Outlet />
			</main>

			<nav className="fixed inset-x-0 bottom-0 z-40 border-t bg-background/90 pb-[env(safe-area-inset-bottom)] backdrop-blur-lg md:hidden">
				<div className="mx-auto grid max-w-md grid-cols-4">
					{links.map((l) => (
						<Link
							key={l.to}
							to={l.to}
							className="flex flex-col items-center gap-1 py-2.5 font-semibold text-[11px] text-muted-foreground transition-colors"
							activeProps={{ className: "text-foreground!" }}
							activeOptions={{ exact: true, includeSearch: false }}
						>
							<l.icon className="size-6" />
							{l.label}
						</Link>
					))}
				</div>
			</nav>
		</div>
	);
}

function UserMenu() {
	const { user } = useRouteContext({ from: "/app" });
	const { theme, setTheme } = useTheme();
	const navigate = useNavigate();
	const logout = useLogout({
		mutation: {
			onSettled: async () => {
				await clearCache();
				navigate({ to: "/login" });
			},
		},
	});

	return (
		<DropdownMenu>
			<DropdownMenuTrigger asChild>
				<Button variant="secondary" size="icon" className="size-10 rounded-xl" aria-label="Account">
					<UserRound className="size-5" />
				</Button>
			</DropdownMenuTrigger>
			<DropdownMenuContent align="end" className="w-52 rounded-2xl p-1.5">
				<DropdownMenuLabel className="font-semibold">{user.username}</DropdownMenuLabel>
				<DropdownMenuSeparator />
				<DropdownMenuRadioGroup value={theme} onValueChange={(v) => setTheme(v as Theme)}>
					<DropdownMenuRadioItem value="light">
						<Sun /> Light
					</DropdownMenuRadioItem>
					<DropdownMenuRadioItem value="dark">
						<Moon /> Dark
					</DropdownMenuRadioItem>
					<DropdownMenuRadioItem value="system">
						<Monitor /> System
					</DropdownMenuRadioItem>
				</DropdownMenuRadioGroup>
				<DropdownMenuSeparator />
				<DropdownMenuItem onSelect={() => logout.mutate()}>
					<LogOut /> Sign out
				</DropdownMenuItem>
			</DropdownMenuContent>
		</DropdownMenu>
	);
}
