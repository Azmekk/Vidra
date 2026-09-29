import { zodResolver } from "@hookform/resolvers/zod";
import { useQueryClient } from "@tanstack/react-query";
import { useRouter } from "@tanstack/react-router";
import { Loader2 } from "lucide-react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { getGetAuthStatusQueryKey, useSetup } from "@/api/gen/auth/auth";
import { AuthCard } from "@/components/auth-card";
import { Button } from "@/components/ui/button";
import {
	Form,
	FormControl,
	FormDescription,
	FormField,
	FormItem,
	FormLabel,
	FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";

const schema = z
	.object({
		code: z.string().trim().min(1, "Enter the setup code from the server log"),
		username: z
			.string()
			.trim()
			.regex(/^[a-zA-Z0-9_.-]{3,32}$/, "3-32 letters, digits, dots, dashes or underscores"),
		password: z.string().min(8, "At least 8 characters").max(256),
		confirm: z.string(),
	})
	.refine((v) => v.password === v.confirm, { path: ["confirm"], message: "Passwords don't match" });

export function SetupPage() {
	const router = useRouter();
	const qc = useQueryClient();
	const form = useForm({
		resolver: zodResolver(schema),
		defaultValues: { code: "", username: "", password: "", confirm: "" },
	});
	const setup = useSetup({
		mutation: {
			onSuccess: async (user) => {
				qc.setQueryData(getGetAuthStatusQueryKey(), { setupRequired: false, user });
				await router.navigate({ to: "/", replace: true });
			},
			onError: (e) => form.setError("root", { message: e.message }),
		},
	});

	return (
		<AuthCard title="Set up Vidra" description="Create the account that owns this library.">
			<Form {...form}>
				<form
					onSubmit={form.handleSubmit(({ confirm: _, ...data }) => setup.mutate({ data }))}
					className="space-y-5"
				>
					<FormField
						control={form.control}
						name="code"
						render={({ field }) => (
							<FormItem>
								<FormLabel>Setup code</FormLabel>
								<FormControl>
									<Input
										autoComplete="one-time-code"
										autoCapitalize="characters"
										autoFocus
										className="font-mono uppercase tracking-widest"
										{...field}
									/>
								</FormControl>
								<FormDescription>Printed in the server log on first start.</FormDescription>
								<FormMessage />
							</FormItem>
						)}
					/>
					<FormField
						control={form.control}
						name="username"
						render={({ field }) => (
							<FormItem>
								<FormLabel>Username</FormLabel>
								<FormControl>
									<Input autoComplete="username" autoCapitalize="none" {...field} />
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
									<Input type="password" autoComplete="new-password" {...field} />
								</FormControl>
								<FormMessage />
							</FormItem>
						)}
					/>
					<FormField
						control={form.control}
						name="confirm"
						render={({ field }) => (
							<FormItem>
								<FormLabel>Confirm password</FormLabel>
								<FormControl>
									<Input type="password" autoComplete="new-password" {...field} />
								</FormControl>
								<FormMessage />
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
						disabled={setup.isPending}
					>
						{setup.isPending && <Loader2 className="animate-spin" />}
						Create account
					</Button>
				</form>
			</Form>
		</AuthCard>
	);
}
