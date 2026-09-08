/**
 * Schemas Zod de los formularios de administración.
 *
 * Las tres pantallas validaban a mano (`if (!email) …`), cada una con su
 * propio criterio. Aquí quedan las mismas reglas y **los mismos mensajes**,
 * escritos una vez.
 *
 * Los ids no se validan como UUID a propósito: el backend también emite ids de
 * invitación y de waitlist con otro formato, y rechazarlos aquí dejaría fuera
 * acciones que el servidor acepta.
 */

import { z } from 'zod';

/** Roles que se pueden asignar al invitar. */
export const inviteRoleSchema = z.enum(['customer', 'admin'], { error: 'Rol inválido' });

/**
 * Invitación de un usuario. El rol llega vacío desde el atajo de la lista de
 * espera, y en ese caso vale `customer`.
 */
export const inviteUserSchema = z.object({
	email: z.string().trim().min(1, 'El correo es requerido'),
	name: z.string().trim().default(''),
	role: z
		.string()
		.transform((v) => v.trim().toLowerCase() || 'customer')
		.pipe(inviteRoleSchema)
});

/** Identificador de una fila sobre la que actúa el admin. */
export const rowIdSchema = z.string().trim().min(1, 'ID requerido');

/**
 * Alta de un activo del catálogo.
 *
 * Los cuatro campos obligatorios comparten mensaje porque el formulario los
 * pide juntos: da igual cuál falte, el aviso es el mismo.
 */
const REQUIRED_ASSET_FIELDS = 'Ticker, nombre, tipo y moneda son requeridos';

export const assetCreateSchema = z.object({
	ticker: z.string().trim().toUpperCase().min(1, REQUIRED_ASSET_FIELDS),
	name: z.string().trim().min(1, REQUIRED_ASSET_FIELDS),
	assetType: z.string().trim().min(1, REQUIRED_ASSET_FIELDS),
	currency: z.string().trim().toUpperCase().min(1, REQUIRED_ASSET_FIELDS),
	exchange: z.string().trim().default('')
});

/**
 * Ajuste del precio manual de un activo.
 *
 * `price` se valida como número pero al backend viaja el texto tal cual lo
 * escribió el admin: convertirlo perdería los decimales de cola («190.00» →
 * «190»), que el backend guarda como los recibe.
 */
export const assetPriceSchema = z.object({
	id: z.string().trim().min(1, 'ID de activo requerido'),
	price: z.coerce.number().positive('Precio inválido'),
	currency: z.string().trim().min(1).default('USD')
});

/**
 * Edición completa de un activo del catálogo.
 *
 * Los cinco campos de identidad son los mismos del alta y comparten su mensaje,
 * porque el formulario los manda siempre enteros: editar el nombre reenvía
 * también el ticker que se conserva.
 *
 * Los dos últimos son de la edición y solo de ella. `isCurated` llega del
 * checkbox, que no envía nada cuando está desmarcado, así que se normaliza a
 * booleano aquí en vez de dejar que «ausente» signifique dos cosas. `price` es
 * opcional —quedarse en blanco deja el precio manual como estaba— y se valida
 * como número, pero al backend viaja el texto tal cual, igual que en el ajuste
 * rápido de la tabla.
 */
export const assetUpdateSchema = z.object({
	id: rowIdSchema,
	ticker: z.string().trim().toUpperCase().min(1, REQUIRED_ASSET_FIELDS),
	name: z.string().trim().min(1, REQUIRED_ASSET_FIELDS),
	assetType: z.string().trim().min(1, REQUIRED_ASSET_FIELDS),
	currency: z.string().trim().toUpperCase().min(1, REQUIRED_ASSET_FIELDS),
	exchange: z.string().trim().default(''),
	isCurated: z
		.union([z.string(), z.boolean(), z.null(), z.undefined()])
		.transform((v) => v === true || v === 'on' || v === 'true'),
	price: z
		.string()
		.trim()
		.default('')
		.refine((v) => v === '' || (Number.isFinite(Number(v)) && Number(v) > 0), 'Precio inválido')
});

/** Alta de una tasa de cambio. */
const REQUIRED_RATE_FIELDS = 'Moneda origen, destino y tasa son requeridos';

export const rateCreateSchema = z.object({
	fromCurrency: z.string().trim().toUpperCase().min(1, REQUIRED_RATE_FIELDS),
	toCurrency: z.string().trim().toUpperCase().min(1, REQUIRED_RATE_FIELDS),
	rate: z.string().trim().min(1, REQUIRED_RATE_FIELDS)
});

/** Ajuste de una tasa; el valor viaja como texto, igual que el precio. */
export const rateUpdateSchema = z.object({
	id: z.string().trim().min(1, 'ID de tasa requerido'),
	rate: z.coerce.number().positive('Tasa inválida')
});
