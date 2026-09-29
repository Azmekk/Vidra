import "@/index.css";
import "@/lib/theme";
import { PersistQueryClientProvider } from "@tanstack/react-query-persist-client";
import { RouterProvider } from "@tanstack/react-router";
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { AuthSkeleton, ShellSkeleton } from "@/components/skeletons";
import { Toaster } from "@/components/ui/sonner";
import { TooltipProvider } from "@/components/ui/tooltip";
import { persister, queryClient } from "@/lib/query";
import { router } from "@/router";

// biome-ignore lint/style/noNonNullAssertion: root element is in index.html
const root = createRoot(document.getElementById("root")!);

root.render(/^\/(login|setup)/.test(location.pathname) ? <AuthSkeleton /> : <ShellSkeleton />);

await router.load().catch(() => {});

root.render(
	<StrictMode>
		<PersistQueryClientProvider
			client={queryClient}
			persistOptions={{ persister, maxAge: 7 * 24 * 60 * 60 * 1000, buster: __APP_VERSION__ }}
		>
			<TooltipProvider>
				<RouterProvider router={router} />
				<Toaster
					position="top-center"
					richColors
					mobileOffset={{ top: "calc(env(safe-area-inset-top) + 12px)" }}
				/>
			</TooltipProvider>
		</PersistQueryClientProvider>
	</StrictMode>,
);
