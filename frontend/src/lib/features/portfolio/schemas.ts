/**
 * Schemas Zod de los formularios de portfolios.
 *
 * Estaban declarados dentro de las actions de `routes/dashboard/portfolios/**`,
 * uno por archivo, con las mismas reglas escritas de formas distintas. Aquí
 * viven juntos para que el contrato de un mismo campo (la cantidad, el precio,
 * la fecha) no dependa de qué formulario lo envía.
 */

import { z } from 'zod';
import { marketProviderSchema } from '$lib/api/schemas';

/** Alta de un portfolio (`routes/dashboard/portfolios/add`). */
export const portfolioCreateSchema = z.object({
	// Con mensaje: son los que ve el usuario cuando el alta se rechaza, y los de
	// Zod por defecto llegan en inglés y hablando de longitudes de cadena.
	name: z.string().min(1, 'Ponle un nombre al portafolio.'),
	description: z.string().nullable(),
	type: z.string().min(1, 'Elige qué guarda el portafolio.'),
	riskId: z.uuid('Elige un nivel de riesgo.'),
	currency: z.string().min(1, 'Elige la moneda del portafolio.'),
	priceValue: z.coerce.number().nonnegative('La meta no puede ser negativa.').default(0),
	isDefault: z.coerce.boolean()
});

/** Edición de un portfolio (`routes/dashboard/portfolios/[id]`). */
export const portfolioUpdateSchema = z.object({
	name: z.string().min(2, 'El nombre debe tener al menos 2 caracteres'),
	description: z.string().optional().default(''),
	type: z.string().min(1),
	riskId: z.uuid(),
	isDefault: z.coerce.boolean()
});

/**
 * Código ISO de tres letras. El precio y su moneda viajan separados, así que
 * una moneda vacía o basura no falla: se guarda y el coste queda etiquetado con
 * algo que no es lo que se pagó.
 */
const currencyCode = z.coerce
	.string()
	.trim()
	.toUpperCase()
	.regex(/^[A-Z]{3}$/, 'Moneda inválida: usa un código ISO de tres letras');

/**
 * Tasa de cambio de la operación.
 *
 * Ausente o vacía significa 1, que es lo que vale cuando la operación y la
 * cuenta usan la misma moneda —la inmensa mayoría—, y es lo que manda el
 * formulario en ese caso. El backend rechaza una tasa distinta de 1 entre una
 * moneda y sí misma, y rechaza que falte cuando las dos difieren, así que aquí
 * solo hace falta normalizar el hueco.
 *
 * Lo que no se normaliza es una tasa inválida. Un `.catch(1)` la convertiría en
 * «sin conversión», que es un número plausible y equivocado: el coste quedaría
 * registrado en la moneda del mercado con la etiqueta de la cuenta, que es
 * exactamente el error que la columna existe para evitar. Cero o negativa
 * tienen que fallar y verse.
 */
const fxRate = z.preprocess(
	(value) => (value === null || value === undefined || value === '' ? 1 : value),
	z.coerce.number().positive('La tasa debe ser mayor que cero')
);

/** Alta de una posición (`routes/dashboard/portfolios/[id]/add`). */
export const portfolioEntrySchema = z.object({
	portfolioId: z.uuid(),
	// Con mensaje: son los que ve el usuario cuando el alta se rechaza, y los de
	// Zod por defecto llegan en inglés hablando de UUID y de números.
	assetId: z.uuid('Elige el activo en el buscador.'),
	sourceId: z.uuid('Elige la plataforma donde tienes el activo.'),
	quantity: z.coerce.number().positive('La cantidad tiene que ser mayor que cero.'),
	price: z.coerce.number().positive('El precio por unidad tiene que ser mayor que cero.'),
	// `costCurrency` es la de la cuenta —en la que el bróker debitó— y
	// `currency` la de la operación, que es la de cotización del activo. Solo
	// difieren cuando el bróker convirtió, y entonces `fxRate` es a cuánto.
	costCurrency: currencyCode,
	currency: currencyCode,
	fxRate,
	entryDate: z.coerce.date(),
	notes: z.coerce.string().optional()
});

/**
 * Campos comunes de una transacción. El precio admite 0 (una entrega o un
 * split no cuestan nada) aunque la cantidad nunca lo sea.
 */
const transactionSchema = z.object({
	type: z.string().min(1),
	quantity: z.coerce.number().positive(),
	price: z.coerce.number().min(0),
	currency: z.string().default('USD'),
	fxRate,
	fees: z.coerce.number().min(0).default(0),
	// Vacía significa «la de la operación», que es donde estuvo la comisión en
	// todas las filas anteriores a la tasa por transacción. El backend rechaza
	// cualquiera que no sea esa o la de la posición.
	feesCurrency: z.coerce.string().trim().toUpperCase().default(''),
	transactionDate: z.coerce.date(),
	notes: z.string().optional()
});

/** Alta de una transacción sobre una posición existente. */
export const transactionCreateSchema = transactionSchema.extend({ entryId: z.uuid() });

/** Edición de una transacción ya registrada. */
export const transactionUpdateSchema = transactionSchema.extend({ txnId: z.uuid() });

/** Borrado de una transacción: solo hace falta identificarla. */
export const transactionDeleteSchema = z.object({ txnId: z.uuid() });

/** Borrado de una posición entera, con todo su historial. */
export const entryDeleteSchema = z.object({ entryId: z.uuid() });

/**
 * Actualización manual del precio de un activo contra un proveedor concreto
 * (`routes/dashboard/assets`).
 *
 * El proveedor sale del enum del contrato de la API y no de una cadena libre:
 * el backend lo rechazaría igual, pero un valor que no es ninguno de los dos no
 * merece un viaje de ida y vuelta para enterarse.
 */
export const assetPriceRefreshSchema = z.object({
	assetId: z.uuid('No se reconoce el activo.'),
	provider: marketProviderSchema
});
