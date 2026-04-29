<script lang="ts">
	import { Button } from 'bits-ui';

	const stats = [
		{ label: 'Inversiones rastreadas', value: '47.000+' },
		{ label: 'Brokers compatibles', value: '15' },
		{ label: 'Para inversionistas serios', value: 'Diseñado' }
	];
</script>

<section class="mx-auto grid w-full max-w-7xl gap-12 px-6 py-24 lg:grid-cols-[1.1fr_0.9fr]">
	<div class="space-y-8">
		<div class="space-y-6">
			<div class="inline-block rounded-full border border-[rgba(212,175,55,0.3)] bg-[rgba(212,175,55,0.05)] px-4 py-2">
				<p class="text-xs font-600 tracking-widest uppercase text-[#d4af37]">Gestión centralizada</p>
			</div>
			<h1 class="font-display text-5xl font-bold leading-tight text-[#ffffff] lg:text-6xl">
				Tu patrimonio, bajo control
			</h1>
			<p class="text-lg leading-relaxed text-[#a0a0a0]">
				Centraliza todas tus inversiones en un solo lugar. Sin dispersión. Sin confusión. Solo claridad financiera.
			</p>
		</div>
		<div class="flex flex-wrap items-center gap-4">
			<Button.Root class="cta-primary" aria-label="Empieza gratis">
				Empieza gratis
			</Button.Root>
			<Button.Root class="cta-ghost" aria-label="Ver demo interactiva" href="#demo">
				Ver demo
			</Button.Root>
		</div>
		<div class="pt-4">
			<p class="mb-4 text-xs font-600 tracking-widest uppercase text-[#a0a0a0]">Sin tarjeta de crédito · Acceso inmediato</p>
			<div class="stats-grid">
				{#each stats as stat (stat.label)}
					<div class="stat-item">
						<span class="text-xs font-600 tracking-widest uppercase text-[#d4af37]">{stat.label}</span>
						<p class="text-2xl font-bold text-[#ffffff]">{stat.value}</p>
					</div>
				{/each}
			</div>
		</div>
	</div>
	<div class="flex flex-col gap-6">
		<div class="dashboard-card">
			<div class="flex items-center justify-between border-b border-[rgba(212,175,55,0.1)] pb-4">
				<p class="text-xs font-600 tracking-widest uppercase text-[#a0a0a0]">Tablero</p>
				<span class="rounded-full bg-[rgba(46,204,113,0.2)] px-3 py-1 text-xs font-600 text-[#2ecc71]">
					Activo
				</span>
			</div>
			<div class="mt-6 space-y-6">
				<div>
					<p class="text-sm font-500 text-[#a0a0a0]">Patrimonio total</p>
					<p class="mt-2 text-3xl font-bold text-[#d4af37]">$128.440</p>
				</div>
				<div class="grid grid-cols-2 gap-4">
					<div class="card-box bg-[rgba(212,175,55,0.05)] p-4">
						<p class="text-sm text-[#a0a0a0]">Riesgo</p>
						<p class="mt-2 text-base font-600 text-[#e0e0e0]">Moderado</p>
					</div>
					<div class="card-box bg-[rgba(46,204,113,0.05)] p-4">
						<p class="text-sm text-[#a0a0a0]">Liquidez</p>
						<p class="mt-2 text-base font-600 text-[#2ecc71]">72h</p>
					</div>
				</div>
				<div class="space-y-4">
					{#each [
						{ label: 'Acciones', value: 40 },
						{ label: 'ETFs', value: 35 },
						{ label: 'Cripto', value: 15 },
						{ label: 'Efectivo', value: 10 }
					] as slice (slice.label)}
						<div>
							<div class="flex items-center justify-between text-sm">
								<span class="text-[#a0a0a0]">{slice.label}</span>
								<span class="font-600 text-[#e0e0e0]">{slice.value}%</span>
							</div>
							<div class="progress-bar mt-2">
								<div
									class="progress-indicator"
									style={`transform: translateX(-${100 - slice.value}%)`}
								></div>
							</div>
						</div>
					{/each}
				</div>
			</div>
		</div>
	</div>
</section>

<style>
	.stats-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
		gap: 2rem;
	}

	.stat-item {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.dashboard-card {
		border: 1px solid rgba(212, 175, 55, 0.2);
		background: linear-gradient(135deg, #1a1f2e 0%, #141a24 100%);
		padding: 2rem;
		box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);
		border-radius: 16px;
		backdrop-filter: blur(10px);
	}

	.card-box {
		border-radius: 12px;
		border: 1px solid rgba(212, 175, 55, 0.1);
	}

	@media (max-width: 1024px) {
		.stats-grid {
			grid-template-columns: 1fr;
		}
	}

	@media (max-width: 768px) {
		.stats-grid {
			grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
			gap: 1.5rem;
		}

		.dashboard-card {
			padding: 1.5rem;
		}
	}
</style>
