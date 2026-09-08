<script lang="ts">
	/**
	 * Catálogo de activos, ordenado por lo que más falta hace mirar.
	 *
	 * El precio de mercado no se toca desde aquí: lo sincroniza cada usuario con
	 * su propia clave. La columna de precio es el respaldo manual del catálogo —el
	 * atajo para corregir muchas filas seguidas, que es lo que se hace a diario;
	 * el resto de la ficha se edita en el modal que abre «Editar»—, y por eso
	 * la tabla se ordena por antigüedad del precio en vez de por ticker: lo que
	 * lleva semanas sin cambiar es justo lo que hay que corregir, y con cien
	 * activos en cinco páginas por orden alfabético no aparecía nunca.
	 */
	import { enhance } from '$app/forms';
	import Badge from '$lib/ui/badge.svelte';
	import DataTable from '$lib/ui/data-table.svelte';
	import EmptyState from '$lib/ui/empty-state.svelte';
	import Pagination from '$lib/ui/pagination.svelte';
	import { formatAssetType } from '$lib/shared/format/asset-type';
	import AdminBlock from './admin-block.svelte';
	import { formatDateTime, formatPrice, type Asset } from '../admin';
	import { describeAssets, formatAge, isStale } from '../desk';

	interface Props {
		assets: Asset[];
		/** `form` de la página, para el resultado del ajuste por fila. */
		form: Record<string, unknown> | null;
		/**
		 * Abre la ficha completa de un activo. El modal lo pone la página, que es
		 * quien tiene el estado que lo abre; aquí solo se dice cuál.
		 */
		onEdit?: (asset: Asset) => void;
	}

	let { assets, form, onEdit }: Props = $props();

	const PER_PAGE = 20;
	let page = $state(1);

	// Sin precio primero (`null` no es «hoy»), después lo más viejo.
	const ordered = $derived(
		[...assets].sort((a, b) => (a.priceUpdatedAt ?? '').localeCompare(b.priceUpdatedAt ?? ''))
	);
	const pagedAssets = $derived(ordered.slice((page - 1) * PER_PAGE, page * PER_PAGE));

	let updatingId = $state<string | null>(null);
	let priceInputs = $state<Record<string, string>>({});

	$effect(() => {
		for (const asset of assets) {
			if (!(asset.id in priceInputs)) {
				priceInputs[asset.id] = asset.currentPrice?.value ?? '';
			}
		}
	});
</script>

<AdminBlock title="Catálogo de activos" summary={describeAssets(assets)}>
	{#if assets.length === 0}
		<EmptyState
			title="El catálogo está vacío"
			description="Crea el primer activo o importa una hoja para que las carteras tengan de dónde elegir."
		/>
	{:else}
		<DataTable caption="Activos del catálogo compartido, su precio manual y cuándo se tocó">
			<thead>
				<tr>
					<th>Ticker</th>
					<th>Nombre</th>
					<th>Tipo</th>
					<th class="num">Precio manual</th>
					<th>Actualizado</th>
					<th class="num">Nuevo precio</th>
					<th><span class="sr-only">Acciones</span></th>
				</tr>
			</thead>
			<tbody>
				{#each pagedAssets as asset (asset.id)}
					{@const stale = isStale(asset.priceUpdatedAt)}
					{@const hasError = form?.updateError && form?.errorId === asset.id}
					{@const saved = form?.updateSuccess && form?.updatedId === asset.id}
					<tr>
						<td class="cell-key">{asset.ticker}</td>
						<td class="cell-name">
							{asset.name}
							<!-- Solo se marca la excepción: un activo aportado por un
							     usuario únicamente lo ve quien lo aportó, y crearlo aquí con
							     el mismo ticker lo cura para todos. -->
							{#if asset.isCurated === false}
								<Badge tone="warning">Aportado</Badge>
							{/if}
						</td>
						<td>{formatAssetType(asset.assetType)}</td>
						<td class="num">{formatPrice(asset.currentPrice, asset.currency)}</td>
						<td class="cell-age" class:aged={stale} title={formatDateTime(asset.priceUpdatedAt)}>
							{formatAge(asset.priceUpdatedAt)}
						</td>
						<td class="num">
							<form
								class="edit"
								method="POST"
								action="?/updatePrice"
								use:enhance={() => {
									updatingId = asset.id;
									return async ({ update }) => {
										updatingId = null;
										await update({ reset: false });
									};
								}}
							>
								<input type="hidden" name="id" value={asset.id} />
								<input
									type="hidden"
									name="currency"
									value={asset.currentPrice?.currency ?? asset.currency ?? 'USD'}
								/>
								<label class="sr-only" for="price-{asset.id}">
									Nuevo precio de {asset.ticker}
								</label>
								<input
									id="price-{asset.id}"
									type="number"
									name="price"
									class="edit-input"
									class:invalid={hasError}
									bind:value={priceInputs[asset.id]}
									min="0.0001"
									step="any"
									placeholder="0.00"
									required
								/>
								<button class="row-action" type="submit" disabled={updatingId === asset.id}>
									{updatingId === asset.id ? 'Guardando…' : 'Guardar'}
								</button>
							</form>
							{#if hasError}
								<p class="row-error">{form.updateError}</p>
							{:else if saved}
								<p class="row-note">Precio guardado</p>
							{/if}
						</td>
						<td class="cell-actions">
							{#if onEdit}
								<button class="row-action" type="button" onclick={() => onEdit?.(asset)}>
									Editar
								</button>
							{/if}
							{#if form?.editError && form?.editId === asset.id}
								<p class="row-error">{form.editError}</p>
							{:else if form?.editSuccess && form?.editedId === asset.id}
								<p class="row-note">Cambios guardados</p>
							{/if}
						</td>
					</tr>
				{/each}
			</tbody>
		</DataTable>
	{/if}

	{#snippet footer()}
		<Pagination bind:page total={assets.length} perPage={PER_PAGE} label="activos" />
	{/snippet}
</AdminBlock>
