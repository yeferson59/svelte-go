import type { Actions, PageServerLoad } from './$types';
import * as market from '$lib/api/market';
import { fail } from '@sveltejs/kit';
import type { Asset } from '$lib/api/types';
import { assetCreateSchema, assetPriceSchema, assetUpdateSchema } from '$lib/features/admin';

export const load: PageServerLoad = async ({ cookies, fetch }) => {
	const res = await market.getAssets({ cookies, fetch }, { page: 1, limit: 100 });

	return {
		assets: res.success && Array.isArray(res.data) ? (res.data as Asset[]) : []
	};
};

export const actions = {
	createAsset: async ({ request, cookies, fetch }) => {
		const fd = await request.formData();

		const parsed = assetCreateSchema.safeParse({
			ticker: fd.get('ticker') ?? '',
			name: fd.get('name') ?? '',
			assetType: fd.get('assetType') ?? '',
			currency: fd.get('currency') ?? '',
			exchange: fd.get('exchange') ?? ''
		});

		if (!parsed.success) {
			return fail(400, { createError: parsed.error.issues[0].message });
		}

		const res = await market.createAsset({ cookies, fetch }, parsed.data);

		if (!res.ok) {
			// `details` lo pone el binder del backend; los errores de dominio
			// (ticker vacío, moneda inválida, cuota) viajan en `action`.
			return fail(res.status, {
				createError: res.details ?? res.action ?? 'No se pudo crear el activo'
			});
		}

		return { createSuccess: true };
	},

	updateAsset: async ({ request, cookies, fetch }) => {
		const fd = await request.formData();
		const id = (fd.get('id') as string | null) ?? '';

		const parsed = assetUpdateSchema.safeParse({
			id,
			ticker: fd.get('ticker') ?? '',
			name: fd.get('name') ?? '',
			assetType: fd.get('assetType') ?? '',
			currency: fd.get('currency') ?? '',
			exchange: fd.get('exchange') ?? '',
			isCurated: fd.get('isCurated'),
			// El texto tal cual, como en `updatePrice`: convertirlo a número
			// perdería los decimales de cola que el backend guarda como llegan.
			price: (fd.get('price') as string | null) ?? ''
		});

		if (!parsed.success) {
			return fail(400, { editError: parsed.error.issues[0].message, editId: id });
		}

		const { id: assetId, price, ...asset } = parsed.data;

		const res = await market.updateAsset({ cookies, fetch }, assetId, {
			...asset,
			// Se omite si el campo quedó en blanco: el backend distingue «no lo
			// toques» de un precio nuevo, y mandar null sería lo segundo.
			...(price === '' ? {} : { price: { value: price, currency: asset.currency } })
		});

		if (!res.ok) {
			return fail(res.status, {
				editError: res.details ?? res.action ?? 'No se pudo guardar el activo',
				editId: assetId
			});
		}

		return { editSuccess: true, editedId: assetId };
	},

	importAssets: async ({ request, cookies, fetch }) => {
		const fd = await request.formData();
		const file = fd.get('file');
		if (!(file instanceof File) || file.size === 0) {
			return fail(400, { importError: 'Selecciona un archivo CSV o Excel' });
		}

		const res = await market.importAssets({ cookies, fetch }, fd);

		if (!res.ok) {
			return fail(res.status, { importError: res.details ?? 'No se pudo importar el archivo' });
		}

		return { importSuccess: true, importResult: res.data };
	},

	updatePrice: async ({ request, cookies, fetch }) => {
		const fd = await request.formData();
		const id = (fd.get('id') as string | null) ?? '';
		// El texto tal cual lo escribió el admin es lo que viaja al backend; el
		// schema solo comprueba que sea un número positivo.
		const priceStr = (fd.get('price') as string | null) ?? '';

		const parsed = assetPriceSchema.safeParse({
			id,
			price: priceStr,
			currency: fd.get('currency') ?? 'USD'
		});

		if (!parsed.success) {
			return fail(400, { updateError: parsed.error.issues[0].message, errorId: id });
		}

		const res = await market.updateAssetPrice({ cookies, fetch }, parsed.data.id, {
			price: { value: priceStr, currency: parsed.data.currency }
		});

		if (!res.ok) {
			return fail(res.status, {
				updateError: res.details ?? 'No se pudo actualizar el precio',
				errorId: parsed.data.id
			});
		}

		return { updateSuccess: true, updatedId: parsed.data.id };
	}
} satisfies Actions;
