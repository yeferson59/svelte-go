<script lang="ts">
	import { onMount } from 'svelte';
	import MetricsChart from './metrics-chart.svelte';

	let metricsEl: HTMLElement;

	onMount(() => {
		const io = new IntersectionObserver(
			(entries) => {
				entries.forEach((entry) => {
					if (entry.isIntersecting) {
						metricsEl.classList.add('in');
						io.unobserve(entry.target);
					}
				});
			},
			{ threshold: 0.15 }
		);
		io.observe(metricsEl);
		return () => io.disconnect();
	});
</script>

<section class="band">
	<div class="wrap block">
		<div class="metrics-head reveal">
			<div class="metrics-head-text">
				<div class="eyebrow">Métricas que importan</div>
				<h2 class="sec-title">Mira crecer tu patrimonio con claridad</h2>
				<p class="sec-desc">
					Valor, peso, rendimiento y distribución de cada portafolio, en una vista hecha para
					entender de un vistazo.
				</p>
			</div>
			<!-- Los mismos rangos que ofrece el panel real. Eran «1M · 6M · 1A ·
			     Todo», que es el juego de antes del rediseño: hoy son cinco y el año
			     se llama «1Y». -->
			<div class="ranges" aria-hidden="true">
				<span>1M</span>
				<span>3M</span>
				<span>6M</span>
				<span class="on">1Y</span>
				<span>Todo</span>
			</div>
		</div>

		<div class="metrics-wrap reveal" bind:this={metricsEl}>
			<div class="panel-top">
				<div>
					<div class="panel-label">Patrimonio total · últimos 12 meses</div>
					<div class="panel-value">
						<b>$248,500</b>
						<em>+12,4%</em>
					</div>
				</div>
				<div class="legend">
					<span><i class="ln amber"></i>Valor de mercado</span>
					<span><i class="ln gray"></i>Capital invertido</span>
				</div>
			</div>

			<MetricsChart />

			<div class="metric-row">
				<div class="mcell">
					<div class="lbl">Capital invertido</div>
					<div class="val">$221,100</div>
				</div>
				<div class="mcell">
					<div class="lbl">Rendimiento</div>
					<div class="val up">+12,4%</div>
				</div>
				<div class="mcell">
					<div class="lbl">Peso mayor</div>
					<div class="val">38%<span class="delta">Acciones</span></div>
				</div>
				<div class="mcell">
					<div class="lbl">Meses en positivo</div>
					<div class="val">10 de 12</div>
				</div>
			</div>
		</div>
	</div>
</section>

<style>
	.metrics-head {
		display: flex;
		align-items: flex-end;
		justify-content: space-between;
		gap: 56px;
		margin-bottom: 44px;
	}

	.metrics-head-text {
		max-width: 620px;
	}

	.ranges {
		display: flex;
		gap: 2px;
		flex-shrink: 0;
		padding: 3px;
		border: 1px solid var(--border);
		border-radius: 8px;
		background: rgba(255, 255, 255, 0.03);
	}

	.ranges span {
		padding: 6px 12px;
		border-radius: 5px;
		font-family: var(--font-mono);
		font-size: 10.5px;
		font-weight: 600;
		letter-spacing: 0.06em;
		color: var(--text-dim);
	}

	.ranges .on {
		background: rgba(212, 145, 42, 0.18);
		color: var(--amber-light);
	}

	.metrics-wrap {
		border-radius: 14px;
		overflow: hidden;
		border: 1px solid var(--border-strong);
		background: var(--surface);
	}

	.panel-top {
		display: flex;
		align-items: flex-end;
		justify-content: space-between;
		gap: 28px;
		flex-wrap: wrap;
		padding: 28px 36px;
		border-bottom: 1px solid var(--border);
	}

	.panel-label {
		font-family: var(--font-mono);
		font-size: 10px;
		font-weight: 500;
		letter-spacing: 0.16em;
		text-transform: uppercase;
		color: var(--text-dim);
	}

	.panel-value {
		display: flex;
		align-items: baseline;
		gap: 12px;
		margin-top: 8px;
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
	}

	.panel-value b {
		font-weight: 600;
		font-size: clamp(24px, 3.2vw, 32px);
		line-height: 1;
		letter-spacing: -0.03em;
	}

	.panel-value em {
		font-style: normal;
		font-size: 14px;
		font-weight: 600;
		color: var(--green);
	}

	.legend {
		display: flex;
		gap: 20px;
		font-family: var(--font-mono);
		font-size: 10px;
		letter-spacing: 0.04em;
		color: var(--text-dim);
	}

	.legend span {
		display: flex;
		align-items: center;
		gap: 7px;
	}

	.ln {
		width: 18px;
		height: 0;
	}

	.ln.amber {
		border-top: 2px solid var(--amber);
	}

	.ln.gray {
		border-top: 1.5px dashed rgba(236, 234, 229, 0.28);
	}

	/* Los chips dejan de flotar junto al titular y cierran el panel en una fila. */
	.metric-row {
		display: grid;
		grid-template-columns: repeat(4, minmax(0, 1fr));
		gap: 1px;
		background: var(--border);
		border-top: 1px solid var(--border);
	}

	.mcell {
		padding: 20px 28px;
		background: #0b0c0d;
	}

	.mcell:first-child {
		padding-left: 36px;
	}

	.lbl {
		font-family: var(--font-mono);
		font-size: 10px;
		font-weight: 500;
		letter-spacing: 0.14em;
		text-transform: uppercase;
		color: var(--text-dim);
	}

	.val {
		display: flex;
		align-items: baseline;
		gap: 8px;
		margin-top: 7px;
		font-family: var(--font-mono);
		font-weight: 600;
		font-size: 21px;
		letter-spacing: -0.02em;
		font-variant-numeric: tabular-nums;
		color: var(--text);
	}

	.val.up {
		color: var(--green);
	}

	.delta {
		font-size: 11px;
		font-weight: 600;
		color: var(--amber);
	}

	@media (max-width: 1040px) {
		.metrics-head {
			flex-direction: column;
			align-items: flex-start;
			gap: 24px;
		}
	}

	@media (max-width: 860px) {
		.metric-row {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}
		.mcell:first-child {
			padding-left: 28px;
		}
	}

	@media (max-width: 620px) {
		.panel-top {
			padding-left: 20px;
			padding-right: 20px;
		}
		.mcell {
			padding: 16px 20px;
		}
		.mcell:first-child {
			padding-left: 20px;
		}
	}
</style>
