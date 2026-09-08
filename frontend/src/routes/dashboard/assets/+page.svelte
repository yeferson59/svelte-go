<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import CurrencySelect from '$lib/ui/currency-select.svelte';
	import PageHeader from '$lib/ui/page-header.svelte';
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import {
		AssetConcentrationBand,
		AssetHoldingsList,
		AssetPricePanel,
		toAssetHoldingRows,
		type AssetHoldingRow
	} from '$lib/features/portfolio';
	import type { PageProps } from './$types';

	const { data, form }: PageProps = $props();

	// De mayor a menor una sola vez, aquí: la barra de concentración y la lista
	// son la misma lectura en dos formas, y si cada una ordenara por su cuenta
	// podrían discrepar sobre cuál es el mayor activo.
	const rows = $derived(toAssetHoldingRows(data.holdings ?? []).sort((a, b) => b.value - a.value));
	const displayCurrency = $derived(data.currency);

	const totalValue = $derived(rows.reduce((sum, row) => sum + row.value, 0));
	// Escala de las barras de la lista. Sale de la cartera entera y no de la
	// página que se esté viendo: si no, la primera fila de cada hoja saldría
	// llena y parecería la mayor posición de todas.
	const maxValue = $derived(rows.reduce((top, row) => Math.max(top, row.value), 0));

	// Posiciones que el backend no pudo convertir: siguen sumadas con su importe
	// nativo, así que el total mezcla monedas y hay que decirlo en vez de
	// presentarlo como una cifra limpia.
	const unconverted = $derived(rows.filter((row) => !row.fxConverted).length);

	/** Activo señalado, compartido por la barra y la lista. */
	let active = $state<string | null>(null);

	/*
	 * Activo abierto en el panel. Se guarda el id y no la fila: al actualizar un
	 * precio la acción recarga los holdings, y una fila copiada aquí seguiría
	 * enseñando el precio viejo con el panel abierto encima.
	 */
	let openedId = $state<string | null>(null);
	const opened = $derived(rows.find((row) => row.assetId === openedId) ?? null);

	function fmt(value: number): string {
		return formatCurrency(value, displayCurrency);
	}

	function goToPortfolios() {
		goto(resolve('/dashboard/portfolios'));
	}
</script>

<svelte:head>
	<title>Mis activos - FINEXIA</title>
	<meta
		name="description"
		content="Todo lo que tienes de cada activo, sumado a través de tus portafolios"
	/>
</svelte:head>

<PageHeader
	title="Mis activos"
	subtitle="Cuánto tienes de cada activo, sumando todos tus portafolios."
>
	{#snippet actions()}
		<CurrencySelect currency={displayCurrency} />
	{/snippet}
</PageHeader>

<section class="total" aria-labelledby="total-value">
	<h2 class="label" id="total-value">Valor total</h2>
	<p class="amount">{privacy.money(fmt(totalValue))}</p>
	<p class="meta">
		{#if rows.length === 1}
			Un solo activo.
		{:else}
			Repartido en {rows.length} activos distintos.
		{/if}
	</p>

	{#if unconverted > 0}
		<p class="fx">
			{unconverted === 1 ? 'Un activo queda' : `${unconverted} activos quedan`} sin convertir a {displayCurrency}:
			no hay tasa para {unconverted === 1 ? 'su moneda' : 'sus monedas'}. {unconverted === 1
				? 'Su importe va'
				: 'Sus importes van'} a valor nominal, así que el total y los pesos mezclan monedas.
		</p>
	{/if}
</section>

<AssetConcentrationBand {rows} formatCurrency={fmt} bind:active />

<AssetHoldingsList
	{rows}
	{maxValue}
	{displayCurrency}
	formatValue={fmt}
	onGoToPortfolios={goToPortfolios}
	onOpen={(row: AssetHoldingRow) => (openedId = row.assetId)}
	bind:active
/>

<AssetPricePanel
	row={opened}
	credentials={data.credentials}
	{form}
	onClose={() => (openedId = null)}
/>

<style>
	/*
	 * La cifra que anclan todos los porcentajes de la página. Más pequeña que la
	 * del panel a propósito: el patrimonio es la portada, y aquí el total solo
	 * está para que un «17,3 %» signifique algo.
	 */
	.total {
		padding-bottom: 2rem;
		border-bottom: 1px solid var(--border);
	}

	.label {
		margin: 0 0 0.5rem;
		font-family: var(--font-body);
		font-size: 0.9rem;
		font-weight: 400;
		color: var(--text-muted);
	}

	.amount {
		margin: 0;
		font-family: var(--font-mono);
		font-size: clamp(2rem, 4.5vw, 2.75rem);
		font-weight: 600;
		line-height: 1;
		letter-spacing: -0.03em;
		color: var(--text);
		overflow-wrap: anywhere;
	}

	.meta {
		margin: 0.7rem 0 0;
		font-size: 0.85rem;
		color: var(--text-dim);
	}

	/* Mismo aviso de «falta una tasa» que el panel: filete ámbar y prosa, no una
	   caja de alerta que compite con la cifra de al lado. */
	.fx {
		max-width: 62ch;
		margin: 1rem 0 0;
		padding-left: 0.75rem;
		border-left: 2px solid rgba(212, 145, 42, 0.45);
		font-size: 0.8rem;
		line-height: 1.5;
		color: var(--text-muted);
	}
</style>
