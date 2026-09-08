<script lang="ts">
	/**
	 * Edición completa de un activo del catálogo.
	 *
	 * La tabla ya permitía corregir el precio en su propia columna, que es el
	 * gesto de todos los días; lo que no había era manera de arreglar el resto de
	 * la ficha —un ticker mal escrito, un tipo equivocado, la moneda— sin volver
	 * a darla de alta con el mismo ticker y confiar en el upsert. Esto es esa
	 * ficha entera, con los mismos campos del alta y en el mismo orden, para que
	 * editar se lea como corregir lo que se escribió al crear.
	 *
	 * Dos campos son solo de aquí. El precio manual se puede dejar en blanco, y
	 * entonces se queda como estaba: el formulario no es la vía rápida de la
	 * tabla, y borrar el respaldo por no rellenar una casilla sería una sorpresa
	 * cara. Y la visibilidad, que es lo que separa un activo curado —lo ve todo
	 * el mundo— de uno aportado, que solo ven quienes lo aportaron.
	 */
	import { enhance } from '$app/forms';
	import { untrack } from 'svelte';
	import Button from '$lib/ui/button.svelte';
	import Checkbox from '$lib/ui/checkbox.svelte';
	import { ASSET_TYPES, type Asset } from '../admin';

	interface Props {
		/** El activo tal como lo devolvió el catálogo; llena el borrador. */
		asset: Asset;
		error?: string;
		/** Se llama cuando el envío sale bien; la página cierra el modal. */
		onSuccess?: () => void;
		/** Cierra el modal sin enviar. */
		onCancel?: () => void;
	}

	let { asset, error = '', onSuccess, onCancel }: Props = $props();

	// El modal monta este formulario al abrirlo y lo desmonta al cerrarlo, así
	// que el borrador se llena una sola vez: reaccionar a la prop pisaría lo que
	// el admin lleve escrito cuando la página se recargue de fondo.
	let draft = $state(
		untrack(() => ({
			ticker: asset.ticker,
			name: asset.name,
			assetType: asset.assetType,
			currency: asset.currency,
			exchange: asset.exchange ?? '',
			// En blanco a propósito: el campo pide un precio *nuevo*, y el que ya
			// hay se ve al lado. Prellenarlo haría que cualquier edición lo
			// reescribiera con la misma cifra y le pusiera fecha de hoy, que es
			// justo lo que la tabla ordena por antigüedad.
			price: '',
			isCurated: asset.isCurated !== false
		}))
	);

	let saving = $state(false);

	/**
	 * Cambiar la moneda sin dar un precio nuevo borra el que hay guardado: el
	 * número está en la tabla sin moneda propia, así que re-denominar el activo
	 * convertiría 190 dólares en 190 pesos sin que nadie lo escribiera. Lo hace
	 * el backend; esto es el aviso antes de pulsar.
	 */
	const currencyChanged = $derived(
		draft.currency.trim().toUpperCase() !== (asset.currency ?? '').toUpperCase()
	);
	const willClearPrice = $derived(
		currencyChanged && draft.price.trim() === '' && asset.currentPrice !== null
	);
</script>

<form
	class="rail-fields"
	method="POST"
	action="?/updateAsset"
	use:enhance={() => {
		saving = true;
		return async ({ result, update }) => {
			saving = false;
			await update({ reset: false });
			if (result.type === 'success') onSuccess?.();
		};
	}}
>
	<input type="hidden" name="id" value={asset.id} />

	<div class="pair">
		<div class="field">
			<label for="edit-ticker">Ticker</label>
			<input
				id="edit-ticker"
				type="text"
				name="ticker"
				bind:value={draft.ticker}
				placeholder="AAPL"
				autocapitalize="characters"
				required
			/>
		</div>
		<div class="field">
			<label for="edit-currency">Moneda</label>
			<input
				id="edit-currency"
				type="text"
				name="currency"
				bind:value={draft.currency}
				placeholder="USD"
				maxlength="3"
				autocapitalize="characters"
				required
			/>
		</div>
	</div>

	<div class="field">
		<label for="edit-name">Nombre</label>
		<input
			id="edit-name"
			type="text"
			name="name"
			bind:value={draft.name}
			placeholder="Apple Inc."
			required
		/>
	</div>

	<div class="pair">
		<div class="field">
			<label for="edit-type">Tipo</label>
			<select id="edit-type" name="assetType" bind:value={draft.assetType} required>
				{#each ASSET_TYPES as t (t.value)}
					<option value={t.value}>{t.label}</option>
				{/each}
			</select>
		</div>
		<div class="field">
			<label for="edit-exchange">Mercado <span class="optional">(opcional)</span></label>
			<input
				id="edit-exchange"
				type="text"
				name="exchange"
				bind:value={draft.exchange}
				placeholder="NASDAQ"
			/>
		</div>
	</div>

	<div class="field">
		<label for="edit-price">
			Precio manual <span class="optional">(en blanco, se queda como está)</span>
		</label>
		<input
			id="edit-price"
			type="number"
			name="price"
			bind:value={draft.price}
			min="0.0001"
			step="any"
			placeholder={asset.currentPrice?.value ?? '0.00'}
		/>
	</div>

	<div class="field visibility">
		<Checkbox
			id="edit-curated"
			name="isCurated"
			bind:checked={draft.isCurated}
			label="Visible para todas las cuentas"
		/>
		<p class="hint">
			Desmarcado, el activo vuelve a ser aportado: solo lo ven los usuarios que lo añadieron, y
			desaparece del catálogo del resto.
		</p>
	</div>

	{#if willClearPrice}
		<p class="feedback warning">
			Cambiar la moneda borra el precio manual guardado, porque la cifra dejaría de significar lo
			mismo. Escribe el precio en {draft.currency.trim().toUpperCase() || 'la nueva moneda'} para conservarlo.
		</p>
	{/if}

	{#if error}
		<p class="feedback error" role="alert">{error}</p>
	{/if}

	<div class="actions">
		{#if onCancel}
			<Button type="button" variant="ghost" onclick={onCancel}>Cancelar</Button>
		{/if}
		<Button type="submit" loading={saving}>Guardar cambios</Button>
	</div>
</form>

<style>
	.visibility {
		gap: 0.5rem;
	}

	.hint {
		margin: 0;
		font-size: 0.75rem;
		line-height: 1.5;
		color: var(--text-dim);
	}

	.actions {
		display: flex;
		justify-content: flex-end;
		gap: 0.75rem;
		margin-top: 0.5rem;
	}
</style>
