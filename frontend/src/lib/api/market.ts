/**
 * Mercado: catálogo de assets y tasas de cambio, las operaciones de
 * administración (alta, import y ajuste manual de precio/tasa), y las claves
 * de proveedor que cada usuario aporta.
 *
 * Los datos de mercado son BYO-key: la aplicación no tiene claves de
 * proveedor, así que ya no existe una sincronización global de administración.
 * Cada usuario sincroniza sus propias tenencias con su propia clave.
 */
import {
	apiRequest,
	apiRequestSafe,
	authedFetchSafe,
	type ApiEvent,
	type ApiResult
} from './client';
import type {
	Asset,
	ExchangeRate,
	ImportResult,
	MarketCredential,
	MarketPrice,
	MarketProvider,
	MarketSyncResult
} from './types';
import {
	assetSchema,
	exchangeRateSchema,
	marketCredentialSchema,
	marketPriceSchema,
	marketSyncResultSchema
} from './schemas';
import { z } from 'zod';

// --- Assets ---------------------------------------------------------------

/** `GET /portfolios/assets` — catálogo de assets (paginado). */
export function getAssets(
	event: ApiEvent,
	opts: { page?: number; limit?: number } = {}
): Promise<ApiResult<Asset[]>> {
	const page = opts.page ?? 1;
	const limit = opts.limit ?? 100;
	return apiRequestSafe(
		event,
		`/portfolios/assets?page=${page}&limit=${limit}`,
		{},
		z.array(assetSchema)
	);
}

/**
 * `GET /portfolios/assets` con búsqueda — para el combobox de activos. Devuelve
 * la `Response` cruda porque el endpoint `/api/assets` la proxya tal cual.
 */
export function searchAssets(
	event: ApiEvent,
	opts: { search?: string; limit?: string } = {}
): Promise<Response | null> {
	const limit = opts.limit ?? '10';
	const search = opts.search?.trim();
	const path = search
		? `/portfolios/assets?search=${encodeURIComponent(search)}&page=1&limit=${limit}`
		: `/portfolios/assets?page=1&limit=${limit}`;
	return authedFetchSafe(event, path);
}

/**
 * `POST /assets` — añade un asset al catálogo.
 *
 * El backend decide qué significa según el rol del llamante: un admin cura la
 * fila (visible para todos) y un usuario la aporta (visible solo para él, y sin
 * sobrescribir un ticker que ya exista). El cuerpo y la respuesta son los
 * mismos en ambos casos.
 */
export function createAsset(
	event: ApiEvent,
	body: Record<string, unknown>
): Promise<ApiResult<Asset>> {
	return apiRequest<Asset>(event, '/assets', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
}

/** `POST /assets/import` — import masivo de assets (multipart). */
export function importAssets(event: ApiEvent, form: FormData): Promise<ApiResult<ImportResult>> {
	return apiRequest<ImportResult>(event, '/assets/import', { method: 'POST', body: form });
}

/** `PATCH /portfolios/assets/:id/price` — fija el precio manual de un asset. */
export function updateAssetPrice(
	event: ApiEvent,
	id: string,
	body: Record<string, unknown>
): Promise<ApiResult<unknown>> {
	return apiRequest<unknown>(event, `/portfolios/assets/${id}/price`, {
		method: 'PATCH',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
}

// --- Exchange rates -------------------------------------------------------

/** `GET /exchange-rates` — tasas de cambio (paginado). */
export function getExchangeRates(
	event: ApiEvent,
	opts: { page?: number; limit?: number } = {}
): Promise<ApiResult<ExchangeRate[]>> {
	const page = opts.page ?? 1;
	const limit = opts.limit ?? 100;
	return apiRequestSafe(
		event,
		`/exchange-rates?page=${page}&limit=${limit}`,
		{},
		z.array(exchangeRateSchema)
	);
}

/**
 * `GET /exchange-rates/latest` — las tasas compartidas, sin paginar.
 *
 * A diferencia del listado de arriba no es de administración: las llena un feed
 * público sin clave (la TRM oficial para USD/COP), así que cualquier usuario
 * puede ver con qué se están convirtiendo sus propias cifras.
 */
export function getLatestExchangeRates(event: ApiEvent): Promise<ApiResult<ExchangeRate[]>> {
	return apiRequestSafe(event, '/exchange-rates/latest', {}, z.array(exchangeRateSchema));
}

/** `POST /exchange-rates` — crea una tasa. */
export function createRate(
	event: ApiEvent,
	body: Record<string, unknown>
): Promise<ApiResult<unknown>> {
	return apiRequest<unknown>(event, '/exchange-rates', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
}

/**
 * `POST /exchange-rates/refresh` — relee el feed público ahora (admin).
 *
 * El job horario es lo que mantiene las tasas al día; esto cubre los dos
 * momentos que su cadencia no alcanza: un despliegue recién creado, que espera
 * un intervalo entero hasta la primera ejecución, y un administrador que
 * necesita la tasa de hoy ya. Sobrescribe los pares que el feed cubre, incluida
 * una tasa escrita a mano.
 */
export function refreshExchangeRates(event: ApiEvent): Promise<ApiResult<ExchangeRate[]>> {
	return apiRequestSafe(
		event,
		'/exchange-rates/refresh',
		{ method: 'POST' },
		z.array(exchangeRateSchema)
	);
}

/** `POST /exchange-rates/import` — import masivo de tasas (multipart). */
export function importRates(event: ApiEvent, form: FormData): Promise<ApiResult<ImportResult>> {
	return apiRequest<ImportResult>(event, '/exchange-rates/import', { method: 'POST', body: form });
}

/** `PATCH /exchange-rates/:id` — actualiza una tasa. */
export function updateRate(
	event: ApiEvent,
	id: string,
	body: Record<string, unknown>
): Promise<ApiResult<unknown>> {
	return apiRequest<unknown>(event, `/exchange-rates/${id}`, {
		method: 'PATCH',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
}

// --- Claves de proveedor (BYO-key) ----------------------------------------

/**
 * `GET /market/credentials` — estado de las claves del usuario.
 *
 * La respuesta nunca contiene la clave: solo su proveedor, sus cuatro últimos
 * caracteres y su estado. No hay endpoint que devuelva una clave guardada.
 */
export function getMarketCredentials(event: ApiEvent): Promise<ApiResult<MarketCredential[]>> {
	return apiRequestSafe(event, '/market/credentials', {}, z.array(marketCredentialSchema));
}

/**
 * `PUT /market/credentials/:provider` — guarda una clave.
 *
 * El backend la verifica contra el proveedor antes de sellarla, así que un
 * 400 aquí significa que el proveedor la rechazó.
 */
export function saveMarketCredential(
	event: ApiEvent,
	provider: MarketProvider,
	apiKey: string
): Promise<ApiResult<MarketCredential>> {
	return apiRequest<MarketCredential>(event, `/market/credentials/${provider}`, {
		method: 'PUT',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ apiKey })
	});
}

/** `POST /market/credentials/:provider/verify` — recomprueba una clave guardada. */
export function verifyMarketCredential(
	event: ApiEvent,
	provider: MarketProvider
): Promise<ApiResult<MarketCredential>> {
	return apiRequest<MarketCredential>(event, `/market/credentials/${provider}/verify`, {
		method: 'POST'
	});
}

/** `DELETE /market/credentials/:provider` — borra una clave. */
export function deleteMarketCredential(
	event: ApiEvent,
	provider: MarketProvider
): Promise<ApiResult<unknown>> {
	return apiRequest<unknown>(event, `/market/credentials/${provider}`, { method: 'DELETE' });
}

/**
 * `POST /market/sync` — sincroniza las tenencias del usuario con sus claves.
 *
 * Devuelve precios y tasas por separado: una posición cotizada en otra moneda
 * no vale nada sin su tasa, y bajo BYO-key no se puede usar la de otro usuario.
 * El backend corta a los 60 s y devuelve lo que dio tiempo a traer, así que una
 * cartera grande puede volver incompleta; el resto lo recoge el job diario.
 */
export function syncMarketData(event: ApiEvent): Promise<ApiResult<MarketSyncResult>> {
	return apiRequest(event, '/market/sync', { method: 'POST' }, marketSyncResultSchema);
}

/**
 * `POST /market/assets/:id/refresh` — vuelve a pedir el precio de un activo a
 * un proveedor concreto.
 *
 * `POST /market/sync` recorre todas las tenencias y deja que la cadena de
 * respaldo decida quién contesta, que es lo que hace falta en un job diario.
 * Esto es lo mismo reducido a un activo y con el proveedor nombrado: quien está
 * mirando una posición quiere saber si el precio está viejo o si el proveedor
 * sigue diciendo lo mismo, y para eso hay que volver a preguntarle *a ese*.
 *
 * Gasta una consulta de la cuota del usuario, así que el backend la limita con
 * la misma puerta que las claves.
 */
export function refreshAssetPrice(
	event: ApiEvent,
	assetId: string,
	provider: MarketProvider
): Promise<ApiResult<MarketPrice>> {
	return apiRequest(
		event,
		`/market/assets/${assetId}/refresh`,
		{
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ provider })
		},
		marketPriceSchema
	);
}
