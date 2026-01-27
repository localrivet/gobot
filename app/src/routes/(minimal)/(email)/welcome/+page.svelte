<script lang="ts">
	import { setSEO } from '$lib/utils/seo';

	const seo = setSEO({
		title: 'Email Confirmed - Gobot',
		description: 'Your email has been confirmed. Welcome to Gobot!',
		url: '/welcome'
	});
</script>

<svelte:head>
	<title>{seo.title}</title>
	<meta name="description" content={seo.description} />
	<link rel="canonical" href={seo.canonical} />
</svelte:head>

<section class="max-w-2xl mx-auto px-4 sm:px-6 lg:px-8 py-20 text-center">
	<div class="animate-fade-in">
		<div class="mb-8">
			<svg
				class="w-16 h-16 mx-auto text-green-500"
				fill="none"
				viewBox="0 0 24 24"
				stroke="currentColor"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
				/>
			</svg>
		</div>
		<h1 class="font-display text-3xl sm:text-4xl font-black text-white leading-tight mb-4">
			Email Confirmed!
		</h1>
		<p class="text-base-content/40 text-lg mb-8">
			Thank you for confirming your email address. You're all set!
		</p>
		<a
			href="/auth/login"
			class="inline-flex items-center px-6 py-3 bg-primary-500 hover:bg-primary-600 text-white font-semibold rounded-lg transition-colors"
		>
			Continue to Login
		</a>
	</div>
</section>
