/**
 * Portfolios — contratos HTTP como schemas Zod.
 *
 * Parte de `lib/api/schemas`: el porqué de que sean schemas y no interfaces
 * está en el `index.ts` de la carpeta.
 */

import { z } from 'zod';

// ---------------------------------------------------------------------------
// Portfolios
// ---------------------------------------------------------------------------

/**
 * Resumen de un portfolio (`GET /portfolios/summary`). Superset de los
 * subconjuntos que antes tipaban por separado el dashboard, el layout de
 * portfolios y el selector de import; `displayCurrency` solo llega cuando se
 * pide `?currency=`.
 */
export const portfolioSummarySchema = z.object({
	id: z.string(),
	name: z.string(),
	description: z.string().optional(),
	type: z.string(),
	baseCurrency: z.string(),
	displayCurrency: z.string().optional(),
	// `false` cuando se pidió `?currency=` y no había tasa: los totales se
	// quedaron en `baseCurrency` en vez de fallar la petición entera, así que
	// esta fila no se puede sumar con las demás. Opcional por el backend previo.
	fxConverted: z.boolean().optional(),
	isDefault: z.boolean().optional(),
	riskId: z.string().optional(),
	riskName: z.string(),
	totalPositions: z.number(),
	totalCostBase: z.string(),
	totalMarketValue: z.string(),
	totalGainLoss: z.string(),
	totalGainLossPct: z.string(),
	createdAt: z.string().optional(),
	// Posiciones cuyos importes no están en la moneda base porque no había
	// tasa: siguen sumadas en los totales, así que un valor > 0 significa que
	// los totales mezclan monedas. Opcional para tolerar un backend anterior.
	positionsUnconverted: z.number().optional()
});

/** Posición dentro de un portfolio (holdings de `GET /portfolios/:id`). */
export const holdingSchema = z.object({
	id: z.string(),
	assetId: z.string(),
	ticker: z.string(),
	name: z.string(),
	assetType: z.string(),
	exchange: z.string(),
	currency: z.string(),
	quantity: z.string(),
	price: z.string(),
	marketPrice: z.string(),
	costCurrency: z.string(),
	category: z.string(),
	entryDate: z.string(),
	notes: z.string(),
	// Totales ya convertidos a la moneda base del portafolio: los únicos
	// importes del holding que se pueden sumar entre posiciones, porque `price`
	// y `marketPrice` vienen cada uno en su propia moneda. `fxConverted: false`
	// significa que faltó la tasa y los dos totales están sin convertir.
	//
	// Opcionales para tolerar un backend anterior a estos campos: quien los
	// consume vuelve al cálculo nativo cuando no llegan.
	costBasisBase: z.string().optional(),
	marketValueBase: z.string().optional(),
	fxConverted: z.boolean().optional()
});

/** Detalle completo de un portfolio (`GET /portfolios/:id`). */
export const portfolioDetailSchema = z.object({
	id: z.string(),
	userId: z.string(),
	name: z.string(),
	description: z.string(),
	type: z.string(),
	baseCurrency: z.string(),
	isDefault: z.boolean(),
	riskId: z.string(),
	riskName: z.string(),
	createdAt: z.string(),
	updatedAt: z.string(),
	holdings: z.array(holdingSchema)
});

/** Nivel de riesgo del catálogo (`GET /portfolios/risks`). */
export const riskSchema = z.object({
	id: z.string(),
	name: z.string(),
	description: z.string()
});

/** Asignación por categoría de activo (`GET /portfolios/allocation`). */
export const allocationItemSchema = z.object({
	category: z.string(),
	marketValue: z.string(),
	percent: z.number(),
	// Moneda en la que están todos los `marketValue` de la respuesta, y
	// posiciones de esta categoría que no pudieron convertirse a ella.
	// Opcionales para tolerar un backend anterior.
	currency: z.string().optional(),
	positionsUnconverted: z.number().optional()
});

/**
 * Un activo con todo lo que el usuario tiene de él, sumado a través de sus
 * portafolios (`GET /portfolios/holdings`).
 *
 * `quantity` son unidades y solo significa algo dentro de la fila: sumar
 * acciones con bitcoins no da nada. `marketValue` es lo que sí compara, y viene
 * en `displayCurrency` para todas las filas —igual que la asignación—, que es
 * lo que hace que `percent` quiera decir algo.
 *
 * `marketPrice` vacío no es precio cero: es una posición valorada a coste, en la
 * que cada entrada pagó el suyo y ningún número representa al activo. Eso lo
 * dice `priceSource`.
 */
export const assetHoldingSchema = z.object({
	assetId: z.string(),
	ticker: z.string(),
	name: z.string(),
	assetType: z.string(),
	exchange: z.string(),
	/** Moneda en la que cotiza el activo, que es la de `marketPrice`. */
	currency: z.string(),
	quantity: z.string(),
	marketPrice: z.string(),
	marketValue: z.string(),
	percent: z.number(),
	/** Moneda de `marketValue`, la misma en todas las filas. */
	displayCurrency: z.string(),
	/** En cuántos portafolios del usuario aparece el activo. */
	portfolios: z.number(),
	priceSource: z.string(),
	/**
	 * Proveedor cuya clave trajo `marketPrice`, y cuándo lo trajo. Solo vienen
	 * cuando `priceSource` es `own`: un precio manual del catálogo y una
	 * posición a coste no salen de ningún proveedor.
	 */
	priceProvider: z.string().optional(),
	priceFetchedAt: z.string().optional(),
	positionsUnconverted: z.number()
});

/** Mayor transacción de un portfolio (`GET /portfolios/:id/top-transaction`). */
export const topTransactionSchema = z.object({
	value: z.string(),
	type: z.string(),
	currency: z.string(),
	assetTicker: z.string(),
	assetName: z.string(),
	transactionDate: z.string()
});

/** Punto de la serie de crecimiento. */
export const growthDataPointSchema = z.object({
	date: z.string(),
	totalValue: z.string(),
	totalCostBase: z.string(),
	gainLoss: z.string(),
	gainLossPct: z.string(),
	// Portafolios sumados a esta fecha sin tasa con la que convertirlos, y por
	// tanto contados a valor nominal. Opcional por si el backend va por detrás.
	portfoliosUnconverted: z.number().optional(),
	// Dinero que el dueño metió (positivo) o sacó (negativo) entre el punto
	// anterior y este, reconstruido de las transacciones. Es lo que hay que
	// descontar de la variación del valor para que quede rentabilidad.
	// Opcional por si el backend va por detrás.
	netFlow: z.string().optional()
});

/**
 * Resumen agregado de la serie de crecimiento.
 *
 * Lleva dos lecturas distintas que no hay que confundir: `totalGrowthPct` mide
 * cuánto se movió el **valor** entre el primer snapshot y el último —abrir un
 * portafolio o añadir una posición cuenta como crecimiento— mientras que
 * `gainLoss`/`gainLossPct` son el beneficio del último punto, mercado menos
 * capital invertido, que es el rendimiento de verdad.
 *
 * Las dos últimas son opcionales para tolerar un backend anterior.
 */
export const growthSummarySchema = z.object({
	firstDate: z.string(),
	initialValue: z.string(),
	currentValue: z.string(),
	totalGrowthPct: z.string(),
	gainLoss: z.string().optional(),
	gainLossPct: z.string().optional(),
	/** Moneda en la que están todos los importes de la serie. */
	currency: z.string().optional()
});

/** Crecimiento (`GET /portfolios/growth` y `GET /portfolios/:id/growth`). */
export const portfolioGrowthSchema = z.object({
	points: z.array(growthDataPointSchema),
	summary: growthSummarySchema
});
