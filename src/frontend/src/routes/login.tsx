import { zodResolver } from "@hookform/resolvers/zod";
import { useQueryClient } from "@tanstack/react-query";
import { getRouteApi, useRouter } from "@tanstack/react-router";
import { Loader2 } from "lucide-react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { getGetAuthStatusQueryKey, useLogin } from "@/api/gen/auth/auth";
import { AuthCard } from "@/components/auth-card";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";

const schema = z.object({
	username: z.string().trim().min(1, "Enter your username"),
	password: z.string().min(1, "Enter your password"),
	remember: z.boolean(),
});

const route = getRouteApi("/login");

export function LoginPage() {
	const { redirect } = route.useSearch();
	const router = useRouter();
	const qc = useQueryClient();
	const form = useForm({
		resolver: zodResolver(schema),
		defaultValues: { username: "", password: "", remember: true },
	});
	const login = useLogin({
		mutation: {
			onSuccess: async (user) => {
				qc.setQueryData(getGetAuthStatusQueryKey(), { setupRequired: false, user });
				await router.navigate({ href: redirect ?? "/", replace: true });
			},
			onError: (e) => form.setError("root", { message: e.message }),
		},
	});

	return (
		<AuthCard title="Welcome back" description="Sign in to your Vidra library.">
			<Form {...form}>
				<form onSubmit={form.handleSubmit((data) => login.mutate({ data }))} className="space-y-5">
					<FormField
						control={form.control}
						name="username"
						render={({ field }) => (
							<FormItem>
								<FormLabel>Username</FormLabel>
								<FormControl>
									<Input autoComplete="username" autoCapitalize="none" autoFocus {...field} />
								</FormControl>
								<FormMessage />
							</FormItem>
						)}
					/>
					<FormField
						control={form.control}
						name="password"
						render={({ field }) => (
							<FormItem>
								<FormLabel>Password</FormLabel>
								<FormControl>
									<Input type="password" autoComplete="current-password" {...field} />
								</FormControl>
								<FormMessage />
							</FormItem>
						)}
					/>
					<FormField
						control={form.control}
						name="remember"
						render={({ field }) => (
							<FormItem className="flex items-center gap-3">
								<FormControl>
									<Checkbox checked={field.value} onCheckedChange={(v) => field.onChange(v === true)} />
								</FormControl>
								<FormLabel className="font-medium">Keep me signed in</FormLabel>
							</FormItem>
						)}
					/>
					{form.formState.errors.root && (
						<p className="rounded-xl bg-destructive/10 px-4 py-3 font-medium text-destructive text-sm">
							{form.formState.errors.root.message}
						</p>
					)}
					<Button
						type="submit"
						size="lg"
						className="h-12 w-full rounded-2xl font-bold text-base"
						disabled={login.isPending}
					>
						{login.isPending && <Loader2 className="animate-spin" />}
						Sign in
					</Button>
				</form>
			</Form>
		</AuthCard>
	);
}
