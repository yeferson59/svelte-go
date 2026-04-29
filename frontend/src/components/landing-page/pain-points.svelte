<script lang="ts">
	const painPoints = [
		{
			icon: '📦',
			title: 'Inversiones dispersas',
			description: 'Acciones en Bono, ETFs en Interactive Brokers, cripto en Kraken, fondos en tu banco local. Todo fragmentado.'
		},
		{
			icon: '👁️',
			title: 'Cero visibilidad del todo',
			description: 'Tienes que revisar cada plataforma, cada login diferente, cada interfaz extraña. Agotador.'
		},
		{
			icon: '❓',
			title: 'Decisiones en la incertidumbre',
			description: 'No ves claramente qué te está ganando dinero, qué pesa demasiado o cómo está realmente compuesto.'
		},
		{
			icon: '📊',
			title: 'Trabajo manual y errores',
			description: 'Excel con actualizaciones manuales. Datos desactualizados. Fórmulas rotas. Frustración.'
		}
	];
</script>

<section id="caracteristicas" class="mx-auto w-full max-w-7xl px-6 py-20">
	<div class="mb-12 space-y-6 text-center">
		<p class="text-xs font-600 tracking-widest uppercase text-[#d4af37]">El problema</p>
		<h2 class="font-display text-4xl font-bold leading-tight text-[#ffffff]">
			¿Tu dinero está en 5 lugares diferentes y no sabes ni cuánto has ganado?
		</h2>
		<p class="mx-auto max-w-2xl text-lg text-[#a0a0a0]">
			Millones de inversionistas viven esta frustración. Múltiples plataformas. Cero visibilidad. Sin control real.
		</p>
	</div>

	<div class="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
		{#each painPoints as pain (pain.title)}
			<div class="pain-card">
				<div class="text-3xl mb-4">{pain.icon}</div>
				<h3 class="mb-3 font-display text-lg font-bold text-[#ffffff]">{pain.title}</h3>
				<p class="text-sm text-[#a0a0a0]">{pain.description}</p>
			</div>
		{/each}
	</div>
</section>

<style>
	.pain-card {
		border: 1px solid rgba(212, 175, 55, 0.15);
		background: linear-gradient(135deg, rgba(26, 31, 46, 0.6) 0%, rgba(20, 26, 36, 0.6) 100%);
		padding: 2rem;
		border-radius: 12px;
		backdrop-filter: blur(10px);
		transition: all 0.3s ease;
	}

	.pain-card:hover {
		border-color: rgba(212, 175, 55, 0.3);
		background: linear-gradient(135deg, rgba(26, 31, 46, 0.8) 0%, rgba(20, 26, 36, 0.8) 100%);
		transform: translateY(-4px);
	}

	@media (max-width: 768px) {
		.pain-card {
			padding: 1.5rem;
		}
	}
</style>
