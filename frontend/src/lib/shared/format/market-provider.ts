/**
 * Nombres de los proveedores de datos de mercado.
 *
 * El vocabulario es el de `marketdata.ProviderName` del backend, que es lo que
 * viaja en `market_credentials.provider`, en el `source` de un precio y en el
 * cuerpo de `POST /market/assets/:id/refresh`. Aquí solo se traduce a cómo se
 * llama la empresa.
 *
 * Vive en `shared` y no en una feature porque lo necesitan dos pantallas que no
 * pueden importarse entre sí: Ajustes, donde se dan de alta las claves, y el
 * panel de un activo, donde se elige con cuál preguntar. La tabla estaba
 * escrita a mano en la primera, y una segunda copia es exactamente cómo empieza
 * a haber dos nombres para el mismo proveedor.
 */

/** Proveedores para los que el backend acepta una clave. */
export const MARKET_PROVIDER_NAMES: Record<string, string> = {
	finnhub: 'Finnhub',
	alphavantage: 'Alpha Vantage'
};

/**
 * Nombre legible de un proveedor. Uno que el backend añada y aquí no esté
 * conserva su identificador crudo, que sigue siendo mejor que un hueco.
 */
export function formatMarketProvider(provider: string): string {
	return MARKET_PROVIDER_NAMES[provider] ?? provider;
}
