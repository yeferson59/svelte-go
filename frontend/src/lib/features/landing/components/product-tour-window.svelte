<script lang="ts">
	/*
	 * Marco de la maqueta: barra de ventana, menú lateral y barra superior,
	 * calcados del panel real. Todo el bloque es `aria-hidden`; el texto
	 * equivalente lo pone el pie de `product-tour.svelte`.
	 *
	 * El chrome enseñaba el panel de antes del rediseño: la marca en versalitas
	 * espaciadas, un antetítulo «MENÚ PRINCIPAL», secciones sin icono, la abierta
	 * marcada a la vez con fondo, borde y texto ámbar, y una barra superior con
	 * migas de pan y dos cápsulas. Hoy la marca va en serif y caja normal, cada
	 * sección lleva su icono, la abierta se marca una sola vez —fondo tenue y un
	 * filete ámbar al canto— y la barra dice el nombre de la sección y quién ha
	 * entrado.
	 */
	import TourViewSummary from './tour-view-summary.svelte';
	import TourViewPortfolios from './tour-view-portfolios.svelte';
	import TourViewTransactions from './tour-view-transactions.svelte';
	import TourViewReports from './tour-view-reports.svelte';
	import { TOUR_ICONS, TOUR_NAV, TOUR_USER, type TourView } from '../product-tour';

	let { view }: { view: TourView } = $props();
</script>

{#snippet icon(name: string, size = 13)}
	<svg
		width={size}
		height={size}
		viewBox="0 0 24 24"
		fill="none"
		stroke="currentColor"
		stroke-width="1.75"
		stroke-linecap="round"
		stroke-linejoin="round"
	>
		{#each TOUR_ICONS[name] ?? [] as d (d)}
			<path {d} />
		{/each}
	</svg>
{/snippet}

<div class="win" aria-hidden="true">
	<div class="win-bar">
		<span class="dot"></span><span class="dot"></span><span class="dot"></span>
		<!-- La ruta de la vista que se está mirando: decía `/dashboard` en las
		     cuatro pestañas, así que la barra de direcciones desmentía a la que
		     estaba abierta. -->
		<div class="win-url">finexia.me{view.path}</div>
	</div>

	<div class="win-body">
		<aside class="win-side">
			<div class="side-brand">
				<svg width="17" height="17" viewBox="0 0 30 30" fill="none">
					<rect width="30" height="30" rx="7" fill="var(--amber)" />
					<path
						d="M7 22L12.5 14.5L16.5 18.5L23 9"
						stroke="#0c0a06"
						stroke-width="3"
						stroke-linecap="round"
						stroke-linejoin="round"
					/>
				</svg>
				<span class="side-name">Finexia</span>
			</div>

			<ul class="side-nav">
				{#each TOUR_NAV as item (item.label)}
					<li class:active={item.label === view.nav}>
						{@render icon(item.icon)}
						{item.label}
					</li>
				{/each}
			</ul>

			<div class="side-foot">
				<span class="side-out">{@render icon('logout', 12)} Cerrar sesión</span>
				<span class="side-version">Finexia v1.0.0</span>
			</div>
		</aside>

		<div class="win-main">
			<div class="win-top">
				<span class="section">{view.nav}</span>
				<span class="top-actions">
					<span class="top-icon">{@render icon('eye', 13)}</span>
					<span class="top-icon">{@render icon('bell', 13)}</span>
					<span class="who">
						<span class="avatar">{TOUR_USER.initial}</span>
						<span class="who-text">
							<b>{TOUR_USER.name}</b>
							<em>{TOUR_USER.email}</em>
						</span>
					</span>
				</span>
			</div>

			<div class="win-view">
				{#if view.id === 'resumen'}
					<TourViewSummary />
				{:else if view.id === 'portafolios'}
					<TourViewPortfolios />
				{:else if view.id === 'transacciones'}
					<TourViewTransactions />
				{:else}
					<TourViewReports />
				{/if}
			</div>
		</div>
	</div>
</div>

<style>
	.win {
		border: 1px solid var(--border-strong);
		border-radius: 14px;
		overflow: hidden;
		background: rgba(255, 255, 255, 0.02);
		box-shadow: 0 30px 80px rgba(0, 0, 0, 0.45);
		backdrop-filter: blur(10px);
	}

	.win-bar {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 10px 14px;
		border-bottom: 1px solid var(--border);
		background: rgba(255, 255, 255, 0.022);
	}

	.dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: rgba(255, 255, 255, 0.13);
	}

	.win-url {
		flex: 1;
		text-align: center;
		font-family: var(--font-mono);
		font-size: 10px;
		letter-spacing: 0.04em;
		color: var(--text-dim);
	}

	/*
	 * El carril del panel a escala: 150 px de los 232 que mide de verdad.
	 *
	 * El alto mínimo acerca las cuatro vistas entre sí, para que la ventana no dé
	 * un salto al cambiar de pestaña y el pie de texto no se mueva debajo del
	 * cursor. No las iguala del todo a propósito: subirlo hasta el resumen dejaba
	 * al listado de portafolios con un palmo de vacío debajo de su tabla, que se
	 * lee como algo que no cargó y no como una página corta.
	 */
	.win-body {
		display: grid;
		grid-template-columns: 150px minmax(0, 1fr);
		min-height: 520px;
	}

	.win-side {
		display: flex;
		flex-direction: column;
		padding: 13px 8px 10px;
		border-right: 1px solid var(--border);
		background: rgba(0, 0, 0, 0.18);
	}

	.side-brand {
		display: flex;
		align-items: center;
		gap: 7px;
		padding: 0 5px 16px;
	}

	/* La única aparición de la serif en el panel: es el logotipo, no un titular. */
	.side-name {
		font-family: var(--font-display);
		font-size: 13px;
		font-weight: 600;
		letter-spacing: 0.02em;
		color: var(--text);
	}

	.side-nav {
		list-style: none;
		margin: 0;
		padding: 0;
		display: flex;
		flex-direction: column;
		gap: 1px;
	}

	.side-nav li {
		position: relative;
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 6px 7px;
		border-radius: 6px;
		font-size: 10.5px;
		color: var(--text-muted);
	}

	/*
	 * Una sola señal para la sección abierta, no cuatro: fondo tenue y un filete
	 * al canto. El ámbar aparece una vez, y por eso significa algo.
	 */
	.side-nav li.active {
		background: rgba(255, 255, 255, 0.04);
		color: var(--text);
		font-weight: 500;
	}

	.side-nav li.active::before {
		content: '';
		position: absolute;
		left: 0;
		top: 6px;
		bottom: 6px;
		width: 2px;
		border-radius: 0 2px 2px 0;
		background: var(--amber);
	}

	.side-foot {
		margin-top: auto;
		padding-top: 16px;
	}

	.side-out {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 6px 7px;
		font-size: 10.5px;
		color: var(--text-muted);
	}

	.side-version {
		display: block;
		margin: 7px 0 0 7px;
		font-size: 8.5px;
		color: var(--text-dim);
	}

	.win-main {
		display: flex;
		flex-direction: column;
		min-width: 0;
	}

	.win-top {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
		padding: 9px 16px;
		border-bottom: 1px solid var(--border);
	}

	/* En qué sección está, dicho como lo dice el panel: en caja normal y en la
	   tipografía de la interfaz. Antes eran migas de pan en versalitas. */
	.section {
		font-size: 11.5px;
		font-weight: 500;
		color: var(--text);
	}

	.top-actions {
		display: flex;
		align-items: center;
		gap: 4px;
	}

	.top-icon {
		display: grid;
		place-items: center;
		width: 22px;
		height: 22px;
		border-radius: 6px;
		color: var(--text-muted);
	}

	.who {
		display: flex;
		align-items: center;
		gap: 7px;
		margin-left: 5px;
		padding: 3px 6px 3px 3px;
		border-radius: 999px;
	}

	.avatar {
		display: grid;
		place-items: center;
		width: 20px;
		height: 20px;
		flex-shrink: 0;
		border-radius: 50%;
		background: rgba(255, 255, 255, 0.08);
		font-size: 9.5px;
		font-weight: 600;
		color: var(--text);
	}

	.who-text {
		display: flex;
		flex-direction: column;
		line-height: 1.25;
	}

	.who-text b {
		font-size: 9.5px;
		font-weight: 400;
		color: var(--text);
	}

	.who-text em {
		font-style: normal;
		font-size: 8.5px;
		color: var(--text-dim);
	}

	.win-view {
		flex: 1;
		padding: 16px 18px;
	}

	@media (max-width: 860px) {
		.win-body {
			grid-template-columns: minmax(0, 1fr);
			min-height: 0;
		}

		.win-side {
			display: none;
		}
	}

	@media (max-width: 560px) {
		.win-view {
			padding: 13px 14px;
		}

		.win-top {
			padding: 8px 13px;
		}

		.who-text {
			display: none;
		}
	}
</style>
