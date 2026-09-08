<script lang="ts">
	/*
	 * La lista consolidada: una fila por activo con lo que el usuario tiene de
	 * él, sumando todos sus portafolios.
	 *
	 * La barra de peso no es una columna: es el fondo de la fila. Como columna
	 * ocupaba metro y medio para no decir nada —el carril iba de 0 a 100 % y la
	 * mayor posición de una cartera repartida pesa un 17 %, así que las quince
	 * barras eran quince muñones idénticos en el sexto izquierdo del carril—.
	 * Aquí el carril es la fila entera y la escala es la mayor posición, que es
	 * la comparación que alguien hace de verdad al recorrer la lista.
	 *
	 * Siete columnas se quedaron en cuatro. «Portafolios» era una columna de
	 * unos: lo que importa es la excepción, y va en la segunda línea del activo
	 * cuando la hay. Cantidad y precio son un solo pensamiento —cuánto tienes y
	 * a cómo está— y van juntos. El punto de color de la clase se fue: la
	 * etiqueta ya dice la clase, y ocho matices más al lado de la escala de la
	 * barra eran ruido sin dato.
	 */
	import EmptyState from '$lib/ui/empty-state.svelte';
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import { formatPercent } from '$lib/shared/format/percent';
	import { formatQuantity, type AssetHoldingRow } from '../asset-holdings';

	let {
		rows,
		maxValue,
		displayCurrency,
		formatValue,
		onGoToPortfolios,
		onOpen,
		active = $bindable(null)
	}: {
		rows: AssetHoldingRow[];
		/**
		 * Valor de la mayor posición de toda la cartera, no de esta página.
		 * La lista se pagina y la barra no puede reescalarse al pasar de hoja:
		 * la primera fila de la página dos parecería la mayor de todas.
		 */
		maxValue: number;
		/** Moneda de la columna «Valor»: la misma en todas las filas. */
		displayCurrency: string;
		formatValue: (value: number) => string;
		/**
		 * Salida del estado vacío. Aquí no hay un portafolio al que agregar —esta
		 * vista los atraviesa todos—, así que lleva a elegir uno.
		 */
		onGoToPortfolios: () => void;
		/**
		 * Abre el activo: de dónde salió su precio y con qué clave volver a
		 * pedirlo. Es la acción de la fila, y por eso cuelga del nombre del
		 * activo y no de un botón en una columna aparte — la fila entera no puede
		 * ser el disparador porque dentro hay texto que se selecciona.
		 */
		onOpen: (row: AssetHoldingRow) => void;
		/** Ticker señalado, compartido con la barra de concentración de arriba. */
		active?: string | null;
	} = $props();

	/*
	 * El precio va en la moneda del activo, no en la de la columna «Valor»: es
	 * lo que cotiza, no lo que se convirtió. Por eso no usa `formatValue`.
	 */
	function fmtPrice(row: AssetHoldingRow): string {
		return privacy.money(formatCurrency(row.marketPrice ?? 0, row.currency));
	}

	/*
	 * `pointerenter` también dispara al tocar, y en una pantalla táctil el
	 * `pointerleave` que lo apagaría puede no llegar nunca: la fila se quedaba
	 * encendida y la barra de arriba, con el resto de sus franjas apagadas, sin
	 * nada evidente que la devolviera a su estado.
	 */
	const canHover = typeof window !== 'undefined' && window.matchMedia('(hover: hover)').matches;

	function point(key: string | null) {
		if (canHover) active = key;
	}

	/** Ancho de la barra de la fila, contra la mayor posición de la cartera. */
	function barWidth(value: number): number {
		if (maxValue <= 0) return 0;

		return Math.max(0, Math.min(100, (value / maxValue) * 100));
	}
</script>

{#if rows.length > 0}
	<table>
		<caption class="sr-only">
			Activos que tienes, con su clase, las unidades sumadas entre portafolios, su precio y cuánto
			pesan sobre el total. Las filas van de mayor a menor valor.
		</caption>
		<thead>
			<tr>
				<th scope="col">Activo</th>
				<th scope="col" class="col-class">Clase</th>
				<th scope="col" class="col-position num">Posición</th>
				<th scope="col" class="col-value num">Valor en {displayCurrency}</th>
				<th scope="col" class="col-weight num">Peso</th>
			</tr>
		</thead>
		<tbody>
			{#each rows as row (row.assetId)}
				<tr
					class:on={active === row.ticker}
					style="--bar: {barWidth(row.value).toFixed(2)}%"
					onpointerenter={() => point(row.ticker)}
					onpointerleave={() => point(null)}
				>
					<th scope="row" class="asset">
						<button
							type="button"
							class="open"
							onclick={() => onOpen(row)}
							aria-label="Ver de dónde sale el precio de {row.ticker} y actualizarlo"
						>
							<span class="ticker">{row.ticker}</span>
							<span class="name">
								{row.name}{#if row.portfolios > 1}<span class="spread">
										, en {row.portfolios} portafolios</span
									>{/if}
							</span>
						</button>
					</th>

					<td class="col-class type">{row.typeLabel}</td>

					<td class="col-position num mono position">
						{formatQuantity(row.quantity)} uds
						{#if row.marketPrice === null}
							<span class="qualifier">a coste, sin precio de mercado</span>
						{:else}
							<span class="unit-price">a {fmtPrice(row)}</span>
						{/if}
					</td>

					<td class="col-value num mono value">
						{privacy.money(formatValue(row.value))}
						{#if !row.fxConverted}
							<span class="qualifier flagged">sin convertir a {displayCurrency}</span>
						{/if}
					</td>

					<td class="col-weight num mono weight">{formatPercent(row.percent)}</td>
				</tr>
			{/each}
		</tbody>
	</table>
{:else}
	<EmptyState
		bordered
		title="Todavía no hay nada que listar"
		description="Cuando registres posiciones en tus portafolios, aquí aparecerá cuánto tienes de cada activo."
	>
		{#snippet action()}
			<button onclick={onGoToPortfolios} class="btn-go-portfolios">
				<svg
					width="18"
					height="18"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
					aria-hidden="true"
				>
					<path d="M12 5v14M5 12h14" />
				</svg>
				Ir a mis portafolios
			</button>
		{/snippet}
	</EmptyState>
{/if}

<style>
	/*
	 * `separate` y no `collapse`: el degradado de la barra se pinta sobre la
	 * fila, y con los bordes fundidos el navegador lo recorta a cada celda y la
	 * barra sale a trozos.
	 */
	table {
		width: 100%;
		border-collapse: separate;
		border-spacing: 0;
	}

	.sr-only {
		position: absolute;
		width: 1px;
		height: 1px;
		padding: 0;
		margin: -1px;
		overflow: hidden;
		clip-path: inset(50%);
		white-space: nowrap;
	}

	/* El nombre no necesita un tercio de la fila: acotarlo acerca la clase a su
	   activo y quita el vacío que quedaba entre los dos. */
	thead th:first-child {
		width: 30%;
	}

	thead th {
		padding: 0 0.75rem 0.6rem;
		border-bottom: 1px solid var(--border);
		font-size: 0.75rem;
		font-weight: 400;
		color: var(--text-dim);
		text-align: left;
		white-space: nowrap;
	}

	thead th.num {
		text-align: right;
	}

	/*
	 * La barra es el fondo de la fila, no un objeto encima: así ninguna celda
	 * puede taparla y no cuesta una columna.
	 *
	 * Se apaga en los últimos píxeles en vez de cortarse en seco. Con el canto
	 * duro, el final de cada barra caía en mitad de una palabra de la columna
	 * «Clase» y esos diez cantos verticales se leían como una división de la
	 * tabla que no existe. Difuminado, la fila se lee como una mancha de calor
	 * —que es lo que se compara al recorrer la lista— y el dato exacto lo da la
	 * columna «Peso», que está para eso.
	 */
	tbody tr {
		background: linear-gradient(
			to right,
			rgba(212, 145, 42, 0.1) 0,
			rgba(212, 145, 42, 0.1) max(0px, calc(var(--bar) - 2.5rem)),
			rgba(212, 145, 42, 0) var(--bar)
		);
	}

	@media (hover: hover) {
		tbody tr.on th,
		tbody tr.on td {
			background: rgba(255, 255, 255, 0.03);
		}
	}

	tbody th,
	tbody td {
		padding: 0.8rem 0.75rem;
		transition: background 0.15s ease;
		border-bottom: 1px solid var(--border);
		font-size: 0.85rem;
		font-weight: 400;
		color: var(--text);
		text-align: left;
		vertical-align: baseline;
	}

	tbody tr:last-child th,
	tbody tr:last-child td {
		border-bottom: none;
	}

	.num {
		text-align: right;
	}

	.mono {
		font-family: var(--font-mono);
		font-variant-numeric: tabular-nums;
	}

	/*
	 * El activo es lo que se pulsa, así que es un botón de verdad y no una fila
	 * con un `onclick`: así llega el foco por teclado, Enter lo abre y el lector
	 * de pantalla lo anuncia como lo que es. Sin ninguna pinta de botón —el
	 * subrayado al pasar por encima es toda la señal que hace falta en una lista
	 * donde las quince filas hacen lo mismo.
	 */
	.open {
		display: block;
		width: 100%;
		padding: 0;
		border: none;
		background: none;
		font: inherit;
		color: inherit;
		text-align: left;
		cursor: pointer;
	}

	.open:hover .ticker,
	.open:focus-visible .ticker {
		text-decoration: underline;
		text-underline-offset: 3px;
	}

	.open:focus-visible {
		outline: 2px solid var(--amber);
		outline-offset: 3px;
		border-radius: 4px;
	}

	.ticker {
		display: block;
		font-family: var(--font-mono);
		font-weight: 600;
		color: var(--text);
	}

	.name {
		display: block;
		margin-top: 0.1rem;
		font-size: 0.78rem;
		line-height: 1.35;
		color: var(--text-muted);
		overflow-wrap: anywhere;
	}

	.spread {
		color: var(--text-dim);
	}

	.type {
		color: var(--text-muted);
		white-space: nowrap;
	}

	.position {
		font-size: 0.8rem;
		color: var(--text-muted);
		white-space: nowrap;
	}

	.unit-price,
	.qualifier {
		display: block;
		margin-top: 0.1rem;
		font-size: 0.72rem;
	}

	.unit-price {
		color: var(--text-dim);
	}

	.qualifier {
		font-family: var(--font-body);
		color: var(--text-dim);
	}

	.qualifier.flagged {
		color: var(--amber);
	}

	.value {
		font-weight: 600;
		white-space: nowrap;
	}

	.weight {
		color: var(--text-muted);
		white-space: nowrap;
	}

	.col-weight {
		width: 4.5rem;
	}

	.col-position {
		width: 12rem;
	}

	.col-value {
		width: 10rem;
	}

	.col-class {
		width: 7rem;
	}

	.btn-go-portfolios {
		display: inline-flex;
		align-items: center;
		gap: 0.6rem;
		padding: 0.8rem 1.4rem;
		border: none;
		border-radius: 10px;
		background: var(--amber);
		color: #0d0800;
		font-weight: 600;
		font-family: var(--font-body);
		font-size: 0.9rem;
		cursor: pointer;
		transition:
			background 0.2s ease,
			transform 0.2s ease;
	}

	.btn-go-portfolios:hover {
		background: var(--amber-light);
		transform: translateY(-1px);
	}

	@media (prefers-reduced-motion: reduce) {
		tbody th,
		tbody td,
		.btn-go-portfolios {
			transition: none;
		}

		.btn-go-portfolios:hover {
			transform: none;
		}
	}

	/*
	 * Debajo de esto la fila no cabe en una línea. En vez de un desplazamiento
	 * lateral que dejaba el valor y el peso fuera de la pantalla —o sea, la
	 * respuesta a la pregunta de la página—, la fila se pliega en dos: activo y
	 * valor arriba, todo lo demás abajo. La barra sigue siendo el fondo.
	 */
	@media (max-width: 760px) {
		thead {
			display: none;
		}

		thead th:first-child {
			width: auto;
		}

		tbody tr {
			display: grid;
			grid-template-columns: minmax(0, 1fr) auto;
			column-gap: 0.75rem;
			padding: 0.85rem 0.75rem;
			border-bottom: 1px solid var(--border);
		}

		tbody tr:last-child {
			border-bottom: none;
		}

		tbody th,
		tbody td {
			padding: 0;
			border-bottom: none;
		}

		tbody th.asset {
			grid-column: 1;
			grid-row: 1;
			box-shadow: none;
		}

		.col-value {
			grid-column: 2;
			grid-row: 1;
			width: auto;
		}

		.col-position {
			grid-column: 1;
			grid-row: 2;
			width: auto;
			margin-top: 0.45rem;
			text-align: left;
		}

		.col-weight {
			grid-column: 2;
			grid-row: 2;
			width: auto;
			margin-top: 0.45rem;
		}

		/* La clase cabe junto a las unidades, que es donde se lee como parte de
		   la descripción de la posición y no como una columna huérfana. */
		.col-class {
			grid-column: 1;
			grid-row: 3;
			width: auto;
			margin-top: 0.15rem;
			font-size: 0.78rem;
		}

		.unit-price {
			display: inline;
			margin-top: 0;
		}
	}
</style>
