<script lang="ts">
	/*
	 * Maqueta de `/dashboard/reports`, la ficha de resultados: la cifra de
	 * cabecera, la matriz de rentabilidad mes a mes, las medidas de movimiento y
	 * los archivos.
	 *
	 * El contenido ya era el del rediseño, pero seguía dentro de tarjetas y a dos
	 * columnas; la página real no tiene ninguna: son secciones a sangre separadas
	 * por un filete, una debajo de otra. Y «Cómo se movió» dice ahora qué mide
	 * cada medida, que es una columna de la tabla y no una ayuda opcional: sin
	 * ella «Ratio de Sharpe 1,43» es un número sin idioma.
	 *
	 * La proyección a cinco años y los archivos descargables se quedan fuera de la
	 * maqueta —están en el pie, entre las capacidades—: cada vista enseña lo alto
	 * de su página, como lo vería quien entra, y en la de reportes esos dos
	 * bloques van muy por debajo del pliegue.
	 */
	import {
		TOUR_KEY_STATS,
		TOUR_MONTHS,
		TOUR_RECORD,
		TOUR_RETURNS,
		tourReturnBackground
	} from '../product-tour';

	/* Coma decimal, como el resto de cifras de la landing y del panel. */
	const fmt = (value: number | null) =>
		value === null ? '–' : `${value > 0 ? '+' : ''}${value.toFixed(1).replace('.', ',')}%`;
</script>

<div class="reports">
	<div class="page-head">
		<p class="mk-page-title">Reportes</p>
		<p class="mk-page-sub">
			Cómo le ha ido a tu dinero desde que lo sigues aquí, y los archivos para llevártelo.
		</p>
	</div>

	<div class="mk-sec">
		<p class="record-label">{TOUR_RECORD.label}</p>
		<p class="record-value">{TOUR_RECORD.value}</p>
		<p class="record-span">{TOUR_RECORD.span}</p>
		<p class="record-money mk-up">{TOUR_RECORD.money}</p>
	</div>

	<div class="mk-sec">
		<p class="mk-h2">Rentabilidad mes a mes</p>

		<!-- Catorce columnas no caben en un móvil: el carril se desplaza de lado,
		     igual que en la página real. -->
		<div class="scroller">
			<table class="mk-table matrix">
				<thead>
					<tr>
						<th class="year">Año</th>
						{#each TOUR_MONTHS as month (month)}
							<th class="mk-num">{month}</th>
						{/each}
						<th class="mk-num">Total</th>
					</tr>
				</thead>
				<tbody>
					{#each TOUR_RETURNS as row (row.year)}
						<tr>
							<th class="year">{row.year}</th>
							{#each row.values as value, index (`${row.year}-${TOUR_MONTHS[index]}`)}
								<td class="cell">
									<span
										class="chip"
										class:empty={value === null}
										style:background-color={tourReturnBackground(value)}
									>
										{fmt(value)}
									</span>
								</td>
							{/each}
							<td class="mk-num" class:mk-up={row.total >= 0} class:mk-dn={row.total < 0}>
								{fmt(row.total)}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>

		<p class="footnote">
			Rendimiento de lo invertido: el dinero que aportas o retiras dentro del mes no cuenta como
			rentabilidad.
		</p>
	</div>

	<div class="mk-sec">
		<p class="mk-h2">Cómo se movió</p>

		<table class="mk-table">
			<thead>
				<tr>
					<th class="col-measure">Medida</th>
					<th class="mk-num col-value">Valor</th>
					<th class="col-meaning">Qué mide</th>
				</tr>
			</thead>
			<tbody>
				{#each TOUR_KEY_STATS as stat (stat.label)}
					<tr>
						<th class="col-measure">{stat.label}</th>
						<td
							class="mk-num col-value"
							class:mk-up={stat.tone === 'up'}
							class:mk-dn={stat.tone === 'down'}
						>
							{stat.value}
							{#if stat.detail}<span class="mk-detail">{stat.detail}</span>{/if}
						</td>
						<td class="meaning">
							{stat.hint}
							<!-- El reparo va con la explicación y no bajo la cifra: alineado a
							     la derecha ocupaba tres renglones en bandera. -->
							{#if stat.note}<span class="note">{stat.note}</span>{/if}
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
</div>

<style>
	.page-head {
		padding-bottom: 16px;
		border-bottom: 1px solid var(--border);
	}

	/* La cifra de cabecera, en el mismo registro que la de la página real: una
	   etiqueta que la nombra, el número, y el periodo en prosa. */
	.record-label {
		font-size: 10.5px;
		color: var(--text-muted);
	}

	.record-value {
		margin-top: 4px;
		font-family: var(--font-mono);
		font-size: 26px;
		font-weight: 600;
		line-height: 1;
		letter-spacing: -0.03em;
		color: var(--text);
	}

	.record-span {
		margin-top: 8px;
		font-size: 9.5px;
		color: var(--text-muted);
	}

	.record-money {
		margin-top: 3px;
		font-size: 9.5px;
	}

	.mk-h2 {
		display: block;
		margin-bottom: 11px;
	}

	/* Trece columnas de cifras: la etiqueta baja a 8 px y las celdas van a tope,
	   que es lo que hace la matriz de verdad cuando aprieta el ancho. */
	.matrix th,
	.matrix td {
		padding-right: 3px;
	}

	.matrix .year {
		font-family: var(--font-mono);
		font-size: 9px;
		color: var(--text-muted);
		white-space: nowrap;
	}

	.matrix thead th {
		font-size: 8px;
		text-align: center;
	}

	.matrix thead th:last-child,
	.matrix tbody td:last-child {
		text-align: right;
	}

	.cell {
		padding: 4px 3px;
		text-align: center;
	}

	/* El color solo dice signo e intensidad: la cifra lleva su propio signo, así
	   que quien no distinga los dos tonos no se pierde nada. */
	.chip {
		display: block;
		padding: 3px 2px;
		border-radius: 3px;
		font-family: var(--font-mono);
		font-size: 8px;
		font-variant-numeric: tabular-nums;
		color: var(--text);
	}

	.chip.empty {
		color: var(--text-dim);
	}

	.matrix tbody td:last-child {
		font-family: var(--font-mono);
		font-size: 9px;
		font-variant-numeric: tabular-nums;
	}

	.footnote {
		margin-top: 9px;
		max-width: 70ch;
		font-size: 8.5px;
		line-height: 1.5;
		color: var(--text-dim);
	}

	/* Sin barra a la vista: la maqueta es decorativa y una barra de desplazamiento
	   dentro de ella se leería como un fallo de la página, no del panel. */
	.scroller {
		overflow-x: auto;
		scrollbar-width: none;
	}

	.scroller::-webkit-scrollbar {
		display: none;
	}

	.matrix {
		min-width: 460px;
	}

	.col-measure {
		width: 24%;
	}

	.col-value {
		width: 14%;
	}

	.meaning {
		font-size: 9.5px;
		line-height: 1.45;
		color: var(--text-muted);
	}

	.note {
		display: block;
		margin-top: 2px;
		font-size: 8.5px;
		color: var(--text-dim);
	}

	.col-value .mk-detail {
		text-align: right;
	}

	/* La explicación es lo primero que sobra cuando aprieta el ancho: la cabecera
	   se va con ella, o queda una columna sin nada debajo. */
	@media (max-width: 780px) {
		.meaning,
		.col-meaning {
			display: none;
		}
	}
</style>
