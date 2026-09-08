/**
 * Feature `portfolio` — superficie pública.
 *
 * Componentes del detalle de portafolio (`routes/dashboard/portfolios/[id]`),
 * del alta (`portfolios/add`) y de la vista consolidada de activos
 * (`dashboard/assets`). `portfolio.ts` aporta los helpers puros (agrupar
 * holdings, distribución por tipo, segmentos del donut) y `asset-holdings.ts`
 * los de la vista consolidada.
 */
export { default as PortfolioEditForm } from './components/portfolio-edit-form.svelte';
export { default as PortfolioHeadline } from './components/portfolio-headline.svelte';
export { default as PortfolioPositions } from './components/portfolio-positions.svelte';
export { default as PortfolioList } from './components/portfolio-list.svelte';
export { default as PortfolioDetailHeader } from './components/portfolio-detail-header.svelte';
export { default as PortfolioAddForm } from './components/portfolio-add-form.svelte';
export { default as PortfolioEntryForm } from './components/portfolio-entry-form.svelte';

// Detalle de un activo dentro del portafolio
// (`portfolios/[id]/assets/[symbol]`). El formulario de alta, el panel de venta
// rápida, la tabla y el modal de edición son internos de
// `asset-transaction-history`; `asset-combobox` lo es de `portfolio-entry-form`.
export { default as AssetPositionHeader } from './components/asset-position-header.svelte';
export { default as AssetPositionHeadline } from './components/asset-position-headline.svelte';
export { default as AssetTransactionHistory } from './components/asset-transaction-history.svelte';
export { default as AssetDeletePosition } from './components/asset-delete-position.svelte';

// Vista consolidada de activos (`dashboard/assets`): lo que el usuario tiene
// de cada activo sumando todos sus portafolios. `asset-holdings-table` es
// interno de la lista, que es la que sabe del buscador y de las hojas.
export { default as AssetConcentrationBand } from './components/asset-concentration-band.svelte';
export { default as AssetHoldingsList } from './components/asset-holdings-list.svelte';
// El panel de un activo: de dónde salió su precio y con qué clave volver a
// pedirlo. Lo abre la lista y lo monta la página, que es quien tiene las claves
// del usuario y el `form` de la acción.
export { default as AssetPricePanel } from './components/asset-price-panel.svelte';

export * from './portfolio';
export * from './asset-holdings';
export * from './asset';
export * from './schemas';
