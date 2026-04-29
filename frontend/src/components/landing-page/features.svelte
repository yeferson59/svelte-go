<script lang="ts">
	const features = [
		{
			icon: '💰',
			name: 'Patrimonio Total',
			description: 'Suma consolidada y en tiempo real de todas tus inversiones, sin importar plataforma. Es el número que importa.'
		},
		{
			icon: '📊',
			name: 'Composición del portafolio',
			description: 'Visualiza qué porcentaje representa cada activo respecto a tu patrimonio total. Identifica concentración y balance.'
		},
		{
			icon: '🏷️',
			name: 'Diversificación por categoría',
			description: 'Agrupa tus inversiones por tipo (Acciones, ETFs, Cripto, Bonos, Fondos, Efectivo) y monitorea tu diversificación.'
		},
		{
			icon: '📈',
			name: 'Rendimiento individual',
			description: 'Cada inversión muestra su retorno (%) desde que fue ingresada, más su valor en ganancia o pérdida absoluta.'
		},
		{
			icon: '💹',
			name: 'P&L (Ganancia/Pérdida)',
			description: 'Resumen claro de cuánto has ganado o perdido en total y por activo. Diferenciación visual inmediata.'
		},
		{
			icon: '📉',
			name: 'Análisis avanzados',
			description: 'Comparativa vs. benchmark, histórico patrimonial, alertas de rebalanceo y mucho más.'
		}
	];
</script>

<section id="caracteristicas-principales" class="mx-auto w-full max-w-7xl px-6 py-20">
	<div class="mb-16 space-y-4 text-center">
		<p class="text-xs font-600 tracking-widest uppercase text-[#d4af37]">Funcionalidades principales</p>
		<h2 class="font-display text-4xl font-bold leading-tight text-[#ffffff]">
			Controla tu patrimonio con precisión
		</h2>
		<p class="mx-auto max-w-2xl text-lg text-[#a0a0a0]">
			Todas las métricas que necesitas para tomar decisiones inversoras inteligentes.
		</p>
	</div>

	<div class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
		{#each features as feature (feature.name)}
			<div class="feature-card">
				<div class="feature-icon">{feature.icon}</div>
				<h3 class="font-display mb-3 text-lg font-bold text-[#ffffff]">{feature.name}</h3>
				<p class="text-sm text-[#a0a0a0]">{feature.description}</p>
			</div>
		{/each}
	</div>
</section>

<style>
	.feature-card {
		border: 1px solid rgba(212, 175, 55, 0.15);
		background: linear-gradient(135deg, rgba(26, 31, 46, 0.6) 0%, rgba(20, 26, 36, 0.6) 100%);
		padding: 2rem;
		border-radius: 16px;
		backdrop-filter: blur(10px);
		transition: all 0.3s ease;
	}

	.feature-card:hover {
		border-color: rgba(212, 175, 55, 0.4);
		background: linear-gradient(135deg, rgba(26, 31, 46, 0.8) 0%, rgba(20, 26, 36, 0.8) 100%);
		transform: translateY(-6px);
		box-shadow: 0 12px 40px rgba(212, 175, 55, 0.15);
	}

	.feature-icon {
		font-size: 2.5rem;
		margin-bottom: 1rem;
		display: inline-block;
	}

	@media (max-width: 768px) {
		.feature-card {
			padding: 1.5rem;
		}
	}
</style>
