<script lang="ts">
	/*
	 * La gráfica de crecimiento de la maqueta del resumen: los dos conmutadores,
	 * las cuatro medidas y las dos series.
	 *
	 * Sale de `tour-view-summary.svelte` por el mismo motivo que en el panel
	 * `portfolio-growth` está aparte de la página: eran la mitad del archivo, y
	 * el reparto del patrimonio se perdía entre los ejes.
	 *
	 * Interno de la feature.
	 */
	import {
		TOUR_GROWTH_COST,
		TOUR_GROWTH_DATES,
		TOUR_GROWTH_MARKET,
		TOUR_GROWTH_SCALE,
		TOUR_GROWTH_STATS
	} from '../product-tour';

	const W = 100;
	const H = 34;

	const { min, max, step } = TOUR_GROWTH_SCALE;

	/** Dónde cae un importe en el lienzo, de 0 arriba a 1 abajo. */
	const at = (value: number) => (max - value) / (max - min);

	/** Los puntos de una serie sobre el lienzo, a partir de sus importes. */
	const plot = (series: number[]) =>
		series
			.map((v, i) => `${((i / (series.length - 1)) * W).toFixed(2)},${(at(v) * H).toFixed(2)}`)
			.join(' ');

	const market = plot(TOUR_GROWTH_MARKET);
	const cost = plot(TOUR_GROWTH_COST);
	const area = `0,${H} ${market} ${W},${H}`;

	/*
	 * Las marcas del eje, de arriba abajo y con su altura. El importe va
	 * abreviado, que es lo que cabe entre marcas y lo que escribe el panel.
	 */
	const ticks = Array.from({ length: (max - min) / step + 1 }, (_, i) => {
		const value = max - i * step;
		return { label: `$${value}k`, at: at(value) };
	});

	const periods = ['1M', '3M', '6M', '1Y', 'Todo'];
</script>

<div class="growth">
	<div class="head">
		<span class="mk-h2">Crecimiento</span>
		<!-- Los dos grupos comparten forma: son la misma clase de control, y verlos
		     distintos haría pensar que uno de ellos navega a otro sitio. -->
		<span class="controls">
			<span class="mk-tabs">
				<span class="mk-tab on">Valor</span>
				<span class="mk-tab">%</span>
			</span>
			<span class="mk-tabs">
				{#each periods as p (p)}
					<span class="mk-tab" class:on={p === 'Todo'}>{p}</span>
				{/each}
			</span>
		</span>
	</div>

	<div class="stats">
		{#each TOUR_GROWTH_STATS as stat (stat.label)}
			<div>
				<span class="stat-label">{stat.label}</span>
				<span
					class="stat-value"
					class:mk-up={stat.tone === 'up'}
					class:mk-amber={stat.tone === 'amber'}
				>
					{stat.value}
				</span>
			</div>
		{/each}
	</div>

	<!-- Lo que enseña la gráfica mientras nadie la señala. Es una instrucción
	     porque la de verdad responde al ratón y al teclado. -->
	<p class="readout">
		Pasa el cursor por la gráfica —o enfócala y usa las flechas— para ver el detalle de cada día.
	</p>

	<div class="plot">
		<div class="y-axis">
			{#each ticks as tick (tick.label)}
				<span style="top: {(tick.at * 100).toFixed(2)}%">{tick.label}</span>
			{/each}
		</div>

		<svg class="chart" viewBox="0 0 {W} {H}" preserveAspectRatio="none">
			<defs>
				<linearGradient id="tourGrowth" x1="0" y1="0" x2="0" y2="1">
					<stop offset="0%" stop-color="rgba(212,145,42,0.22)" />
					<stop offset="100%" stop-color="rgba(212,145,42,0)" />
				</linearGradient>
			</defs>

			<!-- Una línea por marca del eje y a su misma altura: eran tres, repartidas
			     a ojo, así que ninguna decía de qué importe hablaba. -->
			{#each ticks as tick (tick.label)}
				<line
					x1="0"
					y1={tick.at * H}
					x2={W}
					y2={tick.at * H}
					stroke="rgba(255,255,255,0.05)"
					stroke-width="0.5"
					vector-effect="non-scaling-stroke"
				/>
			{/each}

			<polygon points={area} fill="url(#tourGrowth)" />
			<polyline
				points={cost}
				fill="none"
				stroke="var(--cost)"
				stroke-width="0.9"
				stroke-dasharray="3 2"
				vector-effect="non-scaling-stroke"
			/>
			<polyline
				points={market}
				fill="none"
				stroke="var(--amber)"
				stroke-width="1.1"
				stroke-linecap="round"
				stroke-linejoin="round"
				vector-effect="non-scaling-stroke"
			/>
		</svg>
	</div>

	<div class="x-axis">
		{#each TOUR_GROWTH_DATES as date (date)}
			<span>{date}</span>
		{/each}
	</div>

	<div class="legend">
		<span><i class="ln amber"></i>Valor de mercado</span>
		<span><i class="ln cost"></i>Capital invertido</span>
	</div>
</div>

<style>
	.head {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
		margin-bottom: 12px;
	}

	.controls {
		display: flex;
		gap: 5px;
	}

	.stats {
		display: grid;
		grid-template-columns: repeat(4, minmax(0, 1fr));
		gap: 10px;
		margin-bottom: 14px;
	}

	.stat-label {
		display: block;
		font-size: 8.5px;
		color: var(--text-dim);
	}

	.stat-value {
		display: block;
		margin-top: 3px;
		font-family: var(--font-mono);
		font-size: 11.5px;
		font-weight: 600;
		font-variant-numeric: tabular-nums;
		letter-spacing: -0.01em;
		color: var(--text);
	}

	/* El carril de las marcas del eje y el lienzo; los pies de abajo se alinean
	   con el segundo, de ahí sus 32 px de sangría. */
	.plot {
		display: grid;
		grid-template-columns: 26px minmax(0, 1fr);
		gap: 6px;
	}

	/* Cada marca en su altura exacta, no repartidas a partes iguales: así caen
	   sobre su propia línea del lienzo. */
	.y-axis {
		position: relative;
		font-family: var(--font-mono);
		font-size: 7.5px;
		color: var(--text-dim);
	}

	.y-axis span {
		position: absolute;
		right: 0;
		transform: translateY(-50%);
		white-space: nowrap;
	}

	.chart {
		display: block;
		width: 100%;
		height: 116px;
	}

	/* Alto fijo, como en el panel: el detalle aparece y desaparece sin mover la
	   gráfica de sitio. */
	.readout {
		min-height: 20px;
		margin-bottom: 4px;
		font-size: 9px;
		color: var(--text-muted);
	}

	.x-axis {
		display: flex;
		justify-content: space-between;
		margin: 5px 0 0 32px;
		font-family: var(--font-mono);
		font-size: 7.5px;
		color: var(--text-dim);
	}

	.legend {
		display: flex;
		gap: 16px;
		margin: 10px 0 0 32px;
		font-size: 8.5px;
		color: var(--text-dim);
	}

	.legend span {
		display: flex;
		align-items: center;
		gap: 6px;
	}

	.ln {
		width: 14px;
		height: 0;
	}

	.ln.amber {
		border-top: 2px solid var(--amber);
	}

	.ln.cost {
		border-top: 1.5px dashed var(--cost);
	}

	@media (max-width: 620px) {
		.stats {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}
	}
</style>
