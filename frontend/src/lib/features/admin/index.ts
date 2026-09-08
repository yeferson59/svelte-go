/**
 * Feature `admin` — superficie pública.
 *
 * Componentes de `routes/dashboard/admin/**`: la gestión de usuarios
 * (invitaciones, lista de espera y usuarios registrados) y el mantenimiento del
 * catálogo compartido (activos y tasas de cambio). `admin.ts` aporta las
 * constantes y los formateadores, y reexporta los contratos de `$lib/api/types`.
 *
 * `admin-block` es el chrome interno que comparten esas pantallas —el título, la
 * frase de estado y el idioma de las celdas— y no forma parte de la superficie
 * pública. `desk.ts` aporta la edad de los datos y la lista de lo pendiente.
 */
export { default as InviteUserForm } from './components/invite-user-form.svelte';
export { default as InvitationsTable } from './components/invitations-table.svelte';
export { default as WaitlistTable } from './components/waitlist-table.svelte';
export { default as UsersTable } from './components/users-table.svelte';
export { default as AssetCreateForm } from './components/asset-create-form.svelte';
export { default as AssetEditForm } from './components/asset-edit-form.svelte';
export { default as AssetsTable } from './components/assets-table.svelte';
export { default as ExchangeRateCreateForm } from './components/exchange-rate-create-form.svelte';
export { default as ExchangeRatesTable } from './components/exchange-rates-table.svelte';
export { default as ImportCard } from './components/import-card.svelte';
export { default as Worklist } from './components/worklist.svelte';

export * from './admin';
export * from './desk';
export * from './schemas';
