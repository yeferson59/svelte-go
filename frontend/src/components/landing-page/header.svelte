<script lang="ts">
	import { Button } from 'bits-ui';
</script>

<header class="sticky top-0 z-50 border-b border-[rgba(212,175,55,0.1)] bg-[#0f1419]/95 backdrop-blur-md">
	<nav class="mx-auto flex w-full max-w-7xl items-center justify-between px-6 py-4">
		<div class="flex items-center gap-3">
			<div class="logo-mark" aria-hidden="true"></div>
			<span class="font-display text-xl font-bold tracking-tight text-[#d4af37]">FINEXIA</span>
		</div>
		<div class="hidden items-center gap-10 text-sm font-500 tracking-tight md:flex">
			<a href="#caracteristicas" class="nav-link">Características</a>
			<a href="#como-funciona" class="nav-link">Cómo funciona</a>
			<a href="#caracteristicas-principales" class="nav-link">Features</a>
			<a href="#testimonios" class="nav-link">Testimonios</a>
		</div>
		<Button.Root class="cta-secondary" aria-label="Ver demo">Ver demo</Button.Root>
	</nav>
</header>

<style>
	.logo-mark {
		height: 32px;
		width: 32px;
		border-radius: 6px;
		background: linear-gradient(135deg, #d4af37, #2ecc71);
		border: 1px solid #d4af37;
		box-shadow: 0 0 12px rgba(212, 175, 55, 0.3);
	}

	.nav-link {
		position: relative;
		color: #e0e0e0;
		text-decoration: none;
		transition: color 0.25s ease;
	}

	.nav-link::after {
		content: '';
		position: absolute;
		bottom: -4px;
		left: 0;
		width: 0;
		height: 2px;
		background: #d4af37;
		transition: width 0.25s ease;
	}

	.nav-link:hover {
		color: #d4af37;
	}

	.nav-link:hover::after {
		width: 100%;
	}
</style>
