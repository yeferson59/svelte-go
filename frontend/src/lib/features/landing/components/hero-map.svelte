<script lang="ts">
	/*
	 * "Tu mapa": maqueta decorativa del hero (`aria-hidden`).
	 *
	 * Antes era una tarjeta de portafolio genérica, la misma que se ve en
	 * cualquier app de finanzas. El titular promete un mapa y lo que distingue a
	 * Finexia es juntar plataformas sueltas bajo los portafolios que tú defines,
	 * así que eso es lo que dibuja: plataformas → conectores → portafolios →
	 * patrimonio. Las cifras son las de `product-tour.ts`, para que la landing
	 * cuente el mismo ejemplo de principio a fin.
	 */
	import { TOUR_BREAKDOWN, TOUR_PORTFOLIOS } from '../product-tour';

	/*
	 * Las plataformas salen del reparto del recorrido y no de una lista propia:
	 * escritas dos veces, el hero enseñaba cinco marcas y el reparto del panel,
	 * dos pantallas más abajo, otras tres. La geometría de los hilos de abajo da
	 * por hechas cinco, así que si cambia el número hay que rehacerlos.
	 */
	const platforms = TOUR_BREAKDOWN.map((row) => row.name);

	/** Peso de cada portafolio sobre el patrimonio total, y color de su barra. */
	const weights = [
		{ pct: 53, color: 'var(--amber)' },
		{ pct: 28, color: 'var(--green)' },
		{ pct: 19, color: '#6b8cef' }
	];
</script>

<div class="map" aria-hidden="true">
	<div class="map-card reveal">
		<div class="map-top">
			<div>
				<div class="map-eyebrow">Tu mapa</div>
				<div class="map-name">
					{platforms.length} plataformas · {TOUR_PORTFOLIOS.length} portafolios
				</div>
			</div>
			<div class="map-status">Al día</div>
		</div>

		<div class="map-body">
			<div class="map-col">
				<div class="map-col-label">Plataformas</div>
				{#each platforms as platform (platform)}
					<div class="chip">{platform}</div>
				{/each}
			</div>

			<!--
				Geometría fija a propósito: chip de 34px con separación de 12 y tarjeta
				de 82 con separación de 10 dejan las dos columnas en 266px de alto y sus
				centros en 41, 87, 133, 179 y 225. Los extremos de cada hilo caen justo
				sobre esos centros; si cambia una altura, hay que rehacer las curvas.
			-->
			<svg class="wires" viewBox="0 0 54 266" fill="none">
				<path d="M2 41 C 26 41, 28 41, 52 41" stroke="rgba(212,145,42,0.45)" stroke-width="1" />
				<path d="M2 133 C 26 133, 28 41, 52 41" stroke="rgba(212,145,42,0.28)" stroke-width="1" />
				<path d="M2 87 C 26 87, 28 133, 52 133" stroke="rgba(212,145,42,0.45)" stroke-width="1" />
				<path d="M2 179 C 26 179, 28 133, 52 133" stroke="rgba(212,145,42,0.28)" stroke-width="1" />
				<path d="M2 225 C 26 225, 28 225, 52 225" stroke="rgba(212,145,42,0.45)" stroke-width="1" />
				<path d="M2 133 C 26 133, 28 225, 52 225" stroke="rgba(212,145,42,0.28)" stroke-width="1" />
				<circle cx="52" cy="41" r="2.5" fill="var(--amber)" />
				<circle cx="52" cy="133" r="2.5" fill="var(--amber)" />
				<circle cx="52" cy="225" r="2.5" fill="var(--amber)" />
			</svg>

			<div class="map-col">
				<div class="map-col-label">Portafolios</div>
				{#each TOUR_PORTFOLIOS as portfolio, i (portfolio.name)}
					<div class="pf">
						<div class="pf-head">
							<span class="pf-name">{portfolio.name}</span>
							<span class="pf-value">{portfolio.short}</span>
						</div>
						<div class="pf-sub">
							<span class="pf-kind">{portfolio.risk}</span>
							<span class="pf-delta" class:dn={!portfolio.up}>{portfolio.gain}</span>
						</div>
						<div class="pf-bar">
							<span style="width:{weights[i].pct}%; background:{weights[i].color}"></span>
						</div>
					</div>
				{/each}
			</div>
		</div>

		<div class="map-total">
			<span class="map-total-label">Patrimonio total</span>
			<span class="map-total-value">
				<b>$248,500</b>
				<em>+12,4%</em>
			</span>
		</div>
	</div>
</div>

<style>
	.map-card {
		padding: 22px 22px 18px;
		border-radius: 14px;
		background: rgba(255, 255, 255, 0.038);
		border: 1px solid var(--border-strong);
		backdrop-filter: blur(10px);
	}

	.map-top {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 12px;
	}

	.map-eyebrow {
		font-family: var(--font-mono);
		font-size: 10.5px;
		letter-spacing: 0.14em;
		text-transform: uppercase;
		color: var(--text-dim);
	}

	.map-name {
		margin-top: 4px;
		font-family: var(--font-display);
		font-weight: 500;
		font-size: 19px;
		letter-spacing: -0.01em;
	}

	.map-status {
		flex-shrink: 0;
		padding: 4px 9px;
		border-radius: 4px;
		font-family: var(--font-mono);
		font-size: 10px;
		font-weight: 600;
		letter-spacing: 0.1em;
		text-transform: uppercase;
		color: var(--green);
		background: rgba(34, 201, 126, 0.09);
		border: 1px solid rgba(34, 201, 126, 0.2);
	}

	.map-body {
		position: relative;
		display: grid;
		grid-template-columns: 118px 54px minmax(0, 1fr);
		align-items: center;
		margin-top: 44px;
	}

	.map-col {
		position: relative;
		display: flex;
		flex-direction: column;
		gap: 12px;
		min-width: 0;
	}

	.map-col:last-child {
		gap: 10px;
	}

	.map-col-label {
		position: absolute;
		top: -22px;
		font-family: var(--font-mono);
		font-size: 9px;
		letter-spacing: 0.16em;
		text-transform: uppercase;
		color: var(--text-dim);
		padding-bottom: 2px;
	}

	.chip {
		display: flex;
		align-items: center;
		height: 34px;
		box-sizing: border-box;
		padding: 0 10px;
		border: 1px solid var(--border);
		border-radius: 6px;
		background: var(--surface);
		font-family: var(--font-mono);
		font-size: 11px;
		color: var(--text-muted);
	}

	.wires {
		display: block;
		width: 54px;
		height: 266px;
		align-self: center;
	}

	.pf {
		height: 82px;
		box-sizing: border-box;
		padding: 11px 13px;
		border: 1px solid var(--border);
		border-radius: 9px;
		background: rgba(255, 255, 255, 0.022);
	}

	.pf-head {
		display: flex;
		align-items: baseline;
		justify-content: space-between;
		gap: 10px;
	}

	.pf-name {
		font-family: var(--font-display);
		font-weight: 500;
		font-size: 14.5px;
	}

	.pf-value {
		font-family: var(--font-mono);
		font-weight: 600;
		font-size: 14px;
		font-variant-numeric: tabular-nums;
	}

	.pf-sub {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 10px;
		margin-top: 6px;
		font-family: var(--font-mono);
	}

	.pf-kind {
		font-size: 9.5px;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		color: var(--text-dim);
	}

	.pf-delta {
		font-size: 11.5px;
		color: var(--green);
	}

	.pf-delta.dn {
		color: var(--red);
	}

	.pf-bar {
		height: 3px;
		border-radius: 2px;
		background: rgba(255, 255, 255, 0.07);
		overflow: hidden;
		margin-top: 9px;
	}

	.pf-bar > span {
		display: block;
		height: 100%;
		border-radius: 2px;
	}

	.map-total {
		display: flex;
		align-items: baseline;
		justify-content: space-between;
		gap: 12px;
		margin-top: 18px;
		padding-top: 14px;
		border-top: 1px solid var(--border);
	}

	.map-total-label {
		font-family: var(--font-mono);
		font-size: 9.5px;
		letter-spacing: 0.16em;
		text-transform: uppercase;
		color: var(--text-dim);
	}

	.map-total-value {
		display: flex;
		align-items: baseline;
		gap: 10px;
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
	}

	.map-total-value b {
		font-size: 21px;
		font-weight: 600;
		letter-spacing: -0.02em;
	}

	.map-total-value em {
		font-style: normal;
		font-size: 12px;
		font-weight: 600;
		color: var(--green);
	}

	/*
	 * En pantallas estrechas los conectores no caben: las plataformas pasan a
	 * una fila de chips sobre los portafolios y el hilo desaparece, porque a esa
	 * anchura ya no explicaría nada.
	 */
	@media (max-width: 940px) {
		.map-body {
			grid-template-columns: minmax(0, 1fr);
			gap: 16px;
			margin-top: 22px;
		}
		.map-col {
			flex-direction: row;
			flex-wrap: wrap;
			align-items: center;
			gap: 6px;
		}
		.map-col:last-child {
			flex-direction: column;
			align-items: stretch;
			gap: 8px;
		}
		.map-col-label {
			position: static;
			width: 100%;
			padding-bottom: 0;
		}
		.wires {
			display: none;
		}
	}
</style>
