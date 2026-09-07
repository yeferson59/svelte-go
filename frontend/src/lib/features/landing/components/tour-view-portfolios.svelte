<script lang="ts">
	/*
	 * Maqueta de `/dashboard/portfolios`.
	 *
	 * Eran tres tarjetas de portafolio sobre una tabla de posiciones, que es la
	 * pantalla anterior al rediseño y además mezclaba dos rutas. Hoy es una lista:
	 * una fila por portafolio, de mayor a menor, con su perfil de riesgo, la barra
	 * de capital y ganancia, y el total al pie de su columna. Comparar es lo que
	 * se viene a hacer aquí, y en una rejilla de tarjetas las cifras caen a una
	 * altura distinta en cada una.
	 */
	import { TOUR_PORTFOLIOS, TOUR_PORTFOLIO_TOTALS } from '../product-tour';

	/** El carril mide lo que el portafolio más grande de la lista. */
	const scale = Math.max(...TOUR_PORTFOLIOS.map((p) => p.value));

	const width = (amount: number) => `${Math.max(0, Math.min(100, (amount / scale) * 100))}%`;
</script>

<div class="portfolios">
	<div class="page-head">
		<div>
			<p class="mk-page-title">Portafolios</p>
			<p class="mk-page-sub">Cómo tienes agrupado tu dinero, y cómo le va a cada grupo.</p>
		</div>
		<span class="mk-btn">Crear portafolio</span>
	</div>

	<table class="mk-table">
		<thead>
			<tr>
				<th>Portafolio</th>
				<th class="col-risk">Riesgo</th>
				<th class="col-bar">
					<!-- La cabecera de la columna es la leyenda de la barra: es el único
					     sitio donde hace falta decir qué significa cada color. -->
					<span class="legend">
						<i class="key cost"></i>capital
						<i class="key gain"></i>ganancia
					</span>
				</th>
				<th class="mk-num">Valor en USD</th>
				<th class="mk-num">Rendimiento</th>
			</tr>
		</thead>

		<tbody>
			{#each TOUR_PORTFOLIOS as row (row.name)}
				<tr>
					<th>
						<span class="who">
							<span class="mk-name">{row.name}</span>
							{#if row.isDefault}<em class="tag">predeterminado</em>{/if}
						</span>
						<span class="mk-detail">{row.detail}</span>
					</th>

					<td class="col-risk risk">{row.risk}</td>

					<td class="col-bar">
						<!-- La gráfica de crecimiento contraída a un instante: el corte cae
						     siempre en el capital, así que lo que queda dentro es la ganancia
						     y lo que asoma por fuera es lo que falta para recuperarlo. -->
						<span class="bar">
							<span class="held" style="width:{width(Math.min(row.value, row.cost))}"></span>
							{#if row.up}
								<span class="gain" style="width:{width(row.value - row.cost)}"></span>
							{:else}
								<span class="short" style="width:{width(row.cost - row.value)}"></span>
							{/if}
						</span>
					</td>

					<td class="mk-num">{row.money}</td>
					<td class="mk-num" class:mk-up={row.up} class:mk-dn={!row.up}>{row.gain}</td>
				</tr>
			{/each}
		</tbody>

		<!-- El total va al pie de su columna, que es donde vive un total. -->
		<tfoot>
			<tr class="foot">
				<th>
					<span class="mk-name">Total</span>
					<span class="mk-detail">{TOUR_PORTFOLIO_TOTALS.detail}</span>
				</th>
				<td class="col-risk"></td>
				<td class="col-bar cost-total">{TOUR_PORTFOLIO_TOTALS.cost}</td>
				<td class="mk-num">{TOUR_PORTFOLIO_TOTALS.value}</td>
				<td class="mk-num mk-up">{TOUR_PORTFOLIO_TOTALS.gain}</td>
			</tr>
		</tfoot>
	</table>
</div>

<style>
	.portfolios {
		--cost: #73819c;
	}

	.page-head {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 20px;
		padding-bottom: 16px;
		margin-bottom: 18px;
		border-bottom: 1px solid var(--border);
	}

	.who {
		display: flex;
		align-items: baseline;
		gap: 7px;
	}

	.tag {
		font-style: normal;
		font-size: 8.5px;
		color: var(--text-dim);
	}

	.col-risk {
		width: 16%;
	}

	.risk {
		color: var(--text-muted);
	}

	.col-bar {
		width: 30%;
	}

	.legend {
		display: flex;
		align-items: center;
		gap: 5px;
	}

	.key {
		width: 7px;
		height: 7px;
		border-radius: 2px;
	}

	.key.cost {
		background: var(--cost);
	}

	.key.gain {
		background: var(--green);
		margin-left: 6px;
	}

	.bar {
		display: flex;
		height: 7px;
		margin-top: 2px;
		border-radius: 2px;
		background: rgba(255, 255, 255, 0.08);
		overflow: hidden;
	}

	.held {
		background: var(--cost);
	}

	.gain {
		min-width: 3px;
		background: var(--green);
	}

	/*
	 * Lo que falta para recuperar el capital. Va translúcido porque no es barra:
	 * es el hueco que dejó al quedarse corta, y el filete derecho marca el
	 * extremo al que no llegó.
	 */
	.short {
		min-width: 3px;
		background: rgba(224, 90, 90, 0.32);
		box-shadow: inset -2px 0 0 var(--red);
	}

	/* El total va al pie de su columna, separado por un filete propio. */
	.foot th,
	.foot td {
		padding-top: 11px;
		border-top: 1px solid var(--border);
	}

	.cost-total {
		font-size: 9.5px;
		color: var(--text-dim);
	}

	@media (max-width: 620px) {
		.page-head {
			flex-direction: column;
			gap: 12px;
		}

		.col-risk,
		.col-bar {
			display: none;
		}
	}
</style>
