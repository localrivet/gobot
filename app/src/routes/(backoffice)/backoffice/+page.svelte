<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import Card from '$lib/components/ui/Card.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import { Users, CreditCard, TrendingUp, Calendar, LogOut } from 'lucide-svelte';
	import { getAdminStats, adminListUsers } from '$lib/api/gobot';
	import type { AdminStatsResponse, AdminUser } from '$lib/api/gobotComponents';
	import { auth, isAuthenticated } from '$lib/stores';

	let loading = $state(false);
	let stats = $state<AdminStatsResponse | null>(null);
	let recentUsers = $state<AdminUser[]>([]);

	async function loadAdminData() {
		loading = true;
		try {
			const [statsData, usersData] = await Promise.all([
				getAdminStats(),
				adminListUsers({ page: 1, pageSize: 10 })
			]);
			stats = statsData;
			recentUsers = usersData.users;
		} catch (err) {
			console.error('Failed to load admin data:', err);
			// If unauthorized, redirect to login
			if (err instanceof Error && (err.message.includes('401') || err.message.includes('403'))) {
				goto('/auth/login');
			}
		} finally {
			loading = false;
		}
	}

	function handleLogout() {
		auth.logout();
		goto('/auth/login');
	}

	onMount(async () => {
		// Redirect if not authenticated
		if (!$isAuthenticated) {
			goto('/auth/login');
			return;
		}
		await loadAdminData();
	});

	function formatDate(dateString: string): string {
		return new Date(dateString).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric'
		});
	}
</script>

<svelte:head>
	<title>Backoffice - Admin Dashboard</title>
	<meta name="robots" content="noindex, nofollow" />
</svelte:head>

<div class="min-h-screen bg-base-100 p-6">
	<div class="max-w-7xl mx-auto">
		<div class="flex items-center justify-between mb-8">
			<div>
				<h1 class="font-display text-2xl font-bold text-base-content mb-1">Backoffice Dashboard</h1>
				<p class="text-sm text-base-content/60">Your SaaS metrics at a glance</p>
			</div>
			<div class="flex items-center gap-3">
				<Button type="secondary" onclick={loadAdminData} disabled={loading}>
					{loading ? 'Refreshing...' : 'Refresh'}
				</Button>
				<Button type="ghost" onclick={handleLogout}>
					<LogOut class="w-4 h-4 mr-2" />
					Logout
				</Button>
			</div>
		</div>

		{#if stats}
			<div class="grid sm:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
				<Card>
					<div class="flex items-center justify-between mb-4">
						<div class="w-10 h-10 rounded-xl bg-primary/10 flex items-center justify-center">
							<Users class="w-5 h-5 text-primary" />
						</div>
					</div>
					<p class="text-sm font-medium text-base-content/60 mb-1">Total Users</p>
					<p class="font-display text-2xl font-bold text-base-content">{stats.totalUsers}</p>
					<div class="mt-3 pt-3 border-t border-base-300 text-xs text-base-content/40">
						+{stats.newUsersToday} today
					</div>
				</Card>

				<Card>
					<div class="flex items-center justify-between mb-4">
						<div class="w-10 h-10 rounded-xl bg-secondary/10 flex items-center justify-center">
							<CreditCard class="w-5 h-5 text-secondary" />
						</div>
					</div>
					<p class="text-sm font-medium text-base-content/60 mb-1">Active Subscriptions</p>
					<p class="font-display text-2xl font-bold text-base-content">{stats.activeSubscriptions}</p>
					<div class="mt-3 pt-3 border-t border-base-300 text-xs text-base-content/40">
						{stats.trialSubscriptions} on trial
					</div>
				</Card>

				<Card>
					<div class="flex items-center justify-between mb-4">
						<div class="w-10 h-10 rounded-xl bg-tertiary/10 flex items-center justify-center">
							<TrendingUp class="w-5 h-5 text-tertiary" />
						</div>
					</div>
					<p class="text-sm font-medium text-base-content/60 mb-1">This Week</p>
					<p class="font-display text-2xl font-bold text-base-content">+{stats.newUsersThisWeek}</p>
					<div class="mt-3 pt-3 border-t border-base-300 text-xs text-base-content/40">
						New users this week
					</div>
				</Card>

				<Card>
					<div class="flex items-center justify-between mb-4">
						<div class="w-10 h-10 rounded-xl bg-success/10 flex items-center justify-center">
							<Calendar class="w-5 h-5 text-success" />
						</div>
					</div>
					<p class="text-sm font-medium text-base-content/60 mb-1">This Month</p>
					<p class="font-display text-2xl font-bold text-base-content">+{stats.newUsersThisMonth}</p>
					<div class="mt-3 pt-3 border-t border-base-300 text-xs text-base-content/40">
						New users this month
					</div>
				</Card>
			</div>
		{/if}

		{#if recentUsers.length > 0}
			<Card>
				<h2 class="font-display text-lg font-bold text-base-content mb-4">Recent Users</h2>
				<div class="overflow-x-auto">
					<table class="w-full">
						<thead>
							<tr class="border-b border-base-300">
								<th class="text-left py-3 px-2 text-sm font-medium text-base-content/60">Email</th>
								<th class="text-left py-3 px-2 text-sm font-medium text-base-content/60">Name</th>
								<th class="text-left py-3 px-2 text-sm font-medium text-base-content/60">Plan</th>
								<th class="text-left py-3 px-2 text-sm font-medium text-base-content/60">Status</th>
								<th class="text-left py-3 px-2 text-sm font-medium text-base-content/60">Joined</th>
							</tr>
						</thead>
						<tbody>
							{#each recentUsers as user}
								<tr class="border-b border-base-content/10 hover:bg-base-200">
									<td class="py-3 px-2 text-sm text-base-content">{user.email}</td>
									<td class="py-3 px-2 text-sm text-base-content/70">{user.name || '-'}</td>
									<td class="py-3 px-2 text-sm">
										<span class="px-2 py-0.5 rounded-full text-xs font-medium capitalize
											{user.plan === 'free' ? 'bg-base-300 text-base-content/70' :
											 user.plan === 'pro' ? 'bg-primary/20 text-primary' :
											 'bg-secondary/20 text-secondary'}">
											{user.plan}
										</span>
									</td>
									<td class="py-3 px-2 text-sm">
										<span class="px-2 py-0.5 rounded-full text-xs font-medium capitalize
											{user.status === 'active' ? 'bg-success/20 text-success' :
											 user.status === 'trialing' ? 'bg-warning/20 text-warning' :
											 'bg-base-300 text-base-content/60'}">
											{user.status}
										</span>
									</td>
									<td class="py-3 px-2 text-sm text-base-content/60">{formatDate(user.createdAt)}</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			</Card>
		{/if}
	</div>
</div>
