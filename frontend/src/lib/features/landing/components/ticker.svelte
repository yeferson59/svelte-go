<script lang="ts">
	/*
	 * Cinta de cotizaciones. La lista va duplicada para que la animación pueda
	 * dar la vuelta desplazando exactamente un -50%.
	 *
	 * Los importes van en dólares y escritos como los escribe la aplicación
	 * —`formatCurrency` da a cada moneda su locale, y el dólar se lee en en-US—.
	 * Estaban a la española y sin ponerse de acuerdo entre ellos: «$67.240» al
	 * lado de «$189,45», que son el mismo separador diciendo cosas distintas.
	 * Los porcentajes sí van en es-CO, con coma, y su negativo lleva guion, que
	 * es el signo que saca `Intl` en ese idioma.
	 */
	const tickers = [
		{ sym: 'BTC', val: '$67,240', delta: '+2,1%', dir: 'up' },
		{ sym: 'ETH', val: '$3,482', delta: '-0,4%', dir: 'dn' },
		{ sym: 'AAPL', val: '$189.45', delta: '+0,8%', dir: 'up' },
		{ sym: 'S&P500', val: '5,234', delta: '+0,6%', dir: 'up' },
		{ sym: 'EUR/USD', val: '1.0842', delta: '+0,1%', dir: 'up' },
		{ sym: 'ORO', val: '$2,387', delta: '+0,3%', dir: 'up' },
		{ sym: 'MSFT', val: '$415', delta: '+1,2%', dir: 'up' },
		{ sym: 'NVDA', val: '$876', delta: '+3,8%', dir: 'up' }
	];
	const loop = [...tickers, ...tickers];
</script>

<div class="ticker-banner" aria-hidden="true">
	<div class="ticker-track">
		{#each loop as t, i (i)}
			<span>{t.sym} <em>{t.val}</em> <b class={t.dir}>{t.delta}</b></span>
			<span class="sep">·</span>
		{/each}
	</div>
</div>

<style>
	.ticker-banner {
		position: relative;
		z-index: 10;
		background: rgba(255, 255, 255, 0.018);
		border-bottom: 1px solid var(--border);
		overflow: hidden;
		padding: 7px 0;
		font-family: var(--font-mono);
		font-size: 11px;
		letter-spacing: 0.03em;
		color: var(--text-muted);
	}
	.ticker-track {
		display: flex;
		gap: 32px;
		align-items: center;
		white-space: nowrap;
		width: max-content;
		animation: ticker-scroll 40s linear infinite;
	}
	.ticker-track em {
		font-style: normal;
		color: var(--text);
	}
	.ticker-track b {
		font-weight: 600;
	}
	.ticker-track .up {
		color: var(--green);
	}
	.ticker-track .dn {
		color: #e05a5a;
	}
	.ticker-track .sep {
		color: var(--text-dim);
		opacity: 0.5;
	}
	@keyframes ticker-scroll {
		0% {
			transform: translateX(0);
		}
		100% {
			transform: translateX(-50%);
		}
	}
	@media (prefers-reduced-motion: reduce) {
		.ticker-track {
			animation: none;
		}
	}
</style>
