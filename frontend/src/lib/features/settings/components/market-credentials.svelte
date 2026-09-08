<script lang="ts">
	/**
	 * Claves de proveedor de datos de mercado (BYO-key).
	 *
	 * La aplicación no tiene claves de proveedor: cada usuario aporta la suya y
	 * los precios que trae solo los ve quien la puso. Esta sección es el único
	 * sitio donde se introducen.
	 *
	 * La clave nunca vuelve del servidor: se guarda cifrada y de ella solo se
	 * recibe `last4`. Por eso el campo jamás se prellena, y cambiarla significa
	 * escribirla entera de nuevo.
	 */
	import { enhance } from '$app/forms';
	import Input from '$lib/ui/input.svelte';
	import Button from '$lib/ui/button.svelte';
	import SettingsSection from './settings-section.svelte';
	import { formatMarketProvider } from '$lib/shared/format/market-provider';
	import type { MarketCredential, MarketProvider } from '$lib/api/types';

	interface Props {
		credentials: MarketCredential[];
		/** `form` de la página, para leer el resultado de las acciones. */
		form: Record<string, unknown> | null;
	}

	let { credentials, form }: Props = $props();

	/*
	 * Lo que esta pantalla sabe de cada proveedor y ninguna otra necesita: dónde
	 * se saca la clave y qué cuota trae. El nombre no está aquí — sale de
	 * `formatMarketProvider`, que es la misma tabla que usa el panel de un
	 * activo: dos copias escritas a mano es como se acaba llamando de dos formas
	 * distintas al mismo proveedor.
	 */
	const PROVIDERS: {
		id: MarketProvider;
		signupUrl: string;
		hint: string;
	}[] = [
		{
			id: 'finnhub',
			signupUrl: 'https://finnhub.io/register',
			hint: 'Recomendado: su plan gratuito permite 60 consultas por minuto.'
		},
		{
			id: 'alphavantage',
			signupUrl: 'https://www.alphavantage.co/support/#api-key',
			hint: 'Su plan gratuito permite 25 consultas al día, así que se usa como respaldo.'
		}
	];

	const byProvider = $derived(new Map(credentials.map((c) => [c.provider, c])));

	// Un campo por proveedor. Nunca se rellena con nada: no hay valor que leer.
	let keyInputs = $state<Record<string, string>>({ finnhub: '', alphavantage: '' });
	let savingProvider = $state<string | null>(null);
	let verifyingProvider = $state<string | null>(null);
	let deletingProvider = $state<string | null>(null);
	let syncing = $state(false);

	/*
	 * El estado de una clave, dicho en una frase. Eran una píldora en versalitas
	 * («ACTIVA», «NO VÁLIDA») y debajo una nota que explicaba lo mismo: dos
	 * cosas que leer para entender una.
	 */
	const STATUS: Record<MarketCredential['status'], { tone: 'ok' | 'bad' | 'warn'; note: string }> =
		{
			active: { tone: 'ok', note: 'Funciona.' },
			invalid: {
				tone: 'bad',
				note: 'El proveedor la rechazó. Vuelve a introducirla o genera otra.'
			},
			rate_limited: {
				tone: 'warn',
				note: 'La cuota del plan se agotó. Se reintenta en la próxima sincronización.'
			}
		};

	function formatDate(iso: string | null): string {
		if (!iso) return 'nunca';
		return new Intl.DateTimeFormat('es', { dateStyle: 'short', timeStyle: 'short' }).format(
			new Date(iso)
		);
	}

	function errorFor(provider: MarketProvider): string | null {
		if (form?.marketProvider !== provider) return null;
		return (form?.marketError as string) ?? null;
	}

	function successFor(provider: MarketProvider): string | null {
		if (form?.marketProvider !== provider || !form?.marketSuccess) return null;
		return (form?.marketMessage as string) ?? null;
	}

	const hasAnyKey = $derived(credentials.length > 0);
</script>

<SettingsSection
	id="datos-de-mercado"
	title="Datos de mercado"
	description="Finexia no consulta precios con claves propias: usa la tuya, así que la cuota y los datos son tuyos. Los precios que trae solo los ves tú."
>
	{#snippet aside()}
		<p class="privacy">
			Se guarda cifrada y no se puede volver a leer, ni siquiera desde aquí: de ella solo verás sus
			cuatro últimos caracteres.
		</p>
	{/snippet}

	{#if !hasAnyKey}
		<p class="feedback warning">
			Sin ninguna clave, tus posiciones se valoran a su precio de compra y no a precio de mercado.
		</p>
	{/if}

	<div class="providers">
		{#each PROVIDERS as provider (provider.id)}
			{@const stored = byProvider.get(provider.id)}
			{@const status = stored ? STATUS[stored.status] : null}
			{@const error = errorFor(provider.id)}
			{@const success = successFor(provider.id)}

			<div class="provider">
				<div class="provider-head">
					<h4 class="provider-name">{formatMarketProvider(provider.id)}</h4>
					<!-- eslint-disable svelte/no-navigation-without-resolve -- resolve() es para rutas internas; estas salen al sitio del proveedor -->
					<a
						class="provider-link"
						href={provider.signupUrl}
						target="_blank"
						rel="noopener noreferrer"
					>
						Obtener una clave
					</a>
					<!-- eslint-enable svelte/no-navigation-without-resolve -->
				</div>

				{#if stored && status}
					<p class="provider-state" class:is-problem={status.tone !== 'ok'}>
						{status.note} Termina en <code>{stored.last4}</code> y se verificó
						{formatDate(stored.lastVerifiedAt)}.
					</p>
				{:else}
					<p class="provider-state">Sin configurar. {provider.hint}</p>
				{/if}

				<form
					method="POST"
					action="?/saveMarketKey"
					use:enhance={() => {
						savingProvider = provider.id;
						return async ({ update }) => {
							savingProvider = null;
							// reset:false conserva el resto del formulario de ajustes;
							// el campo de la clave se limpia abajo, a mano.
							await update({ reset: false });
							keyInputs[provider.id] = '';
						};
					}}
				>
					<input type="hidden" name="provider" value={provider.id} />
					<div class="key-row">
						<Input
							label={stored ? 'Reemplazar clave' : 'Clave de API'}
							type="password"
							name="apiKey"
							autocomplete="off"
							placeholder="Pega aquí tu clave"
							bind:value={keyInputs[provider.id]}
							required
						/>
						<Button type="submit" size="sm" loading={savingProvider === provider.id}>
							{stored ? 'Reemplazar' : 'Guardar'}
						</Button>
					</div>
				</form>

				{#if error}
					<p class="feedback error">{error}</p>
				{/if}
				{#if success}
					<p class="feedback success">{success}</p>
				{/if}

				{#if stored}
					<div class="provider-actions">
						<form
							method="POST"
							action="?/verifyMarketKey"
							use:enhance={() => {
								verifyingProvider = provider.id;
								return async ({ update }) => {
									verifyingProvider = null;
									await update({ reset: false });
								};
							}}
						>
							<input type="hidden" name="provider" value={provider.id} />
							<button type="submit" class="row-action" disabled={verifyingProvider === provider.id}>
								{verifyingProvider === provider.id ? 'Verificando…' : 'Verificar'}
							</button>
						</form>
						<form
							method="POST"
							action="?/deleteMarketKey"
							use:enhance={() => {
								deletingProvider = provider.id;
								return async ({ update }) => {
									deletingProvider = null;
									await update({ reset: false });
								};
							}}
						>
							<input type="hidden" name="provider" value={provider.id} />
							<button
								type="submit"
								class="row-action danger"
								disabled={deletingProvider === provider.id}
							>
								{deletingProvider === provider.id ? 'Eliminando…' : 'Eliminar'}
							</button>
						</form>
					</div>
				{/if}
			</div>
		{/each}
	</div>

	{#if hasAnyKey}
		<div class="sync-row">
			<div>
				<p class="sync-title">Sincronizar ahora</p>
				<p class="hint">
					Se actualizan los precios de los activos que tienes, con tu clave. Si no, se hace cada día
					automáticamente.
				</p>
			</div>
			<form
				method="POST"
				action="?/syncMarketData"
				use:enhance={() => {
					syncing = true;
					return async ({ update }) => {
						syncing = false;
						await update({ reset: false });
					};
				}}
			>
				<Button type="submit" size="sm" loading={syncing}>Sincronizar</Button>
			</form>
		</div>

		{#if form?.marketSyncError}
			<p class="feedback error">{form.marketSyncError}</p>
		{/if}
		{#if form?.marketSyncSuccess}
			<p class="feedback success">
				{form.marketSyncCount} precio{form.marketSyncCount === 1 ? '' : 's'} actualizado{form.marketSyncCount ===
				1
					? ''
					: 's'}{form.marketSyncRateCount
					? ` y ${form.marketSyncRateCount} tasa${form.marketSyncRateCount === 1 ? '' : 's'} de cambio`
					: ''}.
			</p>
		{/if}
	{/if}
</SettingsSection>

<style>
	/* Lo que le pasa a la clave una vez guardada: va en el carril, con el resto
	   de lo que explica la sección. */
	.privacy {
		max-width: 40ch;
		margin: 0.7rem 0 0;
		font-size: 0.78rem;
		line-height: 1.55;
		color: var(--text-dim);
	}

	.privacy {
		margin-bottom: 1.25rem;
	}

	/* Dos proveedores separados por un filete, no dos cajas dentro de otra: la
	   página tenía tres niveles de marco anidados. */
	.providers {
		display: grid;
	}

	.provider {
		padding: 1.1rem 0;
		border-top: 1px solid var(--border);
	}

	.provider:first-child {
		padding-top: 0;
		border-top: none;
	}

	.provider-head {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem 1rem;
		align-items: baseline;
		justify-content: space-between;
	}

	.provider-name {
		margin: 0;
		font-family: var(--font-body);
		font-size: 0.92rem;
		font-weight: 500;
		color: var(--text);
	}

	.provider-link {
		font-size: 0.8rem;
		color: var(--amber);
		text-decoration: none;
	}

	.provider-link:hover,
	.provider-link:focus-visible {
		text-decoration: underline;
		text-underline-offset: 3px;
	}

	.provider-state {
		max-width: 56ch;
		margin: 0.35rem 0 0;
		font-size: 0.8rem;
		line-height: 1.5;
		color: var(--text-muted);
	}

	/* Los cuatro últimos caracteres, en mono: es la única parte de la clave que
	   se puede enseñar, y hay que poder compararla con la que tienes delante. */
	.provider-state code {
		font-family: var(--font-mono);
		font-size: 0.76rem;
		color: var(--text);
	}

	.provider-state.is-problem {
		color: var(--amber);
	}

	.key-row {
		display: flex;
		gap: 0.75rem;
		align-items: flex-end;
		margin-top: 0.75rem;
	}

	.key-row :global(.input-wrapper) {
		flex: 1;
	}

	.provider-actions {
		display: flex;
		gap: 0.5rem;
		margin-top: 0.75rem;
	}

	.sync-row {
		display: flex;
		flex-wrap: wrap;
		gap: 1rem;
		align-items: center;
		justify-content: space-between;
		margin-top: 1.25rem;
		padding-top: 1.1rem;
		border-top: 1px solid var(--border);
	}

	.sync-title {
		margin: 0 0 0.2rem;
		font-size: 0.88rem;
		font-weight: 500;
		color: var(--text);
	}

	.sync-row .hint {
		margin: 0;
		max-width: 42ch;
	}

	/* El aviso de «sin clave» abre la sección, así que separa de lo que sigue. */
	.feedback.warning {
		margin-bottom: 1.25rem;
	}

	@media (max-width: 640px) {
		.key-row {
			flex-direction: column;
			align-items: stretch;
		}
	}
</style>
