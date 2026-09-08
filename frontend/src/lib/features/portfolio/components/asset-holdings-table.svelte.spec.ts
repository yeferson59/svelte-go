import { page } from 'vitest/browser';
import { describe, it, expect, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import AssetHoldingsTable from './asset-holdings-table.svelte';
import type { AssetHoldingRow } from '../asset-holdings';

const row: AssetHoldingRow = {
	assetId: 'a1',
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
	priceFetchedAt: '2026-09-05T10:00:00Z',
	fxConverted: true
};

const props = {
	maxValue: 9002.7,
	displayCurrency: 'USD',
	formatValue: (v: number) => `$${v.toFixed(2)}`,
	onGoToPortfolios: () => {},
	onOpen: () => {}
};

describe('asset-holdings-table.svelte', () => {
	it('lists the asset with its type, units and weight', async () => {
		render(AssetHoldingsTable, { ...props, rows: [row] });

		await expect.element(page.getByText('AAPL')).toBeInTheDocument();
		await expect.element(page.getByText('Apple Inc.')).toBeInTheDocument();
		await expect.element(page.getByText('Acciones')).toBeInTheDocument();
		await expect.element(page.getByText('42 uds')).toBeInTheDocument();
		// Con la coma de es-CO, como el resto de las cifras de la página: el
		// peso se escapaba por `toFixed` y escribía un punto.
		await expect.element(page.getByText('37,5%')).toBeInTheDocument();
	});

	// La barra es el fondo de la fila y su escala es la mayor posición de toda
	// la cartera, no la de la página que se esté viendo: si se reescalara por
	// hoja, la primera fila de cada una saldría llena.
	it('scales the row bar against the whole portfolio, not the page', async () => {
		render(AssetHoldingsTable, { ...props, maxValue: 18005.4, rows: [row] });

		await expect
			.element(page.getByRole('row', { name: /AAPL/ }))
			.toHaveAttribute('style', '--bar: 50.00%;');
	});

	// Cuántos portafolios comparten el activo es lo que esta vista sabe y el
	// detalle de cada portafolio no puede decir. Solo se dice cuando hay más de
	// uno: una columna entera de «1» no informaba de nada.
	it('says in how many portfolios the asset is held, only when it is shared', async () => {
		const { rerender } = await render(AssetHoldingsTable, { ...props, rows: [row] });
		await expect.element(page.getByText(/, en 2 portafolios/)).toBeInTheDocument();

		await rerender({ ...props, rows: [{ ...row, portfolios: 1 }] });
		await expect.element(page.getByText(/, en \d+ portafolios/)).not.toBeInTheDocument();
	});

	// Sin precio de mercado no hay número que represente al activo: cada
	// entrada pagó el suyo. Un 0 se leería como un activo que no vale nada.
	it('marks a position carried at cost instead of printing a zero price', async () => {
		render(AssetHoldingsTable, {
			...props,
			rows: [{ ...row, marketPrice: null, priceSource: 'cost' }]
		});

		await expect.element(page.getByText('a coste, sin precio de mercado')).toBeInTheDocument();
	});

	// Un importe sin tasa va a valor nominal y mezcla monedas con el resto de
	// la columna: hay que decirlo, no presentarlo como una cifra limpia.
	it('flags a value no rate could convert', async () => {
		render(AssetHoldingsTable, { ...props, rows: [{ ...row, fxConverted: false }] });

		await expect.element(page.getByText('sin convertir a USD')).toBeInTheDocument();
	});

	// El activo es lo que se pulsa para ver de dónde sale su precio, y es un
	// botón de verdad: la fila entera no puede serlo porque dentro hay texto que
	// se selecciona, y un `onclick` sobre el `<tr>` no llega por teclado.
	it('opens the asset when its name is pressed', async () => {
		const onOpen = vi.fn();
		render(AssetHoldingsTable, { ...props, rows: [row], onOpen });

		await page.getByRole('button', { name: /precio de AAPL/ }).click();
		expect(onOpen).toHaveBeenCalledWith(row);
	});

	// Aquí no hay un portafolio al que agregar —la vista los atraviesa todos—,
	// así que la salida del estado vacío es elegir uno.
	it('sends a user with nothing to their portfolios', async () => {
		const onGoToPortfolios = vi.fn();
		render(AssetHoldingsTable, { ...props, rows: [], onGoToPortfolios });

		await expect.element(page.getByText('Todavía no hay nada que listar')).toBeInTheDocument();
		await page.getByRole('button', { name: 'Ir a mis portafolios' }).click();
		expect(onGoToPortfolios).toHaveBeenCalled();
	});
});
