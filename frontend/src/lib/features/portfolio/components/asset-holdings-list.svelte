<script lang="ts">
	/*
	 * La lista de activos con lo que hace falta para encontrar uno: el buscador,
	 * la cuenta de lo que se está viendo y las hojas.
	 *
	 * Van juntos y no en la página porque son un solo comportamiento: al filtrar
	 * hay que volver a la primera hoja, la cuenta tiene que decir «3 de 10» y el
	 * pie de paginación tiene que contar los filtrados, no los que hay. Repartido
	 * entre la ruta y la tabla, ese acuerdo se rompía en cuanto alguien tocara
	 * uno de los tres.
	 */
	import Pagination from '$lib/ui/pagination.svelte';
	import AssetHoldingsTable from './asset-holdings-table.svelte';
	import type { AssetHoldingRow } from '../asset-holdings';

	let {
		rows,
		maxValue,
		displayCurrency,
		formatValue,
		onGoToPortfolios,
		onOpen,
		active = $bindable(null)
	}: {
		/** Todos los activos, ya ordenados de mayor a menor. */
		rows: AssetHoldingRow[];
		/** Valor de la mayor posición: la escala de las barras de las filas. */
		maxValue: number;
		displayCurrency: string;
		formatValue: (value: number) => string;
		onGoToPortfolios: () => void;
		/** Abre el panel de precio de un activo; lo dispara su fila. */
		onOpen: (row: AssetHoldingRow) => void;
		/** Ticker señalado, compartido con la barra de concentración. */
		active?: string | null;
	} = $props();

	/*
	 * El buscador es la respuesta directa a la pregunta de la página —«¿cuánto
	 * tengo de X?»— en una lista que se pagina: con cuarenta activos, encontrar
	 * uno era saber en qué hoja cayó. Sin tildes y sin mayúsculas, porque nadie
	 * escribe «Telefónica» en un buscador.
	 */
	let query = $state('');

	const fold = (text: string) =>
		text
			.toLowerCase()
			.normalize('NFD')
			.replace(/\p{Diacritic}/gu, '');

	const filtered = $derived.by(() => {
		const needle = fold(query.trim());
		if (needle === '') return rows;

		return rows.filter(
			(row) => fold(row.ticker).includes(needle) || fold(row.name).includes(needle)
		);
	});

	const PER_PAGE = 15;
	// `Pagination` devuelve la hoja al rango cuando la lista encoge, así que
	// filtrar hasta dejar tres filas no deja a nadie mirando una hoja vacía.
	let page = $state(1);
	const pagedRows = $derived(filtered.slice((page - 1) * PER_PAGE, page * PER_PAGE));
</script>

<section class="list" aria-labelledby="list-title">
	<header class="head">
		<h2 id="list-title">Tus activos</h2>

		{#if rows.length > 0}
			<div class="tools">
				<label class="search">
					<span class="sr-only">Buscar un activo por nombre o símbolo</span>
					<input
						type="search"
						bind:value={query}
						placeholder="Buscar un activo"
						autocomplete="off"
					/>
				</label>
				<p class="count" aria-live="polite">
					{#if query.trim() === ''}
						{rows.length}
						{rows.length === 1 ? 'activo' : 'activos'}
					{:else}
						{filtered.length} de {rows.length}
					{/if}
				</p>
			</div>
		{/if}
	</header>

	{#if filtered.length === 0 && rows.length > 0}
		<p class="no-match">
			Ningún activo se llama así.
			<button type="button" class="clear" onclick={() => (query = '')}>Ver todos</button>
		</p>
	{:else}
		<AssetHoldingsTable
			rows={pagedRows}
			{maxValue}
			{displayCurrency}
			{formatValue}
			{onGoToPortfolios}
			{onOpen}
			bind:active
		/>

		<Pagination bind:page total={filtered.length} perPage={PER_PAGE} label="activos" />
	{/if}
</section>

<style>
	.list {
		padding-top: 2rem;
	}

	.head {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		justify-content: space-between;
		gap: 0.75rem 1rem;
		margin-bottom: 1.25rem;
	}

	h2 {
		margin: 0;
		font-family: var(--font-body);
		font-size: 1.05rem;
		font-weight: 500;
		color: var(--text);
	}

	.tools {
		display: flex;
		align-items: center;
		gap: 1rem;
	}

	.search input {
		width: 13rem;
		padding: 0.4rem 0.7rem;
		border: 1px solid var(--border);
		border-radius: 8px;
		background: rgba(255, 255, 255, 0.03);
		color: var(--text);
		font-family: inherit;
		font-size: 0.82rem;
	}

	.search input::placeholder {
		color: var(--text-dim);
	}

	.search input:focus-visible {
		border-color: var(--amber);
	}

	.count {
		margin: 0;
		font-family: var(--font-mono);
		font-size: 0.75rem;
		font-variant-numeric: tabular-nums;
		color: var(--text-dim);
		white-space: nowrap;
	}

	.no-match {
		margin: 0;
		padding: 2.5rem 0;
		font-size: 0.9rem;
		color: var(--text-dim);
	}

	.clear {
		margin-left: 0.35rem;
		padding: 0;
		border: none;
		background: none;
		color: var(--text);
		font: inherit;
		text-decoration: underline;
		text-underline-offset: 3px;
		cursor: pointer;
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

	@media (max-width: 560px) {
		.tools {
			width: 100%;
			justify-content: space-between;
		}

		.search {
			flex: 1;
		}

		.search input {
			width: 100%;
		}
	}
</style>
