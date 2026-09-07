<script lang="ts">
	/*
	 * Maqueta de `/dashboard`.
	 *
	 * Enseñaba dos tarjetas —patrimonio y gráfica— con antetítulo en versalitas
	 * monoespaciadas, que es la pantalla anterior al rediseño. La de hoy no tiene
	 * tarjetas: la cifra va suelta y en grande, y debajo la misma cuenta
	 * desglosada en dos secciones separadas por un filete —dónde está el dinero y
	 * cómo ha ido en el tiempo—, con el extracto de movimientos al lado.
	 */
	import TourGrowth from './tour-growth.svelte';
	import TourActivity from './tour-activity.svelte';
	import { TOUR_BREAKDOWN, TOUR_CUTS, TOUR_GROWTH_MARKET, TOUR_NET_WORTH } from '../product-tour';

	/*
	 * La curva de al lado de la cifra: la serie del patrimonio, sin ejes y con
	 * ventana propia. Dibujada sobre la escala de la gráfica grande —que arranca
	 * muy por debajo del primer punto— quedaba una línea casi plana pegada al
	 * borde de arriba, que es justo lo que la miniatura no debe decir.
	 */
	const sparkW = 100;
	const sparkH = 26;
	const sparkPad = 2;
	const sparkMin = Math.min(...TOUR_GROWTH_MARKET);
	const sparkSpan = Math.max(...TOUR_GROWTH_MARKET) - sparkMin;
	const sparkPoints = TOUR_GROWTH_MARKET.map((v, i) => ({
		x: (i / (TOUR_GROWTH_MARKET.length - 1)) * sparkW,
		y: sparkH - sparkPad - ((v - sparkMin) / sparkSpan) * (sparkH - sparkPad * 2)
	}));
	const spark = sparkPoints.map((p) => `${p.x.toFixed(2)},${p.y.toFixed(2)}`).join(' ');
	const sparkEnd = sparkPoints[sparkPoints.length - 1];
</script>

<div class="summary">
	<!-- La cifra: es lo único ruidoso de la página, y sale una sola vez. -->
	<div class="mk-sec headline">
		<div>
			<p class="label">{TOUR_NET_WORTH.label}</p>
			<p class="amount">{TOUR_NET_WORTH.total}</p>
			<p class="delta mk-up">{TOUR_NET_WORTH.delta}</p>
			<p class="meta">{TOUR_NET_WORTH.meta}</p>
		</div>

		<div class="aside">
			<!-- Es un desplegable, no una etiqueta: la moneda en la que se miran los
			     importes de la pantalla se elige aquí, entre once. -->
			<span class="currency">
				{TOUR_NET_WORTH.currency}
				<svg width="9" height="9" viewBox="0 0 24 24" fill="none" stroke="#888" stroke-width="2">
					<polyline points="6 9 12 15 18 9" />
				</svg>
			</span>

			<svg class="spark" viewBox="0 0 {sparkW} {sparkH}" preserveAspectRatio="none">
				<defs>
					<linearGradient id="tourSpark" x1="0" y1="0" x2="0" y2="1">
						<stop offset="0%" stop-color="rgba(212,145,42,0.18)" />
						<stop offset="100%" stop-color="rgba(212,145,42,0)" />
					</linearGradient>
				</defs>
				<polygon points="{spark} {sparkW},{sparkH} 0,{sparkH}" fill="url(#tourSpark)" />
				<polyline
					points={spark}
					fill="none"
					stroke="var(--amber)"
					stroke-width="1.2"
					stroke-linecap="round"
					stroke-linejoin="round"
					vector-effect="non-scaling-stroke"
				/>
				<!--
					El punto de hoy es un trazo de longitud cero con el extremo redondo,
					no un `<circle>`: el lienzo se estira solo en horizontal, y un círculo
					saldría ovalado. Un trazo que no escala sale redondo mida lo que mida
					la caja. Es lo que hace el `Sparkline` del panel.
				-->
				<line
					x1={sparkEnd.x}
					y1={sparkEnd.y}
					x2={sparkEnd.x}
					y2={sparkEnd.y}
					stroke="var(--amber)"
					stroke-width="3.5"
					stroke-linecap="round"
					vector-effect="non-scaling-stroke"
				/>
			</svg>

			<p class="since">{TOUR_NET_WORTH.since}</p>
		</div>
	</div>

	<!-- Dónde está el dinero: el mismo patrimonio leído por plataforma, por
	     portafolio o por clase de activo. -->
	<div class="mk-sec">
		<div class="head">
			<span class="mk-h2">Dónde está</span>
			<span class="mk-tabs">
				{#each TOUR_CUTS as cut, i (cut)}
					<span class="mk-tab" class:on={i === 0}>{cut}</span>
				{/each}
			</span>
		</div>

		<table class="mk-table">
			<thead>
				<tr>
					<th>Plataforma</th>
					<th class="col-bar">Participación</th>
					<th class="mk-num">Valor</th>
					<th class="mk-num">Rendimiento</th>
				</tr>
			</thead>
			<tbody>
				{#each TOUR_BREAKDOWN as row (row.name)}
					<tr>
						<th>
							<span class="mk-name">{row.name}</span>
							<span class="mk-detail">{row.detail}</span>
						</th>
						<td class="col-bar">
							<span class="mk-track"><span class="mk-fill" style="width:{row.share}%"></span></span>
						</td>
						<td class="mk-num">{row.value}</td>
						<td class="mk-num" class:mk-up={row.up} class:mk-dn={!row.up}>{row.gain}</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>

	<!-- Cómo ha ido en el tiempo, y qué ha pasado esta semana. -->
	<div class="mk-sec lower">
		<TourGrowth />
		<TourActivity />
	</div>
</div>

<style>
	/*
	 * El azul frío del capital invertido, frente al ámbar del valor de mercado:
	 * el dinero que pusiste se lee frío y lo que el mercado hizo con él, cálido.
	 * Es el `--cost` del panel, que solo está declarado dentro de su chrome.
	 */
	.summary {
		--cost: #73819c;
	}

	.headline {
		display: grid;
		grid-template-columns: minmax(0, 1fr) 140px;
		align-items: start;
		gap: 20px;
	}

	/* Nombra la cifra que va justo debajo, así que no es una etiqueta de adorno. */
	.label {
		font-size: 10.5px;
		color: var(--text-muted);
	}

	.amount {
		margin-top: 6px;
		font-family: var(--font-mono);
		font-size: clamp(24px, 3.6vw, 34px);
		font-weight: 600;
		line-height: 1;
		letter-spacing: -0.035em;
		color: var(--text);
	}

	.delta {
		margin-top: 8px;
		font-size: 10.5px;
	}

	.meta {
		margin-top: 3px;
		font-size: 9.5px;
		color: var(--text-dim);
	}

	.aside {
		display: flex;
		flex-direction: column;
		align-items: flex-end;
		gap: 10px;
	}

	.currency {
		display: inline-flex;
		align-items: center;
		gap: 5px;
		padding: 3px 7px 3px 8px;
		border: 1px solid var(--border);
		border-radius: 6px;
		background: rgba(255, 255, 255, 0.03);
		font-family: var(--font-mono);
		font-size: 9px;
		font-weight: 600;
		color: var(--amber-light);
	}

	.spark {
		display: block;
		width: 100%;
		height: 34px;
	}

	.since {
		font-size: 8.5px;
		color: var(--text-dim);
	}

	.head {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
		margin-bottom: 12px;
	}

	.col-bar {
		width: 36%;
	}

	/* La gráfica manda, así que se lleva el ancho que sobra; el extracto tiene
	   uno fijo porque son cifras cortas y crecer no lo mejora. */
	.lower {
		display: grid;
		grid-template-columns: minmax(0, 1fr) 170px;
		gap: 26px;
	}

	@media (max-width: 1040px) {
		.lower {
			grid-template-columns: minmax(0, 1fr);
			gap: 22px;
		}
	}

	@media (max-width: 620px) {
		.headline {
			grid-template-columns: minmax(0, 1fr);
			gap: 14px;
		}

		.aside {
			align-items: flex-start;
		}

		.col-bar {
			display: none;
		}
	}
</style>
