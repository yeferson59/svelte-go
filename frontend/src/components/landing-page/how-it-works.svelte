<script lang="ts">
	const steps = [
		{
			number: '01',
			title: 'Importa o vincula',
			description: 'Conecta tus brokers, bancos o plataformas (o sube un Excel) con tus datos de inversión.'
		},
		{
			number: '02',
			title: 'FINEXIA lo centraliza',
			description: 'Nuestro motor de datos consolida todo, deduplicando activos y normalizando nomenclaturas automáticamente.'
		},
		{
			number: '03',
			title: 'Visualiza tu patrimonio',
			description: 'Accede a dashboards, gráficas en tiempo real y métricas de rendimiento al instante.'
		}
	];
</script>

<section id="como-funciona" class="mx-auto w-full max-w-7xl px-6 py-20">
	<div class="mb-16 space-y-4 text-center">
		<p class="text-xs font-600 tracking-widest uppercase text-[#d4af37]">Proceso simple</p>
		<h2 class="font-display text-4xl font-bold leading-tight text-[#ffffff]">
			Conecta en minutos
		</h2>
	</div>

	<div class="grid gap-8 md:grid-cols-3">
		{#each steps as step, index (step.number)}
			<div class="relative">
				<div class="step-card">
					<div class="step-number">{step.number}</div>
					<h3 class="font-display mb-3 text-lg font-bold text-[#ffffff]">{step.title}</h3>
					<p class="text-sm text-[#a0a0a0]">{step.description}</p>
				</div>
				{#if index < steps.length - 1}
					<div class="hidden step-connector md:block">→</div>
				{/if}
			</div>
		{/each}
	</div>
</section>

<style>
	.step-card {
		border: 1px solid rgba(212, 175, 55, 0.2);
		background: linear-gradient(135deg, #1a1f2e 0%, #141a24 100%);
		padding: 2rem;
		border-radius: 16px;
		backdrop-filter: blur(10px);
		position: relative;
		transition: all 0.3s ease;
	}

	.step-card:hover {
		border-color: rgba(212, 175, 55, 0.4);
		transform: translateY(-6px);
		box-shadow: 0 12px 40px rgba(212, 175, 55, 0.2);
	}

	.step-number {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		height: 48px;
		width: 48px;
		margin-bottom: 1rem;
		border-radius: 12px;
		background: linear-gradient(135deg, #d4af37, #2ecc71);
		font-family: 'IBM Plex Mono', monospace;
		font-weight: 700;
		font-size: 1.2rem;
		color: #0f1419;
	}

	.step-connector {
		position: absolute;
		top: 50%;
		right: -3rem;
		transform: translateY(-50%);
		font-size: 2rem;
		color: rgba(212, 175, 55, 0.4);
		font-weight: bold;
	}

	@media (max-width: 768px) {
		.step-card {
			padding: 1.5rem;
		}

		.step-connector {
			display: none;
		}
	}
</style>
