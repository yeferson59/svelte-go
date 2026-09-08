<script lang="ts">
	/**
	 * Lo que hay detrás del precio de un activo, y con qué clave volver a
	 * pedirlo.
	 *
	 * La lista contesta «cuánto tengo de X»; esto contesta la pregunta que viene
	 * justo después —«¿y de dónde sale ese precio?»— y da la única acción que
	 * cabe hacer al respecto. El botón «Sincronizar» de Ajustes recorre todas
	 * las tenencias y deja que la cadena de respaldo elija quién contesta, que
	 * es lo correcto para un job diario y lo contrario de lo que hace falta
	 * mirando una posición: aquí se nombra al proveedor, y el precio que se
	 * guarda es el que ese dio.
	 *
	 * Un proveedor sin clave se enseña igual, apagado. Es la mitad de la
	 * respuesta: saber que Alpha Vantage existe y que no está configurado es lo
	 * que convierte «este proveedor no cubre el activo» en algo que se puede
	 * arreglar.
	 */
	import { enhance } from '$app/forms';
	import { resolve } from '$app/paths';
	import Modal from '$lib/ui/modal.svelte';
	import Button from '$lib/ui/button.svelte';
	import { privacy } from '$lib/shared/privacy.svelte';
	import { formatCurrency } from '$lib/shared/format/money';
	import { formatTimeAgo } from '$lib/shared/format/date';
	import { MARKET_PROVIDER_NAMES, formatMarketProvider } from '$lib/shared/format/market-provider';
	import type { MarketCredential } from '$lib/api/types';
	import type { AssetHoldingRow } from '../asset-holdings';

	let {
		row,
		credentials,
		form,
		onClose
	}: {
		/** Activo señalado, o `null` con el panel cerrado. */
		row: AssetHoldingRow | null;
		/** Claves del usuario. Nunca traen la clave: proveedor, last4 y estado. */
		credentials: MarketCredential[];
		/** `form` de la página, para el resultado de la consulta. */
		form: Record<string, unknown> | null;
		onClose: () => void;
	} = $props();

	const byProvider = $derived(new Map(credentials.map((c) => [c.provider as string, c])));

	/**
	 * Los proveedores que el backend admite, cada uno con la clave del usuario si
	 * la hay. El orden es el de la tabla de nombres, que es el mismo en el que el
	 * backend gasta las cuotas.
	 */
	const providers = $derived(
		Object.keys(MARKET_PROVIDER_NAMES).map((id) => ({
			id,
			name: formatMarketProvider(id),
			credential: byProvider.get(id) ?? null
		}))
	);

	const configured = $derived(providers.filter((p) => p.credential !== null));

	/**
	 * Proveedor propuesto: el que trajo el precio que se está mirando —volver a
	 * preguntar a quien contestó es lo que se quiere nueve de cada diez veces— y
	 * si no, la primera clave que funcione.
	 */
	const suggested = $derived(
		(row?.priceProvider && byProvider.has(row.priceProvider) ? row.priceProvider : null) ??
			configured.find((p) => p.credential?.status === 'active')?.id ??
			configured[0]?.id ??
			''
	);

	/*
	 * Lo que el usuario haya elegido, anotado junto al activo sobre el que lo
	 * eligió. Sin esa segunda mitad haría falta un `$effect` que reescribiera la
	 * elección cada vez que se abre otro activo, y un estado que se corrige a sí
	 * mismo desde un efecto es justo lo que hay que evitar: aquí la elección
	 * simplemente deja de aplicar en cuanto la fila es otra.
	 */
	let pick = $state<{ assetId: string; provider: string } | null>(null);

	const chosen = $derived(pick && pick.assetId === row?.assetId ? pick.provider : suggested);

	let refreshing = $state(false);

	/*
	 * El resultado solo se enseña si es de este activo: el panel se abre y se
	 * cierra sobre la misma página, y `form` sobrevive al cierre.
	 */
	const isMine = $derived(row !== null && form?.refreshAssetId === row.assetId);
	const error = $derived(isMine ? ((form?.refreshError as string) ?? '') : '');
	const justRefreshed = $derived(isMine && form?.refreshSuccess === true);

	/** El precio de arriba: en la moneda del activo, que es como cotiza. */
	const currentPrice = $derived(
		row && row.marketPrice !== null ? formatCurrency(row.marketPrice, row.currency) : ''
	);

	/** De dónde sale el precio que se está mirando, en una frase. */
	const provenance = $derived.by(() => {
		if (!row) return '';

		if (row.priceSource === 'own') {
			const who = row.priceProvider ? formatMarketProvider(row.priceProvider) : 'una de tus claves';
			const when = formatTimeAgo(row.priceFetchedAt);

			return when ? `Lo trajo ${who} ${when}.` : `Lo trajo ${who}.`;
		}

		if (row.priceSource === 'manual') {
			return 'Es el precio de respaldo del catálogo, escrito a mano. Consulta a un proveedor para tener el tuyo.';
		}

		return 'No hay precio de mercado: la posición se valora a lo que costó, así que su ganancia sale cero.';
	});

	function statusNote(credential: MarketCredential | null): string {
		if (!credential) return 'Sin configurar';
		if (credential.status === 'invalid') return 'El proveedor rechazó esta clave';
		if (credential.status === 'rate_limited') return 'Cuota agotada; se repone sola';

		return `Termina en ${credential.last4}`;
	}
</script>

<Modal
	open={row !== null}
	title={row?.ticker ?? ''}
	description={row?.name ?? ''}
	size="md"
	{onClose}
>
	{#if row}
		<section class="current">
			<p class="label">Precio actual</p>
			{#if currentPrice}
				<p class="amount">{privacy.money(currentPrice)}</p>
			{:else}
				<p class="amount none">Sin precio</p>
			{/if}
			<p class="provenance">{provenance}</p>
		</section>

		{#if configured.length === 0}
			<p class="feedback warning">
				No tienes ninguna clave de proveedor, así que no hay a quién preguntar. Se añaden en
				<a href={resolve('/dashboard/settings')}>Ajustes → Datos de mercado</a>.
			</p>
		{:else}
			<form
				method="POST"
				action="?/refreshPrice"
				use:enhance={() => {
					refreshing = true;
					return async ({ update }) => {
						refreshing = false;
						// reset:false conserva el proveedor elegido: si la consulta
						// falló, lo normal es reintentar con el otro, no volver a
						// elegir desde cero.
						await update({ reset: false });
					};
				}}
			>
				<input type="hidden" name="assetId" value={row.assetId} />

				<fieldset>
					<legend>Pedir el precio a</legend>

					{#each providers as provider (provider.id)}
						{@const disabled = provider.credential === null}
						<label class="provider" class:off={disabled}>
							<input
								type="radio"
								name="provider"
								value={provider.id}
								checked={chosen === provider.id}
								onchange={() => (pick = { assetId: row.assetId, provider: provider.id })}
								{disabled}
							/>
							<span class="provider-name">{provider.name}</span>
							<span class="provider-state">{statusNote(provider.credential)}</span>
						</label>
					{/each}
				</fieldset>

				<div class="act">
					<Button type="submit" size="sm" loading={refreshing} disabled={chosen === ''}>
						Actualizar precio
					</Button>
					<p class="hint">
						Gasta una consulta de tu cuota con ese proveedor y el precio que devuelva solo lo ves
						tú.
					</p>
				</div>
			</form>

			{#if error}
				<p class="feedback error">{error}</p>
			{:else if justRefreshed}
				<p class="feedback success">
					{formatMarketProvider((form?.refreshProvider as string) ?? '')} respondió. El precio de arriba
					ya es el nuevo.
				</p>
			{/if}
		{/if}
	{/if}
</Modal>

<style>
	/* El precio abre el panel porque es el dato que se vino a mirar; la frase de
	   debajo es la que decide si hay que hacer algo con él. */
	.current {
		padding-bottom: 1.25rem;
		border-bottom: 1px solid var(--border);
	}

	.label {
		margin: 0 0 0.35rem;
		font-size: 0.8rem;
		color: var(--text-muted);
	}

	.amount {
		margin: 0;
		font-family: var(--font-mono);
		font-size: 1.6rem;
		font-weight: 600;
		line-height: 1;
		letter-spacing: -0.02em;
		color: var(--text);
		overflow-wrap: anywhere;
	}

	.amount.none {
		font-size: 1.1rem;
		font-weight: 400;
		color: var(--text-dim);
	}

	.provenance {
		max-width: 52ch;
		margin: 0.7rem 0 0;
		font-size: 0.8rem;
		line-height: 1.55;
		color: var(--text-muted);
	}

	fieldset {
		margin: 1.25rem 0 0;
		padding: 0;
		border: none;
	}

	legend {
		padding: 0;
		margin-bottom: 0.6rem;
		font-size: 0.8rem;
		color: var(--text-muted);
	}

	/* Cada proveedor es una fila pulsable entera, no un punto de radio con una
	   palabra al lado: el estado de la clave es la mitad de la decisión. */
	.provider {
		display: grid;
		grid-template-columns: auto minmax(0, 1fr) auto;
		align-items: center;
		gap: 0.75rem;
		padding: 0.7rem 0.8rem;
		border: 1px solid var(--border);
		border-radius: 10px;
		cursor: pointer;
	}

	.provider + .provider {
		margin-top: 0.5rem;
	}

	.provider:has(input:checked) {
		border-color: var(--amber);
		background: rgba(212, 145, 42, 0.07);
	}

	.provider:has(input:focus-visible) {
		outline: 2px solid var(--amber);
		outline-offset: 2px;
	}

	.provider.off {
		cursor: not-allowed;
		opacity: 0.55;
	}

	.provider input {
		accent-color: var(--amber);
	}

	.provider-name {
		font-size: 0.88rem;
		color: var(--text);
	}

	.provider-state {
		font-size: 0.75rem;
		color: var(--text-dim);
		text-align: right;
	}

	.act {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 0.75rem 1rem;
		margin-top: 1.1rem;
	}

	.hint {
		flex: 1;
		min-width: 16rem;
		margin: 0;
		font-size: 0.76rem;
		line-height: 1.5;
		color: var(--text-dim);
	}

	.feedback {
		margin-top: 1rem;
	}

	.feedback.warning a {
		color: inherit;
	}

	@media (max-width: 520px) {
		.provider {
			grid-template-columns: auto minmax(0, 1fr);
			row-gap: 0.2rem;
		}

		.provider-state {
			grid-column: 2;
			text-align: left;
		}
	}
</style>
