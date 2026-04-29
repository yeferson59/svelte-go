<script lang="ts">
	import { Tooltip } from 'bits-ui';
	import Header from '$components/landing-page/header.svelte';
	import Hero from '$components/landing-page/hero.svelte';
	import PainPoints from '$components/landing-page/pain-points.svelte';
	import HowItWorks from '$components/landing-page/how-it-works.svelte';
	import Features from '$components/landing-page/features.svelte';
	import SocialProof from '$components/landing-page/social-proof.svelte';
	import FinalCTA from '$components/landing-page/final-cta.svelte';
	import Footer from '$components/landing-page/footer.svelte';
</script>

<svelte:head>
	<title>FINEXIA | Tu patrimonio, bajo control</title>
	<meta
		name="description"
		content="Centraliza todas tus inversiones en un solo lugar. Visibilidad total sobre tu patrimonio, rendimiento y composición. Gestión inteligente de portafolios."
	/>
	<link rel="preconnect" href="https://fonts.googleapis.com" />
	<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin="" />
	<link
		rel="stylesheet"
		href="https://fonts.googleapis.com/css2?family=Poppins:wght@400;500;600;700&family=Lato:wght@400;500;700&family=IBM+Plex+Mono:wght@400;500&display=swap"
	/>
</svelte:head>

<main class="bg-charcoal text-gray-light">
	<Tooltip.Provider delayDuration={200}>
		<div class="page-foundation">
			<Header />

			<Hero />

			<PainPoints />

			<HowItWorks />

			<Features />

			<SocialProof />

			<FinalCTA />

			<Footer />
		</div>
	</Tooltip.Provider>
</main>

<style>
	:global(body) {
		font-family: 'Lato', system-ui, sans-serif;
		background: #0f1419;
		color: #e0e0e0;
		letter-spacing: -0.3px;
	}

	:global(h1),
	:global(h2),
	:global(h3),
	:global(.font-display) {
		font-family: 'Poppins', sans-serif;
		font-weight: 700;
	}

	:global(:root) {
		--charcoal: #0f1419;
		--charcoal-light: #1a1f2e;
		--gray-light: #e0e0e0;
		--gray-medium: #a0a0a0;
		--gold: #d4af37;
		--green: #2ecc71;
		--red: #e74c3c;
		--white: #ffffff;
		--shadow-sm: 0 2px 8px rgba(0, 0, 0, 0.3);
		--shadow-md: 0 8px 24px rgba(0, 0, 0, 0.4);
		--shadow-lg: 0 16px 40px rgba(0, 0, 0, 0.5);
	}

	.page-foundation {
		background: linear-gradient(to bottom, #0f1419 0%, #141a24 50%, #0f1419 100%);
	}

	.bg-charcoal {
		background: #0f1419;
	}

	.text-gray-light {
		color: #e0e0e0;
	}

	/* CTA Primary - Gold accent */
	:global(.cta-primary) {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
		border-radius: 8px;
		background: #d4af37;
		padding: 1rem 2rem;
		font-weight: 700;
		font-size: 0.95rem;
		letter-spacing: 0.5px;
		color: #0f1419;
		box-shadow: 0 4px 16px rgba(212, 175, 55, 0.2);
		transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
		border: none;
		cursor: pointer;
		font-family: 'Poppins', sans-serif;
	}

	:global(.cta-primary:hover) {
		background: #e8c547;
		transform: translateY(-2px);
		box-shadow: 0 6px 24px rgba(212, 175, 55, 0.35);
	}

	:global(.cta-primary:active) {
		transform: translateY(0);
	}

	:global(.cta-primary.invert) {
		background: transparent;
		color: #d4af37;
		border: 2px solid #d4af37;
	}

	:global(.cta-primary.invert:hover) {
		background: rgba(212, 175, 55, 0.1);
		border-color: #e8c547;
	}

	/* CTA Secondary */
	:global(.cta-secondary) {
		border-radius: 8px;
		border: 1px solid #d4af37;
		padding: 0.75rem 1.5rem;
		font-size: 0.85rem;
		text-transform: uppercase;
		letter-spacing: 0.1em;
		font-weight: 600;
		color: #d4af37;
		background: transparent;
		transition: all 0.25s ease;
		cursor: pointer;
		font-family: 'Poppins', sans-serif;
	}

	:global(.cta-secondary:hover) {
		background: rgba(212, 175, 55, 0.1);
		border-color: #e8c547;
	}

	/* CTA Ghost */
	:global(.cta-ghost) {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		border-radius: 8px;
		border: 1px solid #2ecc71;
		padding: 1rem 2rem;
		font-weight: 600;
		color: #2ecc71;
		background: transparent;
		transition: all 0.25s ease;
		cursor: pointer;
		font-family: 'Poppins', sans-serif;
	}

	:global(.cta-ghost:hover) {
		background: rgba(46, 204, 113, 0.1);
		border-color: #2ecc71;
	}

	:global(.cta-ghost.invert) {
		border-color: rgba(255, 255, 255, 0.3);
		color: #ffffff;
	}

	:global(.cta-ghost.invert:hover) {
		background: rgba(255, 255, 255, 0.1);
		border-color: #d4af37;
	}

	/* Separators & UI Elements */
	:global(.separator-rail) {
		height: 48px;
		width: 1px;
		background: rgba(212, 175, 55, 0.2);
		justify-self: center;
	}

	:global(.separator-line) {
		height: 1px;
		width: 100%;
		background: rgba(212, 175, 55, 0.1);
	}

	:global(.progress-bar) {
		height: 6px;
		width: 100%;
		border-radius: 3px;
		background: rgba(212, 175, 55, 0.15);
		overflow: hidden;
	}

	:global(.progress-indicator) {
		display: block;
		height: 100%;
		background: linear-gradient(90deg, #d4af37, #2ecc71);
		border-radius: 3px;
	}

	:global(.info-dot) {
		display: inline-flex;
		height: 24px;
		width: 24px;
		align-items: center;
		justify-content: center;
		border-radius: 50%;
		border: 1px solid #d4af37;
		color: #d4af37;
		font-size: 0.7rem;
		font-weight: 700;
		background: rgba(212, 175, 55, 0.1);
		cursor: help;
	}

	:global(.tooltip-card) {
		max-width: 240px;
		border-radius: 8px;
		border: 1px solid #d4af37;
		background: #1a1f2e;
		color: #e0e0e0;
		padding: 0.75rem 1rem;
		font-size: 0.8rem;
		line-height: 1.6;
		box-shadow: var(--shadow-md);
	}

	:global(.accordion-root) {
		display: grid;
		gap: 1rem;
	}

	:global(.accordion-item) {
		border-bottom: 1px solid rgba(212, 175, 55, 0.1);
		padding-bottom: 1rem;
	}

	:global(.accordion-trigger) {
		width: 100%;
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1rem;
		padding: 0.75rem 0;
		background: transparent;
		text-align: left;
		cursor: pointer;
		color: #d4af37;
		font-weight: 600;
	}

	:global(.accordion-content) {
		padding: 0.5rem 0 0 0;
		color: rgba(224, 224, 224, 0.7);
	}

	/* Animations */
	@keyframes fade-in-up {
		from {
			opacity: 0;
			transform: translateY(20px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	:global(.fade-in-up) {
		animation: fade-in-up 0.6s ease-out forwards;
	}

	@keyframes fade-in {
		from {
			opacity: 0;
		}
		to {
			opacity: 1;
		}
	}

	:global(.fade-in) {
		animation: fade-in 0.6s ease-out forwards;
	}

	/* Responsive */
	@media (max-width: 768px) {
		:global(body) {
			letter-spacing: -0.2px;
		}

		:global(.separator-rail) {
			width: 100%;
			height: 1px;
		}

		:global(.accordion-trigger) {
			flex-direction: column;
			align-items: flex-start;
		}

		:global(.cta-primary) {
			padding: 0.9rem 1.6rem;
			font-size: 0.9rem;
		}
	}
</style>
