<script lang="ts">
	import { setSEO, getFAQSchema } from '$lib/utils/seo';
	import { Button } from '$lib/components/ui';
	import {
		Zap,
		Shield,
		BarChart3,
		ArrowRight,
		Check,
		Quote
	} from 'lucide-svelte';

	const seo = setSEO({
		title: 'Your SaaS Name - Build Something Amazing',
		description:
			'The modern platform for building and scaling your SaaS business. Ship faster with authentication, billing, and email built-in.',
		keywords: [
			'saas boilerplate',
			'saas starter kit',
			'sveltekit template',
			'stripe integration',
			'authentication'
		],
		url: '/'
	});

	const features = [
		{
			icon: Zap,
			title: 'Launch Fast',
			description: 'Ship your MVP in days, not months. Authentication, billing, and email are already wired up.',
			color: 'var(--color-primary)'
		},
		{
			icon: Shield,
			title: 'Secure by Default',
			description: 'Enterprise-grade security with JWT auth, CORS protection, and rate limiting built-in.',
			color: 'var(--color-secondary)'
		},
		{
			icon: BarChart3,
			title: 'Scale Effortlessly',
			description: 'Built on proven technologies that grow with you from 0 to millions of users.',
			color: 'var(--color-tertiary)'
		}
	];

	const steps = [
		{
			step: '1',
			title: 'Clone',
			description: 'Fork the repo and run the setup script'
		},
		{
			step: '2',
			title: 'Configure',
			description: 'Add your API keys and customize the theme'
		},
		{
			step: '3',
			title: 'Launch',
			description: 'Deploy to your favorite platform and start selling'
		}
	];

	const faqs = [
		{
			question: 'What is included in this starter?',
			answer:
				'Everything you need to launch a SaaS: user authentication, Stripe billing integration, email sequences, a beautiful landing page, dashboard, and account management.'
		},
		{
			question: 'What technologies does this use?',
			answer:
				'The frontend is built with SvelteKit 5 and Tailwind CSS v4. The backend uses Go with the go-zero framework. Levee handles auth, billing, and email.'
		},
		{
			question: 'Can I use this for commercial projects?',
			answer:
				'Yes! This is a boilerplate designed to be renamed and customized for your own SaaS product. Use it however you like.'
		},
		{
			question: 'How do I customize the design?',
			answer:
				'All styling is centralized in app.css using CSS custom properties. Change the color palette, fonts, and component styles in one place.'
		}
	];

	const faqSchema = JSON.stringify(getFAQSchema(faqs));
</script>

<svelte:head>
	<title>{seo.title}</title>
	<meta name="description" content={seo.description} />
	<meta name="keywords" content={seo.keywords} />
	<link rel="canonical" href={seo.canonical} />
	<meta property="og:type" content={seo.type} />
	<meta property="og:title" content={seo.title} />
	<meta property="og:description" content={seo.description} />
	<meta property="og:url" content={seo.url} />
	<meta property="og:image" content={seo.image} />
	<meta property="og:site_name" content={seo.siteName} />
	<meta property="og:locale" content={seo.locale} />
	<meta name="twitter:card" content="summary_large_image" />
	<meta name="twitter:site" content={seo.twitterHandle} />
	<meta name="twitter:title" content={seo.title} />
	<meta name="twitter:description" content={seo.description} />
	<meta name="twitter:image" content={seo.image} />
	{@html `<script type="application/ld+json">${faqSchema}</script>`}
</svelte:head>

<!-- Hero Section -->
<section class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-20 sm:py-32 text-center">
	<div class="animate-fade-in">
		<span class="inline-flex items-center gap-2 rounded-full px-4 py-2 text-sm font-medium mb-8 bg-primary/10 border border-primary/20 text-primary">
			<span class="h-2 w-2 rounded-full bg-primary animate-pulse"></span>
			Now in public beta
		</span>

		<h1 class="font-display text-4xl sm:text-5xl lg:text-6xl font-black text-base-content leading-tight tracking-tight mb-6">
			Build your SaaS<br />
			<span class="text-gradient">faster than ever</span>
		</h1>

		<p class="text-lg sm:text-xl text-base-content/70 max-w-2xl mx-auto mb-10 leading-relaxed">
			Stop reinventing the wheel. This boilerplate gives you authentication, billing, email,
			and a polished UI so you can focus on what makes your product unique.
		</p>

		<div class="flex flex-col sm:flex-row items-center justify-center gap-4">
			<Button type="primary" size="lg" href="/auth/register">
				Get Started Free
				<ArrowRight class="w-5 h-5" />
			</Button>
			<Button type="secondary" size="lg" href="/pricing">View Pricing</Button>
		</div>
	</div>
</section>

<!-- Features Grid -->
<section class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-16 sm:py-24" id="features">
	<div class="text-center mb-12 sm:mb-16">
		<h2 class="font-display text-3xl sm:text-4xl font-bold text-base-content mb-4">Everything you need to launch</h2>
		<p class="text-lg text-base-content/70 max-w-2xl mx-auto">
			Built with modern technologies and best practices so you can ship with confidence.
		</p>
	</div>

	<div class="grid gap-6 md:grid-cols-2 lg:grid-cols-3 lg:gap-8">
		{#each features as feature, i}
			{@const FeatureIcon = feature.icon}
			{@const delay = i + 1}
			<div
				class="card bg-base-200/50 border border-base-300 backdrop-blur-sm animate-slide-up opacity-0 stagger-delay"
				style:--delay={delay}
				style:--theme-color={feature.color}
			>
				<div class="card-body">
					<div class="w-14 h-14 rounded-2xl flex items-center justify-center mb-4 themed-bg">
						<FeatureIcon class="w-7 h-7 themed-color" />
					</div>
					<h3 class="font-display text-xl font-bold text-base-content">{feature.title}</h3>
					<p class="text-base-content/70">{feature.description}</p>
				</div>
			</div>
		{/each}
	</div>
</section>

<!-- How It Works -->
<div class="bg-base-200/50 py-16 sm:py-24">
	<section class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
		<div class="text-center mb-12 sm:mb-16">
			<h2 class="font-display text-3xl sm:text-4xl font-bold text-base-content mb-4">Get started in minutes</h2>
			<p class="text-lg text-base-content/70 max-w-2xl mx-auto">Three simple steps to your production-ready SaaS.</p>
		</div>

		<div class="relative">
			<!-- Connecting line (desktop only) -->
			<div class="hidden md:block absolute top-1/2 left-0 right-0 h-0.5 -translate-y-1/2 bg-gradient-to-r from-transparent via-primary/30 to-transparent"></div>

			<div class="relative grid gap-8 md:grid-cols-3">
				{#each steps as item, i}
					{@const delay = i + 1}
					<div class="relative text-center animate-slide-up opacity-0 stagger-delay-150" style:--delay={delay}>
						<div class="relative z-10 mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-base-100 border border-primary/30">
							<span class="font-display text-2xl font-black text-primary">{item.step}</span>
						</div>
						<h3 class="font-display text-lg font-bold text-base-content mb-2">{item.title}</h3>
						<p class="text-sm text-base-content/60">{item.description}</p>
					</div>
				{/each}
			</div>
		</div>
	</section>
</div>

<!-- What's Included -->
<section class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-16 sm:py-24">
	<div class="text-center mb-12 sm:mb-16">
		<h2 class="font-display text-3xl sm:text-4xl font-bold text-base-content mb-4">What's included</h2>
		<p class="text-lg text-base-content/70 max-w-2xl mx-auto">
			A complete foundation for your SaaS business, ready to customize.
		</p>
	</div>

	<div class="grid gap-6 md:grid-cols-2 lg:grid-cols-3 lg:gap-8">
		{#each [
			{ title: 'User Authentication', items: ['Email/password login', 'Password reset flow', 'Email verification', 'Session management'] },
			{ title: 'Stripe Billing', items: ['Subscription management', 'Customer portal', 'Usage-based pricing', 'Webhook handling'] },
			{ title: 'Email Marketing', items: ['Transactional emails', 'Drip campaigns', 'List management', 'Template system'] },
			{ title: 'Beautiful UI', items: ['50+ UI components', 'Dark mode ready', 'Responsive design', 'Accessibility built-in'] },
			{ title: 'Developer Experience', items: ['TypeScript throughout', 'API client generation', 'Hot reload', 'Type-safe routing'] },
			{ title: 'Production Ready', items: ['Security headers', 'Rate limiting', 'Error handling', 'Logging & monitoring'] }
		] as section, i}
			{@const delay = i + 1}
			<div class="card bg-base-200/50 border border-base-300 backdrop-blur-sm animate-slide-up opacity-0 stagger-delay-80" style:--delay={delay}>
				<div class="card-body">
					<h3 class="font-display text-xl font-bold text-base-content">{section.title}</h3>
					<ul class="mt-4 space-y-2">
						{#each section.items as item}
							<li class="flex items-center gap-2 text-base-content/70 text-sm">
								<Check class="w-4 h-4 text-success flex-shrink-0" />
								<span>{item}</span>
							</li>
						{/each}
					</ul>
				</div>
			</div>
		{/each}
	</div>
</section>

<!-- Testimonial -->
<div class="bg-base-200/50 py-16 sm:py-24">
	<section class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
		<div class="max-w-3xl mx-auto">
			<div class="card bg-base-200/80 border border-base-300 backdrop-blur-sm">
				<div class="card-body p-8 lg:p-12">
					<Quote class="w-8 h-8 text-primary mb-4" />
					<p class="text-lg text-base-content italic leading-relaxed mb-6">
						"This boilerplate saved us months of development time. We launched our MVP in under two weeks
						and have been iterating on customer feedback ever since."
					</p>
					<div class="flex items-center gap-4">
						<div class="w-12 h-12 rounded-full flex items-center justify-center text-white font-bold text-lg bg-gradient-to-br from-primary to-secondary">
							JD
						</div>
						<div>
							<p class="font-semibold text-base-content">Jane Doe</p>
							<p class="text-sm text-base-content/60">Founder, Example Startup</p>
						</div>
					</div>
				</div>
			</div>
		</div>
	</section>
</div>

<!-- FAQ Section -->
<section class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-16 sm:py-24" id="faq">
	<div class="text-center mb-12 sm:mb-16">
		<h2 class="font-display text-3xl sm:text-4xl font-bold text-base-content mb-4">Frequently Asked Questions</h2>
	</div>

	<div class="max-w-3xl mx-auto space-y-6">
		{#each faqs as faq}
			<div class="card bg-base-200/50 border border-base-300 backdrop-blur-sm">
				<div class="card-body p-6">
					<h3 class="font-display text-lg font-semibold text-base-content mb-2">{faq.question}</h3>
					<p class="text-base-content/70 leading-relaxed">{faq.answer}</p>
				</div>
			</div>
		{/each}
	</div>
</section>

<!-- CTA Section -->
<div class="bg-base-200/50 py-16 sm:py-24">
	<section class="text-center max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
		<h2 class="font-display text-3xl sm:text-4xl font-bold text-base-content mb-6">
			Ready to build your SaaS?
		</h2>
		<p class="text-xl text-base-content/70 mb-10 max-w-2xl mx-auto">
			Start with a solid foundation and focus on what makes your product unique.
		</p>
		<Button type="primary" size="lg" href="/auth/register">
			Get Started Free
			<ArrowRight class="w-5 h-5" />
		</Button>
	</section>
</div>
