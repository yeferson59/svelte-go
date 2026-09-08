import type { Actions, PageServerLoad } from './$types';
import { fail } from '@sveltejs/kit';
import * as portfolio from '$lib/api/portfolio';
import * as market from '$lib/api/market';
import { resolveDisplayCurrency } from '$lib/shared/currency';
import { assetPriceRefreshSchema } from '$lib/features/portfolio';
import type { AssetHolding, MarketCredential } from '$lib/api/types';

export const load: PageServerLoad = async ({ cookies, fetch, url, locals }) => {
	// Mismo contrato que el panel: sin `?currency=` manda la moneda de la
	// cuenta. Aquí no es opcional que haya una — las filas suman posiciones de
	// portafolios que pueden estar en monedas distintas, así que el peso de cada
	// activo solo significa algo si todo llega convertido a la misma.
	const currency = resolveDisplayCurrency(
		url.searchParams.get('currency'),
		locals.user?.preferredCurrency
	);

	const event = { cookies, fetch };

	// Las claves acompañan a las tenencias porque son la mitad de la respuesta a
	// «¿de dónde sale este precio?»: sin ninguna, las posiciones van a coste y
	// el panel del activo no tiene a quién preguntar. Nunca traen la clave, solo
	// el proveedor, sus cuatro últimos caracteres y su estado.
	const [holdingsRes, credentialsRes] = await Promise.all([
		portfolio.getAssetHoldings(event, currency),
		market.getMarketCredentials(event)
	]);

	const holdings: AssetHolding[] =
		holdingsRes.ok && holdingsRes.success && Array.isArray(holdingsRes.data)
			? holdingsRes.data
			: [];

	// Un fallo aquí no vacía la página: sin claves el panel dice que no hay
	// ninguna configurada, que es exactamente lo que se ve si la llamada falla.
	const credentials: MarketCredential[] = credentialsRes.ok ? (credentialsRes.data ?? []) : [];

	return { holdings, currency, credentials };
};

export const actions = {
	/**
	 * Vuelve a pedir el precio de un activo a uno de los proveedores del
	 * usuario.
	 *
	 * Es lo mismo que el botón «Sincronizar» de Ajustes reducido a un activo y
	 * con el proveedor elegido a mano: gasta una consulta de la cuota del
	 * usuario, así que el backend lo limita igual que las claves.
	 */
	refreshPrice: async ({ request, cookies, fetch }) => {
		const fd = await request.formData();
		const assetId = (fd.get('assetId') as string | null) ?? '';

		const parsed = assetPriceRefreshSchema.safeParse({
			assetId,
			provider: fd.get('provider')
		});

		if (!parsed.success) {
			return fail(400, {
				refreshAssetId: assetId,
				refreshError: 'Elige uno de tus proveedores para consultar el precio.'
			});
		}

		const res = await market.refreshAssetPrice(
			{ cookies, fetch },
			parsed.data.assetId,
			parsed.data.provider
		);

		if (!res.ok) {
			// `details` viene del backend y ya está redactado para el usuario: una
			// consulta puede fallar de seis maneras distintas —cuota agotada, clave
			// rechazada, activo que ese proveedor no cubre— y el código de estado no
			// alcanza para distinguirlas.
			return fail(res.status, {
				refreshAssetId: parsed.data.assetId,
				refreshError: res.details ?? 'No se pudo consultar el precio. Inténtalo de nuevo.'
			});
		}

		return {
			refreshSuccess: true,
			refreshAssetId: parsed.data.assetId,
			refreshProvider: parsed.data.provider,
			refreshedPrice: res.data?.price ?? '',
			refreshedAt: res.data?.fetchedAt ?? ''
		};
	}
} satisfies Actions;
