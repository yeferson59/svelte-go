import { page } from 'vitest/browser';
import { describe, it, expect } from 'vitest';
import { render } from 'vitest-browser-svelte';
import AssetPricePanel from './asset-price-panel.svelte';
import type { MarketCredential } from '$lib/api/types';
import type { AssetHoldingRow } from '../asset-holdings';

function row(overrides: Partial<AssetHoldingRow> = {}): AssetHoldingRow {
	return {
		assetId: '7c3e8b5d-1f2a-4b6c-9d0e-3a4b5c6d7e8f',
		ticker: 'AAPL',
		name: 'Apple Inc.',
		assetType: 'stock',
		typeLabel: 'Acciones',
		quantity: 42,
		marketPrice: 214.35,
		currency: 'USD',
		value: 9002.7,
		percent: 37.5,
		portfolios: 2,
		priceSource: 'own',
		priceProvider: 'finnhub',
		priceFetchedAt: new Date(Date.now() - 3 * 86_400_000).toISOString(),
		fxConverted: true,
		...overrides
	};
}

function credential(overrides: Partial<MarketCredential> = {}): MarketCredential {
	return {
		provider: 'finnhub',
		last4: '3f9a',
		status: 'active',
		lastVerifiedAt: '2026-09-05T10:00:00Z',
		createdAt: '2026-09-01T10:00:00Z',
		updatedAt: '2026-09-05T10:00:00Z',
		...overrides
	};
}

const props = { form: null, onClose: () => {} };

describe('asset-price-panel.svelte', () => {
	// La pregunta que trae a alguien aquí no es cuánto vale, que ya lo vio en la
	// lista: es de dónde salió ese número y cuándo.
	it('names the provider behind the price and how old it is', async () => {
		render(AssetPricePanel, { ...props, row: row(), credentials: [credential()] });

		await expect.element(page.getByText('Lo trajo Finnhub hace 3 días.')).toBeInTheDocument();
	});

	// Un precio manual del catálogo no salió de ninguna clave, y decir «lo trajo
	// tu proveedor» sería falso sobre el único dato que el panel existe para
	// explicar.
	it('says a catalog price is not the user’s own', async () => {
		render(AssetPricePanel, {
			...props,
			row: row({ priceSource: 'manual', priceProvider: null, priceFetchedAt: null }),
			credentials: [credential()]
		});

		await expect.element(page.getByText(/precio de respaldo del catálogo/)).toBeInTheDocument();
	});

	// Una posición a coste no tiene precio: un 0 se leería como un activo que no
	// vale nada, y la ganancia que sale de ahí es exactamente cero por
	// construcción.
	it('does not print a price for a position carried at cost', async () => {
		render(AssetPricePanel, {
			...props,
			row: row({
				marketPrice: null,
				priceSource: 'cost',
				priceProvider: null,
				priceFetchedAt: null
			}),
			credentials: [credential()]
		});

		await expect.element(page.getByText('Sin precio')).toBeInTheDocument();
		await expect.element(page.getByText(/se valora a lo que costó/)).toBeInTheDocument();
	});

	// Volver a preguntar a quien contestó es lo que se quiere casi siempre, así
	// que esa es la opción que ya viene marcada.
	it('preselects the provider that produced the current price', async () => {
		render(AssetPricePanel, {
			...props,
			row: row(),
			credentials: [credential(), credential({ provider: 'alphavantage', last4: '1c2d' })]
		});

		await expect.element(page.getByRole('radio', { name: /Finnhub/ })).toBeChecked();
		await expect.element(page.getByRole('radio', { name: /Alpha Vantage/ })).not.toBeChecked();
	});

	// Un proveedor sin clave se enseña igual, apagado: saber que Alpha Vantage
	// existe y que no está configurado es lo que convierte «este proveedor no
	// cubre el activo» en algo que se puede arreglar.
	it('shows a provider with no key, disabled', async () => {
		render(AssetPricePanel, { ...props, row: row(), credentials: [credential()] });

		await expect.element(page.getByRole('radio', { name: /Alpha Vantage/ })).toBeDisabled();
		await expect.element(page.getByText('Sin configurar')).toBeInTheDocument();
	});

	// Sin ninguna clave no hay a quién preguntar, así que no se enseña un botón
	// que solo puede fallar: se enseña dónde se arregla.
	it('sends a user with no keys to settings instead of offering the button', async () => {
		render(AssetPricePanel, { ...props, row: row(), credentials: [] });

		await expect.element(page.getByText(/No tienes ninguna clave/)).toBeInTheDocument();
		await expect
			.element(page.getByRole('button', { name: 'Actualizar precio' }))
			.not.toBeInTheDocument();
	});

	// El panel se abre y se cierra sobre la misma página, y `form` le sobrevive:
	// el error de un activo no puede aparecer sobre otro.
	it('only shows the result of the asset it belongs to', async () => {
		const { rerender } = await render(AssetPricePanel, {
			...props,
			row: row(),
			credentials: [credential()],
			form: {
				refreshAssetId: 'otro-activo',
				refreshError: 'La cuota de tu clave con este proveedor se agotó.'
			}
		});

		await expect.element(page.getByText(/La cuota de tu clave/)).not.toBeInTheDocument();

		await rerender({
			...props,
			row: row(),
			credentials: [credential()],
			form: {
				refreshAssetId: row().assetId,
				refreshError: 'La cuota de tu clave con este proveedor se agotó.'
			}
		});

		await expect.element(page.getByText(/La cuota de tu clave/)).toBeInTheDocument();
	});
});
