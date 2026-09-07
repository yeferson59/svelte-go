<script lang="ts">
	/*
	 * Gráfico de la sección "Métricas".
	 *
	 * Vive aparte de `metrics.svelte` porque el trazado, los ejes y la animación
	 * de entrada ocupan tanto como el resto de la sección junta. La clase `.in`
	 * que dispara la animación la pone el observador del componente padre sobre
	 * `.metrics-wrap`, de ahí los selectores `:global`.
	 */
	import {
		METRICS_MONTHS,
		METRICS_MARKET_VALUE,
		METRICS_INVESTED,
		METRICS_SCALE,
		METRICS_TICKS,
		chartPoints,
		chartY,
		type ChartBox
	} from '../landing';

	/* Área de trazado dentro del viewBox de 1040×300: el margen izquierdo deja
	   sitio a las etiquetas del eje y el inferior, a los meses. */
	const BOX: ChartBox = { left: 62, right: 1010, top: 30, bottom: 250 };

	const marketLine = chartPoints(METRICS_MARKET_VALUE, METRICS_SCALE, BOX);
	const investedLine = chartPoints(METRICS_INVESTED, METRICS_SCALE, BOX);
	const marketArea = `${BOX.left},${BOX.bottom} ${marketLine} ${BOX.right},${BOX.bottom}`;
	const ticks = METRICS_TICKS.map((value) => ({ value, y: chartY(value, METRICS_SCALE, BOX) }));
	const monthX = METRICS_MONTHS.map(
		(_, i) => BOX.left + ((BOX.right - BOX.left) * i) / (METRICS_MONTHS.length - 1)
	);
	const lastX = BOX.right;
	const lastY = chartY(METRICS_MARKET_VALUE[METRICS_MARKET_VALUE.length - 1], METRICS_SCALE, BOX);
</script>

<div class="chart-zone">
	<svg class="chart-svg" viewBox="0 0 1040 300" preserveAspectRatio="none" aria-hidden="true">
		<defs>
			<linearGradient id="metricsFill" x1="0" y1="0" x2="0" y2="1">
				<stop offset="0%" stop-color="rgba(212,145,42,0.2)" />
				<stop offset="100%" stop-color="rgba(212,145,42,0)" />
			</linearGradient>
		</defs>

		<g class="chart-grid">
			{#each ticks as tick (tick.value)}
				<line x1={BOX.left} y1={tick.y} x2={BOX.right} y2={tick.y} />
			{/each}
			<line class="axis" x1={BOX.left} y1={BOX.bottom} x2={BOX.right} y2={BOX.bottom} />
		</g>

		<g class="tick-label">
			{#each ticks as tick (tick.value)}
				<text x="0" y={tick.y + 3.5}>${tick.value}K</text>
			{/each}
		</g>

		<polygon class="chart-area" points={marketArea} />
		<polyline class="chart-cost" points={investedLine} />
		<polyline class="chart-line" points={marketLine} />

		<g class="chart-dot">
			<circle class="pulse-ring" cx={lastX} cy={lastY} r="6" fill="rgba(212,145,42,0.35)" />
			<circle cx={lastX} cy={lastY} r="5" fill="var(--amber)" stroke="#08090a" stroke-width="2.5" />
		</g>

		<g class="month-label">
			{#each METRICS_MONTHS as month, i (month)}
				<text x={monthX[i]} y="278">{month}</text>
			{/each}
		</g>
	</svg>

	<div class="chart-flag">
		<div class="t">Este año</div>
		<div class="v">+ $27,400</div>
	</div>
</div>

<p class="axis-note">Eje vertical recortado desde $170,000 para que la variación sea legible.</p>

<style>
	.chart-zone {
		position: relative;
		padding: 24px 36px 0;
	}

	.chart-svg {
		display: block;
		width: 100%;
		height: 300px;
	}

	.chart-grid line {
		stroke: rgba(255, 255, 255, 0.05);
		stroke-width: 1;
	}

	.chart-grid .axis {
		stroke: rgba(255, 255, 255, 0.09);
	}

	.tick-label text,
	.month-label text {
		fill: var(--text-dim);
		font-family: var(--font-mono);
		font-size: 10px;
		letter-spacing: 0.04em;
	}

	.month-label text {
		text-anchor: middle;
		letter-spacing: 0.1em;
	}

	.chart-area {
		fill: url(#metricsFill);
		opacity: 0;
		transition: opacity 1.2s ease 0.5s;
	}

	:global(.metrics-wrap.in) .chart-area {
		opacity: 1;
	}

	.chart-cost {
		fill: none;
		stroke: rgba(236, 234, 229, 0.26);
		stroke-width: 1.4;
		stroke-dasharray: 5 4;
		opacity: 0;
		transition: opacity 0.8s ease 0.9s;
	}

	:global(.metrics-wrap.in) .chart-cost {
		opacity: 1;
	}

	.chart-line {
		fill: none;
		stroke: var(--amber);
		stroke-width: 2;
		stroke-linecap: round;
		stroke-linejoin: round;
		stroke-dasharray: 1600;
		stroke-dashoffset: 1600;
		transition: stroke-dashoffset 1.8s cubic-bezier(0.4, 0, 0.2, 1) 0.2s;
	}

	:global(.metrics-wrap.in) .chart-line {
		stroke-dashoffset: 0;
	}

	.chart-dot {
		opacity: 0;
		transition: opacity 0.4s ease 1.6s;
	}

	:global(.metrics-wrap.in) .chart-dot {
		opacity: 1;
	}

	.pulse-ring {
		transform-box: fill-box;
		transform-origin: center;
		animation: dotpulse 2.2s ease-out infinite;
	}

	@keyframes dotpulse {
		0% {
			transform: scale(0.6);
			opacity: 0.7;
		}
		70% {
			transform: scale(2.4);
			opacity: 0;
		}
		100% {
			opacity: 0;
		}
	}

	.chart-flag {
		position: absolute;
		top: 42px;
		right: 8%;
		padding: 9px 13px;
		border-radius: 8px;
		background: rgba(8, 9, 10, 0.9);
		border: 1px solid rgba(34, 201, 126, 0.28);
		backdrop-filter: blur(8px);
		opacity: 0;
		transform: translateY(6px);
		transition:
			opacity 0.5s ease 1.5s,
			transform 0.5s ease 1.5s;
	}

	:global(.metrics-wrap.in) .chart-flag {
		opacity: 1;
		transform: none;
	}

	.chart-flag .t {
		font-family: var(--font-mono);
		font-size: 10px;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		color: var(--text-dim);
	}

	.chart-flag .v {
		font-family: var(--font-mono);
		font-weight: 600;
		font-size: 17px;
		color: var(--green);
		margin-top: 3px;
		font-variant-numeric: tabular-nums;
	}

	.axis-note {
		margin: 0;
		padding: 0 36px 18px;
		font-family: var(--font-mono);
		font-size: 9.5px;
		letter-spacing: 0.06em;
		color: var(--text-dim);
	}

	@media (max-width: 620px) {
		.chart-zone,
		.axis-note {
			padding-left: 20px;
			padding-right: 20px;
		}
		.chart-svg {
			height: 220px;
		}
		.chart-flag {
			right: 6%;
			top: 34px;
		}
	}
</style>
