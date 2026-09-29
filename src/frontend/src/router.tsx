import type { QueryClient } from "@tanstack/react-query";
import {
	createRootRouteWithContext,
	createRoute,
	createRouter,
	lazyRouteComponent,
	Outlet,
	redirect,
} from "@tanstack/react-router";
import { z } from "zod";
import { setUnauthorizedHandler } from "@/api/fetcher";
import { getGetAuthStatusQueryOptions } from "@/api/gen/auth/auth";
import { AppShell } from "@/components/app-shell";
import { AuthSkeleton, LibrarySkeleton, PageSkeleton, ShellSkeleton } from "@/components/skeletons";
import { clearCache, queryClient } from "@/lib/query";
import { LibraryPage } from "@/routes/library";
import { LoginPage } from "@/routes/login";
import { SetupPage } from "@/routes/setup";

const authStatus = () => queryClient.ensureQueryData(getGetAuthStatusQueryOptions());

const rootRoute = createRootRouteWithContext<{ queryClient: QueryClient }>()({
	component: Outlet,
});

const loginRoute = createRoute({
	getParentRoute: () => rootRoute,
	path: "/login",
	validateSearch: z.object({ redirect: z.string().optional() }),
	beforeLoad: async ({ search }) => {
		const status = await authStatus();
		if (status.setupRequired) throw redirect({ to: "/setup" });
		if (status.user) throw redirect({ href: search.redirect ?? "/" });
	},
	pendingComponent: AuthSkeleton,
	component: LoginPage,
});

const setupRoute = createRoute({
	getParentRoute: () => rootRoute,
	path: "/setup",
	beforeLoad: async () => {
		if (!(await authStatus()).setupRequired) throw redirect({ to: "/login" });
	},
	pendingComponent: AuthSkeleton,
	component: SetupPage,
});

const appRoute = createRoute({
	getParentRoute: () => rootRoute,
	id: "app",
	beforeLoad: async ({ location }) => {
		const status = await authStatus();
		if (status.setupRequired) throw redirect({ to: "/setup" });
		if (!status.user) throw redirect({ to: "/login", search: { redirect: location.href } });
		return { user: status.user };
	},
	pendingComponent: ShellSkeleton,
	component: AppShell,
});

const libraryRoute = createRoute({
	getParentRoute: () => appRoute,
	path: "/",
	validateSearch: z.object({
		q: z.string().optional(),
		order: z.enum(["created_at_desc", "created_at_asc", "name_asc", "name_desc"]).optional(),
	}),
	pendingComponent: LibrarySkeleton,
	component: LibraryPage,
});

const downloadRoute = createRoute({
	getParentRoute: () => appRoute,
	path: "/download",
	validateSearch: z.object({
		url: z.string().optional(),
		text: z.string().optional(),
		title: z.string().optional(),
		quick: z.coerce.boolean().optional(),
	}),
	pendingComponent: PageSkeleton,
	component: lazyRouteComponent(() => import("@/routes/download"), "DownloadPage"),
});

const settingsRoute = createRoute({
	getParentRoute: () => appRoute,
	path: "/settings",
	pendingComponent: PageSkeleton,
	component: lazyRouteComponent(() => import("@/routes/settings"), "SettingsPage"),
});

const errorsRoute = createRoute({
	getParentRoute: () => appRoute,
	path: "/errors",
	pendingComponent: PageSkeleton,
	component: lazyRouteComponent(() => import("@/routes/errors"), "ErrorsPage"),
});

const routeTree = rootRoute.addChildren([
	loginRoute,
	setupRoute,
	appRoute.addChildren([libraryRoute, downloadRoute, settingsRoute, errorsRoute]),
]);

export const router = createRouter({
	routeTree,
	context: { queryClient },
	defaultPreload: "intent",
	defaultPreloadStaleTime: 0,
	defaultPendingMs: 300,
	defaultPendingMinMs: 400,
	scrollRestoration: true,
});

setUnauthorizedHandler(async () => {
	if (router.state.location.pathname === "/login") return;
	await clearCache();
	router.navigate({ to: "/login", search: { redirect: router.state.location.href } });
});

declare module "@tanstack/react-router" {
	interface Register {
		router: typeof router;
	}
}
